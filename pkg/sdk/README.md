# Nautikus SDK

SDK for defining DAG workflows in Go and Python with support for sequential, parallel, and conditional task execution.

## Directory Structure

```
sdk/
├── go/                    # Go SDK
│   ├── api.go             # Public API (DAGBuilder, Task, Serve)
│   └── processing/        # Internal processing logic
│       ├── types.go       # TaskDef, TaskType
│       ├── executor.go    # Task execution (runs inside Pod)
│       └── generator.go   # Manifest generation
├── python/                # Python SDK
│   ├── __init__.py
│   ├── api.py             # Public API (DAGBuilder, Task, serve)
│   └── processing/        # Internal processing logic
│       ├── types.py       # TaskDef, TaskType
│       ├── executor.py    # Task execution
│       └── generator.py   # Manifest generation
└── README.md
```

## Quick Start

### Go

```go
package main

import sdk "github.com/kination/nautikus/internal/sdk/go"

func task1() { /* ... */ }
func task2() { /* ... */ }

func main() {
    sdk.NewDAG("my-dag").
        AddSequential(
            sdk.Task{Name: "task1", Fn: task1},
            sdk.Task{Name: "task2", Fn: task2},
        ).
        Serve()
}
```

### Python

```python
from internal.sdk.python import DAGBuilder, Task

def task1(): pass
def task2(): pass

DAGBuilder("my-dag").add_sequential(
    Task(name="task1", fn=task1),
    Task(name="task2", fn=task2),
).serve()
```

## API Reference

### DAGBuilder

| Method | Description |
|--------|-------------|
| `NewDAG(name)` / `DAGBuilder(name)` | Create new DAG builder |
| `AddTask(name, fn, deps...)` | Add single task with optional dependencies |
| `AddSequential(tasks...)` | Add tasks in sequence (auto-chained) |
| `AddParallel(afterTask, tasks...)` | Add tasks running in parallel |
| `AddBranch(name, conditionFn, branches)` | Add conditional branching |
| `AddJoin(name, fn, waitFor...)` | Add join point for branches |
| `Serve()` | Generate manifest or execute task |

### Task Types

#### Sequential Execution
Tasks run one after another. Each task depends on the previous.

```go
// Go
AddSequential(
    sdk.Task{Name: "extract", Fn: extract},
    sdk.Task{Name: "transform", Fn: transform},
    sdk.Task{Name: "load", Fn: load},
)
```

```python
# Python
.add_sequential(
    Task(name="extract", fn=extract),
    Task(name="transform", fn=transform),
    Task(name="load", fn=load),
)
```

#### Parallel Execution
Tasks run concurrently after a common predecessor.

```go
// Go
AddParallel("extract",
    sdk.Task{Name: "validate_a", Fn: validateA},
    sdk.Task{Name: "validate_b", Fn: validateB},
)
```

```python
# Python
.add_parallel("extract",
    Task(name="validate_a", fn=validate_a),
    Task(name="validate_b", fn=validate_b),
)
```

#### Conditional Branching (like Airflow BranchPythonOperator)
Execute different task paths based on runtime condition.

```go
// Go - condition function returns branch name
func checkQuality() string {
    if score >= 70 {
        return "high_quality"
    }
    return "low_quality"
}

AddBranch("check", checkQuality, map[string][]sdk.Task{
    "high_quality": {{Name: "fast_process", Fn: fastProcess}},
    "low_quality":  {{Name: "clean", Fn: clean}, {Name: "retry", Fn: retry}},
})
```

```python
# Python
def check_quality() -> str:
    return "high_quality" if score >= 70 else "low_quality"

.add_branch("check", check_quality, {
    "high_quality": [Task(name="fast_process", fn=fast_process)],
    "low_quality": [Task(name="clean", fn=clean), Task(name="retry", fn=retry)],
})
```

#### Join Point
Wait for any of the specified upstream tasks (useful after branches).

```go
// Go
AddJoin("finalize", finalize, "fast_process", "retry")
```

```python
# Python
.add_join("finalize", finalize, ["fast_process", "retry"])
```

## Complete Example

ETL pipeline with quality-based branching:

```
extract → validate → check_quality ─┬─ high_quality → process_high ──────────┬─→ load
                                    └─ low_quality → process_low → clean ────┘
```

### Go

```go
package main

import (
    "fmt"
    sdk "github.com/kination/nautikus/internal/sdk/go"
)

func extract()     { fmt.Println("Extracting...") }
func validate()    { fmt.Println("Validating...") }
func checkQuality() string {
    if qualityScore >= 70 { return "high_quality" }
    return "low_quality"
}
func processHigh() { fmt.Println("High quality processing") }
func processLow()  { fmt.Println("Low quality processing") }
func clean()       { fmt.Println("Cleaning data") }
func load()        { fmt.Println("Loading to destination") }

func main() {
    sdk.NewDAG("etl-pipeline").
        AddSequential(
            sdk.Task{Name: "extract", Fn: extract},
            sdk.Task{Name: "validate", Fn: validate},
        ).
        AddBranch("check_quality", checkQuality, map[string][]sdk.Task{
            "high_quality": {{Name: "process_high", Fn: processHigh}},
            "low_quality":  {{Name: "process_low", Fn: processLow}, {Name: "clean", Fn: clean}},
        }).
        AddJoin("load", load, "process_high", "clean").
        Serve()
}
```

### Python

```python
from internal.sdk.python import DAGBuilder, Task

def extract(): print("Extracting...")
def validate(): print("Validating...")
def check_quality() -> str:
    return "high_quality" if quality_score >= 70 else "low_quality"
def process_high(): print("High quality processing")
def process_low(): print("Low quality processing")
def clean(): print("Cleaning data")
def load(): print("Loading to destination")

(DAGBuilder("etl-pipeline")
    .add_sequential(
        Task(name="extract", fn=extract),
        Task(name="validate", fn=validate),
    )
    .add_branch("check_quality", check_quality, {
        "high_quality": [Task(name="process_high", fn=process_high)],
        "low_quality": [Task(name="process_low", fn=process_low), Task(name="clean", fn=clean)],
    })
    .add_join("load", load, ["process_high", "clean"])
    .serve()
)
```

## Legacy API

For backward compatibility with existing code:

```go
// Go
sdk.Serve("dag-name", []func(){task1, task2})
```

```python
# Python
from internal.sdk.python import serve
serve("dag-name", [task1, task2])
```

## Architecture

The SDK separates concerns into two layers:

1. **API Layer** (`api.go` / `api.py`): Fluent builder interface for users
2. **Processing Layer** (`processing/`): Internal logic for manifest generation and task execution

This separation allows:
- Easy extension of processing logic for large-scale DAGs
- Independent scaling of parsing/scheduling components
- Clean API surface for users

## Environment Variables

Used internally during Pod execution:

| Variable | Description |
|----------|-------------|
| `NAUTIKUS_TASK_NAME` | Current task to execute |
| `NAUTIKUS_TASK_TYPE` | Task type (branch/join) |
| `NAUTIKUS_BRANCH_RESULT` | Selected branch from condition |
| `NAUTIKUS_SELECTED_BRANCH` | Active branch for conditional tasks |
