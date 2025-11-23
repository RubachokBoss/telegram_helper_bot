package database

import (
	"database/sql"
	"fmt"
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

func NewPostgresConnection(config Config) (*sql.DB, error) {
	configstr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	db, err := sql.Open("postgres", configstr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	err = createTable(db)
	if err != nil {
		return nil, err
	}
	log.Println("Successfully connected to database")
	return db, nil
}

func createTable(db *sql.DB) error {
	sqltablequery := `
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
	_, err := db.Exec(sqltablequery)
	return err
}
