// Package scheduler provides scheduling implementations.
package scheduler

import (
	"context"
	"sync"

	workflowv1 "github.com/kination/nautikus/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("scheduler")

// DefaultScheduler implements the Scheduler interface with configurable policies.
type DefaultScheduler struct {
	mu     sync.RWMutex
	config SchedulerConfig

	// Track active tasks for concurrency control
	activeTasksPerDAG map[string]int
	totalActiveTasks  int
}

// NewScheduler creates a new DefaultScheduler with the given configuration
func NewScheduler(config SchedulerConfig) *DefaultScheduler {
	return &DefaultScheduler{
		config:            config,
		activeTasksPerDAG: make(map[string]int),
	}
}

// NewDefaultScheduler creates a scheduler with default configuration
func NewDefaultScheduler() *DefaultScheduler {
	return NewScheduler(DefaultSchedulerConfig())
}

// Name returns the scheduler name
func (s *DefaultScheduler) Name() string {
	return "default-scheduler"
}

// Policy returns the scheduling policy
func (s *DefaultScheduler) Policy() Policy {
	return s.config.Policy
}

// Config returns the scheduler configuration
func (s *DefaultScheduler) Config() SchedulerConfig {
	return s.config
}

// Schedule determines which tasks should be executed next for a given DAG.
func (s *DefaultScheduler) Schedule(ctx context.Context, dag *workflowv1.Dag) ([]workflowv1.TaskSpec, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build status map for quick lookup
	statusMap := make(map[string]workflowv1.TaskState)
	for _, ts := range dag.Status.TaskStatuses {
		statusMap[ts.Name] = ts.State
	}

	// Count currently active tasks for this DAG
	activeCount := 0
	for _, ts := range dag.Status.TaskStatuses {
		if ts.State == workflowv1.StateRunning || ts.State == workflowv1.StatePending {
			activeCount++
		}
	}

	var candidates []workflowv1.TaskSpec

	// Find tasks that are ready to run
	for _, task := range dag.Spec.Tasks {
		if _, exists := statusMap[task.Name]; exists {
			continue
		}

		allDepsCompleted := true
		for _, dep := range task.Dependencies {
			if statusMap[dep] != workflowv1.StateCompleted {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			candidates = append(candidates, task)
		}
	}

	// Apply scheduling policy
	s.sortByPolicy(candidates, dag)

	// Apply concurrency limit
	availableSlots := int(s.config.MaxActiveTasks) - activeCount
	if availableSlots <= 0 {
		return nil, nil
	}

	if len(candidates) > availableSlots {
		candidates = candidates[:availableSlots]
	}

	return candidates, nil
}

// CanSchedule checks if a task can be scheduled
func (s *DefaultScheduler) CanSchedule(ctx context.Context, task *TaskInfo) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.totalActiveTasks >= int(s.config.MaxActiveTasks) {
		return false, nil
	}
	return true, nil
}

// NotifyTaskStarted updates internal state when a task starts
func (s *DefaultScheduler) NotifyTaskStarted(dagName, taskName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeTasksPerDAG[dagName]++
	s.totalActiveTasks++
}

// NotifyTaskCompleted updates internal state when a task completes
func (s *DefaultScheduler) NotifyTaskCompleted(dagName, taskName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeTasksPerDAG[dagName] > 0 {
		s.activeTasksPerDAG[dagName]--
	}
	if s.totalActiveTasks > 0 {
		s.totalActiveTasks--
	}

	if s.activeTasksPerDAG[dagName] == 0 {
		delete(s.activeTasksPerDAG, dagName)
	}
}

// sortByPolicy sorts candidates based on the configured policy
func (s *DefaultScheduler) sortByPolicy(candidates []workflowv1.TaskSpec, dag *workflowv1.Dag) {
	switch s.config.Policy {
	case PolicyFIFO:
		return
	case PolicyPriority:
		// TODO: Sort by priority when TaskSpec has Priority field
		return
	case PolicyFairShare:
		// TODO: Implement fair share across DAGs
		return
	}
}

// GetActiveTaskCount returns the number of currently active tasks
func (s *DefaultScheduler) GetActiveTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalActiveTasks
}
