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

	// Output 디렉토리 초기화 (없으면 생성)
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

			// 확장자에 따른 처리
			ext := filepath.Ext(d.Name())
			switch ext {
			case ".py":
				return generateJSON("python3", []string{path}, path, outputDir)
			case ".go":
				// Go 파일은 'go run'으로 실행
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

// generateJSON은 스크립트(py/go)를 실행하고 표준 출력을 파일로 저장합니다.
func generateJSON(cmdName string, cmdArgs []string, srcPath string, outputDir string) error {
	cmd := exec.Command(cmdName, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 실행
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution failed for %s\n[Stderr]: %s", srcPath, stderr.String())
	}

	output := stdout.Bytes()
	if len(output) == 0 {
		log.Printf("⚠️  Warning: %s produced no output. Skipping.", srcPath)
		return nil
	}

	// 파일 저장 (example.py -> example.json)
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
