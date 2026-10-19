# task-cli

[![CI](https://github.com/stonalampa/go-mastery-task-tracker/actions/workflows/ci.yml/badge.svg)](https://github.com/stonalampa/go-mastery-task-tracker/actions/workflows/ci.yml)

A small command-line task tracker written in Go. It keeps a to-do list in a JSON file in the current directory. It is my solution to the [roadmap.sh Task Tracker](https://roadmap.sh/projects/task-tracker) project and part of my `go-mastery` series. It uses the standard library only and has no third-party dependencies.

## Install

```sh
go install github.com/stonalampa/go-mastery-task-tracker/cmd/task-cli@latest
```

Or build from source (Go 1.27+):

```sh
git clone https://github.com/stonalampa/go-mastery-task-tracker.git
cd go-mastery-task-tracker
make build          # -> bin/task-cli
```

## Usage

```text
task-cli add <description>
task-cli update <id> <description>
task-cli delete <id>
task-cli mark-in-progress <id>
task-cli mark-done <id>
task-cli list [todo|in-progress|done]
```

A real session:

```text
$ task-cli add "Buy milk"
Task added successfully (ID: 1)
$ task-cli add "Fix the bike"
Task added successfully (ID: 2)
$ task-cli add "Write the README"
Task added successfully (ID: 3)
$ task-cli mark-in-progress 2
Task 2 marked in-progress
$ task-cli mark-done 3
Task 3 marked done
$ task-cli list
ID  STATUS       DESCRIPTION       CREATED           UPDATED
1   todo         Buy milk          2026-10-19 13:54  2026-10-19 13:54
2   in-progress  Fix the bike      2026-10-19 13:54  2026-10-19 13:54
3   done         Write the README  2026-10-19 13:54  2026-10-19 13:54
$ task-cli list in-progress
ID  STATUS       DESCRIPTION   CREATED           UPDATED
2   in-progress  Fix the bike  2026-10-19 13:54  2026-10-19 13:54
$ task-cli update 1 "Buy milk and bread"
Task 1 updated
$ task-cli delete 3
Task 3 deleted
$ task-cli add "Call mom"
Task added successfully (ID: 4)
$ task-cli list
ID  STATUS       DESCRIPTION         CREATED           UPDATED
1   todo         Buy milk and bread  2026-10-19 13:54  2026-10-19 13:54
2   in-progress  Fix the bike        2026-10-19 13:54  2026-10-19 13:54
4   todo         Call mom            2026-10-19 13:54  2026-10-19 13:54
$ task-cli mark-done 9
error: task 9 not found
```

Success messages go to stdout. Errors go to stderr as `error: ...`. Mistakes in the command line itself (unknown command or status, wrong number of arguments, invalid ID) are followed by the usage text. The exit code is `0` on success and `1` on any error.

## The task file

Tasks live in `tasks.json` in the current directory. The file is created by the first `add`; `list` never creates it.

```json
{
  "nextId": 5,
  "tasks": [
    {
      "id": 1,
      "description": "Buy milk and bread",
      "status": "todo",
      "createdAt": "2026-10-19T11:54:03.183748Z",
      "updatedAt": "2026-10-19T11:54:03.266072Z"
    }
  ]
}
```

Times are stored in UTC and shown in local time.

Guarantees:

- **IDs are never reused.** `nextId` only increases. After deleting task 3, the next task gets ID 4, so an ID you remember never points at a different task ([ADR-0001](docs/adr/0001-task-file-keeps-a-monotonic-id-counter.md)). This holds as long as two `task-cli` processes don't run at the same moment (see the limitations below).
- **Saves are atomic.** A command that changes tasks writes a temporary file in the same directory, syncs it, and renames it over `tasks.json`; a command that fails or only reads never writes. A crash leaves the old file or the new one, never a half-written file. An existing file keeps its permissions, and a new file is private (`0600`) ([ADR-0002](docs/adr/0002-whole-file-atomic-rewrite-without-locking.md)).
- **A damaged file is never overwritten.** Invalid JSON, or IDs that could collide (`nextId` not above every task ID, duplicate IDs, IDs below 1), make every command fail with an error. Nothing is written.
- **Marking a task with the status it already has changes nothing**, and its `updatedAt` stays the same.

Known limitations:

- There is no file locking. Two `task-cli` processes running at the same moment can lose one of the two changes, and two simultaneous `add`s can both report the same ID. This is acceptable for a single-user CLI and is documented in ADR-0002.
- Statuses read from `tasks.json` are not validated: a hand-edited unknown status (e.g. `"blocked"`) loads and is listed as is.
- `list` shows control characters in a description (tab, newline, terminal escapes) as visible escapes such as `\t` or `\x1b`. The task file keeps the description exactly as it was typed.

## Go concepts this project practises

- **Testable main.** `main` only calls `run(args, stdout, stderr) int`. All behaviour sits behind that function, so tests drive the real program with in-memory buffers and a temporary directory, without starting a process.
- **One test seam.** Every test goes through `run` and checks only what a user can observe: output, exit code and the task file. Most tests are table-driven with `t.Run` subtests. Each test was either seen failing before its code existed, or mutation-checked: the guarded code was broken on purpose to confirm the test fails.
- **Atomic file writes** with `os.CreateTemp`, `Sync` and `os.Rename`, and deferred cleanup that wraps the original error.
- **Typed string status.** `type Status string` with constants. A status given on the command line is parsed once at the edge (`ParseStatus` for the `list` filter, the command name itself for `mark-*`) and trusted inside.
- **Errors as values.** The domain exposes sentinel errors so that a future caller, such as an HTTP API, could tell them apart with `errors.Is` (for example, 404 vs 400) without parsing message text. `ErrNotFound` is wrapped with the task ID using `%w` ("task 7 not found"), and `ErrEmptyDescription` is returned as is. The CLI itself just prints them.
- **Slices and copies.** `List` sorts a new slice instead of the task file's own, and `Delete` uses `slices.Delete`.

## Project layout

```text
cmd/task-cli/        entry point: argument dispatch, output, tests through run
internal/task/       domain (Task, Status, File operations) and task-file persistence
docs/adr/            architecture decision records
GLOSSARY.md          domain vocabulary (Task, Task ID, Status, Mark, Task file)
```

## Development

```sh
make test     # go test -race ./...
make vet      # go vet ./...
make lint     # golangci-lint run (v2; standard linters + revive, gofmt, goimports)
make vuln     # govulncheck ./...
make build    # bin/task-cli
```

CI runs vet, race-enabled tests, golangci-lint and govulncheck on every pull request, and `main` accepts changes only through pull requests with a green CI run.

The work was planned as a spec ([#1](https://github.com/stonalampa/go-mastery-task-tracker/issues/1)) and split into small tickets. Each ticket became one PR, built test-first, reviewed by an independent AI reviewer (OpenAI Codex), and walked through line by line before merge.
