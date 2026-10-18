package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Load reads the task file at path. A missing file is an empty task list.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &File{NextID: 1}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &f, nil
}

// Save writes f to path atomically: it writes a temporary file in the same
// directory and renames it over path, so a crash leaves either the old or
// the new task file, never a half-written one (ADR-0002).
func Save(path string, f *File) (err error) {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("saving %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tasks-*.json")
	if err != nil {
		return fmt.Errorf("saving %s: %w", path, err)
	}
	defer func() {
		if err != nil {
			// Best-effort cleanup; the original error is the one to report.
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			err = fmt.Errorf("saving %s: %w", path, err)
		}
	}()
	// CreateTemp uses 0600; the task file is ordinary user data.
	if err = tmp.Chmod(0o644); err != nil {
		return err
	}
	if _, err = tmp.Write(data); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
