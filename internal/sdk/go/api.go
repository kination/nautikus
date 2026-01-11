package sdk

import (
	"os"

	"github.com/kination/nautikus/internal/sdk/go/processing"
)

// Task represents a unit of work in a DAG
type Task struct {
	Name         string
	Fn           func()
	Dependencies []string
}

// BranchCondition defines a conditional branch
type BranchCondition struct {
	Condition func() bool // Returns true to execute TrueBranch
	TrueBranch  []Task
	FalseBranch []Task
}

// DAGBuilder provides fluent API for building DAGs
type DAGBuilder struct {
	name  string
	tasks []processing.TaskDef
}

// NewDAG creates a new DAG builder
func NewDAG(name string) *DAGBuilder {
	return &DAGBuilder{name: name}
}

// AddTask adds a simple task with optional dependencies
func (b *DAGBuilder) AddTask(name string, fn func(), deps ...string) *DAGBuilder {
	b.tasks = append(b.tasks, processing.TaskDef{
		Name:         name,
		Fn:           fn,
		Dependencies: deps,
		TaskType:     processing.TaskTypeSimple,
	})
	return b
}

// AddSequential adds tasks that run sequentially (each depends on previous)
func (b *DAGBuilder) AddSequential(tasks ...Task) *DAGBuilder {
	var prevName string
	for _, t := range tasks {
		deps := t.Dependencies
		if prevName != "" {
			deps = append([]string{prevName}, deps...)
		}
		b.tasks = append(b.tasks, processing.TaskDef{
			Name:         t.Name,
			Fn:           t.Fn,
			Dependencies: deps,
			TaskType:     processing.TaskTypeSimple,
		})
		prevName = t.Name
	}
	return b
}

// AddParallel adds tasks that run in parallel (same dependencies)
func (b *DAGBuilder) AddParallel(afterTask string, tasks ...Task) *DAGBuilder {
	for _, t := range tasks {
		deps := t.Dependencies
		if afterTask != "" {
			deps = append([]string{afterTask}, deps...)
		}
		b.tasks = append(b.tasks, processing.TaskDef{
			Name:         t.Name,
			Fn:           t.Fn,
			Dependencies: deps,
			TaskType:     processing.TaskTypeSimple,
		})
	}
	return b
}

// AddBranch adds conditional branching (like Airflow's BranchPythonOperator)
// The condition task determines which branch to execute
func (b *DAGBuilder) AddBranch(conditionTaskName string, conditionFn func() string, branches map[string][]Task) *DAGBuilder {
	// Add the condition task that returns which branch to take
	b.tasks = append(b.tasks, processing.TaskDef{
		Name:           conditionTaskName,
		BranchFn:       conditionFn,
		TaskType:       processing.TaskTypeBranch,
		BranchTargets:  getBranchNames(branches),
	})

	// Add all branch tasks with skip conditions
	for branchName, branchTasks := range branches {
		var prevName string
		for i, t := range branchTasks {
			deps := t.Dependencies
			if i == 0 {
				// First task in branch depends on condition task
				deps = append([]string{conditionTaskName}, deps...)
			} else if prevName != "" {
				deps = append([]string{prevName}, deps...)
			}
			b.tasks = append(b.tasks, processing.TaskDef{
				Name:            t.Name,
				Fn:              t.Fn,
				Dependencies:    deps,
				TaskType:        processing.TaskTypeSimple,
				BranchCondition: branchName,
				ConditionSource: conditionTaskName,
			})
			prevName = t.Name
		}
	}
	return b
}

// AddJoin adds a join task that waits for any of the specified tasks
func (b *DAGBuilder) AddJoin(name string, fn func(), waitFor ...string) *DAGBuilder {
	b.tasks = append(b.tasks, processing.TaskDef{
		Name:         name,
		Fn:           fn,
		Dependencies: waitFor,
		TaskType:     processing.TaskTypeJoin,
	})
	return b
}

// Serve executes the DAG (either generates manifest or runs task based on env)
func (b *DAGBuilder) Serve() {
	if targetTask := os.Getenv("NAUTIKUS_TASK_NAME"); targetTask != "" {
		processing.ExecuteTask(targetTask, b.tasks)
		return
	}
	processing.GenerateManifest(b.name, b.tasks)
}

// Legacy API for backward compatibility
func Serve(dagName string, tasks []func()) {
	builder := NewDAG(dagName)
	for i, fn := range tasks {
		task := Task{Name: processing.GetFuncName(fn), Fn: fn}
		if i == 0 {
			builder.AddTask(task.Name, task.Fn)
		} else {
			prevName := processing.GetFuncName(tasks[i-1])
			builder.AddTask(task.Name, task.Fn, prevName)
		}
	}
	builder.Serve()
}

func getBranchNames(branches map[string][]Task) []string {
	names := make([]string, 0, len(branches))
	for name := range branches {
		names = append(names, name)
	}
	return names
}
