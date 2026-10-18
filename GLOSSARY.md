# Task Tracker

A command-line tool for keeping a personal to-do list in a file in the current directory.

## Language

**Task**:
A single piece of work the user wants to track, with a description and a status.
_Avoid_: Todo, item, entry, note

**Task ID**:
A positive whole number that identifies one task for its whole life; once used, it is never given to another task, even after deletion.
_Avoid_: Index, number, position

**Description**:
The short free-text summary of what a task is about.
_Avoid_: Title, name, text

**Status**:
Where a task stands: exactly one of `todo`, `in-progress` or `done`. A new task starts as `todo`.
_Avoid_: State, stage, progress

**Mark**:
To change a task's status (to `in-progress` or `done`).
_Avoid_: Complete, finish, set, toggle

**Task file**:
The `tasks.json` file in the current directory that holds all tasks.
_Avoid_: Database, store, DB
