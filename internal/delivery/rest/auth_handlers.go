package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
)

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string          `json:"token"`
	User  *domain.WebUser `json:"user"`
}

// Register handles user registration
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Email, password, and first name are required")
		return
	}

	if len(req.Password) < 6 {
		s.writeJSONError(w, http.StatusBadRequest, "Password must be at least 6 characters long")
		return
	}

	if !strings.Contains(req.Email, "@") {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	registration := domain.UserRegistration{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := s.authSvc.Register(registration)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Generate token for the newly registered user
	token, err := s.authSvc.Login(domain.UserCredentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	s.writeJSONSuccess(w, http.StatusCreated, response)
}

// Login handles user authentication
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	token, err := s.authSvc.Login(domain.UserCredentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get user details
	user, err := s.authSvc.ValidateToken(token)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "Failed to get user details")
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	s.writeJSONSuccess(w, http.StatusOK, response)
}

// GetUserProfile returns the current user's profile
func (s *Server) getUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user, exists := getUserFromContext(r)
	if !exists {
		s.writeJSONError(w, http.StatusUnauthorized, "User not found in context")
		return
	}

	s.writeJSONSuccess(w, http.StatusOK, map[string]interface{}{"user": user})
}
