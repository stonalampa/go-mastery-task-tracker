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
		return nil, fmt.Errorf("%s is not valid JSON: %w", path, err)
	}
	if err := f.checkIDs(); err != nil {
		return nil, fmt.Errorf("%s is invalid: %w", path, err)
	}
	return &f, nil
}

// checkIDs rejects a task file whose IDs could be reused or are unusable
// (ADR-0001): nextId must be at least 1 and above every task ID, and task
// IDs must be positive and unique.
func (f *File) checkIDs() error {
	if f.NextID < 1 {
		return fmt.Errorf("nextId %d must be at least 1", f.NextID)
	}
	seen := make(map[int]bool, len(f.Tasks))
	for _, t := range f.Tasks {
		switch {
		case t.ID < 1:
			return fmt.Errorf("task ID %d must be at least 1", t.ID)
		case t.ID >= f.NextID:
			return fmt.Errorf("nextId %d is not greater than task ID %d", f.NextID, t.ID)
		case seen[t.ID]:
			return fmt.Errorf("task ID %d appears more than once", t.ID)
		}
		seen[t.ID] = true
	}
	return nil
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
	// Keep the mode of an existing task file; a new one stays at
	// CreateTemp's private 0600.
	if info, statErr := os.Stat(path); statErr == nil {
		if err = tmp.Chmod(info.Mode().Perm()); err != nil {
			return err
		}
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
