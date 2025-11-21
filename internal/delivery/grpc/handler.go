package grpc

import (
	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	"google.golang.org/grpc"
)

type taskHandler struct {
	pb.UnimplementedTaskServer
	taskService domain.TaskService
}

func NewTaskHandler(taskService domain.TaskService) *taskHandler {
	return &taskHandler{
		taskService: taskService,
	}
}

func (h *taskHandler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	return nil, nil
}

func (h *taskHandler) DomainToPb(task *domain.Task) *pb.Task{
	return &pb.Task{
		Id
	}
}
