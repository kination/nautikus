package executor

import (
	"context"
	"testing"

	workflowv1 "github.com/kination/nautikus/api/v1"
)

// MockExecutor is a mock implementation for testing
type MockExecutor struct {
	taskTypes []workflowv1.TaskType
}

func (m *MockExecutor) Type() []workflowv1.TaskType {
	return m.taskTypes
}

func (m *MockExecutor) Execute(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) error {
	return nil
}

func (m *MockExecutor) GetStatus(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (workflowv1.TaskState, error) {
	return workflowv1.StateCompleted, nil
}

func (m *MockExecutor) Cleanup(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) error {
	return nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if registry.executors == nil {
		t.Fatal("executors map is nil")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	mockExec := &MockExecutor{
		taskTypes: []workflowv1.TaskType{workflowv1.TaskTypeBash, workflowv1.TaskTypePython},
	}

	registry.Register(mockExec)

	// Verify both task types are registered
	if !registry.Has(workflowv1.TaskTypeBash) {
		t.Error("TaskTypeBash should be registered")
	}
	if !registry.Has(workflowv1.TaskTypePython) {
		t.Error("TaskTypePython should be registered")
	}
	if registry.Has(workflowv1.TaskTypeGo) {
		t.Error("TaskTypeGo should not be registered")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	mockExec := &MockExecutor{
		taskTypes: []workflowv1.TaskType{workflowv1.TaskTypeBash},
	}

	registry.Register(mockExec)

	// Test getting registered executor
	exec, err := registry.Get(workflowv1.TaskTypeBash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec == nil {
		t.Fatal("executor should not be nil")
	}

	// Test getting unregistered executor
	_, err = registry.Get(workflowv1.TaskTypePython)
	if err == nil {
		t.Error("expected error for unregistered task type")
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()

	mockExec := &MockExecutor{
		taskTypes: []workflowv1.TaskType{workflowv1.TaskTypeBash},
	}

	registry.Register(mockExec)

	if !registry.Has(workflowv1.TaskTypeBash) {
		t.Error("Has should return true for registered type")
	}
	if registry.Has(workflowv1.TaskTypePython) {
		t.Error("Has should return false for unregistered type")
	}
}

func TestRegistry_Types(t *testing.T) {
	registry := NewRegistry()

	mockExec := &MockExecutor{
		taskTypes: []workflowv1.TaskType{workflowv1.TaskTypeBash, workflowv1.TaskTypePython, workflowv1.TaskTypeGo},
	}

	registry.Register(mockExec)

	types := registry.Types()
	if len(types) != 3 {
		t.Errorf("expected 3 types, got %d", len(types))
	}

	// Verify all types are present
	typeSet := make(map[workflowv1.TaskType]bool)
	for _, tt := range types {
		typeSet[tt] = true
	}

	if !typeSet[workflowv1.TaskTypeBash] {
		t.Error("TaskTypeBash should be in types")
	}
	if !typeSet[workflowv1.TaskTypePython] {
		t.Error("TaskTypePython should be in types")
	}
	if !typeSet[workflowv1.TaskTypeGo] {
		t.Error("TaskTypeGo should be in types")
	}
}

func TestRegistry_Concurrent(t *testing.T) {
	registry := NewRegistry()

	// Test concurrent registration and access
	done := make(chan bool)

	// Concurrent registration
	go func() {
		for i := 0; i < 100; i++ {
			mockExec := &MockExecutor{
				taskTypes: []workflowv1.TaskType{workflowv1.TaskTypeBash},
			}
			registry.Register(mockExec)
		}
		done <- true
	}()

	// Concurrent read
	go func() {
		for i := 0; i < 100; i++ {
			registry.Has(workflowv1.TaskTypeBash)
			registry.Types()
		}
		done <- true
	}()

	<-done
	<-done
}
