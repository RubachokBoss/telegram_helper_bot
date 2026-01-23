package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RubachokBoss/telegram_helper_bot/config"
	"github.com/RubachokBoss/telegram_helper_bot/internal/delivery/rest"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/database"
	"github.com/RubachokBoss/telegram_helper_bot/internal/repository/postgres"
	"github.com/RubachokBoss/telegram_helper_bot/internal/service"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключаемся к PostgreSQL для пользователей
	var db *sql.DB
	for i := 0; i < 10; i++ {
		db, err = database.NewPostgresConnection(database.Config{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			User:     cfg.Postgres.User,
			Password: cfg.Postgres.Password,
			DBName:   cfg.Postgres.DBName,
			SSLMode:  cfg.Postgres.SSLMode,
		})
		if err != nil {
			log.Printf("Attempt %d: Failed to connect to database: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}
	defer db.Close()

	log.Println("✅ Successfully connected to database")

	// Подключаемся к gRPC серверу для работы с задачами
	var grpcConn *grpc.ClientConn
	for i := 0; i < 10; i++ {
		grpcConn, err = grpc.Dial(cfg.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	defer grpcConn.Close()

	log.Println("✅ Successfully connected to gRPC server")

	// Создаем репозитории и сервисы
	userRepo := postgres.NewWebUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)

	// gRPC клиент для работы с задачами
	taskClient := pb.NewTaskServiceClient(grpcConn)

	// REST API сервер
	restServer := rest.NewServer(authService, taskClient, cfg)

	// Запускаем сервер в горутине
	go func() {
		log.Printf("🚀 Web API server starting on port %s", cfg.WebAPI.Port)
		if err := restServer.Start(); err != nil {
			log.Fatal("Failed to start REST server:", err)
		}
	}()

	// Ожидаем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down web API server...")
	restServer.Stop()
	log.Println("Server stopped")
}
