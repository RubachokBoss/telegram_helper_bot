package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/database"
	"github.com/redis/go-redis/v9"
)

type taskCacheRepository struct {
	client *redis.Client
	base   domain.TaskRepository // Базовый репозиторий (PostgreSQL)
}

// NewTaskCacheRepository создает новый кэш-репозиторий
func NewTaskCacheRepository(client *redis.Client, base domain.TaskRepository) domain.TaskRepository {
	return &taskCacheRepository{
		client: client,
		base:   base,
	}
}

func (r *taskCacheRepository) Create(task *domain.Task) error {
	// Создаем задачу в БД
	err := r.base.Create(task)
	if err != nil {
		return err
	}

	// Пытаемся обновить кэш задач владельца вместо полной инвалидации
	ctx := context.Background()
	err = r.updateOwnerCacheWithNewTask(ctx, task)
	if err != nil {
		log.Printf("⚠️ Failed to update cache for owner %s: %v, falling back to invalidation", task.OwnerID, err)
		// Fallback: инвалидируем кэш если не удалось обновить
		r.invalidateOwnerCache(ctx, task.OwnerID)
	}

	log.Printf("✅ Task created and owner cache updated for owner: %s", task.OwnerID)
	return nil
}

func (r *taskCacheRepository) FindByID(id string) (*domain.Task, error) {
	ctx := context.Background()
	cacheKey := database.TaskKey(id)

	// Проверяем кэш
	if cached, err := r.getCachedTask(ctx, cacheKey); err == nil && cached != nil {
		log.Printf("✅ Task found in cache: %s", id)
		return cached, nil
	}

	// Если нет в кэше, получаем из БД
	task, err := r.base.FindByID(id)
	if err != nil || task == nil {
		return task, err
	}

	// Кэшируем результат
	r.cacheTask(ctx, cacheKey, task, database.TaskTTL)
	log.Printf("✅ Task cached: %s", id)

	return task, nil
}

func (r *taskCacheRepository) FindByUserID(userID string) ([]*domain.Task, error) {
	ctx := context.Background()
	cacheKey := database.UserTasksKey(userID)

	// Проверяем кэш
	if cached, err := r.getCachedTasks(ctx, cacheKey); err == nil && cached != nil {
		log.Printf("✅ User tasks found in cache for user: %s", userID)
		return cached, nil
	}

	// Если нет в кэше, получаем из БД
	tasks, err := r.base.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Кэшируем результат
	r.cacheTasks(ctx, cacheKey, tasks, database.UserTasksTTL)
	log.Printf("✅ User tasks cached for user: %s (%d tasks)", userID, len(tasks))

	return tasks, nil
}

func (r *taskCacheRepository) FindByOwnerID(ownerID string) ([]*domain.Task, error) {
	ctx := context.Background()
	cacheKey := database.OwnerTasksKey(ownerID)

	// Проверяем кэш
	if cached, err := r.getCachedTasks(ctx, cacheKey); err == nil && cached != nil {
		log.Printf("✅ Owner tasks found in cache for owner: %s", ownerID)
		return cached, nil
	}

	// Если нет в кэше, получаем из БД
	tasks, err := r.base.FindByOwnerID(ownerID)
	if err != nil {
		return nil, err
	}

	// Кэшируем результат
	r.cacheTasks(ctx, cacheKey, tasks, database.OwnerTasksTTL)
	log.Printf("✅ Owner tasks cached for owner: %s (%d tasks)", ownerID, len(tasks))

	return tasks, nil
}

func (r *taskCacheRepository) Update(task *domain.Task) error {
	// Обновляем в БД
	err := r.base.Update(task)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Инвалидируем кэш для конкретной задачи
	taskKey := database.TaskKey(task.ID)
	r.client.Del(ctx, taskKey)

	// Инвалидируем кэш для владельца и исполнителя
	r.invalidateOwnerCache(ctx, task.OwnerID)
	if task.AssignedID != "" {
		r.invalidateUserCache(ctx, task.AssignedID)
	}

	log.Printf("✅ Task updated and cache invalidated: %s", task.ID)
	return nil
}

func (r *taskCacheRepository) Delete(task *domain.Task) error {
	// Удаляем из БД
	err := r.base.Delete(task)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Инвалидируем весь кэш связанный с задачей
	taskKey := database.TaskKey(task.ID)
	r.client.Del(ctx, taskKey)

	r.invalidateOwnerCache(ctx, task.OwnerID)
	if task.AssignedID != "" {
		r.invalidateUserCache(ctx, task.AssignedID)
	}

	log.Printf("✅ Task deleted and cache invalidated: %s", task.ID)
	return nil
}

// Helper methods for caching

func (r *taskCacheRepository) cacheTask(ctx context.Context, key string, task *domain.Task, ttl time.Duration) {
	if task == nil {
		return
	}

	data, err := json.Marshal(task)
	if err != nil {
		log.Printf("❌ Failed to marshal task for cache: %v", err)
		return
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		log.Printf("❌ Failed to cache task: %v", err)
	}
}

func (r *taskCacheRepository) cacheTasks(ctx context.Context, key string, tasks []*domain.Task, ttl time.Duration) {
	data, err := json.Marshal(tasks)
	if err != nil {
		log.Printf("❌ Failed to marshal tasks for cache: %v", err)
		return
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		log.Printf("❌ Failed to cache tasks: %v", err)
	}
}

func (r *taskCacheRepository) getCachedTask(ctx context.Context, key string) (*domain.Task, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var task domain.Task
	err = json.Unmarshal([]byte(data), &task)
	if err != nil {
		log.Printf("❌ Failed to unmarshal cached task: %v", err)
		return nil, err
	}

	return &task, nil
}

func (r *taskCacheRepository) getCachedTasks(ctx context.Context, key string) ([]*domain.Task, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tasks []*domain.Task
	err = json.Unmarshal([]byte(data), &tasks)
	if err != nil {
		log.Printf("❌ Failed to unmarshal cached tasks: %v", err)
		return nil, err
	}

	return tasks, nil
}

func (r *taskCacheRepository) invalidateUserCache(ctx context.Context, userID string) {
	key := database.UserTasksKey(userID)
	r.client.Del(ctx, key)
	log.Printf("🗑️ Invalidated user cache: %s", key)
}

// updateOwnerCacheWithNewTask пытается добавить новую задачу к существующему кэшу владельца
func (r *taskCacheRepository) updateOwnerCacheWithNewTask(ctx context.Context, task *domain.Task) error {
	cacheKey := database.OwnerTasksKey(task.OwnerID)

	// Проверяем, существует ли кэш
	exists, err := r.client.Exists(ctx, cacheKey).Result()
	if err != nil {
		return err
	}

	// Если кэша нет, ничего не делаем - он обновится при следующем запросе
	if exists == 0 {
		log.Printf("ℹ️ Owner cache doesn't exist for %s, skipping update", task.OwnerID)
		return nil
	}

	// Получаем существующие задачи из кэша
	cachedTasks, err := r.getCachedTasks(ctx, cacheKey)
	if err != nil {
		return err
	}

	if cachedTasks == nil {
		// Кэш существует, но пустой - просто инвалидируем
		r.invalidateOwnerCache(ctx, task.OwnerID)
		return nil
	}

	// Добавляем новую задачу в начало массива (задачи сортируются по created_at DESC)
	updatedTasks := append([]*domain.Task{task}, cachedTasks...)

	// Сохраняем обновленный массив в кэш
	r.cacheTasks(ctx, cacheKey, updatedTasks, database.OwnerTasksTTL)

	log.Printf("✅ Added new task to owner cache for %s, total tasks: %d", task.OwnerID, len(updatedTasks))
	return nil
}

func (r *taskCacheRepository) invalidateOwnerCache(ctx context.Context, ownerID string) {
	key := database.OwnerTasksKey(ownerID)
	r.client.Del(ctx, key)
	log.Printf("🗑️ Invalidated owner cache: %s", key)
}
