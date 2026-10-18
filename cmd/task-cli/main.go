// Command task-cli tracks tasks in a JSON file in the current directory.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/stonalampa/go-mastery-task-tracker/internal/task"
)

// cellEscaper shows characters that would break the list table as visible
// escapes; the task file keeps the description exactly as typed.
var cellEscaper = strings.NewReplacer("\t", `\t`, "\n", `\n`, "\r", `\r`)

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
	case "list":
		return list(args[1:], stdout, stderr)
	default:
		return failUsage(stderr, fmt.Errorf("unknown command %q", args[0]))
	}
}

func add(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return failUsage(stderr, errors.New("add takes exactly one argument: <description>"))
	}
	f, err := task.Load(taskFileName)
	if err != nil {
		return fail(stderr, err)
	}
	t, err := f.Add(args[0], time.Now().UTC())
	if err != nil {
		return fail(stderr, err)
	}
	if err := task.Save(taskFileName, f); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Task added successfully (ID: %d)\n", t.ID)
	return 0
}

func list(args []string, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		return failUsage(stderr, errors.New("list takes at most one argument: [todo|in-progress|done]"))
	}
	var status task.Status
	if len(args) == 1 {
		var err error
		if status, err = task.ParseStatus(args[0]); err != nil {
			return failUsage(stderr, err)
		}
	}
	f, err := task.Load(taskFileName)
	if err != nil {
		return fail(stderr, err)
	}
	tasks := f.List(status)
	if len(tasks) == 0 {
		fmt.Fprintln(stdout, "No tasks found.")
		return 0
	}
	const timeLayout = "2006-01-02 15:04"
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tDESCRIPTION\tCREATED\tUPDATED")
	for _, t := range tasks {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", t.ID, t.Status, cellEscaper.Replace(t.Description),
			t.CreatedAt.Local().Format(timeLayout), t.UpdatedAt.Local().Format(timeLayout))
	}
	if err := w.Flush(); err != nil {
		return fail(stderr, err)
	}
	return 0
}

// fail reports err on stderr and returns the error exit code.
func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "error: %v\n", err)
	return 1
}

// failUsage reports err followed by the usage text and returns the error exit code.
func failUsage(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "error: %v\n%s", err, usage)
	return 1
}
