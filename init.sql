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

-- Web users table for REST API authentication
CREATE TABLE IF NOT EXISTS web_users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_web_users_email ON web_users(email);