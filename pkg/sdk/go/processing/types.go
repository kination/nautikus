package processing

// TaskType defines the execution behavior of a task
type TaskType int

const (
	TaskTypeSimple TaskType = iota // Normal task
	TaskTypeBranch                 // Conditional branch selector
	TaskTypeJoin                   // Waits for any upstream branch
)

// TaskDef is the internal representation of a task for processing
type TaskDef struct {
	Name            string
	Fn              func()
	BranchFn        func() string // For branch tasks: returns selected branch name
	Dependencies    []string
	TaskType        TaskType
	BranchTargets   []string // For branch tasks: possible branch names
	BranchCondition string   // For conditional tasks: which branch this belongs to
	ConditionSource string   // For conditional tasks: which branch task determines execution
}
