package processing

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	workflowv1 "github.com/kination/nautikus/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateManifest creates the DAG JSON manifest from task definitions
func GenerateManifest(dagName string, tasks []TaskDef) {
	scriptContent := readCallerSource()

	taskSpecs := make([]workflowv1.TaskSpec, 0, len(tasks))

	for _, task := range tasks {
		spec := workflowv1.TaskSpec{
			Name:         task.Name,
			Type:         workflowv1.TaskTypeGo,
			Script:       scriptContent,
			Dependencies: task.Dependencies,
			Env: map[string]string{
				"NAUTIKUS_TASK_NAME": task.Name,
			},
		}

		// Add branch metadata for conditional tasks
		if task.BranchCondition != "" {
			spec.Env["NAUTIKUS_BRANCH_CONDITION"] = task.BranchCondition
			spec.Env["NAUTIKUS_CONDITION_SOURCE"] = task.ConditionSource
		}

		// Mark branch selector tasks
		if task.TaskType == TaskTypeBranch {
			spec.Env["NAUTIKUS_TASK_TYPE"] = "branch"
			// Store branch targets for controller to use
			if len(task.BranchTargets) > 0 {
				spec.Env["NAUTIKUS_BRANCH_TARGETS"] = joinStrings(task.BranchTargets, ",")
			}
		}

		// Mark join tasks
		if task.TaskType == TaskTypeJoin {
			spec.Env["NAUTIKUS_TASK_TYPE"] = "join"
		}

		taskSpecs = append(taskSpecs, spec)
	}

	dag := workflowv1.Dag{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "workflow.nautikus.io/v1",
			Kind:       "Dag",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: dagName,
		},
		Spec: workflowv1.DagSpec{
			Tasks: taskSpecs,
		},
	}

	output, err := json.MarshalIndent(dag, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling DAG: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func readCallerSource() string {
	// Walk up the stack to find the original caller (user's dag file)
	for i := 1; i < 10; i++ {
		_, callerFile, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Skip internal SDK files
		if !isSDKFile(callerFile) {
			content, err := os.ReadFile(callerFile)
			if err == nil {
				return string(content)
			}
		}
	}

	// Fallback: try common paths
	wd, _ := os.Getwd()
	paths := []string{
		wd + "/test/dags/go_dag.go",
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content)
		}
	}

	fmt.Fprintf(os.Stderr, "Error: could not read source file\n")
	os.Exit(1)
	return ""
}

func isSDKFile(path string) bool {
	return len(path) > 0 && (contains(path, "/sdk/go/") || contains(path, "/sdk/python/"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
