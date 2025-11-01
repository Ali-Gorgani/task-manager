package repository

import (
	"context"

	"github.com/Ali-Gorgani/task-manager/internal/models"
)

// TaskRepository defines the interface for task storage operations
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id string) (*models.Task, error)
	GetAll(ctx context.Context, filter *models.TaskFilter) ([]models.Task, int, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}
