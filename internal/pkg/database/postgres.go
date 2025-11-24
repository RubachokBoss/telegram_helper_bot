package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresConnection(cfg Config) (*sql.DB, error) {
	// Собираем все параметры
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10 application_name=task_manager",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	log.Printf("Connecting to: host=%s, port=%s, dbname=%s", cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open failed: %v", err)
	}

	// Тестовый запрос для проверки
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("test query failed: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database successfully!")
	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
		text TEXT NOT NULL,
		owner_id VARCHAR(100) NOT NULL,
		assigned_id VARCHAR(100),
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_tasks_owner_id ON tasks(owner_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_assigned_id ON tasks(assigned_id);
	`

	_, err := db.Exec(query)
	return err
}
