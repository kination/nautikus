package runner

import (
	"context"
	"fmt"

	workflowv1 "github.com/kination/nautikus/api/v1"
	"github.com/kination/nautikus/internal/executor"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("runner")

// DefaultRunner implements the Runner interface using the executor registry.
type DefaultRunner struct {
	executorRegistry *executor.Registry
	config           RunnerConfig
}

// NewRunner creates a new DefaultRunner with the given executor registry
func NewRunner(registry *executor.Registry, config RunnerConfig) *DefaultRunner {
	return &DefaultRunner{
		executorRegistry: registry,
		config:           config,
	}
}

// NewDefaultRunner creates a runner with default configuration
func NewDefaultRunner(registry *executor.Registry) *DefaultRunner {
	return NewRunner(registry, DefaultRunnerConfig())
}

// Run executes a task using the appropriate executor
func (r *DefaultRunner) Run(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (*RunResult, error) {
	// Get the executor for this task type
	exec, err := r.executorRegistry.Get(task.Type)
	if err != nil {
		return nil, fmt.Errorf("no executor found for task type %s: %w", task.Type, err)
	}

	log.Info("Running task", "dag", dag.Name, "task", task.Name, "type", task.Type)

	// Execute the task
	if err := exec.Execute(ctx, dag, task); err != nil {
		return &RunResult{
			TaskName: task.Name,
			PodName:  fmt.Sprintf("%s-%s", dag.Name, task.Name),
			State:    workflowv1.StateFailed,
			Message:  err.Error(),
		}, err
	}

	return &RunResult{
		TaskName: task.Name,
		PodName:  fmt.Sprintf("%s-%s", dag.Name, task.Name),
		State:    workflowv1.StatePending,
		Message:  "Task started",
	}, nil
}

// GetStatus retrieves the current status of a task
func (r *DefaultRunner) GetStatus(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (workflowv1.TaskState, error) {
	exec, err := r.executorRegistry.Get(task.Type)
	if err != nil {
		return workflowv1.StateFailed, fmt.Errorf("no executor found for task type %s: %w", task.Type, err)
	}

	return exec.GetStatus(ctx, dag, task)
}

// Config returns the runner configuration
func (r *DefaultRunner) Config() RunnerConfig {
	return r.config
}

// ExecutorRegistry returns the executor registry
func (r *DefaultRunner) ExecutorRegistry() *executor.Registry {
	return r.executorRegistry
}
