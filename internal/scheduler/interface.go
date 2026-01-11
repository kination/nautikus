// Package scheduler provides scheduling interfaces and implementations.
// This is the core abstraction for Phase 4 (Scheduler/Runner Separation).
package scheduler

import (
	"context"

	workflowv1 "github.com/kination/nautikus/api/v1"
)

// Policy defines the scheduling policy type
type Policy string

const (
	// PolicyFIFO schedules tasks in first-in-first-out order
	PolicyFIFO Policy = "FIFO"
	// PolicyPriority schedules tasks based on priority
	PolicyPriority Policy = "Priority"
	// PolicyFairShare schedules tasks with fair resource sharing
	PolicyFairShare Policy = "FairShare"
)

// TaskInfo contains task information for scheduling decisions
type TaskInfo struct {
	DagName   string
	TaskName  string
	Priority  int32
	Resources ResourceRequirements
}

// ResourceRequirements defines resource requests/limits
type ResourceRequirements struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
	GPUCount      int32
}

// Scheduler defines the interface for task scheduling.
// Implementations provide different scheduling strategies.
type Scheduler interface {
	// Name returns the scheduler name
	Name() string

	// Policy returns the scheduling policy
	Policy() Policy

	// Schedule determines which tasks should be executed next
	Schedule(ctx context.Context, dag *workflowv1.Dag) ([]workflowv1.TaskSpec, error)

	// CanSchedule checks if a task can be scheduled based on resources and constraints
	CanSchedule(ctx context.Context, task *TaskInfo) (bool, error)
}

// Queue defines the interface for task queuing
type Queue interface {
	// Enqueue adds a task to the queue
	Enqueue(task *TaskInfo) error

	// Dequeue removes and returns the next task from the queue
	Dequeue() (*TaskInfo, error)

	// Peek returns the next task without removing it
	Peek() (*TaskInfo, error)

	// Len returns the queue length
	Len() int
}

// SchedulerConfig holds common configuration for schedulers
type SchedulerConfig struct {
	Policy           Policy
	MaxConcurrentDAGs int32
	MaxActiveTasks    int32
}

// DefaultSchedulerConfig returns the default scheduler configuration
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		Policy:           PolicyFIFO,
		MaxConcurrentDAGs: 100,
		MaxActiveTasks:    10,
	}
}
