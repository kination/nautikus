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

	workflowv1 "github.com/kination/gostration/api/v1"
)

// DagReconciler reconciles a Dag object
type DagReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DagReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. DAG CR 가져오기
	var dag workflowv1.Dag
	if err := r.Get(ctx, req.NamespacedName, &dag); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. 현재 실행 중인 Pod들과 Status 동기화
	if err := r.syncStatus(ctx, &dag); err != nil {
		return ctrl.Result{}, err
	}

	// 3. 실행할 다음 Task 찾기 (의존성 체크)
	nextTasks := r.getNextTasks(&dag)

	// 4. Pod 생성
	for _, task := range nextTasks {
		pod := r.buildPod(&dag, task)
		// Pod 소유권 설정 (DAG 삭제 시 Pod도 삭제되도록)
		if err := controllerutil.SetControllerReference(&dag, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating a new Pod", "Pod.Name", pod.Name, "Task.Name", task.Name)
		if err := r.Create(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 5. Status 업데이트
	if err := r.Status().Update(ctx, &dag); err != nil {
		return ctrl.Result{}, err
	}

	// DAG가 아직 안 끝났다면 계속 Reconcile
	if dag.Status.State == workflowv1.StateRunning {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// 현재 Pod 상태를 보고 DAG Status 업데이트
func (r *DagReconciler) syncStatus(ctx context.Context, dag *workflowv1.Dag) error {
	// 맵 초기화
	if dag.Status.TaskStatuses == nil {
		dag.Status.TaskStatuses = []workflowv1.TaskStatus{}
	}

	// 전체 상태가 없으면 Running으로 시작
	if dag.Status.State == "" {
		dag.Status.State = workflowv1.StateRunning
	}

	// 실제 Pod 상태 조회
	statusMap := make(map[string]*workflowv1.TaskStatus)
	for i := range dag.Status.TaskStatuses {
		t := &dag.Status.TaskStatuses[i]
		statusMap[t.Name] = t
	}

	for _, taskSpec := range dag.Spec.Tasks {
		// 이미 완료/실패된 건 패스
		currentStatus, exists := statusMap[taskSpec.Name]
		if !exists {
			continue
		}
		if currentStatus.State == workflowv1.StateCompleted || currentStatus.State == workflowv1.StateFailed {
			continue
		}

		// Pod 조회
		pod := &corev1.Pod{}
		err := r.Get(ctx, types.NamespacedName{Name: currentStatus.PodName, Namespace: dag.Namespace}, pod)
		if err != nil {
			if errors.IsNotFound(err) {
				// Pod가 없으면 아직 생성 전이거나 삭제됨
				continue
			}
			return err
		}

		// Pod 상태 반영
		if pod.Status.Phase == corev1.PodSucceeded {
			currentStatus.State = workflowv1.StateCompleted
		} else if pod.Status.Phase == corev1.PodFailed {
			currentStatus.State = workflowv1.StateFailed
			dag.Status.State = workflowv1.StateFailed // 하나라도 실패하면 DAG 실패
		} else {
			currentStatus.State = workflowv1.StateRunning
		}
	}
	return nil
}

// 실행 가능한(의존성이 해결된) Task 찾기
func (r *DagReconciler) getNextTasks(dag *workflowv1.Dag) []workflowv1.TaskSpec {
	var nextTasks []workflowv1.TaskSpec

	// 상태 룩업 맵
	statusMap := make(map[string]workflowv1.TaskState)
	for _, s := range dag.Status.TaskStatuses {
		statusMap[s.Name] = s.State
	}

	for _, task := range dag.Spec.Tasks {
		// 이미 실행 중이거나 완료된 Task는 제외
		if state, ok := statusMap[task.Name]; ok && state != "" {
			continue
		}

		// 의존성 체크
		allDepsCompleted := true
		for _, dep := range task.Dependencies {
			if statusMap[dep] != workflowv1.StateCompleted {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			nextTasks = append(nextTasks, task)
			// 중복 실행 방지를 위해 상태를 미리 Pending으로 추가 (실제 업데이트는 Reconcile 끝에서)
			dag.Status.TaskStatuses = append(dag.Status.TaskStatuses, workflowv1.TaskStatus{
				Name:    task.Name,
				State:   workflowv1.StatePending,
				PodName: fmt.Sprintf("%s-%s", dag.Name, task.Name), // Pod 이름 규칙
			})
		}
	}
	return nextTasks
}

// TaskSpec을 Pod로 변환 (Bash, Python, Go Operator 로직)
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
		// Python 코드를 인라인으로 실행
		command = []string{"python", "-c"}
		args = []string{task.Script}
	case workflowv1.TaskTypeGo:
		image = "golang:1.20-alpine"
		// Go 코드는 파일로 저장 후 실행하거나, 여기서는 간단히 go run을 위해 sh 감싸기
		// 실제로는 ConfigMap으로 코드를 마운트하는 것이 좋습니다.
		command = []string{"/bin/sh", "-c"}
		// 매우 간단한 인라인 실행 예시 (복잡한 코드는 ConfigMap 권장)
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
		Owns(&corev1.Pod{}). // Pod 상태 변화를 감지
		Complete(r)
}
