package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskType defines the type of task
type TaskType string

const (
	TaskTypeBash   TaskType = "Bash"
	TaskTypePython TaskType = "Python"
	TaskTypeGo     TaskType = "Go"
)

// TaskSpec defines the task spec
type TaskSpec struct {
	Name         string   `json:"name"`
	Type         TaskType `json:"type"`
	Dependencies []string `json:"dependencies,omitempty"` // Parent tasks that must be completed before this task can run

	// Execution details for each task type
	Command string            `json:"command,omitempty"` // For Bash tasks
	Script  string            `json:"script,omitempty"`  // For Python/Go code
	Image   string            `json:"image,omitempty"`   // Custom container image if needed
	Env     map[string]string `json:"env,omitempty"`     // Environment variables for the task
}

// DagSpec defines the complete specification of a DAG as defined by the user.
type DagSpec struct {
	Tasks []TaskSpec `json:"tasks"`
}

// TaskState represents the current state of an individual task.
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

// DagStatus represents the overall status of the entire DAG.
type DagStatus struct {
	State        TaskState    `json:"state"` // Overall DAG state (Running, Completed, etc.)
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
