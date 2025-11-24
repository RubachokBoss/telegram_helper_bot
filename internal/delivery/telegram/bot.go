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
		text := strings.TrimSpace(message.CommandArguments())
		if text == "" {
			b.sendMessage(message.Chat.ID, "Использоване: /new <текст задачи>")
			return
		}
		b.createTask(ctx, message.Chat.ID, text, userId)
	case "my":
		b.getUserTasks(ctx, message.Chat.ID, userId)
	case "owner":
		b.getOwnerTasks(ctx, message.Chat.ID, userId)
	default:
		command := message.Command()
		if strings.HasPrefix(command, "assign_") {
			taskID := strings.TrimPrefix(command, "assign_")
			b.assigntask(ctx, message.Chat.ID, userId, taskID)
		} else if strings.HasPrefix(command, "unassign_") {
			taskID := strings.TrimPrefix(command, "unassign_")
			b.unassigntask(ctx, message.Chat.ID, taskID)
		} else if strings.HasPrefix(command, "resolve_") {
			taskID := strings.TrimPrefix(command, "resolve_")
			b.resolvetask(ctx, message.Chat.ID, taskID)
		} else {
			b.sendMessage(message.Chat.ID, "Неизвестная команда")
		}
	}

}
