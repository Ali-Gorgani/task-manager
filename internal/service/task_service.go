package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ali-Gorgani/task-manager/internal/cache"
	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/Ali-Gorgani/task-manager/internal/repository"
)

// TaskService handles business logic for tasks
type TaskService struct {
	repo  repository.TaskRepository
	cache *cache.RedisCache
}

// NewTaskService creates a new task service
func NewTaskService(repo repository.TaskRepository, cache *cache.RedisCache) *TaskService {
	return &TaskService{
		repo:  repo,
		cache: cache,
	}
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.Task, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	if req.Status != "" && !models.IsValidStatus(req.Status) {
		return nil, errors.New("invalid status")
	}

	task := models.NewTask(req.Title, req.Description, req.Assignee, req.Status)

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Invalidate list cache
	if s.cache != nil {
		_ = s.cache.InvalidateTaskList(ctx)
	}

	return task, nil
}

// GetTask retrieves a task by ID (with caching)
func (s *TaskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	// Try cache first
	if s.cache != nil {
		cachedTask, err := s.cache.GetTask(ctx, id)
		if err == nil && cachedTask != nil {
			return cachedTask, nil
		}
	}

	// Cache miss, get from database
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if s.cache != nil {
		_ = s.cache.SetTask(ctx, task)
	}

	return task, nil
}

// ListTasks retrieves all tasks with filtering and pagination (with caching)
func (s *TaskService) ListTasks(ctx context.Context, filter *models.TaskFilter) (*models.TaskListResponse, error) {
	if filter == nil {
		filter = &models.TaskFilter{}
	}

	// Set default pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// Validate filter
	if filter.Status != nil && !models.IsValidStatus(*filter.Status) {
		return nil, errors.New("invalid status filter")
	}

	// Try cache first (only for GET requests with specific filters)
	if s.cache != nil {
		cacheKey := cache.GenerateCacheKey(filter)
		cachedTasks, err := s.cache.GetTaskList(ctx, cacheKey)
		if err == nil && cachedTasks != nil {
			total := len(cachedTasks)
			totalPages := (total + filter.PageSize - 1) / filter.PageSize
			return &models.TaskListResponse{
				Tasks:      cachedTasks,
				Total:      total,
				Page:       filter.Page,
				PageSize:   filter.PageSize,
				TotalPages: totalPages,
			}, nil
		}
	}

	// Cache miss, get from database
	tasks, total, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Store in cache
	if s.cache != nil {
		cacheKey := cache.GenerateCacheKey(filter)
		_ = s.cache.SetTaskList(ctx, cacheKey, tasks)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &models.TaskListResponse{
		Tasks:      tasks,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(ctx context.Context, id string, req *models.UpdateTaskRequest) (*models.Task, error) {
	// Get existing task
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if !models.IsValidStatus(*req.Status) {
			return nil, errors.New("invalid status")
		}
		task.Status = *req.Status
	}
	if req.Assignee != nil {
		task.Assignee = *req.Assignee
	}

	task.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Invalidate caches
	if s.cache != nil {
		_ = s.cache.DeleteTask(ctx, id)
		_ = s.cache.InvalidateTaskList(ctx)
	}

	return task, nil
}

// DeleteTask deletes a task by ID
func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	if s.cache != nil {
		_ = s.cache.DeleteTask(ctx, id)
		_ = s.cache.InvalidateTaskList(ctx)
	}

	return nil
}

// GetTaskCount returns the total number of tasks
func (s *TaskService) GetTaskCount(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}
