package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeleteRemovesOnlyThatTask(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTaskFile(t, threeTasks)
	var stdout, stderr bytes.Buffer

	code := run([]string{"delete", "3"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "Task 3 deleted\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	f := readTaskFile(t)
	var ids []int
	for _, task := range f.Tasks {
		ids = append(ids, task.ID)
	}
	if len(ids) != 2 || ids[0] != 5 || ids[1] != 1 {
		t.Errorf("remaining task IDs = %v, want [5 1]", ids)
	}
	if f.NextID != 6 {
		t.Errorf("nextId = %d, want 6 (unchanged)", f.NextID)
	}
}

func TestDeletedIDIsNeverReused(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	for _, d := range []string{"Buy milk", "Fix the bike", "Write README"} {
		run([]string{"add", d}, &stdout, &stderr)
	}
	run([]string{"delete", "3"}, &stdout, &stderr)
	stdout.Reset()

	code := run([]string{"add", "Call mom"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "Task added successfully (ID: 4)\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestDeleteUnknownIDLeavesTaskFileUnchanged(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTaskFile(t, oneTask)
	var stdout, stderr bytes.Buffer

	code := run([]string{"delete", "7"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got, want := stderr.String(), "error: task 7 not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	assertTaskFileIs(t, oneTask)
}

func TestDeleteRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "non-numeric ID", args: []string{"delete", "abc"}},
		{name: "zero ID", args: []string{"delete", "0"}},
		{name: "no arguments", args: []string{"delete"}},
		{name: "extra argument", args: []string{"delete", "1", "2"}},
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
			if !strings.HasPrefix(stderr.String(), "error: ") || !strings.Contains(stderr.String(), "Usage:") {
				t.Errorf("stderr should be an error plus usage:\n%s", stderr.String())
			}
			assertTaskFileIs(t, oneTask)
		})
	}
}
