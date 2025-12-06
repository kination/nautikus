package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	workflowv1 "github.com/kination/pequod/api/v1"
)

// DagReconciler reconciles a Dag object
// +kubebuilder:rbac:groups=workflow.pequod.io,resources=dags,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workflow.pequod.io,resources=dags/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workflow.pequod.io,resources=dags/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
type DagReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DagReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Bring DAG CR
	var dag workflowv1.Dag
	if err := r.Get(ctx, req.NamespacedName, &dag); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Sync current running Pod and Status
	if err := r.syncStatus(ctx, &dag); err != nil {
		return ctrl.Result{}, err
	}

	// Find next task to run (dependency check)
	nextTasks := r.getNextTasks(&dag)

	// Create Pod
	for _, task := range nextTasks {
		pod := r.buildPod(&dag, task)
		// Set owner reference (Pod will be deleted when DAG is deleted)
		if err := controllerutil.SetControllerReference(&dag, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Pod", "Pod.Name", pod.Name, "Task.Name", task.Name)
		if err := r.Create(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update Status
	if err := r.Status().Update(ctx, &dag); err != nil {
		return ctrl.Result{}, err
	}

	// If DAG is not finished, continue Reconcile
	if dag.Status.State == workflowv1.StateRunning {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// Sync Pod status to DAG status
func (r *DagReconciler) syncStatus(ctx context.Context, dag *workflowv1.Dag) error {
	// Initialize map
	if dag.Status.TaskStatuses == nil {
		dag.Status.TaskStatuses = []workflowv1.TaskStatus{}
	}

	// If no state, set to Running
	if dag.Status.State == "" {
		dag.Status.State = workflowv1.StateRunning
	}

	// Sync actual Pod status
	statusMap := make(map[string]*workflowv1.TaskStatus)
	for i := range dag.Status.TaskStatuses {
		t := &dag.Status.TaskStatuses[i]
		statusMap[t.Name] = t
	}

	for _, taskSpec := range dag.Spec.Tasks {
		// Skip if already completed or failed
		currentStatus, exists := statusMap[taskSpec.Name]
		if !exists {
			continue
		}
		if currentStatus.State == workflowv1.StateCompleted || currentStatus.State == workflowv1.StateFailed {
			continue
		}

		// Bring Pod
		pod := &corev1.Pod{}
		err := r.Get(ctx, types.NamespacedName{Name: currentStatus.PodName, Namespace: dag.Namespace}, pod)
		if err != nil {
			if errors.IsNotFound(err) {
				// Skip if Pod not found
				continue
			}
			return err
		}

		// Update Pod status
		if pod.Status.Phase == corev1.PodSucceeded {
			currentStatus.State = workflowv1.StateCompleted
		} else if pod.Status.Phase == corev1.PodFailed {
			currentStatus.State = workflowv1.StateFailed
			dag.Status.State = workflowv1.StateFailed // If one task fails, mark DAG as failed
		} else {
			currentStatus.State = workflowv1.StateRunning
		}
	}
	return nil
}

// Find next task to run (dependency check)
func (r *DagReconciler) getNextTasks(dag *workflowv1.Dag) []workflowv1.TaskSpec {
	var nextTasks []workflowv1.TaskSpec

	// Initialize status map
	statusMap := make(map[string]workflowv1.TaskState)
	for _, s := range dag.Status.TaskStatuses {
		statusMap[s.Name] = s.State
	}

	for _, task := range dag.Spec.Tasks {
		// Skip if already running or completed
		if state, ok := statusMap[task.Name]; ok && state != "" {
			continue
		}

		// Dependency check
		allDepsCompleted := true
		for _, dep := range task.Dependencies {
			if statusMap[dep] != workflowv1.StateCompleted {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			nextTasks = append(nextTasks, task)
			// Prevent duplicate execution by adding status to Pending (actual update at Reconcile end)
			dag.Status.TaskStatuses = append(dag.Status.TaskStatuses, workflowv1.TaskStatus{
				Name:    task.Name,
				State:   workflowv1.StatePending,
				PodName: fmt.Sprintf("%s-%s", dag.Name, task.Name), // Pod name rule
			})
		}
	}
	return nextTasks
}

// Convert TaskSpec to Pod (Bash, Python, Go Operator logic)
func (r *DagReconciler) buildPod(dag *workflowv1.Dag, task workflowv1.TaskSpec) *corev1.Pod {
	podName := fmt.Sprintf("%s-%s", dag.Name, task.Name)

	var image string
	var command []string
	var args []string

	switch task.Type {
	case workflowv1.TaskTypeBash:
		image = "ubuntu:latest"
		command = []string{"/bin/bash", "-c"}
		args = []string{task.Command}
	case workflowv1.TaskTypePython:
		image = "python:3.9-slim"
		// Python code inline execution
		command = []string{"python", "-c"}
		args = []string{task.Script}
	case workflowv1.TaskTypeGo:
		image = "golang:1.20-alpine"
		// Go code inline execution
		// TODO: Use ConfigMap to mount code
		command = []string{"/bin/sh", "-c"}
		// Simple inline execution example (complex code should use ConfigMap)
		goCmd := fmt.Sprintf("echo '%s' > main.go && go run main.go", task.Script)
		args = []string{goCmd}
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: dag.Namespace,
			Labels: map[string]string{
				"dag":  dag.Name,
				"task": task.Name,
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
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DagReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workflowv1.Dag{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
