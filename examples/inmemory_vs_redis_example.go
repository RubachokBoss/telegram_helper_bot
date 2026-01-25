package examples

// Этот файл демонстрирует проблему с обычным in-memory store
// в микросервисной архитектуре

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================
// ❌ ПРОБЛЕМА: Обычный In-Memory Store
// ============================================

type InMemoryCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]interface{}),
	}
}

func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data[key]
	return value, ok
}

func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// ПРОБЛЕМА 1: Изолированное состояние
// Каждый сервис имеет свой кеш
func ExampleIsolatedState() {
	// Telegram Bot (контейнер 1)
	telegramBotCache := NewInMemoryCache()
	telegramBotCache.Set("task:123", "Task from Telegram Bot")

	// Notes Service (контейнер 2)
	notesServiceCache := NewInMemoryCache()
	// ❌ Notes Service НЕ видит кеш Telegram Bot!
	value, ok := notesServiceCache.Get("task:123")
	fmt.Printf("Notes Service found task: %v (ok=%v)\n", value, ok)
	// Output: Notes Service found task: <nil> (ok=false)
	// ❌ Кеш не найден, хотя Telegram Bot его кешировал!
}

// ПРОБЛЕМА 2: Потеря данных при перезапуске
func ExampleDataLoss() {
	cache := NewInMemoryCache()
	cache.Set("task:1", "Task 1")
	cache.Set("task:2", "Task 2")
	cache.Set("task:3", "Task 3")

	fmt.Printf("Cache before restart: %d items\n", len(cache.data))
	// Output: Cache before restart: 3 items

	// Сервис перезапускается...
	// ❌ Все данные теряются!
	cache = NewInMemoryCache() // Новый экземпляр

	fmt.Printf("Cache after restart: %d items\n", len(cache.data))
	// Output: Cache after restart: 0 items
	// ❌ Все данные потеряны!
}

// ПРОБЛЕМА 3: Нет синхронизации при масштабировании
func ExampleNoSynchronization() {
	// Запускаем 3 инстанса Notes Service
	instance1Cache := NewInMemoryCache()
	instance2Cache := NewInMemoryCache()
	instance3Cache := NewInMemoryCache()

	// Запрос 1 → Instance 1
	instance1Cache.Set("task:123", "Task 123")
	fmt.Println("Instance 1 cached task:123")

	// Запрос 2 → Instance 2
	value, ok := instance2Cache.Get("task:123")
	fmt.Printf("Instance 2 found task:123: %v (ok=%v)\n", value, ok)
	// Output: Instance 2 found task:123: <nil> (ok=false)
	// ❌ Instance 2 НЕ видит кеш Instance 1!

	// Запрос 3 → Instance 3
	value, ok = instance3Cache.Get("task:123")
	fmt.Printf("Instance 3 found task:123: %v (ok=%v)\n", value, ok)
	// Output: Instance 3 found task:123: <nil> (ok=false)
	// ❌ Instance 3 НЕ видит кеш Instance 1!

	// Результат: кеш-хит = 33% (плохо!)
}

// ============================================
// ✅ РЕШЕНИЕ: Redis
// ============================================

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// РЕШЕНИЕ 1: Разделяемое состояние
func ExampleSharedState(redisClient *redis.Client) {
	ctx := context.Background()

	// Telegram Bot (контейнер 1)
	telegramBotCache := NewRedisCache(redisClient)
	telegramBotCache.Set(ctx, "task:123", "Task from Telegram Bot", 10*time.Minute)

	// Notes Service (контейнер 2)
	notesServiceCache := NewRedisCache(redisClient)
	// ✅ Notes Service видит тот же Redis!
	value, err := notesServiceCache.Get(ctx, "task:123")
	if err == nil {
		fmt.Printf("Notes Service found task: %s\n", value)
		// Output: Notes Service found task: Task from Telegram Bot
		// ✅ Кеш найден!
	}
}

// РЕШЕНИЕ 2: Персистентность
func ExamplePersistence(redisClient *redis.Client) {
	ctx := context.Background()
	cache := NewRedisCache(redisClient)

	// Кешируем данные
	cache.Set(ctx, "task:1", "Task 1", 10*time.Minute)
	cache.Set(ctx, "task:2", "Task 2", 10*time.Minute)
	cache.Set(ctx, "task:3", "Task 3", 10*time.Minute)

	fmt.Println("Cache before restart: 3 items cached in Redis")
	// Redis сохраняет данные на диск (RDB, AOF)

	// Сервис перезапускается...
	// ✅ Redis продолжает работать!
	// ✅ Данные НЕ теряются!

	// Новый инстанс подключается к тому же Redis
	newCache := NewRedisCache(redisClient)
	value, err := newCache.Get(ctx, "task:1")
	if err == nil {
		fmt.Printf("Cache after restart: found task:1 = %s\n", value)
		// Output: Cache after restart: found task:1 = Task 1
		// ✅ Данные сохранены!
	}
}

// РЕШЕНИЕ 3: Синхронизация при масштабировании
func ExampleSynchronization(redisClient *redis.Client) {
	ctx := context.Background()

	// Запускаем 3 инстанса Notes Service
	instance1Cache := NewRedisCache(redisClient)
	instance2Cache := NewRedisCache(redisClient)
	instance3Cache := NewRedisCache(redisClient)

	// Запрос 1 → Instance 1
	instance1Cache.Set(ctx, "task:123", "Task 123", 10*time.Minute)
	fmt.Println("Instance 1 cached task:123 in Redis")

	// Запрос 2 → Instance 2
	value, err := instance2Cache.Get(ctx, "task:123")
	if err == nil {
		fmt.Printf("Instance 2 found task:123: %s\n", value)
		// Output: Instance 2 found task:123: Task 123
		// ✅ Instance 2 видит кеш Instance 1!
	}

	// Запрос 3 → Instance 3
	value, err = instance3Cache.Get(ctx, "task:123")
	if err == nil {
		fmt.Printf("Instance 3 found task:123: %s\n", value)
		// Output: Instance 3 found task:123: Task 123
		// ✅ Instance 3 видит кеш Instance 1!
	}

	// Результат: кеш-хит = 100% (отлично!)
}

// ============================================
// ПРАКТИЧЕСКИЙ ПРИМЕР ИЗ ВАШЕГО ПРОЕКТА
// ============================================

// Сценарий: Пользователь создает задачу через Telegram Bot,
// затем запрашивает список задач через Web API

// ❌ С обычным in-memory store:
func ExampleWithInMemoryStore() {
	// Telegram Bot (контейнер 1)
	telegramBotCache := NewInMemoryCache()
	telegramBotCache.Set("user_tasks:user123", []string{"task:1", "task:2"})

	// Notes Service (контейнер 2)
	notesServiceCache := NewInMemoryCache()
	// ❌ Notes Service НЕ видит кеш Telegram Bot!

	// Web API запрашивает задачи
	// → Notes Service НЕ находит в своем кеше
	// → Идет в PostgreSQL (медленно) 🐌
	fmt.Println("❌ Cache miss! Going to PostgreSQL...")
}

// ✅ С Redis:
func ExampleWithRedis(redisClient *redis.Client) {
	ctx := context.Background()

	// Telegram Bot (контейнер 1)
	telegramBotCache := NewRedisCache(redisClient)
	telegramBotCache.Set(ctx, "user_tasks:user123", "task:1,task:2", 10*time.Minute)

	// Notes Service (контейнер 2)
	notesServiceCache := NewRedisCache(redisClient)
	// ✅ Notes Service видит тот же Redis!

	// Web API запрашивает задачи
	value, err := notesServiceCache.Get(ctx, "user_tasks:user123")
	if err == nil {
		fmt.Printf("✅ Cache hit! Tasks: %s\n", value)
		// ✅ Быстрый ответ из кеша!
	}
}

// ============================================
// ДОПОЛНИТЕЛЬНЫЕ ПРЕИМУЩЕСТВА REDIS
// ============================================

// 1. TTL (автоматическое истечение)
func ExampleTTL(redisClient *redis.Client) {
	ctx := context.Background()
	cache := NewRedisCache(redisClient)

	// Кеш автоматически истечет через 10 минут
	cache.Set(ctx, "task:123", "Task 123", 10*time.Minute)

	// Через 10 минут ключ автоматически удалится
	// ✅ Не нужно самому управлять TTL!
}

// 2. Pipeline (батчинг операций)
func ExamplePipeline(redisClient *redis.Client) {
	ctx := context.Background()

	// ❌ Без Pipeline: 3 отдельных запроса
	redisClient.Set(ctx, "key1", "value1", 0)
	redisClient.Set(ctx, "key2", "value2", 0)
	redisClient.Set(ctx, "key3", "value3", 0)
	// = 3 сетевых round-trips

	// ✅ С Pipeline: 1 запрос
	pipe := redisClient.Pipeline()
	pipe.Set(ctx, "key1", "value1", 0)
	pipe.Set(ctx, "key2", "value2", 0)
	pipe.Set(ctx, "key3", "value3", 0)
	pipe.Exec(ctx)
	// = 1 сетевой round-trip (в 3 раза быстрее!)
}

// 3. Атомарные операции
func ExampleAtomicOperations(redisClient *redis.Client) {
	ctx := context.Background()

	// ✅ Атомарный счетчик
	count, _ := redisClient.Incr(ctx, "task_count").Result()
	fmt.Printf("Task count: %d\n", count)

	// ✅ Атомарное обновление
	redisClient.HSet(ctx, "task:123", "status", "completed")
	// Гарантированно атомарно!
}

// 4. Продвинутые структуры данных
func ExampleAdvancedStructures(redisClient *redis.Client) {
	ctx := context.Background()

	// ✅ Hash (для объектов)
	redisClient.HSet(ctx, "task:123",
		"id", "123",
		"text", "Do something",
		"owner_id", "user456",
	)

	// ✅ Sorted Set (для списков с сортировкой)
	redisClient.ZAdd(ctx, "user_tasks:user456", redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: "task:123",
	})

	// ✅ List (для очередей)
	redisClient.LPush(ctx, "task_queue", "task:123")

	// ✅ Set (для множеств)
	redisClient.SAdd(ctx, "active_users", "user456")
}

