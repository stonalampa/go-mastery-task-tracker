// Command task-cli tracks tasks in a JSON file in the current directory.
package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/stonalampa/go-mastery-task-tracker/internal/task"
)

// taskFileName is the task file, relative to the current directory.
const taskFileName = "tasks.json"

const usage = `Usage:
  task-cli add <description>
  task-cli update <id> <description>
  task-cli delete <id>
  task-cli mark-in-progress <id>
  task-cli mark-done <id>
  task-cli list [todo|in-progress|done]
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one task-cli command and returns the process exit code.
// args excludes the program name.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 1
	}
	switch args[0] {
	case "add":
		return add(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n%s", args[0], usage)
		return 1
	}
}

func add(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintf(stderr, "error: add takes exactly one argument: <description>\n%s", usage)
		return 1
	}
	f, err := task.Load(taskFileName)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	t, err := f.Add(args[0], time.Now().UTC())
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if err := task.Save(taskFileName, f); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Task added successfully (ID: %d)\n", t.ID)
	return 0
}
