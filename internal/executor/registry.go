package executor

import (
	"fmt"
	"sync"

	workflowv1 "github.com/kination/nautikus/api/v1"
)

// Registry manages executor registration and lookup
type Registry struct {
	mu        sync.RWMutex
	executors map[workflowv1.TaskType]Executor
}

// NewRegistry creates a new executor registry
func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[workflowv1.TaskType]Executor),
	}
}

// Register adds an executor to the registry
func (r *Registry) Register(exec Executor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, taskType := range exec.Type() {
		r.executors[taskType] = exec
	}
}

// Get retrieves an executor for the given task type
func (r *Registry) Get(taskType workflowv1.TaskType) (Executor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exec, ok := r.executors[taskType]
	if !ok {
		return nil, fmt.Errorf("no executor registered for task type: %s", taskType)
	}
	return exec, nil
}

// Has checks if an executor is registered for the given task type
func (r *Registry) Has(taskType workflowv1.TaskType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.executors[taskType]
	return ok
}

// Types returns all registered task types
func (r *Registry) Types() []workflowv1.TaskType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]workflowv1.TaskType, 0, len(r.executors))
	for t := range r.executors {
		types = append(types, t)
	}
	return types
}
