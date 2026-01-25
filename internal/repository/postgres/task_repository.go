package postgres

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/google/uuid"
)


type TaskRepository struct  {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) Create(task *domain.Task) error {
	task.ID = strings.ReplaceAll(uuid.NewString(), "-", "")
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	query := `INSERT INTO tasks (id, text, owner_id, assigned_id, created_at, updates_at)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, &task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskRepository) FindByID(id string) (*domain.Task, error) {
	log.Println("Fidning")
	var task domain.Task
	query := `SELECT id, text, owner_id, assigned_id, created_at, updated_at FROM tasks WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt)

	if err == sql.ErrNoRows {
		log.Printf("Не нашел я строку с id:%s", id)
		return nil, nil
	} else if err != nil {
		log.Printf("%s, %s", err, id)
	}
	
	return &task, nil
}

func (r *TaskRepository) FindByUserID(UserId string) ([]*domain.Task, error) {
	query := `SELECT id, text, owner_id, assigned_id, created_at, updated_at FROM tasks WHERE assigned_id = $1 Order by created_at desc`

	rows, err := r.db.Query(query, UserId)
	if err != nil {
		return nil ,err
	}
	defer rows.Close()
	tasks := []*domain.Task{}

	for rows.Next() {
		task := domain.Task{}

		if err := rows.Scan(&task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt); err != nil {
			log.Printf("%s", err)
			return nil ,err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func (r *TaskRepository) FindByOwnerID(OwnerId string) ([]*domain.Task, error) {
	query := `SELECT id, text, owner_id, assigned_id, created_at, updated_at FROM tasks WHERE owner_id = $1 Order by created_at desc`

	rows, err := r.db.Query(query, OwnerId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	tasks := []*domain.Task{}

	for rows.Next() {
		task := domain.Task{}

		if err := rows.Scan(&task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(task *domain.Task) error {
	query := `UPDATE tasks SET text = $1, owner_id = $2, assigned_id = $3, updated_at = $4 WHERE id = $5`

	task.UpdatedAt = time.Now()
	result, err := r.db.Exec(query, &task.Text, &task.OwnerID, &task.AssignedID, &task.UpdatedAt, &task.ID)
	if err != nil {
		return err
	}

	count, _ := result.RowsAffected()
	if count == 0 {
		log.Print("No task founded to update")
		return nil
	}
	return nil
}


func (r *TaskRepository) Delete(task *domain.Task) error {
	query := `DELETE FROM tasks WHERE id = $1`

	_, err := r.db.Exec(query, &task.ID)

	if err != nil {
		log.Printf("Error to delete id: %s", task.ID)
		return err
	}
	return nil
}