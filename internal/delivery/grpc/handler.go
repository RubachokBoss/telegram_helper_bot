package grpc

import (
	"context"
	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type taskHandler struct {
	pb.UnimplementedTaskServiceServer
	taskService domain.TaskService
}

func NewTaskHandler(taskService domain.TaskService) pb.TaskServiceServer {
	return &taskHandler{
		taskService: taskService,
	}
}

func (h *taskHandler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	newtask, err := h.taskService.CreateTask(req.Text, req.OwnerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateTaskResponse{
		Task: h.domainToPB(newtask),
	}, nil
}

func (h *taskHandler) AssignTask(ctx context.Context, req *pb.AssignTaskRequest) (*pb.AssignTaskResponse, error) {
	assignedtask, err := h.taskService.AssignTask(req.TaskId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AssignTaskResponse{
		Task: h.domainToPB(assignedtask),
	}, nil
}

func (h *taskHandler) UnassignTAsk(ctx context.Context, req *pb.UnassignTaskRequest) (*pb.UnassignTaskResponse, error) {
	unassignedtask, err := h.taskService.UnassignTask(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UnassignTaskResponse{
		Task: h.domainToPB(unassignedtask),
	}, nil
}
func (h *taskHandler) ResolveTask(ctx context.Context, req *pb.ResolveTaskRequest) (*pb.ResolveTaskResponse, error) {
	err := h.taskService.ResolveTask(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ResolveTaskResponse{
		Success: true,
	}, nil
}

func (h *taskHandler) GetUserTasks(ctx context.Context, req *pb.GetUserTasksRequest) (*pb.GetUserTasksResponse, error) {
	usertasks, err := h.taskService.GetUserTasks(req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetUserTasksResponse{
		Tasks: h.domainsToPB(usertasks),
	}, nil
}

func (h *taskHandler) GetOwnerTasks(ctx context.Context, req *pb.GetOwnerTasksRequest) (*pb.GetOwnerTasksResponse, error) {
	ownertasks, err := h.taskService.GetOwnerTasks(req.OwnerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetOwnerTasksResponse{
		Tasks: h.domainsToPB(ownertasks),
	}, nil
}
func (h *taskHandler) domainToPB(task *domain.Task) *pb.Task {
	return &pb.Task{
		Id:         task.ID,
		Text:       task.Text,
		OwnerId:    task.OwnerID,
		AssignedId: task.AssignedID,
		UpdatedAt:  task.UpdatedAt.Format(time.RFC3339),
		CreatedAt:  task.CreatedAt.Format(time.RFC3339),
	}
}

func (h *taskHandler) domainsToPB(tasks []*domain.Task) []*pb.Task {
	pbTasks := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		pbTasks[i] = h.domainToPB(task)
	}
	return pbTasks
}
