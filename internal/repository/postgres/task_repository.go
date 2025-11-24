package postgres

import (
	"database/sql"
	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *taskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) Create(task *domain.Task) error {
	task.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	query := `INSERT INTO tasks (id, text, owner_id, assigned_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, task.ID, task.Text, task.OwnerID, task.AssignedID, task.CreatedAt, task.UpdatedAt)
	return err
}

func (r *taskRepository) FindByID(id string) (*domain.Task, error) {
	log.Printf("🔍 Finding task by ID: %s", id)

	query := `SELECT id, text, owner_id, assigned_id, created_at, updated_at FROM tasks WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var task domain.Task
	err := row.Scan(&task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Task not found: %s", id)
			return nil, nil
		}
		log.Printf("❌ Error finding task %s: %v", id, err)
		return nil, err
	}

	log.Printf("✅ Task found: %s - %s", task.ID, task.Text)
	return &task, nil
}

func (r *taskRepository) FindByUserID(userID string) ([]*domain.Task, error) {
	query := `SELECT id, text, owner_id, assigned_id, created_at, updated_at FROM tasks WHERE assigned_id = $1 Order by created_at desc`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(&task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func (r *taskRepository) FindByOwnerID(ownerID string) ([]*domain.Task, error) {
	query := `SELECT id, text, owner_id, assigned_id, created_at, updated_at from tasks WHERE owner_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		err = rows.Scan(&task.ID, &task.Text, &task.OwnerID, &task.AssignedID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}

func (r *taskRepository) Update(task *domain.Task) error {
	task.UpdatedAt = time.Now()

	query := `UPDATE tasks SET text = $1, owner_id = $2, assigned_id = $3, updated_at = $4 WHERE id = $5`

	log.Printf("🔍 SQL: %s", query)
	log.Printf("🔍 Params: text=%s, owner_id=%s, assigned_id=%s, updated_at=%s, id=%s",
		task.Text, task.OwnerID, task.AssignedID, task.UpdatedAt, task.ID)

	result, err := r.db.Exec(query, task.Text, task.OwnerID, task.AssignedID, task.UpdatedAt, task.ID)
	if err != nil {
		log.Printf("❌ SQL error: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Task updated successfully. Rows affected: %d", rowsAffected)

	return nil
}
func (r *taskRepository) Delete(task *domain.Task) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, task.ID)
	return err
}
