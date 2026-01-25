package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/database"
	"github.com/redis/go-redis/v9"
)

// TaskCacheClient - клиент для работы с общим кешем задач в Redis
// Используется всеми микросервисами для чтения из общего кеша
type TaskCacheClient struct {
	client *redis.Client
}

// NewTaskCacheClient создает новый клиент для работы с кешем задач
func NewTaskCacheClient(client *redis.Client) *TaskCacheClient {
	if client == nil {
		return nil // Возвращаем nil, если Redis недоступен
	}
	return &TaskCacheClient{client: client}
}

// GetTask получает задачу из кеша по ID
func (c *TaskCacheClient) GetTask(ctx context.Context, taskID string) (*domain.Task, error) {
	if c == nil || c.client == nil {
		return nil, redis.Nil // Кеш недоступен
	}

	key := database.TaskKey(taskID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var task domain.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		log.Printf("❌ Failed to unmarshal cached task: %v", err)
		return nil, err
	}

	return &task, nil
}

// GetUserTasks получает список задач пользователя из кеша
func (c *TaskCacheClient) GetUserTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	if c == nil || c.client == nil {
		return nil, redis.Nil // Кеш недоступен
	}

	key := database.UserTasksKey(userID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tasks []*domain.Task
	if err := json.Unmarshal([]byte(data), &tasks); err != nil {
		log.Printf("❌ Failed to unmarshal cached tasks: %v", err)
		return nil, err
	}

	return tasks, nil
}

// GetOwnerTasks получает список задач владельца из кеша
func (c *TaskCacheClient) GetOwnerTasks(ctx context.Context, ownerID string) ([]*domain.Task, error) {
	if c == nil || c.client == nil {
		return nil, redis.Nil // Кеш недоступен
	}

	key := database.OwnerTasksKey(ownerID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tasks []*domain.Task
	if err := json.Unmarshal([]byte(data), &tasks); err != nil {
		log.Printf("❌ Failed to unmarshal cached tasks: %v", err)
		return nil, err
	}

	return tasks, nil
}

// InvalidateTask инвалидирует кеш задачи (используется при обновлении/удалении)
func (c *TaskCacheClient) InvalidateTask(ctx context.Context, taskID string) error {
	if c == nil || c.client == nil {
		return nil // Кеш недоступен, игнорируем
	}

	key := database.TaskKey(taskID)
	return c.client.Del(ctx, key).Err()
}

// InvalidateUserCache инвалидирует кеш задач пользователя
func (c *TaskCacheClient) InvalidateUserCache(ctx context.Context, userID string) error {
	if c == nil || c.client == nil {
		return nil // Кеш недоступен, игнорируем
	}

	key := database.UserTasksKey(userID)
	return c.client.Del(ctx, key).Err()
}

// InvalidateOwnerCache инвалидирует кеш задач владельца
func (c *TaskCacheClient) InvalidateOwnerCache(ctx context.Context, ownerID string) error {
	if c == nil || c.client == nil {
		return nil // Кеш недоступен, игнорируем
	}

	key := database.OwnerTasksKey(ownerID)
	return c.client.Del(ctx, key).Err()
}

// IsAvailable проверяет доступность кеша
func (c *TaskCacheClient) IsAvailable() bool {
	return c != nil && c.client != nil
}

// Ping проверяет подключение к Redis
func (c *TaskCacheClient) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("cache client is not available")
	}
	return c.client.Ping(ctx).Err()
}

