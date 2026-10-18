package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// commandsUsingTaskFile holds one invocation of every command that loads the task file.
var commandsUsingTaskFile = [][]string{
	{"list"},
	{"list", "todo"},
	{"add", "Buy milk"},
	{"update", "1", "Buy bread"},
	{"delete", "1"},
}

func TestCorruptTaskFileIsReportedAndLeftUntouched(t *testing.T) {
	const truncated = `{"nextId": 2, "tasks": [{"id": 1, "descr`
	for _, args := range commandsUsingTaskFile {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Chdir(t.TempDir())
			writeTaskFile(t, truncated)
			var stdout, stderr bytes.Buffer

			code := run(args, &stdout, &stderr)

			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if want := "error: tasks.json is not valid JSON: "; !strings.HasPrefix(stderr.String(), want) {
				t.Errorf("stderr = %q, want it to start with %q", stderr.String(), want)
			}
			assertTaskFileIs(t, truncated)
			assertOnlyTaskFileInDir(t)
		})
	}
}

func TestTaskFileWithUnsafeIDsIsReportedAndLeftUntouched(t *testing.T) {
	const task = `"description": "Buy milk", "status": "todo",
		"createdAt": "2026-10-17T08:15:00Z", "updatedAt": "2026-10-17T08:15:00Z"`
	files := map[string]string{
		"no nextId":           `{"tasks": []}`,
		"nextId equals an ID": `{"nextId": 1, "tasks": [{"id": 1, ` + task + `}]}`,
		"nextId below an ID":  `{"nextId": 1, "tasks": [{"id": 5, ` + task + `}]}`,
		"task ID zero":        `{"nextId": 2, "tasks": [{"id": 0, ` + task + `}]}`,
		"duplicate task IDs":  `{"nextId": 3, "tasks": [{"id": 2, ` + task + `}, {"id": 2, ` + task + `}]}`,
	}
	for name, content := range files {
		for _, args := range commandsUsingTaskFile {
			t.Run(name+"/"+strings.Join(args, " "), func(t *testing.T) {
				t.Chdir(t.TempDir())
				writeTaskFile(t, content)
				var stdout, stderr bytes.Buffer

				code := run(args, &stdout, &stderr)

				if code != 1 {
					t.Errorf("exit code = %d, want 1", code)
				}
				if want := "error: tasks.json is invalid: "; !strings.HasPrefix(stderr.String(), want) {
					t.Errorf("stderr = %q, want it to start with %q", stderr.String(), want)
				}
				assertTaskFileIs(t, content)
				assertOnlyTaskFileInDir(t)
			})
		}
	}
}

// assertOnlyTaskFileInDir fails if anything besides the task file (such as
// a leftover temp file) is in the current directory.
func assertOnlyTaskFileInDir(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "tasks.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory contains %v, want only tasks.json", names)
	}
}
