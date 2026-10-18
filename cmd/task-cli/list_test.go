package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestListWithoutTaskFileSaysNoTasksAndCreatesNothing(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer

	code := run([]string{"list"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "No tasks found.\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if _, err := os.Stat("tasks.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("task file was created (stat err = %v), want none", err)
	}
}

// threeTasks is a task file with one task per status, stored out of ID order.
const threeTasks = `{
  "nextId": 6,
  "tasks": [
    {"id": 5, "description": "Write README", "status": "done",
     "createdAt": "2026-10-18T09:00:00Z", "updatedAt": "2026-10-18T10:30:00Z"},
    {"id": 1, "description": "Buy milk", "status": "todo",
     "createdAt": "2026-10-17T08:15:00Z", "updatedAt": "2026-10-17T08:15:00Z"},
    {"id": 3, "description": "Fix the bike", "status": "in-progress",
     "createdAt": "2026-10-17T12:00:00Z", "updatedAt": "2026-10-18T07:45:00Z"}
  ]
}`

// listRow matches one table row: ID, status, description, created, updated.
var listRow = regexp.MustCompile(`^(\d+)\s+(\S+)\s+(.+?)\s+(\d{4}-\d{2}-\d{2} \d{2}:\d{2})\s+(\d{4}-\d{2}-\d{2} \d{2}:\d{2})$`)

// listedTasks parses list output into "ID status description" rows,
// checking the header and the time format along the way.
func listedTasks(t *testing.T, out string) []string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if got := strings.Fields(lines[0]); strings.Join(got, " ") != "ID STATUS DESCRIPTION CREATED UPDATED" {
		t.Fatalf("header = %q, want ID STATUS DESCRIPTION CREATED UPDATED", lines[0])
	}
	statusCol, descCol := strings.Index(lines[0], "STATUS"), strings.Index(lines[0], "DESCRIPTION")
	var rows []string
	for _, line := range lines[1:] {
		m := listRow.FindStringSubmatch(line)
		if m == nil {
			t.Fatalf("row %q does not match ID, status, description and two YYYY-MM-DD HH:MM times", line)
		}
		if !strings.HasPrefix(line[statusCol:], m[2]) || !strings.HasPrefix(line[descCol:], m[3]) {
			t.Errorf("row %q is not aligned with header %q", line, lines[0])
		}
		rows = append(rows, m[1]+" "+m[2]+" "+m[3])
	}
	return rows
}

func writeTaskFile(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile("tasks.json", []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestListShowsAllTasksSortedByID(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTaskFile(t, threeTasks)
	var stdout, stderr bytes.Buffer

	code := run([]string{"list"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	want := []string{"1 todo Buy milk", "3 in-progress Fix the bike", "5 done Write README"}
	if got := listedTasks(t, stdout.String()); !slices.Equal(got, want) {
		t.Errorf("listed tasks = %q, want %q", got, want)
	}
}

func TestListFiltersByStatus(t *testing.T) {
	tests := []struct {
		filter string
		want   []string
	}{
		{filter: "todo", want: []string{"1 todo Buy milk"}},
		{filter: "in-progress", want: []string{"3 in-progress Fix the bike"}},
		{filter: "done", want: []string{"5 done Write README"}},
	}
	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeTaskFile(t, threeTasks)
			var stdout, stderr bytes.Buffer

			code := run([]string{"list", tt.filter}, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
			}
			if got := listedTasks(t, stdout.String()); !slices.Equal(got, tt.want) {
				t.Errorf("listed tasks = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListWithNoMatchingTasksSaysNoTasks(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTaskFile(t, `{"nextId": 2, "tasks": [{"id": 1, "description": "Buy milk", "status": "todo",
		"createdAt": "2026-10-17T08:15:00Z", "updatedAt": "2026-10-17T08:15:00Z"}]}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"list", "done"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "No tasks found.\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestListRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown filter", args: []string{"list", "blocked"}},
		{name: "extra argument", args: []string{"list", "todo", "done"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeTaskFile(t, threeTasks)
			var stdout, stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)

			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if !strings.HasPrefix(stderr.String(), "error: ") || !strings.Contains(stderr.String(), "Usage:") {
				t.Errorf("stderr should be an error plus usage:\n%s", stderr.String())
			}
		})
	}
}

func TestListEscapesControlCharactersInDescriptions(t *testing.T) {
	t.Chdir(t.TempDir())
	// JSON "\t", "\n" and "\r" decode to a real tab, newline and carriage return.
	writeTaskFile(t, `{"nextId": 2, "tasks": [{"id": 1, "description": "a\tb\nc\rd", "status": "todo",
		"createdAt": "2026-10-17T08:15:00Z", "updatedAt": "2026-10-17T08:15:00Z"}]}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"list"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	want := []string{`1 todo a\tb\nc\rd`}
	if got := listedTasks(t, stdout.String()); !slices.Equal(got, want) {
		t.Errorf("listed tasks = %q, want %q", got, want)
	}
}
