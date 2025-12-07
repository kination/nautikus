package compiler

import (
	"bytes"
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
				return generateJSON("python3", []string{path}, path, outputDir)
			case ".go":
				return generateJSON("go", []string{"run", path}, path, outputDir)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("walk error in %s: %w", src.Location, err)
		}
	}
	return nil
}

// generateJSON is a helper function that runs a command and saves the standard output to a file.
func generateJSON(cmdName string, cmdArgs []string, srcPath string, outputDir string) error {
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

	// convert output to file (blablabla.py -> blablabla.json)
	baseName := filepath.Base(srcPath)
	ext := filepath.Ext(baseName)
	fileName := strings.TrimSuffix(baseName, ext) + ".json"
	savePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(savePath, output, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	fmt.Printf("   ✨ Compiled: %s -> %s\n", baseName, fileName)
	return nil
}
