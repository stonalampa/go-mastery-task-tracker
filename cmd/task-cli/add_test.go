package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
)

// taskFile mirrors the task file schema from the spec, independently of the
// internal types, so the tests catch accidental schema changes.
type taskFile struct {
	NextID int `json:"nextId"`
	Tasks  []struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
		Status      string `json:"status"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	} `json:"tasks"`
}

func readTaskFile(t *testing.T) taskFile {
	t.Helper()
	data, err := os.ReadFile("tasks.json")
	if err != nil {
		t.Fatalf("reading task file: %v", err)
	}
	var f taskFile
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("parsing task file: %v\n%s", err, data)
	}
	return f
}

func TestAddCreatesTaskFileWithFirstTask(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer

	code := run([]string{"add", "Buy milk"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "Task added successfully (ID: 1)\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	f := readTaskFile(t)
	if len(f.Tasks) != 1 {
		t.Fatalf("task file has %d tasks, want 1", len(f.Tasks))
	}
	if got := f.Tasks[0]; got.ID != 1 || got.Description != "Buy milk" || got.Status != "todo" {
		t.Errorf("task = %+v, want ID 1, description %q, status todo", got, "Buy milk")
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory has %d entries, want only the task file (no leftover temp files)", len(entries))
	}
}

func TestAddAssignsIncreasingIDsAndTimestamps(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	run([]string{"add", "Buy milk"}, &stdout, &stderr)
	stdout.Reset()

	code := run([]string{"add", "Write README"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got, want := stdout.String(), "Task added successfully (ID: 2)\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	f := readTaskFile(t)
	if f.NextID != 3 {
		t.Errorf("nextId = %d, want 3", f.NextID)
	}
	if len(f.Tasks) != 2 || f.Tasks[1].ID != 2 || f.Tasks[1].Description != "Write README" {
		t.Fatalf("tasks = %+v, want second task with ID 2 and description %q", f.Tasks, "Write README")
	}
	created, err := time.Parse(time.RFC3339Nano, f.Tasks[1].CreatedAt)
	if err != nil {
		t.Fatalf("createdAt %q is not RFC 3339: %v", f.Tasks[1].CreatedAt, err)
	}
	if _, offset := created.Zone(); offset != 0 {
		t.Errorf("createdAt %q is not UTC", f.Tasks[1].CreatedAt)
	}
	if f.Tasks[1].UpdatedAt != f.Tasks[1].CreatedAt {
		t.Errorf("updatedAt %q != createdAt %q for a new task", f.Tasks[1].UpdatedAt, f.Tasks[1].CreatedAt)
	}
}

func TestAddRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantUsage bool
	}{
		{name: "empty description", args: []string{"add", ""}},
		{name: "whitespace description", args: []string{"add", "   "}},
		{name: "missing description", args: []string{"add"}, wantUsage: true},
		{name: "extra argument", args: []string{"add", "Buy milk", "now"}, wantUsage: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			var stdout, stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)

			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if !strings.HasPrefix(stderr.String(), "error: ") {
				t.Errorf("stderr does not start with %q:\n%s", "error: ", stderr.String())
			}
			if got := strings.Contains(stderr.String(), "Usage:"); got != tt.wantUsage {
				t.Errorf("usage shown = %v, want %v:\n%s", got, tt.wantUsage, stderr.String())
			}
			if _, err := os.Stat("tasks.json"); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("task file was created (stat err = %v), want none", err)
			}
		})
	}
}

func TestAddKeepsTaskFilePermissions(t *testing.T) {
	t.Run("new task file is private", func(t *testing.T) {
		t.Chdir(t.TempDir())
		var stdout, stderr bytes.Buffer

		run([]string{"add", "Buy milk"}, &stdout, &stderr)

		assertMode(t, 0o600)
	})
	t.Run("existing task file keeps its mode", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := os.WriteFile("tasks.json", []byte(`{"nextId": 1, "tasks": []}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod("tasks.json", 0o640); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer

		run([]string{"add", "Buy milk"}, &stdout, &stderr)

		assertMode(t, 0o640)
	})
}

func assertMode(t *testing.T, want fs.FileMode) {
	t.Helper()
	info, err := os.Stat("tasks.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("task file mode = %v, want %v", got, want)
	}
}
