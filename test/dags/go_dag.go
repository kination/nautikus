package main

import (
	"fmt"
	"math/rand"

	sdk "github.com/kination/nautikus/pkg/sdk/go"
)

// -----------------------------------------------------------
// User Logic (Pure Go)
// -----------------------------------------------------------

func extractData() {
	fmt.Println("Extracting data from source...")
	fmt.Println("Data extraction complete.")
}

func validateData() {
	fmt.Println("Validating extracted data...")
	fmt.Println("Validation passed.")
}

func checkDataQuality() string {
	// Simulate quality check - returns branch name
	score := rand.Intn(100)
	fmt.Printf("Data quality score: %d\n", score)
	if score >= 70 {
		return "high_quality"
	}
	return "low_quality"
}

func processHighQuality() {
	fmt.Println("Processing high quality data - full pipeline")
}

func processLowQuality() {
	fmt.Println("Processing low quality data - applying corrections")
}

func cleanData() {
	fmt.Println("Cleaning and normalizing data...")
}

func loadData() {
	fmt.Println("Loading processed data to destination...")
	fmt.Println("Pipeline complete!")
}

func main() {
	sdk.NewDAG("go-etl-pipeline").
		// Sequential tasks: extract -> validate
		AddSequential(
			sdk.Task{Name: "extract", Fn: extractData},
			sdk.Task{Name: "validate", Fn: validateData},
		).
		// Conditional branching based on data quality
		AddBranch("check_quality", checkDataQuality, map[string][]sdk.Task{
			"high_quality": {
				{Name: "process_high", Fn: processHighQuality},
			},
			"low_quality": {
				{Name: "process_low", Fn: processLowQuality},
				{Name: "clean", Fn: cleanData},
			},
		}).
		// Join: wait for either branch to complete
		AddJoin("load", loadData, "process_high", "clean").
		Serve()
}
