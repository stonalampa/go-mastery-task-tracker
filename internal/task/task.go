// Package task holds the task tracker's domain model and its task file.
package task

import (
	"errors"
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
