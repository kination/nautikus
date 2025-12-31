package processing

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
)

// ExecuteTask runs a specific task by name (called inside Pod)
func ExecuteTask(targetTask string, tasks []TaskDef) {
	for _, task := range tasks {
		if task.Name == targetTask {
			fmt.Printf("🚀 Starting task: %s\n", targetTask)

			// Handle branch condition check
			if task.BranchCondition != "" {
				selectedBranch := os.Getenv("NAUTIKUS_SELECTED_BRANCH")
				if selectedBranch != "" && selectedBranch != task.BranchCondition {
					fmt.Printf("⏭️  Skipping task %s (branch %s not selected, selected: %s)\n",
						targetTask, task.BranchCondition, selectedBranch)
					return
				}
			}

			// Handle branch selector task
			if task.TaskType == TaskTypeBranch && task.BranchFn != nil {
				selectedBranch := task.BranchFn()
				fmt.Printf("🔀 Branch selected: %s\n", selectedBranch)
				// Output branch selection for downstream tasks
				fmt.Printf("NAUTIKUS_BRANCH_RESULT=%s\n", selectedBranch)
				return
			}

			// Execute normal task
			if task.Fn != nil {
				task.Fn()
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "❌ Unknown task: %s\n", targetTask)
	os.Exit(1)
}

// GetFuncName extracts function name from func pointer
func GetFuncName(fn interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}
