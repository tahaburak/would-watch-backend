package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/burak/would_watch/backend/internal/database"
	"github.com/burak/would_watch/backend/internal/middleware"
	"github.com/google/uuid"
)

// SessionHandler handles session-related API endpoints
type SessionHandler struct {
	sessionRepo *database.SessionRepository
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionRepo *database.SessionRepository) *SessionHandler {
	return &SessionHandler{
		sessionRepo: sessionRepo,
	}
}

// CreateSessionResponse represents the response when creating a session
type CreateSessionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// CreateSession handles POST /api/sessions
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	// Parse user ID as UUID
	creatorID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Create session in database
	session, err := h.sessionRepo.CreateSession(ctx, creatorID)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := CreateSessionResponse{
		ID:     session.ID.String(),
		Status: session.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// GetSession handles GET /api/sessions/{id}
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	// Expected format: /api/sessions/{id}
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	sessionIDStr := parts[2]
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get session from database
	session, err := h.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
