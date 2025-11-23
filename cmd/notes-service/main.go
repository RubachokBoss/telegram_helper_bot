package notes_service

import (
	"github.com/RubachokBoss/telegram_helper_bot/config"
	"github.com/RubachokBoss/telegram_helper_bot/internal/delivery/grpc"
	"github.com/RubachokBoss/telegram_helper_bot/internal/pkg/database"
	"github.com/RubachokBoss/telegram_helper_bot/internal/repository/postgres"
	"github.com/RubachokBoss/telegram_helper_bot/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgresConnection(database.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DBName:   cfg.Postgres.DBName,
		SSLMode:  cfg.Postgres.SSLMode,
	})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	taskRepo := postgres.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	grpcServer := grpc.NewServer(taskService)

	go func() {
		if err := grpcServer.Start(cfg.GRPC.Port); err != nil {
			log.Fatal(err)
		}

	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	grpcServer.Stop()
	log.Println("Server stopped")
}
