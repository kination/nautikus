"""
Nautikus Python SDK - Minimal API layer
"""
import os
from typing import Callable, List, Dict, Optional
from dataclasses import dataclass, field

from .processing.types import TaskDef, TaskType
from .processing.executor import execute_task
from .processing.generator import generate_manifest


@dataclass
class Task:
    """Represents a unit of work in a DAG"""
    name: str
    fn: Callable
    dependencies: List[str] = field(default_factory=list)


class DAGBuilder:
    """Fluent API for building DAGs"""

    def __init__(self, name: str):
        self.name = name
        self.tasks: List[TaskDef] = []

    def add_task(self, name: str, fn: Callable, deps: List[str] = None) -> 'DAGBuilder':
        """Add a simple task with optional dependencies"""
        self.tasks.append(TaskDef(
            name=name,
            fn=fn,
            dependencies=deps or [],
            task_type=TaskType.SIMPLE
        ))
        return self

    def add_sequential(self, *tasks: Task) -> 'DAGBuilder':
        """Add tasks that run sequentially (each depends on previous)"""
        prev_name = None
        for t in tasks:
            deps = list(t.dependencies)
            if prev_name:
                deps.insert(0, prev_name)
            self.tasks.append(TaskDef(
                name=t.name,
                fn=t.fn,
                dependencies=deps,
                task_type=TaskType.SIMPLE
            ))
            prev_name = t.name
        return self

    def add_parallel(self, after_task: str, *tasks: Task) -> 'DAGBuilder':
        """Add tasks that run in parallel (same dependencies)"""
        for t in tasks:
            deps = list(t.dependencies)
            if after_task:
                deps.insert(0, after_task)
            self.tasks.append(TaskDef(
                name=t.name,
                fn=t.fn,
                dependencies=deps,
                task_type=TaskType.SIMPLE
            ))
        return self

    def add_branch(self, condition_task_name: str, condition_fn: Callable[[], str],
                   branches: Dict[str, List[Task]]) -> 'DAGBuilder':
        """
        Add conditional branching (like Airflow's BranchPythonOperator)

        Args:
            condition_task_name: Name of the condition evaluation task
            condition_fn: Function that returns the branch name to execute
            branches: Dict mapping branch names to list of tasks
        """
        # Add the condition task
        self.tasks.append(TaskDef(
            name=condition_task_name,
            branch_fn=condition_fn,
            task_type=TaskType.BRANCH,
            branch_targets=list(branches.keys())
        ))

        # Add all branch tasks with skip conditions
        for branch_name, branch_tasks in branches.items():
            prev_name = None
            for i, t in enumerate(branch_tasks):
                deps = list(t.dependencies)
                if i == 0:
                    deps.insert(0, condition_task_name)
                elif prev_name:
                    deps.insert(0, prev_name)

                self.tasks.append(TaskDef(
                    name=t.name,
                    fn=t.fn,
                    dependencies=deps,
                    task_type=TaskType.SIMPLE,
                    branch_condition=branch_name,
                    condition_source=condition_task_name
                ))
                prev_name = t.name
        return self

    def add_join(self, name: str, fn: Callable, wait_for: List[str]) -> 'DAGBuilder':
        """Add a join task that waits for any of the specified tasks"""
        self.tasks.append(TaskDef(
            name=name,
            fn=fn,
            dependencies=wait_for,
            task_type=TaskType.JOIN
        ))
        return self

    def serve(self):
        """Execute the DAG (either generates manifest or runs task based on env)"""
        target_task = os.environ.get("NAUTIKUS_TASK_NAME")
        if target_task:
            execute_task(target_task, self.tasks)
            return
        generate_manifest(self.name, self.tasks)


def serve(dag_name: str, tasks: List[Callable]):
    """Legacy API for backward compatibility"""
    builder = DAGBuilder(dag_name)
    for i, fn in enumerate(tasks):
        if i == 0:
            builder.add_task(fn.__name__, fn)
        else:
            prev_name = tasks[i-1].__name__
            builder.add_task(fn.__name__, fn, [prev_name])
    builder.serve()
