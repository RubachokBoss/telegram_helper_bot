package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedisClient(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем подключение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("✅ Connected to Redis successfully!")
	return client, nil
}

// Cache helper functions
func SetJSON(ctx context.Context, client *redis.Client, key string, value interface{}, ttl time.Duration) error {
	return client.Set(ctx, key, value, ttl).Err()
}

func GetJSON(ctx context.Context, client *redis.Client, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func DeleteKeys(ctx context.Context, client *redis.Client, pattern string) error {
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}
	return nil
}

// Cache key generators
func UserTasksKey(userID string) string {
	return fmt.Sprintf("user_tasks:%s", userID)
}

func OwnerTasksKey(ownerID string) string {
	return fmt.Sprintf("owner_tasks:%s", ownerID)
}

func TaskKey(taskID string) string {
	return fmt.Sprintf("task:%s", taskID)
}

// Cache TTL constants
const (
	UserTasksTTL  = 10 * time.Minute // Списки задач пользователя
	OwnerTasksTTL = 10 * time.Minute // Списки задач владельца
	TaskTTL       = 30 * time.Minute // Индивидуальные задачи
)
