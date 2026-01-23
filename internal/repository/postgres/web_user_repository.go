package postgres

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/google/uuid"
)

type webUserRepository struct {
	db *sql.DB
}

func NewWebUserRepository(db *sql.DB) domain.WebUserRepository {
	return &webUserRepository{
		db: db,
	}
}

func (r *webUserRepository) Create(user *domain.WebUser) error {
	user.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO web_users (id, email, password_hash, first_name, last_name, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(query, user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		log.Printf("❌ Error creating web user: %v", err)
		return err
	}

	log.Printf("✅ Web user created: %s (%s)", user.ID, user.Email)
	return nil
}

func (r *webUserRepository) FindByID(id string) (*domain.WebUser, error) {
	log.Printf("🔍 Finding web user by ID: %s", id)

	query := `SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
	          FROM web_users WHERE id = $1`

	row := r.db.QueryRow(query, id)

	var user domain.WebUser
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Web user not found: %s", id)
			return nil, nil
		}
		log.Printf("❌ Error finding web user %s: %v", id, err)
		return nil, err
	}

	log.Printf("✅ Web user found: %s - %s", user.ID, user.Email)
	return &user, nil
}

func (r *webUserRepository) FindByEmail(email string) (*domain.WebUser, error) {
	log.Printf("🔍 Finding web user by email: %s", email)

	query := `SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
	          FROM web_users WHERE email = $1`

	row := r.db.QueryRow(query, email)

	var user domain.WebUser
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Web user not found by email: %s", email)
			return nil, nil
		}
		log.Printf("❌ Error finding web user by email %s: %v", email, err)
		return nil, err
	}

	log.Printf("✅ Web user found by email: %s - %s", user.ID, user.Email)
	return &user, nil
}

func (r *webUserRepository) Update(user *domain.WebUser) error {
	user.UpdatedAt = time.Now()

	query := `UPDATE web_users SET email = $1, password_hash = $2, first_name = $3, last_name = $4, updated_at = $5 WHERE id = $6`

	log.Printf("🔍 SQL: %s", query)
	log.Printf("🔍 Params: email=%s, first_name=%s, last_name=%s, updated_at=%s, id=%s",
		user.Email, user.FirstName, user.LastName, user.UpdatedAt, user.ID)

	result, err := r.db.Exec(query, user.Email, user.Password, user.FirstName, user.LastName, user.UpdatedAt, user.ID)
	if err != nil {
		log.Printf("❌ SQL error: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Web user updated successfully. Rows affected: %d", rowsAffected)

	return nil
}

func (r *webUserRepository) Delete(id string) error {
	query := `DELETE FROM web_users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("❌ Error deleting web user %s: %v", id, err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Web user deleted successfully. Rows affected: %d", rowsAffected)

	return nil
}
