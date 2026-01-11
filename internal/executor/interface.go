// Package executor provides the Executor interface and registry for task execution.
// This is the core abstraction for Phase 2 (Executor Interface).
package executor

import (
	"context"

	workflowv1 "github.com/kination/nautikus/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Executor defines the interface for executing tasks.
// Different implementations handle different task types (Pod, Spark, etc.)
type Executor interface {
	// Type returns the task type(s) this executor handles
	Type() []workflowv1.TaskType

	// Execute creates the resources needed to run the task
	Execute(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) error

	// GetStatus retrieves the current status of the task
	GetStatus(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (workflowv1.TaskState, error)

	// Cleanup removes the resources created for the task (optional, OwnerReference handles most cases)
	Cleanup(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) error
}

// ExecutorConfig holds common configuration for executors
type ExecutorConfig struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// BaseExecutor provides common functionality for executors
type BaseExecutor struct {
	Config ExecutorConfig
}

// NewBaseExecutor creates a new BaseExecutor
func NewBaseExecutor(cfg ExecutorConfig) BaseExecutor {
	return BaseExecutor{Config: cfg}
}
