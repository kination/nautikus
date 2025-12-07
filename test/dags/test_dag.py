import json
import sys

class Task:
    def __init__(self, name, task_type, command=None, script=None, dependencies=None):
        self.name = name
        self.type = task_type
        self.command = command
        self.script = script
        self.dependencies = dependencies or []

    def to_dict(self):
        d = {
            "name": self.name,
            "type": self.type,
            "dependencies": self.dependencies
        }
        if self.command: d["command"] = self.command
        if self.script: d["script"] = self.script
        return d

    # dependency with overloading
    def __rshift__(self, other):
        other.dependencies.append(self.name)
        return other

class DAG:
    def __init__(self, name, tasks):
        self.name = name
        self.tasks = tasks

    def generate_json(self):
        manifest = {
            "apiVersion": "workflow.my.domain/v1",
            "kind": "Dag",
            "metadata": {"name": self.name},
            "spec": {
                "tasks": [t.to_dict() for t in self.tasks]
            }
        }
        return json.dumps(manifest, indent=2)

if __name__ == "__main__":
    # test bash task
    t1 = Task("print-date", "Bash", command="date")

    # test python task
    py_code = """
import random
print(f"Random number: {random.randint(1, 100)}")
"""
    t2 = Task("random-py", "Python", script=py_code)

    # test go task
    go_code = """
package main
import "fmt"
func main() {
    fmt.Println("Hello from Go Operator!")
}
"""
    t3 = Task("hello-go", "Go", script=go_code)

    t1 >> t2 >> t3

    my_dag = DAG("example-dag", [t1, t2, t3])
    print(my_dag.generate_json())
    