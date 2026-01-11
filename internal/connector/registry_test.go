package connector

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workflowv1 "github.com/kination/nautikus/api/v1"
)

// MockExternalConnector is a mock implementation for testing
type MockExternalConnector struct {
	connType string
}

func (m *MockExternalConnector) Type() string {
	return m.connType
}

func (m *MockExternalConnector) BuildResource(task *workflowv1.TaskSpec, dag *workflowv1.Dag) (*unstructured.Unstructured, error) {
	return &unstructured.Unstructured{}, nil
}

func (m *MockExternalConnector) GetStatus(ctx context.Context, cl client.Client, task *workflowv1.TaskSpec, dag *workflowv1.Dag) (workflowv1.TaskState, error) {
	return workflowv1.StateCompleted, nil
}

func (m *MockExternalConnector) Cleanup(ctx context.Context, cl client.Client, task *workflowv1.TaskSpec, dag *workflowv1.Dag) error {
	return nil
}

// MockCloudConnector is a mock implementation for testing
type MockCloudConnector struct {
	connType string
}

func (m *MockCloudConnector) Type() string {
	return m.connType
}

func (m *MockCloudConnector) Submit(ctx context.Context, task *workflowv1.TaskSpec, dag *workflowv1.Dag) (string, error) {
	return "job-123", nil
}

func (m *MockCloudConnector) GetStatus(ctx context.Context, jobID string) (workflowv1.TaskState, error) {
	return workflowv1.StateCompleted, nil
}

func (m *MockCloudConnector) Cancel(ctx context.Context, jobID string) error {
	return nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if registry.externalConnectors == nil {
		t.Fatal("externalConnectors map is nil")
	}
	if registry.cloudConnectors == nil {
		t.Fatal("cloudConnectors map is nil")
	}
}

func TestRegistry_RegisterExternal(t *testing.T) {
	registry := NewRegistry()

	mockConn := &MockExternalConnector{connType: "kubeflow"}
	registry.RegisterExternal(mockConn)

	if !registry.HasExternal("kubeflow") {
		t.Error("kubeflow connector should be registered")
	}
	if registry.HasExternal("spark") {
		t.Error("spark connector should not be registered")
	}
}

func TestRegistry_RegisterCloud(t *testing.T) {
	registry := NewRegistry()

	mockConn := &MockCloudConnector{connType: "sagemaker"}
	registry.RegisterCloud(mockConn)

	if !registry.HasCloud("sagemaker") {
		t.Error("sagemaker connector should be registered")
	}
	if registry.HasCloud("vertex") {
		t.Error("vertex connector should not be registered")
	}
}

func TestRegistry_GetExternal(t *testing.T) {
	registry := NewRegistry()

	mockConn := &MockExternalConnector{connType: "kubeflow"}
	registry.RegisterExternal(mockConn)

	// Test getting registered connector
	conn, err := registry.GetExternal("kubeflow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("connector should not be nil")
	}
	if conn.Type() != "kubeflow" {
		t.Errorf("expected type 'kubeflow', got %s", conn.Type())
	}

	// Test getting unregistered connector
	_, err = registry.GetExternal("spark")
	if err == nil {
		t.Error("expected error for unregistered connector type")
	}
}

func TestRegistry_GetCloud(t *testing.T) {
	registry := NewRegistry()

	mockConn := &MockCloudConnector{connType: "sagemaker"}
	registry.RegisterCloud(mockConn)

	// Test getting registered connector
	conn, err := registry.GetCloud("sagemaker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("connector should not be nil")
	}
	if conn.Type() != "sagemaker" {
		t.Errorf("expected type 'sagemaker', got %s", conn.Type())
	}

	// Test getting unregistered connector
	_, err = registry.GetCloud("vertex")
	if err == nil {
		t.Error("expected error for unregistered connector type")
	}
}

func TestRegistry_HasExternal(t *testing.T) {
	registry := NewRegistry()

	mockConn := &MockExternalConnector{connType: "kubeflow"}
	registry.RegisterExternal(mockConn)

	if !registry.HasExternal("kubeflow") {
		t.Error("HasExternal should return true for registered type")
	}
	if registry.HasExternal("spark") {
		t.Error("HasExternal should return false for unregistered type")
	}
}

func TestRegistry_HasCloud(t *testing.T) {
	registry := NewRegistry()

	mockConn := &MockCloudConnector{connType: "sagemaker"}
	registry.RegisterCloud(mockConn)

	if !registry.HasCloud("sagemaker") {
		t.Error("HasCloud should return true for registered type")
	}
	if registry.HasCloud("vertex") {
		t.Error("HasCloud should return false for unregistered type")
	}
}

func TestRegistry_ExternalTypes(t *testing.T) {
	registry := NewRegistry()

	mockConn1 := &MockExternalConnector{connType: "kubeflow"}
	mockConn2 := &MockExternalConnector{connType: "spark"}
	registry.RegisterExternal(mockConn1)
	registry.RegisterExternal(mockConn2)

	types := registry.ExternalTypes()
	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}

	typeSet := make(map[string]bool)
	for _, tt := range types {
		typeSet[tt] = true
	}

	if !typeSet["kubeflow"] {
		t.Error("kubeflow should be in types")
	}
	if !typeSet["spark"] {
		t.Error("spark should be in types")
	}
}

func TestRegistry_CloudTypes(t *testing.T) {
	registry := NewRegistry()

	mockConn1 := &MockCloudConnector{connType: "sagemaker"}
	mockConn2 := &MockCloudConnector{connType: "vertex"}
	registry.RegisterCloud(mockConn1)
	registry.RegisterCloud(mockConn2)

	types := registry.CloudTypes()
	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}

	typeSet := make(map[string]bool)
	for _, tt := range types {
		typeSet[tt] = true
	}

	if !typeSet["sagemaker"] {
		t.Error("sagemaker should be in types")
	}
	if !typeSet["vertex"] {
		t.Error("vertex should be in types")
	}
}

func TestRegistry_Concurrent(t *testing.T) {
	registry := NewRegistry()

	done := make(chan bool)

	// Concurrent registration
	go func() {
		for i := 0; i < 100; i++ {
			mockExt := &MockExternalConnector{connType: "kubeflow"}
			mockCloud := &MockCloudConnector{connType: "sagemaker"}
			registry.RegisterExternal(mockExt)
			registry.RegisterCloud(mockCloud)
		}
		done <- true
	}()

	// Concurrent read
	go func() {
		for i := 0; i < 100; i++ {
			registry.HasExternal("kubeflow")
			registry.HasCloud("sagemaker")
			registry.ExternalTypes()
			registry.CloudTypes()
		}
		done <- true
	}()

	<-done
	<-done
}
