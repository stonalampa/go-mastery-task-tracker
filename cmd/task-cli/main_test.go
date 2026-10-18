package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWithoutArgumentsPrintsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(nil, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	for _, cmd := range []string{"add", "update", "delete", "mark-in-progress", "mark-done", "list"} {
		if !strings.Contains(stderr.String(), cmd) {
			t.Errorf("usage on stderr does not mention %q:\n%s", cmd, stderr.String())
		}
	}
}

func TestRunWithUnknownCommandReportsErrorAndUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"foo"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if !strings.HasPrefix(stderr.String(), `error: unknown command "foo"`+"\n") {
		t.Errorf("stderr does not start with the unknown-command error:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("stderr does not include usage:\n%s", stderr.String())
	}
}
