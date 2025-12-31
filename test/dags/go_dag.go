package main

import (
	"fmt"

	"github.com/kination/nautikus/internal/sdk"
)

const (
// Placeholder to ensure imports are kept if needed, though they aren't for the SDK usage.
)

// -----------------------------------------------------------
// 1. User Logic (Pure Go)
// -----------------------------------------------------------

func task1() {
	fmt.Println("This is task 1 running inside the Pod!")
	fmt.Println("Doing some work...")
}

func task2() {
	fmt.Println("This is task 2 running inside the Pod!")
	fmt.Println("Task 2 completed.")
}

func main() {
	// Define DAG and task execution order naturally
	sdk.Serve(
		"go-generated-dag",
		[]func(){task1, task2}, // Sequential execution: task1 -> task2
	)
}
