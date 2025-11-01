package cache

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		filter   *models.TaskFilter
		expected string
	}{
		{
			name:     "Nil filter",
			filter:   nil,
			expected: "tasks:list:all",
		},
		{
			name: "With status",
			filter: &models.TaskFilter{
				Status:   ptrTaskStatus(models.TaskStatusPending),
				Page:     1,
				PageSize: 10,
			},
			expected: "tasks:list:status:pending:page:1:size:10",
		},
		{
			name: "With assignee",
			filter: &models.TaskFilter{
				Assignee: ptrString("test@example.com"),
				Page:     2,
				PageSize: 20,
			},
			expected: "tasks:list:assignee:test@example.com:page:2:size:20",
		},
		{
			name: "With both",
			filter: &models.TaskFilter{
				Status:   ptrTaskStatus(models.TaskStatusCompleted),
				Assignee: ptrString("user@example.com"),
				Page:     1,
				PageSize: 10,
			},
			expected: "tasks:list:status:completed:assignee:user@example.com:page:1:size:10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCacheKey(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func ptrTaskStatus(s models.TaskStatus) *models.TaskStatus {
	return &s
}

func ptrString(s string) *string {
	return &s
}

// Mock Redis client test
func TestRedisCache_MockOperations(t *testing.T) {
	// These tests would require a Redis instance or mock
	// For now, we just test the cache key generation logic
	t.Run("Cache key generation", func(t *testing.T) {
		filter := &models.TaskFilter{
			Page:     1,
			PageSize: 10,
		}
		key := GenerateCacheKey(filter)
		assert.NotEmpty(t, key)
	})
}

func TestRedisCache_GetTask(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewRedisCache(db)
	ctx := context.Background()

	t.Run("Cache hit", func(t *testing.T) {
		task := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)
		taskData, _ := json.Marshal(task)

		mock.ExpectGet("task:" + task.ID).SetVal(string(taskData))

		result, err := cache.GetTask(ctx, task.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.ID, result.ID)
		assert.Equal(t, task.Title, result.Title)
	})

	t.Run("Cache miss", func(t *testing.T) {
		mock.ExpectGet("task:nonexistent").RedisNil()

		result, err := cache.GetTask(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Redis error", func(t *testing.T) {
		mock.ExpectGet("task:error").SetErr(assert.AnError)

		result, err := cache.GetTask(ctx, "error")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRedisCache_SetTask(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewRedisCache(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		task := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)
		taskData, _ := json.Marshal(task)

		mock.ExpectSet("task:"+task.ID, taskData, cacheTTL).SetVal("OK")

		err := cache.SetTask(ctx, task)
		assert.NoError(t, err)
	})

	t.Run("Redis error", func(t *testing.T) {
		task := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)
		taskData, _ := json.Marshal(task)

		mock.ExpectSet("task:"+task.ID, taskData, cacheTTL).SetErr(assert.AnError)

		err := cache.SetTask(ctx, task)
		assert.Error(t, err)
	})
}

func TestRedisCache_DeleteTask(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewRedisCache(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		taskID := "test-id"
		mock.ExpectDel("task:" + taskID).SetVal(1)

		err := cache.DeleteTask(ctx, taskID)
		assert.NoError(t, err)
	})

	t.Run("Redis error", func(t *testing.T) {
		taskID := "error-id"
		mock.ExpectDel("task:" + taskID).SetErr(assert.AnError)

		err := cache.DeleteTask(ctx, taskID)
		assert.Error(t, err)
	})
}

func TestRedisCache_GetTaskList(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewRedisCache(db)
	ctx := context.Background()

	t.Run("Cache hit", func(t *testing.T) {
		tasks := []models.Task{
			*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
			*models.NewTask("Task 2", "Desc 2", "user2@example.com", models.TaskStatusCompleted),
		}
		tasksData, _ := json.Marshal(tasks)
		cacheKey := "tasks:list:all"

		mock.ExpectGet(cacheKey).SetVal(string(tasksData))

		result, err := cache.GetTaskList(ctx, cacheKey)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
	})

	t.Run("Cache miss", func(t *testing.T) {
		cacheKey := "tasks:list:empty"
		mock.ExpectGet(cacheKey).RedisNil()

		result, err := cache.GetTaskList(ctx, cacheKey)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Redis error", func(t *testing.T) {
		cacheKey := "tasks:list:error"
		mock.ExpectGet(cacheKey).SetErr(assert.AnError)

		result, err := cache.GetTaskList(ctx, cacheKey)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRedisCache_SetTaskList(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewRedisCache(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		tasks := []models.Task{
			*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
		}
		tasksData, _ := json.Marshal(tasks)
		cacheKey := "tasks:list:test"

		mock.ExpectSet(cacheKey, tasksData, cacheTTL).SetVal("OK")

		err := cache.SetTaskList(ctx, cacheKey, tasks)
		assert.NoError(t, err)
	})

	t.Run("Redis error", func(t *testing.T) {
		tasks := []models.Task{
			*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
		}
		tasksData, _ := json.Marshal(tasks)
		cacheKey := "tasks:list:error"

		mock.ExpectSet(cacheKey, tasksData, cacheTTL).SetErr(assert.AnError)

		err := cache.SetTaskList(ctx, cacheKey, tasks)
		assert.Error(t, err)
	})
}

func TestRedisCache_InvalidateTaskList(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewRedisCache(db)
	ctx := context.Background()

	t.Run("Success with keys", func(t *testing.T) {
		keys := []string{"tasks:list:1", "tasks:list:2"}

		mock.ExpectScan(0, "tasks:list*", 0).SetVal(keys, 0)
		mock.ExpectDel(keys[0]).SetVal(1)
		mock.ExpectDel(keys[1]).SetVal(1)

		err := cache.InvalidateTaskList(ctx)
		assert.NoError(t, err)
	})

	t.Run("Success with no keys", func(t *testing.T) {
		mock.ExpectScan(0, "tasks:list*", 0).SetVal([]string{}, 0)

		err := cache.InvalidateTaskList(ctx)
		assert.NoError(t, err)
	})
}

func TestNewRedisCache(t *testing.T) {
	db, _ := redismock.NewClientMock()
	cache := NewRedisCache(db)

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.client)
}
