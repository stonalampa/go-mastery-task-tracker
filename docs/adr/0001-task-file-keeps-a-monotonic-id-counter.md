# Task file keeps a monotonic ID counter

The task file is a JSON object `{"nextId": N, "tasks": [...]}` rather than a bare array, and new tasks get `nextId`, which only ever increases. With the obvious alternative (`max(id) + 1`) deleting the highest task and adding a new one reuses its ID, so a user who remembered "task 5" would silently act on a different task. A UUID would also avoid reuse but makes IDs unpleasant to type on the command line.
