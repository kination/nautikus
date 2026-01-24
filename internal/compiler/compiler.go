package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type DagSource struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
}

func CompileDags(configPath string, outputDir string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config error: %w", err)
	}

	var sources []DagSource
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return fmt.Errorf("yaml parse error: %w", err)
	}

	// Generate output directory if not exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	for _, src := range sources {
		fmt.Printf("📂 Scanning source: %s (%s)\n", src.Name, src.Location)

		err := filepath.WalkDir(src.Location, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			// support python and go file
			ext := filepath.Ext(d.Name())
			switch ext {
			case ".py":
				return generateYAML("python3", []string{path}, path, outputDir)
			case ".go":
				return generateYAML("go", []string{"run", path}, path, outputDir)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("walk error in %s: %w", src.Location, err)
		}
	}
	return nil
}

// generateYAML is a helper function that runs a command, parses JSON output, and saves it as YAML.
func generateYAML(cmdName string, cmdArgs []string, srcPath string, outputDir string) error {
	cmd := exec.Command(cmdName, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution failed for %s\n[Stderr]: %s", srcPath, stderr.String())
	}

	output := stdout.Bytes()
	if len(output) == 0 {
		log.Printf("⚠️  Warning: %s produced no output. Skipping.", srcPath)
		return nil
	}

	// Parse JSON output from the script
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		log.Printf("⚠️  Skip: %s is not a valid DAG (failed to parse JSON: %v)", srcPath, err)
		return nil
	}

	// Convert to YAML with 2-space indentation
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // Set indent to 2 spaces (Kubernetes standard)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}

	yamlOutput := buf.Bytes()

	// Save as YAML file (blablabla.py -> blablabla.yaml)
	baseName := filepath.Base(srcPath)
	ext := filepath.Ext(baseName)
	fileName := strings.TrimSuffix(baseName, ext) + ".yaml"
	savePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(savePath, yamlOutput, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	fmt.Printf("   ✨ Compiled: %s -> %s\n", baseName, fileName)
	return nil
}
