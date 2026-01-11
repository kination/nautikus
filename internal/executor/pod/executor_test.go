package pod

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	workflowv1 "github.com/kination/nautikus/api/v1"
	"github.com/kination/nautikus/internal/executor"
)

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = workflowv1.AddToScheme(scheme)
	return scheme
}

func TestNew(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	if exec == nil {
		t.Fatal("New returned nil")
	}
}

func TestExecutor_Type(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	types := exec.Type()
	if len(types) != 3 {
		t.Errorf("expected 3 task types, got %d", len(types))
	}

	// Verify all expected types
	typeSet := make(map[workflowv1.TaskType]bool)
	for _, tt := range types {
		typeSet[tt] = true
	}

	if !typeSet[workflowv1.TaskTypeBash] {
		t.Error("should support TaskTypeBash")
	}
	if !typeSet[workflowv1.TaskTypePython] {
		t.Error("should support TaskTypePython")
	}
	if !typeSet[workflowv1.TaskTypeGo] {
		t.Error("should support TaskTypeGo")
	}
}

func TestExecutor_BuildPod_Bash(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	dag := &workflowv1.Dag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag",
			Namespace: "default",
		},
	}

	task := &workflowv1.TaskSpec{
		Name:    "test-task",
		Type:    workflowv1.TaskTypeBash,
		Command: "echo hello",
	}

	pod := exec.buildPod(dag, task)

	// Verify pod name
	expectedName := "test-dag-test-task"
	if pod.Name != expectedName {
		t.Errorf("expected pod name %s, got %s", expectedName, pod.Name)
	}

	// Verify image
	if pod.Spec.Containers[0].Image != "ubuntu:latest" {
		t.Errorf("expected ubuntu:latest image for Bash, got %s", pod.Spec.Containers[0].Image)
	}

	// Verify command
	if pod.Spec.Containers[0].Command[0] != "/bin/bash" {
		t.Errorf("expected /bin/bash command, got %v", pod.Spec.Containers[0].Command)
	}

	// Verify args contain the command
	if pod.Spec.Containers[0].Args[0] != "echo hello" {
		t.Errorf("expected 'echo hello' in args, got %v", pod.Spec.Containers[0].Args)
	}
}

func TestExecutor_BuildPod_Python(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	dag := &workflowv1.Dag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag",
			Namespace: "default",
		},
	}

	task := &workflowv1.TaskSpec{
		Name:   "python-task",
		Type:   workflowv1.TaskTypePython,
		Script: "print('hello')",
	}

	pod := exec.buildPod(dag, task)

	// Verify image
	if pod.Spec.Containers[0].Image != "python:3.9-slim" {
		t.Errorf("expected python:3.9-slim image, got %s", pod.Spec.Containers[0].Image)
	}

	// Verify command
	if pod.Spec.Containers[0].Command[0] != "python" {
		t.Errorf("expected python command, got %v", pod.Spec.Containers[0].Command)
	}
}

func TestExecutor_BuildPod_Go(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	dag := &workflowv1.Dag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag",
			Namespace: "default",
		},
	}

	task := &workflowv1.TaskSpec{
		Name:   "go-task",
		Type:   workflowv1.TaskTypeGo,
		Script: "package main\nfunc main() {}",
	}

	pod := exec.buildPod(dag, task)

	// Verify image
	if pod.Spec.Containers[0].Image != "golang:1.20-alpine" {
		t.Errorf("expected golang:1.20-alpine image, got %s", pod.Spec.Containers[0].Image)
	}
}

func TestExecutor_BuildPod_WithEnv(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	dag := &workflowv1.Dag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag",
			Namespace: "default",
		},
	}

	task := &workflowv1.TaskSpec{
		Name:    "env-task",
		Type:    workflowv1.TaskTypeBash,
		Command: "echo $MY_VAR",
		Env: map[string]string{
			"MY_VAR": "test-value",
		},
	}

	pod := exec.buildPod(dag, task)

	// Verify env var is set
	found := false
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == "MY_VAR" && env.Value == "test-value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected MY_VAR env var to be set")
	}
}

func TestExecutor_GetStatus(t *testing.T) {
	scheme := newTestScheme()

	tests := []struct {
		name          string
		podPhase      corev1.PodPhase
		expectedState workflowv1.TaskState
	}{
		{
			name:          "Succeeded",
			podPhase:      corev1.PodSucceeded,
			expectedState: workflowv1.StateCompleted,
		},
		{
			name:          "Failed",
			podPhase:      corev1.PodFailed,
			expectedState: workflowv1.StateFailed,
		},
		{
			name:          "Running",
			podPhase:      corev1.PodRunning,
			expectedState: workflowv1.StateRunning,
		},
		{
			name:          "Pending",
			podPhase:      corev1.PodPending,
			expectedState: workflowv1.StatePending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dag-test-task",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: tt.podPhase,
				},
			}

			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()

			exec := New(executor.ExecutorConfig{
				Client: client,
				Scheme: scheme,
			})

			dag := &workflowv1.Dag{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dag",
					Namespace: "default",
				},
			}

			task := &workflowv1.TaskSpec{
				Name: "test-task",
				Type: workflowv1.TaskTypeBash,
			}

			state, err := exec.GetStatus(context.Background(), dag, task)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if state != tt.expectedState {
				t.Errorf("expected state %s, got %s", tt.expectedState, state)
			}
		})
	}
}

func TestExecutor_GetStatus_PodNotFound(t *testing.T) {
	scheme := newTestScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	dag := &workflowv1.Dag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag",
			Namespace: "default",
		},
	}

	task := &workflowv1.TaskSpec{
		Name: "test-task",
		Type: workflowv1.TaskTypeBash,
	}

	state, err := exec.GetStatus(context.Background(), dag, task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When pod is not found, should return Pending
	if state != workflowv1.StatePending {
		t.Errorf("expected StatePending for missing pod, got %s", state)
	}
}

func TestExecutor_Cleanup(t *testing.T) {
	scheme := newTestScheme()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag-test-task",
			Namespace: "default",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()

	exec := New(executor.ExecutorConfig{
		Client: client,
		Scheme: scheme,
	})

	dag := &workflowv1.Dag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dag",
			Namespace: "default",
		},
	}

	task := &workflowv1.TaskSpec{
		Name: "test-task",
		Type: workflowv1.TaskTypeBash,
	}

	err := exec.Cleanup(context.Background(), dag, task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify pod is deleted
	deletedPod := &corev1.Pod{}
	err = client.Get(context.Background(), types.NamespacedName{Name: "test-dag-test-task", Namespace: "default"}, deletedPod)
	if err == nil {
		t.Error("pod should be deleted")
	}
}
