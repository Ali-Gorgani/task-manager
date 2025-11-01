package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	taskCachePrefix = "task:"
	taskListKey     = "tasks:list"
	cacheTTL        = 5 * time.Minute
)

// RedisCache implements a Redis-based cache for tasks
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// GetTask retrieves a task from cache
func (c *RedisCache) GetTask(ctx context.Context, id string) (*models.Task, error) {
	key := taskCachePrefix + id
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// SetTask stores a task in cache
func (c *RedisCache) SetTask(ctx context.Context, task *models.Task) error {
	key := taskCachePrefix + task.ID
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := c.client.Set(ctx, key, data, cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// DeleteTask removes a task from cache
func (c *RedisCache) DeleteTask(ctx context.Context, id string) error {
	key := taskCachePrefix + id
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// GetTaskList retrieves task list from cache
func (c *RedisCache) GetTaskList(ctx context.Context, cacheKey string) ([]models.Task, error) {
	data, err := c.client.Get(ctx, cacheKey).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list from cache: %w", err)
	}

	var tasks []models.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return tasks, nil
}

// SetTaskList stores task list in cache
func (c *RedisCache) SetTaskList(ctx context.Context, cacheKey string, tasks []models.Task) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := c.client.Set(ctx, cacheKey, data, cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set list cache: %w", err)
	}

	return nil
}

// InvalidateTaskList invalidates all task list caches
func (c *RedisCache) InvalidateTaskList(ctx context.Context) error {
	// Delete all keys matching the pattern
	iter := c.client.Scan(ctx, 0, "tasks:list*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to iterate keys: %w", err)
	}

	return nil
}

// GenerateCacheKey generates a cache key for task list with filters
func GenerateCacheKey(filter *models.TaskFilter) string {
	key := taskListKey
	if filter == nil {
		return key + ":all"
	}

	if filter.Status != nil {
		key += fmt.Sprintf(":status:%s", *filter.Status)
	}
	if filter.Assignee != nil {
		key += fmt.Sprintf(":assignee:%s", *filter.Assignee)
	}
	key += fmt.Sprintf(":page:%d:size:%d", filter.Page, filter.PageSize)

	return key
}
