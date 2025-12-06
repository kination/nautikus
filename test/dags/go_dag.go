package main

import (
	"encoding/json"
	"fmt"
	"os"

	workflowv1 "github.com/kination/pequod/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Define Tasks
	t1 := workflowv1.TaskSpec{
		Name:    "go-task-1",
		Type:    workflowv1.TaskTypeBash,
		Command: "echo 'Hello from Go DAG'",
	}

	t2 := workflowv1.TaskSpec{
		Name:         "go-task-2",
		Type:         workflowv1.TaskTypeGo,
		Dependencies: []string{"go-task-1"},
		Script: `package main
import "fmt"
func main() {
	fmt.Println("This is a Go task running inside the Pod")
}`,
	}

	// Define DAG
	dag := workflowv1.Dag{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "workflow.pequod.io/v1",
			Kind:       "Dag",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "go-generated-dag",
		},
		Spec: workflowv1.DagSpec{
			Tasks: []workflowv1.TaskSpec{t1, t2},
		},
	}

	// Output JSON (which is valid YAML)
	output, err := json.MarshalIndent(dag, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling DAG: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
