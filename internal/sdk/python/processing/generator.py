"""DAG manifest generation logic"""
import json
import sys
import inspect
from typing import List

from .types import TaskDef, TaskType


def generate_manifest(dag_name: str, tasks: List[TaskDef]):
    """Create the DAG JSON manifest from task definitions"""
    script_content = _read_caller_source()

    task_specs = []

    for task in tasks:
        spec = {
            "name": task.name,
            "type": "Python",
            "script": script_content,
            "dependencies": task.dependencies,
            "env": {
                "NAUTIKUS_TASK_NAME": task.name
            }
        }

        # Add branch metadata for conditional tasks
        if task.branch_condition:
            spec["env"]["NAUTIKUS_BRANCH_CONDITION"] = task.branch_condition
            spec["env"]["NAUTIKUS_CONDITION_SOURCE"] = task.condition_source

        # Mark branch selector tasks
        if task.task_type == TaskType.BRANCH:
            spec["env"]["NAUTIKUS_TASK_TYPE"] = "branch"
            if task.branch_targets:
                spec["env"]["NAUTIKUS_BRANCH_TARGETS"] = ",".join(task.branch_targets)

        # Mark join tasks
        if task.task_type == TaskType.JOIN:
            spec["env"]["NAUTIKUS_TASK_TYPE"] = "join"

        task_specs.append(spec)

    manifest = {
        "apiVersion": "workflow.nautikus.io/v1",
        "kind": "Dag",
        "metadata": {
            "name": dag_name
        },
        "spec": {
            "tasks": task_specs
        }
    }

    print(json.dumps(manifest, indent=2))


def _read_caller_source() -> str:
    """Read the source file that called the SDK"""
    # Walk up the stack to find the user's dag file (not SDK files)
    for frame_info in inspect.stack():
        caller_file = frame_info.filename
        # Skip SDK files and built-in locations
        if not _is_sdk_file(caller_file) and not caller_file.startswith('<'):
            try:
                with open(caller_file, 'r') as f:
                    return f.read()
            except Exception:
                continue

    sys.stderr.write("Error: could not read source file\n")
    sys.exit(1)


def _is_sdk_file(path: str) -> bool:
    """Check if path is part of the SDK"""
    return '/sdk/python/' in path or '/sdk/go/' in path
