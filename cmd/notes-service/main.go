package main

import (
	"database/sql"
	"github.com/RubachokBoss/telegram_helper_bot/config"
	"github.com/RubachokBoss/telegram_helper_bot/internal/delivery/grpc"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/database"
	"github.com/RubachokBoss/telegram_helper_bot/internal/repository/postgres"
	"github.com/RubachokBoss/telegram_helper_bot/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Ждем пока база данных будет готова
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

	taskRepo := postgres.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	grpcServer := grpc.NewServer(taskService)

	go func() {
		if err := grpcServer.Start(cfg.GRPC.Port); err != nil {
			log.Fatal("Failed to start gRPC server:", err)
		}
	}()

	log.Println("✅ gRPC server started on", cfg.GRPC.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	grpcServer.Stop()
	log.Println("Server stopped")
}
