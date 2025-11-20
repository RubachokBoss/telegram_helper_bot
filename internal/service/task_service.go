package service

import (
	"errors"
	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
)

type taskService struct {
	repo domain.TaskRepository
}

func NewTaskService(repo domain.TaskRepository) domain.TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(text, ownerID string) (*domain.Task, error) {
	if text == "" {
		return nil, errors.New("text can't be empty")
	}
	if ownerID == "" {
		return nil, errors.New("ownerID can't be empty")
	}
	task := &domain.Task{
		Text:       text,
		OwnerID:    ownerID,
		AssignedID: "",
	}
	err := s.repo.Create(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) AssignTask(taskID, userID string) (*domain.Task, error) {
	task, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	task.AssignedID = userID
	err = s.repo.Update(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) UnassignTask(taskID string) (*domain.Task, error) {
	task, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	task.AssignedID = ""
	err = s.repo.Update(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) ResolveTask(taskID string) error {
	task, err := s.repo.FindByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	return s.repo.Delete(task)
}

func (s *taskService) GetUserTasks(userID string) ([]*domain.Task, error) {
	return s.repo.FindByUserID(userID)
}
func (s *taskService) GetOwnerTasks(ownerID string) ([]*domain.Task, error) {
	return s.repo.FindByOwnerID(ownerID)
}
