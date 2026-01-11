// Package connector provides interfaces for external resource connectors.
// This is the core abstraction for Phase 6-2 (AI Training Connector).
package connector

import (
	"context"

	workflowv1 "github.com/kination/nautikus/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ExternalResourceConnector defines the interface for K8s CR-based connectors.
// Used for Kubeflow, Katib, KServe, Ray, etc.
type ExternalResourceConnector interface {
	// Type returns the task type this connector handles (e.g., "kubeflow/pytorchjob")
	Type() string

	// BuildResource creates an unstructured CR from task spec
	BuildResource(task *workflowv1.TaskSpec, dag *workflowv1.Dag) (*unstructured.Unstructured, error)

	// GetStatus checks the external resource status and maps to TaskState
	GetStatus(ctx context.Context, cl client.Client, task *workflowv1.TaskSpec, dag *workflowv1.Dag) (workflowv1.TaskState, error)

	// Cleanup removes the external resource (optional, for OwnerReference fallback)
	Cleanup(ctx context.Context, cl client.Client, task *workflowv1.TaskSpec, dag *workflowv1.Dag) error
}

// CloudServiceConnector defines the interface for external cloud API connectors.
// Used for SageMaker, Vertex AI, Azure ML, OpenAI, etc.
type CloudServiceConnector interface {
	// Type returns the task type this connector handles (e.g., "aws/sagemaker")
	Type() string

	// Submit submits a job to the cloud service, returns job ID
	Submit(ctx context.Context, task *workflowv1.TaskSpec, dag *workflowv1.Dag) (jobID string, err error)

	// GetStatus checks the job status and maps to TaskState
	GetStatus(ctx context.Context, jobID string) (workflowv1.TaskState, error)

	// Cancel cancels a running job
	Cancel(ctx context.Context, jobID string) error
}

// ConnectorConfig holds common configuration for connectors
type ConnectorConfig struct {
	Client    client.Client
	Namespace string
}
