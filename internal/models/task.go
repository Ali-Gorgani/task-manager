package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task represents a to-do task
type Task struct {
	ID          string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title       string     `json:"title" example:"Complete project documentation" binding:"required"`
	Description string     `json:"description" example:"Write comprehensive README and API docs"`
	Status      TaskStatus `json:"status" example:"pending"`
	Assignee    string     `json:"assignee" example:"john.doe@example.com"`
	CreatedAt   time.Time  `json:"created_at" example:"2025-11-01T10:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2025-11-01T12:00:00Z"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required" example:"Complete project documentation"`
	Description string     `json:"description" example:"Write comprehensive README and API docs"`
	Status      TaskStatus `json:"status" example:"pending"`
	Assignee    string     `json:"assignee" example:"john.doe@example.com"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       *string     `json:"title,omitempty" example:"Updated task title"`
	Description *string     `json:"description,omitempty" example:"Updated description"`
	Status      *TaskStatus `json:"status,omitempty" example:"in_progress"`
	Assignee    *string     `json:"assignee,omitempty" example:"jane.doe@example.com"`
}

// TaskFilter represents filtering options for tasks
type TaskFilter struct {
	Status   *TaskStatus `form:"status" example:"pending"`
	Assignee *string     `form:"assignee" example:"john.doe@example.com"`
	Page     int         `form:"page" example:"1"`
	PageSize int         `form:"page_size" example:"10"`
}

// TaskListResponse represents a paginated list of tasks
type TaskListResponse struct {
	Tasks      []Task `json:"tasks"`
	Total      int    `json:"total" example:"100"`
	Page       int    `json:"page" example:"1"`
	PageSize   int    `json:"page_size" example:"10"`
	TotalPages int    `json:"total_pages" example:"10"`
}

// NewTask creates a new task with default values
func NewTask(title, description, assignee string, status TaskStatus) *Task {
	now := time.Now()
	if status == "" {
		status = TaskStatusPending
	}

	return &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      status,
		Assignee:    assignee,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsValidStatus checks if the status is valid
func IsValidStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusCancelled:
		return true
	default:
		return false
	}
}
