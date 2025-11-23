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
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	conn, err := grpc.Dial("localhost"+cfg.GRPC.Port, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to dial:", err)
	}
	defer conn.Close()

	client := pb.NewTaskServiceClient(conn)

	bot, err := telegram.NewBot(cfg.Telegram.Token, client)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	log.Println("Bot is now running.  Press CTRL-C to exit.")

	go bot.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down bot...")
}
