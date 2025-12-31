package main

import (
	"fmt"
	"os"

	"github.com/kination/nautikus/internal/compiler"
	"github.com/spf13/cobra"
)

var (
	configPath string
	outputDir  string
)

var rootCmd = &cobra.Command{
	Use:   "dag-cli",
	Short: "Nautikus DAG Compiler - Convert Go/Python code to Kubernetes DAG manifests",
	Long: `Nautikus DAG Compiler is a tool that compiles DAG definitions written in 
Go or Python into Kubernetes-compatible YAML manifests.

The compiler reads your code, executes it to generate JSON output, 
and converts it to properly formatted YAML files.`,
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile DAG definitions from code to YAML manifests",
	Long: `Compile DAG definitions from Go or Python source files into 
Kubernetes YAML manifests. The compiler will:
  1. Scan the configured source directories
  2. Execute .go and .py files
  3. Capture JSON output
  4. Convert to YAML format
  5. Save to the output directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Starting Nautikus DAG Compiler...")
		fmt.Printf("   - Config: %s\n", configPath)
		fmt.Printf("   - Output: %s\n", outputDir)

		if err := compiler.CompileDags(configPath, outputDir); err != nil {
			return fmt.Errorf("compilation failed: %w", err)
		}

		fmt.Println("✅ All DAGs compiled successfully!")
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of dag-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Nautikus DAG CLI v0.1.0")
	},
}

func init() {
	// Add flags to compile command
	compileCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to the configuration file")
	compileCmd.Flags().StringVarP(&outputDir, "out", "o", "dist", "Directory to save generated YAML files")

	// Add commands to root
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}
