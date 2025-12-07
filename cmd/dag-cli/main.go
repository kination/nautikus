package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kination/nautikus/internal/compiler"
)

func main() {
	// 플래그 설정
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	outputDir := flag.String("out", "dist", "Directory to save generated JSON files")
	flag.Parse()

	fmt.Println("🚀 Starting Nautikus DAG Compiler...")
	fmt.Printf("   - Config: %s\n", *configPath)
	fmt.Printf("   - Output: %s\n", *outputDir)

	// 컴파일 로직 실행
	if err := compiler.CompileDags(*configPath, *outputDir); err != nil {
		fmt.Printf("❌ Compilation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All DAGs compiled successfully!")
}
