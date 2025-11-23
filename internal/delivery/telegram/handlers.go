package telegram

import (
	"context"
	"fmt"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func (b *Bot) createTask(ctx context.Context, chatID int64, text, userID string) {
	task, err := b.client.CreateTask(ctx, &pb.CreateTaskRequest{
		Text:    text,
		OwnerId: userID,
	})
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to create task: %s", err))
		return
	}
	msg := fmt.Sprintf("Задача создана!\nID: %s\nТекст: %s", task.Task.Id, task.Task.Text)
	b.sendMessage(chatID, msg)
}

func (b *Bot) assigntask(ctx context.Context, chatID int64, userID, taskID string) {
	task, err := b.client.AssignTask(ctx, &pb.AssignTaskRequest{
		TaskId: taskID,
		UserId: userID,
	})
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to assign task: %s", err))
		return
	}
	msg := fmt.Sprintf("Задача назначена! \nID: %s\nИсполнитель: %s", task.Task.Id, task.Task.AssignedId)
	b.sendMessage(chatID, msg)
}
func (b *Bot) unassigntask(ctx context.Context, chatID int64, userID, taskID string) {
	task, err := b.client.UnassignTask(ctx, &pb.UnassignTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to unassign task: %s", err))
		return
	}
	msg := fmt.Sprintf("Задача снята!\nID: %s", task.Task.Id)
	b.sendMessage(chatID, msg)
}
func (b *Bot) resolvetask(ctx context.Context, chatID int64, userID, taskID string) {
	task, err := b.client.ResolveTask(ctx, &pb.ResolveTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to resolve task: %s", err))
		return
	}
	if task.Success {
		b.sendMessage(chatID, "Задача выполнена и удалена")
	} else {
		b.sendMessage(chatID, "Ошибка выполнения задачи")
	}
}
func (b *Bot) getUserTasks(ctx context.Context, chatID int64, userID string) {
	tasks, err := b.client.GetUserTasks(ctx, &pb.GetUserTasksRequest{
		UserId: userID,
	})
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to get tasks: %s", err))
		return
	}
	b.showTasks(chatID, tasks.Tasks, "Ваши Задачи")
}
func (b *Bot) getOwnerTasks(ctx context.Context, chatID int64, userId string) {
	tasks, err := b.client.GetOwnerTasks(ctx, &pb.GetOwnerTasksRequest{
		OwnerId: userId,
	})
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to get tasks: %s", err))
		return
	}
	b.showTasks(chatID, tasks.Tasks, "Задачи созданные вами")
}

func (b *Bot) showTasks(chatID int64, tasks []*pb.Task, header string) {
	if len(tasks) == 0 {
		b.sendMessage(chatID, fmt.Sprint("У тебя нет задач бразэ!"))
		return
	}
	var message strings.Builder
	message.WriteString(header + "\n\n")
	for i, task := range tasks {
		message.WriteString(fmt.Sprintf("Задача #%d\n", i+1))
		message.WriteString(fmt.Sprintf("ID: %s\n", task.Id))
		message.WriteString(fmt.Sprintf("Текст: %s\n", task.Text))

		if task.AssignedId != "" {
			message.WriteString(fmt.Sprintf("Исполнитель: %s\n", task.AssignedId))
		} else {
			message.WriteString("Исполнитель: не назначен\n")
		}

		message.WriteString(fmt.Sprintf("Создана: %s\n", formatTime(task.CreatedAt)))
		message.WriteString("---\n")
	}
}
func (b *Bot) sendMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	b.bot.Send(msg)
}

func formatTime(time string) string {
	if len(time) < 10 {
		return time[:10]
	}
	return time
}
