// Command task-cli tracks tasks in a JSON file in the current directory.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"

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
	case "list":
		return list(args[1:], stdout, stderr)
	case "update":
		return update(args[1:], stdout, stderr)
	case "delete":
		return remove(args[1:], stdout, stderr)
	case "mark-in-progress":
		return mark(task.StatusInProgress, args[1:], stdout, stderr)
	case "mark-done":
		return mark(task.StatusDone, args[1:], stdout, stderr)
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
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", t.ID, t.Status, escapeControl(t.Description),
			t.CreatedAt.Local().Format(timeLayout), t.UpdatedAt.Local().Format(timeLayout))
	}
	if err := w.Flush(); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func update(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		return failUsage(stderr, errors.New("update takes exactly two arguments: <id> <description>"))
	}
	id, err := parseID(args[0])
	if err != nil {
		return failUsage(stderr, err)
	}
	f, err := task.Load(taskFileName)
	if err != nil {
		return fail(stderr, err)
	}
	if err := f.Update(id, args[1], time.Now().UTC()); err != nil {
		return fail(stderr, err)
	}
	if err := task.Save(taskFileName, f); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Task %d updated\n", id)
	return 0
}

// remove implements the delete command (delete is a Go builtin).
func remove(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return failUsage(stderr, errors.New("delete takes exactly one argument: <id>"))
	}
	id, err := parseID(args[0])
	if err != nil {
		return failUsage(stderr, err)
	}
	f, err := task.Load(taskFileName)
	if err != nil {
		return fail(stderr, err)
	}
	if err := f.Delete(id); err != nil {
		return fail(stderr, err)
	}
	if err := task.Save(taskFileName, f); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Task %d deleted\n", id)
	return 0
}

// parseID parses a task ID argument, which must be a positive whole number.
func parseID(s string) (int, error) {
	// ParseUint rejects any sign ("+1", "-1"); IntSize-1 bits always fit an int.
	id, err := strconv.ParseUint(s, 10, strconv.IntSize-1)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid task ID %q: must be a positive whole number", s)
	}
	return int(id), nil
}

// escapeControl shows control characters (tab, newline, ESC, ...) as Go
// escapes such as \t or \x1b, so a description can neither break the list
// table nor send commands to the terminal. The task file keeps it as typed.
func escapeControl(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsControl(r) {
			b.WriteRune(r)
			continue
		}
		q := strconv.QuoteRune(r) // e.g. '\t' or '\x1b', with the quotes
		b.WriteString(q[1 : len(q)-1])
	}
	return b.String()
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

// mark implements mark-in-progress and mark-done: the command name decides
// the status, the only argument is the task ID.
func mark(status task.Status, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return failUsage(stderr, fmt.Errorf("mark-%s takes exactly one argument: <id>", status))
	}
	id, err := parseID(args[0])
	if err != nil {
		return failUsage(stderr, err)
	}
	f, err := task.Load(taskFileName)
	if err != nil {
		return fail(stderr, err)
	}
	if err := f.Mark(id, status, time.Now().UTC()); err != nil {
		return fail(stderr, err)
	}
	if err := task.Save(taskFileName, f); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Task %d marked %s\n", id, status)
	return 0
}
