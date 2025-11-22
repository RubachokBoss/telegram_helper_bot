package telegram

import (
	"context"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	bot    *tgbotapi.BotAPI
	client pb.TaskServiceClient
}

func NewBot(token string, client pb.TaskServiceClient) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return &Bot{
		bot:    bot,
		client: client,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(u)

	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			b.HandlerCommands(update.Message)
		}
	}
}

func (b *Bot) HandlerCommands(message *tgbotapi.Message) {
	ctx := context.Background()

	userId := message.From.UserName

	switch message.Command() {

	case "tasks":
		b.sendMessage(message.Chat.ID, "Доступные команды:\n/new - создать задачу\n/my - мои задачи\n/owner - задачи, созданные мной")
	case "new":
		text := strings.TrimSpace(msg.CommandArguments())
		if text == "" {
			b.sendMessage(message.Chat.ID, "Использоване: /new <текст задачи>")
			return
		}
		b.CreateTask(ctx, message.Chat.ID, text, userId)
	case "my":
		b.GetUserTasks(ctx, message.Chat.ID, userId)
	case "owner":
		b.GetOwnerTasks(ctx, message.Chat.ID, userId)
	default:
		command := message.Command()
		switch command {
		case strings.HasPrefix(command, "assign_"):
			taskID := strings.TrimPrefix(command, "assign_")
			b.assignTask(ctx, message.Chat.ID, taskID)
		case strings.HasPrefix(command, "unassign_"):
			taskID := strings.TrimPrefix(command, "unassign_")
			b.unassignTask(ctx, message.Chat.ID, taskID)

		case strings.HasPrefix(command, "resolve_"):
			taskID := strings.TrimPrefix(command, "resolve_")
			b.resolveTask(ctx, message.Chat.ID, taskID)

		default:
			b.sendMessage(message.Chat.ID, "Неизвестная команда")
		}
	}

}
