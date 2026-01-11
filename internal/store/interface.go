// Package store provides storage interfaces for DAG and task persistence.
// This is the core abstraction for Phase 5 (Database Strategy).
package store

import (
	"context"
	"time"

	workflowv1 "github.com/kination/nautikus/api/v1"
)

// Store defines the interface for DAG and task persistence.
// Implementations can use etcd (default), PostgreSQL, or other backends.
type Store interface {
	// DAG operations
	SaveDAG(ctx context.Context, dag *workflowv1.Dag) error
	GetDAG(ctx context.Context, namespace, name string) (*workflowv1.Dag, error)
	ListDAGs(ctx context.Context, namespace string, opts ListOptions) ([]*workflowv1.Dag, error)
	DeleteDAG(ctx context.Context, namespace, name string) error

	// Task status operations
	SaveTaskStatus(ctx context.Context, dagNamespace, dagName, taskName string, status *workflowv1.TaskStatus) error
	GetTaskStatus(ctx context.Context, dagNamespace, dagName, taskName string) (*workflowv1.TaskStatus, error)
	ListTaskStatuses(ctx context.Context, dagNamespace, dagName string) ([]workflowv1.TaskStatus, error)

	// History operations (for completed DAG runs)
	SaveDAGRun(ctx context.Context, run *DAGRun) error
	GetDAGRun(ctx context.Context, runID string) (*DAGRun, error)
	ListDAGRuns(ctx context.Context, dagNamespace, dagName string, opts ListOptions) ([]*DAGRun, error)

	// Health check
	Ping(ctx context.Context) error

	// Close releases resources
	Close() error
}

// ListOptions defines options for listing operations
type ListOptions struct {
	// Limit is the maximum number of items to return
	Limit int
	// Offset is the number of items to skip
	Offset int
	// LabelSelector filters by labels
	LabelSelector string
	// State filters by DAG state
	State workflowv1.TaskState
}

// DAGRun represents a completed or historical DAG execution
type DAGRun struct {
	// RunID is a unique identifier for this run
	RunID string
	// DAGNamespace is the namespace of the DAG
	DAGNamespace string
	// DAGName is the name of the DAG
	DAGName string
	// StartTime is when the DAG run started
	StartTime time.Time
	// EndTime is when the DAG run completed
	EndTime *time.Time
	// State is the final state of the DAG run
	State workflowv1.TaskState
	// TaskRuns contains the status of each task in this run
	TaskRuns []TaskRun
	// Metadata contains additional run metadata
	Metadata map[string]string
}

// TaskRun represents a task execution within a DAG run
type TaskRun struct {
	// TaskName is the name of the task
	TaskName string
	// State is the final state of the task
	State workflowv1.TaskState
	// StartTime is when the task started
	StartTime time.Time
	// EndTime is when the task completed
	EndTime *time.Time
	// PodName is the name of the Pod that ran this task
	PodName string
	// Message contains any status message or error
	Message string
}

// StoreConfig holds configuration for creating a store
type StoreConfig struct {
	// Type is the store backend type (etcd, postgres, memory)
	Type StoreType
	// ConnectionString is the connection string for the backend
	ConnectionString string
	// MaxConnections is the maximum number of connections
	MaxConnections int
	// Timeout is the default operation timeout
	Timeout time.Duration
}

// StoreType defines the type of store backend
type StoreType string

const (
	// StoreTypeEtcd uses etcd as the backend (default, via Kubernetes)
	StoreTypeEtcd StoreType = "etcd"
	// StoreTypePostgres uses PostgreSQL as the backend
	StoreTypePostgres StoreType = "postgres"
	// StoreTypeMemory uses in-memory storage (for testing)
	StoreTypeMemory StoreType = "memory"
)

// DefaultStoreConfig returns the default store configuration
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		Type:           StoreTypeEtcd,
		MaxConnections: 10,
		Timeout:        30 * time.Second,
	}
}

// EventStore defines the interface for event persistence and streaming.
// Used for audit logging and event-driven architectures.
type EventStore interface {
	// Publish publishes an event
	Publish(ctx context.Context, event *Event) error

	// Subscribe subscribes to events matching the filter
	Subscribe(ctx context.Context, filter EventFilter) (<-chan *Event, error)

	// GetEvents retrieves historical events
	GetEvents(ctx context.Context, filter EventFilter, opts ListOptions) ([]*Event, error)
}

// Event represents a workflow event
type Event struct {
	// ID is the unique event identifier
	ID string
	// Type is the event type
	Type EventType
	// Timestamp is when the event occurred
	Timestamp time.Time
	// DAGNamespace is the namespace of the related DAG
	DAGNamespace string
	// DAGName is the name of the related DAG
	DAGName string
	// TaskName is the name of the related task (if applicable)
	TaskName string
	// Data contains event-specific data
	Data map[string]interface{}
}

// EventType defines the type of event
type EventType string

const (
	// EventTypeDAGCreated is emitted when a DAG is created
	EventTypeDAGCreated EventType = "dag.created"
	// EventTypeDAGStarted is emitted when a DAG starts running
	EventTypeDAGStarted EventType = "dag.started"
	// EventTypeDAGCompleted is emitted when a DAG completes successfully
	EventTypeDAGCompleted EventType = "dag.completed"
	// EventTypeDAGFailed is emitted when a DAG fails
	EventTypeDAGFailed EventType = "dag.failed"
	// EventTypeTaskStarted is emitted when a task starts
	EventTypeTaskStarted EventType = "task.started"
	// EventTypeTaskCompleted is emitted when a task completes
	EventTypeTaskCompleted EventType = "task.completed"
	// EventTypeTaskFailed is emitted when a task fails
	EventTypeTaskFailed EventType = "task.failed"
)

// EventFilter defines criteria for filtering events
type EventFilter struct {
	// Types filters by event types
	Types []EventType
	// DAGNamespace filters by DAG namespace
	DAGNamespace string
	// DAGName filters by DAG name
	DAGName string
	// Since filters events after this time
	Since *time.Time
	// Until filters events before this time
	Until *time.Time
}
