# Nautikus

A Kubernetes-native workflow engine built with the Operator pattern, similar to Apache Airflow. Define DAGs (Directed Acyclic Graphs) as code and let Kubernetes orchestrate your workflows.

## Features

- 📝 **Define DAGs in Code**: Write workflows in Go or Python
- 🔄 **Automatic Compilation**: Convert code to Kubernetes manifests (YAML)
- 🎯 **Dependency Management**: Define task dependencies with ease
- 🚀 **Native Kubernetes**: Runs as a Kubernetes controller
- 🔧 **Multiple Task Types**: Support for Bash, Python, and Go tasks

## Getting Started

### Prerequisites
- Go (version v1.24.6+)
- Docker (version 17.03+)
- kubectl (version v1.11.3+)
- kind (for local testing)

## Quick Start: Local Development

### 1. Create a local cluster
```sh
kind create cluster --name nautikus
```

### 2. Build and load image
```sh
make docker-build IMG=nautikus:v1
kind load docker-image nautikus:v1 --name nautikus
```

### 3. Install CRDs and deploy controller
```sh
make install
make deploy IMG=nautikus:v1
```

### 4. Verify deployment
```sh
kubectl get pods -n nautikus-system
```

### 5. Compile and run a DAG

**Option A: Use pre-defined YAML**
```sh
kubectl apply -f config/samples/workflow_v1_dag_test.yaml
```

**Option B: Define DAG in code and compile**
```sh
# Using go run
go run cmd/dag-cli/main.go compile

# Or build the CLI first
make build-cli
./bin/dag-cli compile

# Or use make target
make compile-dags

# Apply the generated YAML
kubectl apply -f dist/go_dag.yaml
```

### CLI Options
```sh
# View all available commands
./bin/dag-cli --help

# View compile command options
./bin/dag-cli compile --help

# Compile with custom config and output directory
./bin/dag-cli compile --config my-config.yaml --out output/

# Short flags
./bin/dag-cli compile -c my-config.yaml -o output/

# Check version
./bin/dag-cli version
```

### 6. Monitor DAG execution
```sh
# Check DAG status
kubectl get dags

# View detailed status
kubectl describe dag go-generated-dag

# Check task pods
kubectl get pods

# View controller logs
kubectl logs -n nautikus-system -l control-plane=controller-manager -f
```

### 7. Cleanup
```sh
kubectl delete dags --all
make undeploy
make uninstall
kind delete cluster --name nautikus
```

## Writing DAGs in Code

### Go Example (`test/dags/my_workflow.go`)

```go
package main

import sdk "github.com/kination/nautikus/pkg/sdk/go"

func task1() { println("Hello from Task 1") }
func task2() { println("Hello from Task 2") }

func main() {
    sdk.NewDAG("my-workflow").
        AddSequential(
            sdk.Task{Name: "task-1", Fn: task1},
            sdk.Task{Name: "task-2", Fn: task2},
        ).
        Serve()
}
```

Then compile and apply:
```sh
# Using the CLI binary
make build-cli
./bin/dag-cli compile

# Or using go run
go run cmd/dag-cli/main.go compile

# Or using make target
make compile-dags

# Apply the generated manifest
kubectl apply -f dist/my_workflow.yaml
```

## Production Deployment

### 1. Build & push your image
```sh
make docker-build docker-push IMG=<your-registry>/nautikus:tag
```

### 2. Install CRDs
```sh
make install
```

### 3. Deploy controller
```sh
make deploy IMG=<your-registry>/nautikus:tag
```

### 4. Create DAG instances
```sh
kubectl apply -k config/samples/
```

## Uninstall

### 1. Delete DAG instances
```sh
kubectl delete -k config/samples/
```

### 2. Remove CRDs
```sh
make uninstall
```

### 3. Undeploy controller
```sh
make undeploy
```

## Development

### Run tests
```sh
make test
```

### Run E2E tests
```sh
make test-e2e
```

### Run controller locally (without Docker)
```sh
make run
```

### Generate code and manifests
```sh
make manifests  # Generate CRDs
make generate   # Generate DeepCopy methods
```

## Project Structure

```
nautikus/
├── api/v1/              # CRD definitions
├── cmd/
│   ├── manager/         # Controller entrypoint
│   └── dag-cli/         # DAG compiler CLI
├── internal/
│   ├── controller/      # Orchestration logic (DagReconciler)
│   ├── scheduler/       # Task dependency & scheduling logic
│   ├── runner/          # Task execution & status monitoring
│   ├── executor/        # Execution engines (Pod, etc.)
│   └── compiler/        # Code-to-YAML compiler
├── pkg/
│   └── sdk/             # Go/Python SDK for users
├── config/              # Kubernetes manifests
│   ├── crd/            # CRD definitions
│   ├── rbac/           # RBAC rules
│   └── samples/        # Example DAGs
├── test/dags/          # Example DAG definitions
└── dist/               # Compiled YAML output (gitignored)
```

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0.
