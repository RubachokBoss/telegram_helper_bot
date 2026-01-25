package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
)

// CreateTaskRequest represents the create task request body
type CreateTaskRequest struct {
	Text string `json:"text"`
}

// CreateTask handles task creation
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user, exists := getUserFromContext(r)
	if !exists {
		s.writeJSONError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Basic validation
	if req.Text == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Task text is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	response, err := s.taskClient.CreateTask(ctx, &pb.CreateTaskRequest{
		Text:    req.Text,
		OwnerId: user.ID,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSONSuccess(w, http.StatusCreated, map[string]interface{}{"task": response.Task})
}

// GetUserTasks handles getting tasks assigned to a user
func (s *Server) getUserTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract userID from URL path: /api/v1/tasks/user/{userId}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/user/")
	userID := strings.TrimSuffix(path, "/")

	if userID == "" {
		s.writeJSONError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Сначала пытаемся получить из общего кеша
	if s.cacheClient != nil && s.cacheClient.IsAvailable() {
		cachedTasks, err := s.cacheClient.GetUserTasks(ctx, userID)
		if err == nil && len(cachedTasks) > 0 {
			// Конвертируем domain.Task в pb.Task для ответа
			pbTasks := make([]*pb.Task, len(cachedTasks))
			for i, task := range cachedTasks {
				pbTasks[i] = domainTaskToPB(task)
			}
			s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"tasks": pbTasks, "source": "cache"})
			return
		}
	}

	// Если нет в кеше, идем через gRPC
	response, err := s.taskClient.GetUserTasks(ctx, &pb.GetUserTasksRequest{
		UserId: userID,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"tasks": response.Tasks, "source": "database"})
}

// GetOwnerTasks handles getting tasks owned by a user
func (s *Server) getOwnerTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract ownerID from URL path: /api/v1/tasks/owner/{ownerId}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/owner/")
	ownerID := strings.TrimSuffix(path, "/")

	if ownerID == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Owner ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Сначала пытаемся получить из общего кеша
	if s.cacheClient != nil && s.cacheClient.IsAvailable() {
		cachedTasks, err := s.cacheClient.GetOwnerTasks(ctx, ownerID)
		if err == nil && len(cachedTasks) > 0 {
			// Конвертируем domain.Task в pb.Task для ответа
			pbTasks := make([]*pb.Task, len(cachedTasks))
			for i, task := range cachedTasks {
				pbTasks[i] = domainTaskToPB(task)
			}
			s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"tasks": pbTasks, "source": "cache"})
			return
		}
	}

	// Если нет в кеше, идем через gRPC
	response, err := s.taskClient.GetOwnerTasks(ctx, &pb.GetOwnerTasksRequest{
		OwnerId: ownerID,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"tasks": response.Tasks, "source": "database"})
}

// AssignTask handles task assignment to a user
func (s *Server) assignTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract taskID and userID from URL path: /api/v1/tasks/{taskId}/assign/{userId}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[1] != "assign" {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid URL format")
		return
	}

	taskID := parts[0]
	userID := parts[2]

	if taskID == "" || userID == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Task ID and User ID are required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	response, err := s.taskClient.AssignTask(ctx, &pb.AssignTaskRequest{
		TaskId: taskID,
		UserId: userID,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"task": response.Task})
}

// UnassignTask handles task unassignment
func (s *Server) unassignTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract taskID from URL path: /api/v1/tasks/{taskId}/unassign
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] != "unassign" {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid URL format")
		return
	}

	taskID := parts[0]
	if taskID == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	response, err := s.taskClient.UnassignTask(ctx, &pb.UnassignTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"task": response.Task})
}

// ResolveTask handles task resolution (deletion)
func (s *Server) resolveTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract taskID from URL path: /api/v1/tasks/{taskId}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	taskID := strings.TrimSuffix(path, "/")

	if taskID == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	response, err := s.taskClient.ResolveTask(ctx, &pb.ResolveTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if response.Success {
		s.writeJSONSuccess(w, http.StatusOK, map[string]string{"message": "Task resolved and deleted successfully"})
	} else {
		s.writeJSONError(w, http.StatusInternalServerError, "Failed to resolve task")
	}
}

// Helper function to convert protobuf task to domain task (if needed)
func pbTaskToDomain(pbTask *pb.Task) *domain.Task {
	if pbTask == nil {
		return nil
	}

	// Note: pbTask.CreatedAt and pbTask.UpdatedAt are strings in the current protobuf definition
	// For proper conversion, we would need timestamp fields in protobuf
	// For now, returning nil times (this function is not currently used)
	return &domain.Task{
		ID:         pbTask.Id,
		Text:       pbTask.Text,
		OwnerID:    pbTask.OwnerId,
		AssignedID: pbTask.AssignedId,
		CreatedAt:  time.Time{}, // Would need proper conversion from string
		UpdatedAt:  time.Time{}, // Would need proper conversion from string
	}
}

// Helper function to convert domain task to protobuf task (if needed)
func domainTaskToPB(task *domain.Task) *pb.Task {
	if task == nil {
		return nil
	}

	// Note: pb.Task expects string timestamps, not Unix timestamps
	// This depends on your protobuf definition
	return &pb.Task{
		Id:         task.ID,
		Text:       task.Text,
		OwnerId:    task.OwnerID,
		AssignedId: task.AssignedID,
		CreatedAt:  task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  task.UpdatedAt.Format(time.RFC3339),
	}
}
