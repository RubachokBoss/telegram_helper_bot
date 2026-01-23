package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/RubachokBoss/telegram_helper_bot/config"
	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/RubachokBoss/telegram_helper_bot/pkg/pb"
)

type Server struct {
	mux        *http.ServeMux
	authSvc    domain.AuthService
	taskClient pb.TaskServiceClient
	config     *config.Config
	server     *http.Server
}

type contextKey string

const userContextKey contextKey = "user"

func NewServer(authSvc domain.AuthService, taskClient pb.TaskServiceClient, cfg *config.Config) *Server {
	server := &Server{
		mux:        http.NewServeMux(),
		authSvc:    authSvc,
		taskClient: taskClient,
		config:     cfg,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// CORS middleware wrapper
	withCORS := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			s.enableCORS(w, r)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			handler(w, r)
		}
	}

	// Public routes
	s.mux.HandleFunc("/api/v1/auth/register", withCORS(s.register))
	s.mux.HandleFunc("/api/v1/auth/login", withCORS(s.login))

	// Protected routes (require authentication)
	s.mux.HandleFunc("/api/v1/tasks", withCORS(s.authMiddleware(s.createTask)))
	s.mux.HandleFunc("/api/v1/tasks/user/", withCORS(s.authMiddleware(s.getUserTasks)))
	s.mux.HandleFunc("/api/v1/tasks/owner/", withCORS(s.authMiddleware(s.getOwnerTasks)))
	s.mux.HandleFunc("/api/v1/tasks/", withCORS(s.authMiddleware(s.handleTasks)))
	s.mux.HandleFunc("/api/v1/user/profile", withCORS(s.authMiddleware(s.getUserProfile)))
}

func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:    s.config.WebAPI.Port,
		Handler: s.mux,
	}

	log.Printf("🚀 REST API server starting on port %s", s.config.WebAPI.Port)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	log.Println("🛑 REST API server stopping")
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

// enableCORS sets CORS headers
func (s *Server) enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
}

// authMiddleware wraps handlers with authentication
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeJSONError(w, http.StatusUnauthorized, "Authorization token required")
			return
		}

		// Remove "Bearer " prefix if present
		token := authHeader
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		user, err := s.authSvc.ValidateToken(token)
		if err != nil {
			s.writeJSONError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Set user in request context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// writeJSONError writes a JSON error response
func (s *Server) writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// writeJSONSuccess writes a JSON success response
func (s *Server) writeJSONSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// getUserFromContext extracts user from request context
func getUserFromContext(r *http.Request) (*domain.WebUser, bool) {
	user, exists := r.Context().Value(userContextKey).(*domain.WebUser)
	return user, exists
}

// handleTasks routes task-specific operations based on HTTP method
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid task path")
		return
	}

	switch r.Method {
	case http.MethodPut:
		if len(parts) >= 3 && parts[1] == "assign" {
			s.assignTask(w, r)
		} else if len(parts) >= 2 && parts[1] == "unassign" {
			s.unassignTask(w, r)
		} else {
			s.writeJSONError(w, http.StatusMethodNotAllowed, "Invalid PUT operation")
		}
	case http.MethodDelete:
		s.resolveTask(w, r)
	default:
		s.writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
