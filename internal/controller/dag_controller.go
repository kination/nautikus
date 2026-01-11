package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	workflowv1 "github.com/kination/nautikus/api/v1"
	"github.com/kination/nautikus/internal/executor"
	podexecutor "github.com/kination/nautikus/internal/executor/pod"
	"github.com/kination/nautikus/internal/runner"
	"github.com/kination/nautikus/internal/scheduler"
)

// DagReconciler reconciles a Dag object
// +kubebuilder:rbac:groups=workflow.nautikus.io,resources=dags,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workflow.nautikus.io,resources=dags/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workflow.nautikus.io,resources=dags/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
type DagReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Separated components for Phase B architecture
	ExecutorRegistry *executor.Registry
	Scheduler        *scheduler.DefaultScheduler
	Runner           *runner.DefaultRunner
}

// SetupComponents initializes the executor registry, scheduler, and runner
func (r *DagReconciler) SetupComponents() {
	// Setup executor registry
	if r.ExecutorRegistry == nil {
		r.ExecutorRegistry = executor.NewRegistry()
	}

	// Register the Pod executor for Bash, Python, Go task types
	podExec := podexecutor.New(executor.ExecutorConfig{
		Client: r.Client,
		Scheme: r.Scheme,
	})
	r.ExecutorRegistry.Register(podExec)

	// Setup scheduler with default config
	if r.Scheduler == nil {
		r.Scheduler = scheduler.NewDefaultScheduler()
	}

	// Setup runner with executor registry
	if r.Runner == nil {
		r.Runner = runner.NewDefaultRunner(r.ExecutorRegistry)
	}
}

// SetupExecutors is kept for backward compatibility
func (r *DagReconciler) SetupExecutors() {
	r.SetupComponents()
}

func (r *DagReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Ensure components are set up
	if r.Scheduler == nil || r.Runner == nil {
		r.SetupComponents()
	}

	// Bring DAG CR
	var dag workflowv1.Dag
	if err := r.Get(ctx, req.NamespacedName, &dag); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Sync current running Pod and Status
	if err := r.syncStatus(ctx, &dag); err != nil {
		return ctrl.Result{}, err
	}

	// Check if DAG is already completed or failed
	if dag.Status.State == workflowv1.StateCompleted || dag.Status.State == workflowv1.StateFailed {
		return ctrl.Result{}, nil
	}

	// Use Scheduler to determine next tasks
	nextTasks, err := r.Scheduler.Schedule(ctx, &dag)
	if err != nil {
		log.Error(err, "Failed to schedule tasks")
		return ctrl.Result{}, err
	}

	// Use Runner to execute scheduled tasks
	for _, task := range nextTasks {
		// Add task to status as Pending before execution
		dag.Status.TaskStatuses = append(dag.Status.TaskStatuses, workflowv1.TaskStatus{
			Name:    task.Name,
			State:   workflowv1.StatePending,
			PodName: fmt.Sprintf("%s-%s", dag.Name, task.Name),
		})

		// Notify scheduler that task is starting
		r.Scheduler.NotifyTaskStarted(dag.Name, task.Name)

		log.Info("Executing task", "dag", dag.Name, "task", task.Name, "type", task.Type)

		// Execute task using runner
		result, err := r.Runner.Run(ctx, &dag, &task)
		if err != nil {
			log.Error(err, "Failed to run task", "task", task.Name)
			// Update task status to failed
			for i := range dag.Status.TaskStatuses {
				if dag.Status.TaskStatuses[i].Name == task.Name {
					dag.Status.TaskStatuses[i].State = workflowv1.StateFailed
					dag.Status.TaskStatuses[i].Message = err.Error()
					break
				}
			}
			dag.Status.State = workflowv1.StateFailed
			r.Scheduler.NotifyTaskCompleted(dag.Name, task.Name)

			// Update status and return error
			if updateErr := r.Status().Update(ctx, &dag); updateErr != nil {
				log.Error(updateErr, "Failed to update DAG status")
			}
			return ctrl.Result{}, err
		}

		// Update task status from result
		for i := range dag.Status.TaskStatuses {
			if dag.Status.TaskStatuses[i].Name == task.Name {
				dag.Status.TaskStatuses[i].State = result.State
				dag.Status.TaskStatuses[i].PodName = result.PodName
				dag.Status.TaskStatuses[i].Message = result.Message
				break
			}
		}
	}

	// Check if all tasks are completed
	r.updateDAGState(&dag)

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

// syncStatus syncs Pod status to DAG status
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

		// Use Runner to get status
		state, err := r.Runner.GetStatus(ctx, dag, &taskSpec)
		if err != nil {
			// Fallback to legacy Pod status check
			pod := &corev1.Pod{}
			err := r.Get(ctx, types.NamespacedName{Name: currentStatus.PodName, Namespace: dag.Namespace}, pod)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return err
			}

			if pod.Status.Phase == corev1.PodSucceeded {
				currentStatus.State = workflowv1.StateCompleted
				r.Scheduler.NotifyTaskCompleted(dag.Name, taskSpec.Name)
			} else if pod.Status.Phase == corev1.PodFailed {
				currentStatus.State = workflowv1.StateFailed
				dag.Status.State = workflowv1.StateFailed
				r.Scheduler.NotifyTaskCompleted(dag.Name, taskSpec.Name)
			} else {
				currentStatus.State = workflowv1.StateRunning
			}
			continue
		}

		// Update status if changed
		if currentStatus.State != state {
			previousState := currentStatus.State
			currentStatus.State = state

			// Notify scheduler when task completes
			if (state == workflowv1.StateCompleted || state == workflowv1.StateFailed) &&
				(previousState == workflowv1.StateRunning || previousState == workflowv1.StatePending) {
				r.Scheduler.NotifyTaskCompleted(dag.Name, taskSpec.Name)
			}

			if state == workflowv1.StateFailed {
				dag.Status.State = workflowv1.StateFailed
			}
		}
	}
	return nil
}

// updateDAGState checks if all tasks are completed and updates DAG state
func (r *DagReconciler) updateDAGState(dag *workflowv1.Dag) {
	if dag.Status.State == workflowv1.StateFailed {
		return
	}

	allCompleted := true
	for _, ts := range dag.Status.TaskStatuses {
		if ts.State != workflowv1.StateCompleted {
			allCompleted = false
			break
		}
	}

	// Check if all tasks have status entries
	if allCompleted && len(dag.Status.TaskStatuses) == len(dag.Spec.Tasks) {
		dag.Status.State = workflowv1.StateCompleted
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DagReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize components before setting up
	r.SetupComponents()

	return ctrl.NewControllerManagedBy(mgr).
		For(&workflowv1.Dag{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
