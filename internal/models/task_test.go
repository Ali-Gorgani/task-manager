package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	description := "Test Description"
	assignee := "test@example.com"
	status := TaskStatusPending

	task := NewTask(title, description, assignee, status)

	assert.NotEmpty(t, task.ID)
	assert.Equal(t, title, task.Title)
	assert.Equal(t, description, task.Description)
	assert.Equal(t, assignee, task.Assignee)
	assert.Equal(t, status, task.Status)
	assert.NotZero(t, task.CreatedAt)
	assert.NotZero(t, task.UpdatedAt)
}

func TestNewTask_DefaultStatus(t *testing.T) {
	task := NewTask("Test", "Description", "test@example.com", "")

	assert.Equal(t, TaskStatusPending, task.Status)
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"Valid Pending", TaskStatusPending, true},
		{"Valid InProgress", TaskStatusInProgress, true},
		{"Valid Completed", TaskStatusCompleted, true},
		{"Valid Cancelled", TaskStatusCancelled, true},
		{"Invalid Status", TaskStatus("invalid"), false},
		{"Empty Status", TaskStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
