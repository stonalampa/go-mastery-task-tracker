package main

import (
	"bytes"
	"strings"
	"testing"
)

// updatedAtIn holds each task's updatedAt in the threeTasks fixture
// (ID 1 todo, ID 3 in-progress, ID 5 done).
var updatedAtIn = map[int]string{
	1: "2026-10-17T08:15:00Z",
	3: "2026-10-18T07:45:00Z",
	5: "2026-10-18T10:30:00Z",
}

func TestMark(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		id          int
		wantStdout  string
		wantStatus  string
		wantUpdated bool // whether updatedAt must change
	}{
		{"todo to in-progress", []string{"mark-in-progress", "1"}, 1, "Task 1 marked in-progress\n", "in-progress", true},
		{"todo to done", []string{"mark-done", "1"}, 1, "Task 1 marked done\n", "done", true},
		{"in-progress to done", []string{"mark-done", "3"}, 3, "Task 3 marked done\n", "done", true},
		{"done reopened as in-progress", []string{"mark-in-progress", "5"}, 5, "Task 5 marked in-progress\n", "in-progress", true},
		{"in-progress stays in-progress", []string{"mark-in-progress", "3"}, 3, "Task 3 marked in-progress\n", "in-progress", false},
		{"done stays done", []string{"mark-done", "5"}, 5, "Task 5 marked done\n", "done", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeTaskFile(t, threeTasks)
			var stdout, stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", got, tt.wantStdout)
			}
			for _, task := range readTaskFile(t).Tasks {
				if task.ID != tt.id {
					continue
				}
				if task.Status != tt.wantStatus {
					t.Errorf("status = %q, want %q", task.Status, tt.wantStatus)
				}
				if updated := task.UpdatedAt != updatedAtIn[tt.id]; updated != tt.wantUpdated {
					t.Errorf("updatedAt changed = %v, want %v (updatedAt %s)", updated, tt.wantUpdated, task.UpdatedAt)
				}
			}
		})
	}
}

func TestMarkRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string // prefix
	}{
		{"unknown ID", []string{"mark-done", "7"}, "error: task 7 not found\n"},
		{"non-numeric ID", []string{"mark-done", "abc"}, `error: invalid task ID "abc"`},
		{"zero ID", []string{"mark-in-progress", "0"}, `error: invalid task ID "0"`},
		{"no ID", []string{"mark-done"}, "error: mark-done takes exactly one argument"},
		{"extra argument", []string{"mark-in-progress", "1", "2"}, "error: mark-in-progress takes exactly one argument"},
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
			if !strings.HasPrefix(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to start with %q", stderr.String(), tt.wantStderr)
			}
			assertTaskFileIs(t, threeTasks)
		})
	}
}
