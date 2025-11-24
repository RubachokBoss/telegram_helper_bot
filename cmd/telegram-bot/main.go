package main

import (
	"github.com/RubachokBoss/telegram_helper_bot/config"
	"github.com/RubachokBoss/telegram_helper_bot/internal/delivery/telegram"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
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

	client := pb.NewTaskServiceClient(conn)

	bot, err := telegram.NewBot(cfg.Telegram.Token, client)
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
