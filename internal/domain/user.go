package domain

import "time"

// WebUser represents a web application user
type WebUser struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"` // Never send password in JSON
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserCredentials for login requests
type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserRegistration for registration requests
type UserRegistration struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// WebUserRepository interface for user data operations
type WebUserRepository interface {
	Create(user *WebUser) error
	FindByID(id string) (*WebUser, error)
	FindByEmail(email string) (*WebUser, error)
	Update(user *WebUser) error
	Delete(id string) error
}

// AuthService interface for authentication operations
type AuthService interface {
	Register(creds UserRegistration) (*WebUser, error)
	Login(creds UserCredentials) (string, error) // returns JWT token
	ValidateToken(token string) (*WebUser, error)
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}
