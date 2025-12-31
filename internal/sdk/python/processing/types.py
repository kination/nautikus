"""Type definitions for the processing layer"""
from dataclasses import dataclass, field
from enum import Enum
from typing import Callable, List, Optional


class TaskType(Enum):
    """Defines the execution behavior of a task"""
    SIMPLE = "simple"   # Normal task
    BRANCH = "branch"   # Conditional branch selector
    JOIN = "join"       # Waits for any upstream branch


@dataclass
class TaskDef:
    """Internal representation of a task for processing"""
    name: str
    fn: Optional[Callable] = None
    branch_fn: Optional[Callable[[], str]] = None  # For branch tasks
    dependencies: List[str] = field(default_factory=list)
    task_type: TaskType = TaskType.SIMPLE
    branch_targets: List[str] = field(default_factory=list)  # For branch tasks
    branch_condition: Optional[str] = None  # For conditional tasks: which branch this belongs to
    condition_source: Optional[str] = None  # For conditional tasks: which branch task determines execution
