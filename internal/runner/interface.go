// Package runner provides task execution capabilities.
// Runner is responsible for actually executing scheduled tasks using executors.
package runner

import (
	"context"

	workflowv1 "github.com/kination/nautikus/api/v1"
)

// Runner defines the interface for task execution.
// It receives scheduled tasks and uses executors to run them.
type Runner interface {
	// Run executes a task and returns the result
	Run(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (*RunResult, error)

	// GetStatus checks the current status of a running task
	GetStatus(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (workflowv1.TaskState, error)
}

// RunResult contains the result of task execution
type RunResult struct {
	// TaskName is the name of the executed task
	TaskName string

	// PodName is the name of the created Pod (or other resource)
	PodName string

	// State is the initial state after execution started
	State workflowv1.TaskState

	// Message contains any additional information
	Message string
}

// RunnerConfig holds configuration for the runner
type RunnerConfig struct {
	// MaxRetries is the maximum number of retries for failed tasks
	MaxRetries int

	// RetryBackoff is the backoff duration between retries (in seconds)
	RetryBackoffSeconds int
}

// DefaultRunnerConfig returns the default runner configuration
func DefaultRunnerConfig() RunnerConfig {
	return RunnerConfig{
		MaxRetries:          0,
		RetryBackoffSeconds: 30,
	}
}
