// Package pod provides the PodExecutor for running tasks as Kubernetes Pods.
package pod

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	workflowv1 "github.com/kination/nautikus/api/v1"
	"github.com/kination/nautikus/internal/executor"
)

// Executor implements the executor.Executor interface for Pod-based tasks
type Executor struct {
	executor.BaseExecutor
}

// New creates a new PodExecutor
func New(cfg executor.ExecutorConfig) *Executor {
	return &Executor{
		BaseExecutor: executor.NewBaseExecutor(cfg),
	}
}

// Type returns the task types this executor handles
func (e *Executor) Type() []workflowv1.TaskType {
	return []workflowv1.TaskType{
		workflowv1.TaskTypeBash,
		workflowv1.TaskTypePython,
		workflowv1.TaskTypeGo,
	}
}

// Execute creates a Pod to run the task
func (e *Executor) Execute(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) error {
	pod := e.buildPod(dag, task)

	// Set owner reference (Pod will be deleted when DAG is deleted)
	if err := controllerutil.SetControllerReference(dag, pod, e.Config.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	if err := e.Config.Client.Create(ctx, pod); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil // Pod already exists, skip
		}
		return fmt.Errorf("failed to create pod: %w", err)
	}

	return nil
}

// GetStatus retrieves the current status of the task by checking the Pod status
func (e *Executor) GetStatus(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) (workflowv1.TaskState, error) {
	podName := e.getPodName(dag.Name, task.Name)

	pod := &corev1.Pod{}
	err := e.Config.Client.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: dag.Namespace,
	}, pod)

	if err != nil {
		if errors.IsNotFound(err) {
			return workflowv1.StatePending, nil
		}
		return workflowv1.StateFailed, err
	}

	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		return workflowv1.StateCompleted, nil
	case corev1.PodFailed:
		return workflowv1.StateFailed, nil
	case corev1.PodRunning:
		return workflowv1.StateRunning, nil
	default:
		return workflowv1.StatePending, nil
	}
}

// Cleanup removes the Pod (usually handled by OwnerReference)
func (e *Executor) Cleanup(ctx context.Context, dag *workflowv1.Dag, task *workflowv1.TaskSpec) error {
	podName := e.getPodName(dag.Name, task.Name)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: dag.Namespace,
		},
	}

	if err := e.Config.Client.Delete(ctx, pod); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

// buildPod converts TaskSpec to Pod
func (e *Executor) buildPod(dag *workflowv1.Dag, task *workflowv1.TaskSpec) *corev1.Pod {
	podName := e.getPodName(dag.Name, task.Name)

	image, command, args := e.getContainerSpec(task)

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: dag.Namespace,
			Labels: map[string]string{
				"dag":                      dag.Name,
				"task":                     task.Name,
				"app.kubernetes.io/name":  "nautikus",
				"app.kubernetes.io/part-of": "nautikus",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "task-runner",
					Image:   image,
					Command: command,
					Args:    args,
					Env:     e.buildEnv(task.Env),
				},
			},
		},
	}
}

// getContainerSpec returns image, command, and args based on task type
func (e *Executor) getContainerSpec(task *workflowv1.TaskSpec) (string, []string, []string) {
	var image string
	var command []string
	var args []string

	// Use custom image if specified
	if task.Image != "" {
		image = task.Image
	}

	switch task.Type {
	case workflowv1.TaskTypeBash:
		if image == "" {
			image = "ubuntu:latest"
		}
		command = []string{"/bin/bash", "-c"}
		args = []string{task.Command}

	case workflowv1.TaskTypePython:
		if image == "" {
			image = "python:3.9-slim"
		}
		command = []string{"python", "-c"}
		args = []string{task.Script}

	case workflowv1.TaskTypeGo:
		if image == "" {
			image = "golang:1.20-alpine"
		}
		command = []string{"/bin/sh", "-c"}
		goCmd := fmt.Sprintf("echo '%s' > main.go && go mod init dag && go mod tidy && go run main.go", task.Script)
		args = []string{goCmd}
	}

	return image, command, args
}

// buildEnv converts map to EnvVar slice
func (e *Executor) buildEnv(envMap map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for k, v := range envMap {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envVars
}

// getPodName generates the pod name from dag and task names
func (e *Executor) getPodName(dagName, taskName string) string {
	return fmt.Sprintf("%s-%s", dagName, taskName)
}
