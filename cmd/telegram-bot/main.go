package main

import (
	"github.com/RubachokBoss/telegram_helper_bot/config"
	"github.com/RubachokBoss/telegram_helper_bot/internal/delivery/telegram"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/cache"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/database"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Ждем пока gRPC сервер будет готов
	var conn *grpc.ClientConn
	for i := 0; i < 10; i++ {
		conn, err = grpc.Dial(cfg.GRPC.Port, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Printf("Attempt %d: Failed to connect to gRPC server: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		log.Fatal("Failed to connect to gRPC server after retries:", err)
	}
	defer conn.Close()

	log.Println("✅ Successfully connected to gRPC server")

	// Подключаемся к Redis для общего кеша
	var redisClient *redis.Client
	for i := 0; i < 10; i++ {
		redisClient, err = database.NewRedisClient(database.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err != nil {
			log.Printf("Attempt %d: Failed to connect to Redis: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		log.Printf("⚠️ Failed to connect to Redis after retries: %v (continuing without cache)", err)
		redisClient = nil // Продолжаем без кеша
	} else {
		defer redisClient.Close()
		log.Println("✅ Successfully connected to Redis (shared cache)")
	}

	client := pb.NewTaskServiceClient(conn)

	// Создаем кеш-клиент для общего кеша
	cacheClient := cache.NewTaskCacheClient(redisClient)
	if cacheClient != nil && cacheClient.IsAvailable() {
		log.Println("✅ Shared cache client initialized")
	} else {
		log.Println("⚠️ Shared cache unavailable, using gRPC only")
	}

	bot, err := telegram.NewBot(cfg.Telegram.Token, client, redisClient, cacheClient)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	log.Println("✅ Bot created successfully")
	log.Println("Bot is now running. Press CTRL-C to exit.")

	go bot.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down bot...")
}
