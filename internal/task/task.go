// Package task holds the task tracker's domain model and its task file.
package task

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Status is where a task stands.
type Status string

// The three statuses a task can have.
const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

// ParseStatus returns the status named s, or an error if there is none.
func ParseStatus(s string) (Status, error) {
	switch st := Status(s); st {
	case StatusTodo, StatusInProgress, StatusDone:
		return st, nil
	default:
		return "", fmt.Errorf("unknown status %q", s)
	}
}

// ErrNotFound is returned, wrapped with the task ID, when no task has that ID.
var ErrNotFound = errors.New("not found")

// ErrEmptyDescription is returned when a description is empty or only spaces.
var ErrEmptyDescription = errors.New("description must not be empty")

// Task is a single piece of work the user wants to track.
type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// File is the content of the task file: all tasks plus the next task ID,
// which only ever increases (ADR-0001).
type File struct {
	NextID int    `json:"nextId"`
	Tasks  []Task `json:"tasks"`
}

// find returns the index of the task with the given ID.
func (f *File) find(id int) (int, error) {
	for i, t := range f.Tasks {
		if t.ID == id {
			return i, nil
		}
	}
	return 0, fmt.Errorf("task %d %w", id, ErrNotFound)
}

// Add records a new task with status todo, created and updated at now.
func (f *File) Add(description string, now time.Time) (Task, error) {
	if strings.TrimSpace(description) == "" {
		return Task{}, ErrEmptyDescription
	}
	t := Task{
		ID:          f.NextID,
		Description: description,
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	f.NextID++
	f.Tasks = append(f.Tasks, t)
	return t, nil
}

// List returns the tasks with the given status, sorted by ID.
// The empty status lists all tasks.
func (f *File) List(status Status) []Task {
	var tasks []Task
	for _, t := range f.Tasks {
		if status == "" || t.Status == status {
			tasks = append(tasks, t)
		}
	}
	slices.SortFunc(tasks, func(a, b Task) int { return cmp.Compare(a.ID, b.ID) })
	return tasks
}

// Update replaces the description of the task with the given ID and sets
// its updated time to now.
func (f *File) Update(id int, description string, now time.Time) error {
	if strings.TrimSpace(description) == "" {
		return ErrEmptyDescription
	}
	i, err := f.find(id)
	if err != nil {
		return err
	}
	f.Tasks[i].Description = description
	f.Tasks[i].UpdatedAt = now
	return nil
}

// Delete removes the task with the given ID. Its ID is never reused,
// because NextID is left as is (ADR-0001).
func (f *File) Delete(id int) error {
	i, err := f.find(id)
	if err != nil {
		return err
	}
	f.Tasks = slices.Delete(f.Tasks, i, i+1)
	return nil
}

// Mark sets the status of the task with the given ID and its updated time
// to now. Marking a task with the status it already has changes nothing.
func (f *File) Mark(id int, status Status, now time.Time) error {
	i, err := f.find(id)
	if err != nil {
		return err
	}
	if f.Tasks[i].Status == status {
		return nil
	}
	f.Tasks[i].Status = status
	f.Tasks[i].UpdatedAt = now
	return nil
}
