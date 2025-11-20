package domain

import "time"

type Task struct {
	ID         string    `json:"id"`
	Text       string    `json:"text"`
	OwnerID    string    `json:"owner_id"`
	AssignedID string    `json:"assigned_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TaskRepository interface {
	Create(task *Task) error
	FindByID(id string) (*Task, error)
	FindByUserID(userID string) ([]*Task, error)
	FindByOwnerID(ownerID string) ([]*Task, error)
	Update(task *Task) error
	Delete(task *Task) error
}

type TaskService interface {
	CreateTask(text, ownerId string) (*Task, error)
	AssignTask(taskID, userId string) (*Task, error)
	UnassignTask(taskID string) (*Task, error)
	ResolveTask(taskID string) error
	GetUserTasks(userID string) ([]*Task, error)
	GetOwnerTasks(ownerID string) ([]*Task, error)
}
