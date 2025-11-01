package main

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/Ali-Gorgani/task-manager/internal/cache"
	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/Ali-Gorgani/task-manager/internal/repository"
	"github.com/Ali-Gorgani/task-manager/internal/service"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests require a running PostgreSQL instance
// These tests are designed to run with the test script

func setupTestDB(t *testing.T) (*sql.DB, *repository.PostgresTaskRepository) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	repo := repository.NewPostgresTaskRepository(db)
	err = repo.InitSchema(context.Background())
	require.NoError(t, err, "Failed to initialize schema")

	// Clean up existing data
	_, err = db.Exec("DELETE FROM tasks")
	require.NoError(t, err, "Failed to clean up test data")

	return db, repo
}

func setupTestRedis(t *testing.T) *cache.RedisCache {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   1, // Use DB 1 for tests
	})

	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		t.Log("Redis not available, running tests without cache")
		return nil
	}

	// Clean up test cache
	redisClient.FlushDB(ctx)

	return cache.NewRedisCache(redisClient)
}

func TestIntegration_TaskLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, repo := setupTestDB(t)
	defer db.Close()

	redisCache := setupTestRedis(t)
	taskService := service.NewTaskService(repo, redisCache)

	ctx := context.Background()

	t.Run("Create, Read, Update, Delete task", func(t *testing.T) {
		// 1. Create a task
		createReq := &models.CreateTaskRequest{
			Title:       "Integration Test Task",
			Description: "This is a test task",
			Status:      models.TaskStatusPending,
			Assignee:    "test@example.com",
		}

		createdTask, err := taskService.CreateTask(ctx, createReq)
		require.NoError(t, err)
		assert.NotEmpty(t, createdTask.ID)
		assert.Equal(t, "Integration Test Task", createdTask.Title)
		assert.Equal(t, models.TaskStatusPending, createdTask.Status)

		// 2. Read the task
		retrievedTask, err := taskService.GetTask(ctx, createdTask.ID)
		require.NoError(t, err)
		assert.Equal(t, createdTask.ID, retrievedTask.ID)
		assert.Equal(t, createdTask.Title, retrievedTask.Title)

		// 3. Update the task
		newStatus := models.TaskStatusInProgress
		updateReq := &models.UpdateTaskRequest{
			Status: &newStatus,
		}

		updatedTask, err := taskService.UpdateTask(ctx, createdTask.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, models.TaskStatusInProgress, updatedTask.Status)

		// 4. Delete the task
		err = taskService.DeleteTask(ctx, createdTask.ID)
		require.NoError(t, err)

		// 5. Verify deletion
		_, err = taskService.GetTask(ctx, createdTask.ID)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrTaskNotFound, err)
	})
}

func TestIntegration_CacheInvalidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, repo := setupTestDB(t)
	defer db.Close()

	redisCache := setupTestRedis(t)
	if redisCache == nil {
		t.Skip("Redis not available, skipping cache test")
	}

	taskService := service.NewTaskService(repo, redisCache)
	ctx := context.Background()

	t.Run("Cache invalidation on update", func(t *testing.T) {
		// Create a task
		createReq := &models.CreateTaskRequest{
			Title:       "Cache Test Task",
			Description: "Testing cache invalidation",
			Status:      models.TaskStatusPending,
			Assignee:    "cache@example.com",
		}

		task, err := taskService.CreateTask(ctx, createReq)
		require.NoError(t, err)

		// First read - should populate cache
		task1, err := taskService.GetTask(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, models.TaskStatusPending, task1.Status)

		// Update the task
		newStatus := models.TaskStatusCompleted
		updateReq := &models.UpdateTaskRequest{
			Status: &newStatus,
		}

		_, err = taskService.UpdateTask(ctx, task.ID, updateReq)
		require.NoError(t, err)

		// Second read - should get updated value from DB (cache invalidated)
		task2, err := taskService.GetTask(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, models.TaskStatusCompleted, task2.Status)

		// Clean up
		err = taskService.DeleteTask(ctx, task.ID)
		require.NoError(t, err)
	})

	t.Run("List cache invalidation on create", func(t *testing.T) {
		// Get initial list - populates cache
		filter := &models.TaskFilter{
			Page:     1,
			PageSize: 10,
		}
		list1, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		initialCount := list1.Total

		// Create a new task - should invalidate list cache
		createReq := &models.CreateTaskRequest{
			Title:       "New Task",
			Description: "This should invalidate list cache",
			Status:      models.TaskStatusPending,
			Assignee:    "new@example.com",
		}
		newTask, err := taskService.CreateTask(ctx, createReq)
		require.NoError(t, err)

		// Get list again - should reflect new task
		list2, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, initialCount+1, list2.Total)

		// Clean up
		err = taskService.DeleteTask(ctx, newTask.ID)
		require.NoError(t, err)
	})
}

func TestIntegration_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, repo := setupTestDB(t)
	defer db.Close()

	redisCache := setupTestRedis(t)
	taskService := service.NewTaskService(repo, redisCache)

	ctx := context.Background()

	t.Run("Pagination works correctly", func(t *testing.T) {
		// Create 15 tasks
		taskIDs := make([]string, 15)
		for i := 0; i < 15; i++ {
			createReq := &models.CreateTaskRequest{
				Title:       "Pagination Test Task",
				Description: "Testing pagination",
				Status:      models.TaskStatusPending,
				Assignee:    "pagination@example.com",
			}
			task, err := taskService.CreateTask(ctx, createReq)
			require.NoError(t, err)
			taskIDs[i] = task.ID
		}

		// Test page 1
		filter := &models.TaskFilter{
			Page:     1,
			PageSize: 10,
		}
		page1, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 10, len(page1.Tasks))
		assert.Equal(t, 15, page1.Total)
		assert.Equal(t, 2, page1.TotalPages)

		// Test page 2
		filter.Page = 2
		page2, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 5, len(page2.Tasks))
		assert.Equal(t, 15, page2.Total)

		// Clean up
		for _, id := range taskIDs {
			err = taskService.DeleteTask(ctx, id)
			require.NoError(t, err)
		}
	})

	t.Run("Page size limits", func(t *testing.T) {
		// Create 5 tasks
		taskIDs := make([]string, 5)
		for i := 0; i < 5; i++ {
			createReq := &models.CreateTaskRequest{
				Title:       "Page Size Test",
				Description: "Testing page size limits",
				Status:      models.TaskStatusPending,
				Assignee:    "pagesize@example.com",
			}
			task, err := taskService.CreateTask(ctx, createReq)
			require.NoError(t, err)
			taskIDs[i] = task.ID
		}

		// Test default page size
		filter := &models.TaskFilter{}
		result, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result.Tasks), 10) // Default page size

		// Test max page size (should cap at 100)
		filter.PageSize = 150
		result, err = taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, result.PageSize, 100)

		// Clean up
		for _, id := range taskIDs {
			err = taskService.DeleteTask(ctx, id)
			require.NoError(t, err)
		}
	})
}

func TestIntegration_Filtering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, repo := setupTestDB(t)
	defer db.Close()

	redisCache := setupTestRedis(t)
	taskService := service.NewTaskService(repo, redisCache)

	ctx := context.Background()

	t.Run("Filtering by status", func(t *testing.T) {
		// Create tasks with different statuses
		pendingReq := &models.CreateTaskRequest{
			Title:    "Pending Task",
			Status:   models.TaskStatusPending,
			Assignee: "filter@example.com",
		}
		pendingTask, err := taskService.CreateTask(ctx, pendingReq)
		require.NoError(t, err)

		completedReq := &models.CreateTaskRequest{
			Title:    "Completed Task",
			Status:   models.TaskStatusCompleted,
			Assignee: "filter@example.com",
		}
		completedTask, err := taskService.CreateTask(ctx, completedReq)
		require.NoError(t, err)

		// Filter by pending status
		pendingStatus := models.TaskStatusPending
		filter := &models.TaskFilter{
			Status:   &pendingStatus,
			Page:     1,
			PageSize: 10,
		}
		result, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, 1)
		for _, task := range result.Tasks {
			assert.Equal(t, models.TaskStatusPending, task.Status)
		}

		// Filter by completed status
		completedStatus := models.TaskStatusCompleted
		filter.Status = &completedStatus
		result, err = taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, 1)
		for _, task := range result.Tasks {
			assert.Equal(t, models.TaskStatusCompleted, task.Status)
		}

		// Clean up
		err = taskService.DeleteTask(ctx, pendingTask.ID)
		require.NoError(t, err)
		err = taskService.DeleteTask(ctx, completedTask.ID)
		require.NoError(t, err)
	})

	t.Run("Filtering by assignee", func(t *testing.T) {
		// Create tasks with different assignees
		user1Req := &models.CreateTaskRequest{
			Title:    "User 1 Task",
			Status:   models.TaskStatusPending,
			Assignee: "user1@example.com",
		}
		user1Task, err := taskService.CreateTask(ctx, user1Req)
		require.NoError(t, err)

		user2Req := &models.CreateTaskRequest{
			Title:    "User 2 Task",
			Status:   models.TaskStatusPending,
			Assignee: "user2@example.com",
		}
		user2Task, err := taskService.CreateTask(ctx, user2Req)
		require.NoError(t, err)

		// Filter by user1
		assignee := "user1@example.com"
		filter := &models.TaskFilter{
			Assignee: &assignee,
			Page:     1,
			PageSize: 10,
		}
		result, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, 1)
		for _, task := range result.Tasks {
			assert.Equal(t, "user1@example.com", task.Assignee)
		}

		// Clean up
		err = taskService.DeleteTask(ctx, user1Task.ID)
		require.NoError(t, err)
		err = taskService.DeleteTask(ctx, user2Task.ID)
		require.NoError(t, err)
	})

	t.Run("Combined filters", func(t *testing.T) {
		// Create tasks with specific status and assignee
		createReq := &models.CreateTaskRequest{
			Title:    "Combined Filter Task",
			Status:   models.TaskStatusInProgress,
			Assignee: "combined@example.com",
		}
		task, err := taskService.CreateTask(ctx, createReq)
		require.NoError(t, err)

		// Filter by both status and assignee
		status := models.TaskStatusInProgress
		assignee := "combined@example.com"
		filter := &models.TaskFilter{
			Status:   &status,
			Assignee: &assignee,
			Page:     1,
			PageSize: 10,
		}
		result, err := taskService.ListTasks(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, 1)
		for _, task := range result.Tasks {
			assert.Equal(t, models.TaskStatusInProgress, task.Status)
			assert.Equal(t, "combined@example.com", task.Assignee)
		}

		// Clean up
		err = taskService.DeleteTask(ctx, task.ID)
		require.NoError(t, err)
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, repo := setupTestDB(t)
	defer db.Close()

	redisCache := setupTestRedis(t)
	taskService := service.NewTaskService(repo, redisCache)

	ctx := context.Background()

	t.Run("Get non-existent task", func(t *testing.T) {
		_, err := taskService.GetTask(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Equal(t, repository.ErrTaskNotFound, err)
	})

	t.Run("Update non-existent task", func(t *testing.T) {
		newStatus := models.TaskStatusCompleted
		updateReq := &models.UpdateTaskRequest{
			Status: &newStatus,
		}
		_, err := taskService.UpdateTask(ctx, "non-existent-id", updateReq)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrTaskNotFound, err)
	})

	t.Run("Delete non-existent task", func(t *testing.T) {
		err := taskService.DeleteTask(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Equal(t, repository.ErrTaskNotFound, err)
	})

	t.Run("Create task with empty title", func(t *testing.T) {
		createReq := &models.CreateTaskRequest{
			Title:    "",
			Status:   models.TaskStatusPending,
			Assignee: "error@example.com",
		}
		_, err := taskService.CreateTask(ctx, createReq)
		assert.Error(t, err)
	})

	t.Run("Create task with invalid status", func(t *testing.T) {
		createReq := &models.CreateTaskRequest{
			Title:    "Invalid Status Task",
			Status:   "invalid_status",
			Assignee: "error@example.com",
		}
		_, err := taskService.CreateTask(ctx, createReq)
		assert.Error(t, err)
	})
}
