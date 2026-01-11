"""Task execution logic"""
import os
import sys
from typing import List

from .types import TaskDef, TaskType


def execute_task(target_task: str, tasks: List[TaskDef]):
    """Execute a specific task by name (called inside Pod)"""
    for task in tasks:
        if task.name == target_task:
            print(f"🚀 Starting task: {target_task}")

            # Handle branch condition check
            if task.branch_condition:
                selected_branch = os.environ.get("NAUTIKUS_SELECTED_BRANCH", "")
                if selected_branch and selected_branch != task.branch_condition:
                    print(f"⏭️  Skipping task {target_task} "
                          f"(branch {task.branch_condition} not selected, selected: {selected_branch})")
                    return

            # Handle branch selector task
            if task.task_type == TaskType.BRANCH and task.branch_fn:
                selected_branch = task.branch_fn()
                print(f"🔀 Branch selected: {selected_branch}")
                # Output branch selection for downstream tasks
                print(f"NAUTIKUS_BRANCH_RESULT={selected_branch}")
                return

            # Execute normal task
            if task.fn:
                try:
                    task.fn()
                except Exception as e:
                    print(f"❌ Task failed: {e}")
                    sys.exit(1)
            return

    print(f"❌ Unknown task: {target_task}", file=sys.stderr)
    sys.exit(1)
