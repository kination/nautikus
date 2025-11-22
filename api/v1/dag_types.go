package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskType은 지원할 Operator 종류를 정의합니다.
type TaskType string

const (
	TaskTypeBash   TaskType = "Bash"
	TaskTypePython TaskType = "Python"
	TaskTypeGo     TaskType = "Go"
)

// TaskSpec은 DAG 내부의 개별 작업 단위입니다.
type TaskSpec struct {
	Name         string   `json:"name"`
	Type         TaskType `json:"type"`
	Dependencies []string `json:"dependencies,omitempty"` // 이 Task가 실행되기 위해 완료되어야 할 부모 Task들

	// Operator별 실행 내용
	Command string `json:"command,omitempty"` // Bash용
	Script  string `json:"script,omitempty"`  // Python/Go 코드 본문
	Image   string `json:"image,omitempty"`   // 커스텀 이미지 사용 시
}

// DagSpec은 사용자가 정의하는 DAG의 전체 명세입니다.
type DagSpec struct {
	Tasks []TaskSpec `json:"tasks"`
}

// TaskStatus는 개별 Task의 현재 상태입니다.
type TaskState string

const (
	StatePending   TaskState = "Pending"
	StateRunning   TaskState = "Running"
	StateCompleted TaskState = "Completed"
	StateFailed    TaskState = "Failed"
)

type TaskStatus struct {
	Name    string    `json:"name"`
	State   TaskState `json:"state"`
	PodName string    `json:"podName,omitempty"`
	Message string    `json:"message,omitempty"`
}

// DagStatus는 DAG 전체의 상태입니다.
type DagStatus struct {
	State        TaskState    `json:"state"` // DAG 전체 상태 (Running, Completed...)
	TaskStatuses []TaskStatus `json:"taskStatuses,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Dag is the Schema for the dags API
type Dag struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DagSpec   `json:"spec,omitempty"`
	Status DagStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DagList contains a list of Dag
type DagList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dag `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dag{}, &DagList{})
}
