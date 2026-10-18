package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// oneTask is a task file with a single in-progress task created and last
// updated in the past, so a fresh update is distinguishable.
const oneTask = `{"nextId": 2, "tasks": [{"id": 1, "description": "Buy milk", "status": "in-progress",
	"createdAt": "2026-10-17T08:15:00Z", "updatedAt": "2026-10-17T09:00:00Z"}]}`

func TestUpdateChangesDescriptionAndUpdatedTime(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTaskFile(t, oneTask)
	var stdout, stderr bytes.Buffer

	code := run([]string{"update", "1", "Buy milk and bread"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "Task 1 updated\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	got := readTaskFile(t).Tasks[0]
	if got.Description != "Buy milk and bread" {
		t.Errorf("description = %q, want %q", got.Description, "Buy milk and bread")
	}
	if got.Status != "in-progress" || got.CreatedAt != "2026-10-17T08:15:00Z" {
		t.Errorf("status/createdAt = %q/%q, want them unchanged (in-progress/2026-10-17T08:15:00Z)", got.Status, got.CreatedAt)
	}
	if got.UpdatedAt == "2026-10-17T09:00:00Z" {
		t.Errorf("updatedAt = %q, want it refreshed", got.UpdatedAt)
	}
}

func TestUpdateUnknownIDLeavesTaskFileUnchanged(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTaskFile(t, oneTask)
	var stdout, stderr bytes.Buffer

	code := run([]string{"update", "7", "Anything"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got, want := stderr.String(), "error: task 7 not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	assertTaskFileIs(t, oneTask)
}

func TestUpdateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantUsage bool
	}{
		{name: "non-numeric ID", args: []string{"update", "abc", "x"}, wantUsage: true},
		{name: "zero ID", args: []string{"update", "0", "x"}, wantUsage: true},
		{name: "negative ID", args: []string{"update", "-1", "x"}, wantUsage: true},
		{name: "plus-signed ID", args: []string{"update", "+1", "x"}, wantUsage: true},
		{name: "empty description", args: []string{"update", "1", ""}},
		{name: "whitespace description", args: []string{"update", "1", "   "}},
		{name: "no arguments", args: []string{"update"}, wantUsage: true},
		{name: "missing description", args: []string{"update", "1"}, wantUsage: true},
		{name: "extra argument", args: []string{"update", "1", "x", "y"}, wantUsage: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeTaskFile(t, oneTask)
			var stdout, stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)

			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if !strings.HasPrefix(stderr.String(), "error: ") {
				t.Errorf("stderr does not start with %q:\n%s", "error: ", stderr.String())
			}
			if got := strings.Contains(stderr.String(), "Usage:"); got != tt.wantUsage {
				t.Errorf("usage shown = %v, want %v:\n%s", got, tt.wantUsage, stderr.String())
			}
			assertTaskFileIs(t, oneTask)
		})
	}
}

// assertTaskFileIs fails unless the task file holds exactly want.
func assertTaskFileIs(t *testing.T, want string) {
	t.Helper()
	got, err := os.ReadFile("tasks.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("task file changed:\n%s\nwant:\n%s", got, want)
	}
}
