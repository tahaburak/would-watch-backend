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

// VoteHandler handles vote-related API endpoints
type VoteHandler struct {
	voteRepo    *database.VoteRepository
	sessionRepo *database.SessionRepository
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(voteRepo *database.VoteRepository, sessionRepo *database.SessionRepository) *VoteHandler {
	return &VoteHandler{
		voteRepo:    voteRepo,
		sessionRepo: sessionRepo,
	}
}

// VoteRequest represents the request body for casting a vote
type VoteRequest struct {
	MediaID string `json:"media_id"`
	Vote    string `json:"vote"`
}

// VoteResponse represents the response after casting a vote
type VoteResponse struct {
	Success bool `json:"success"`
	IsMatch bool `json:"is_match"`
}

// CastVote handles POST /api/sessions/{id}/vote
func (h *VoteHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract session ID from URL path
	// Expected format: /api/sessions/{id}/vote
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[3] != "vote" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	sessionIDStr := parts[2]
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID format", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate vote value
	if req.Vote != "yes" && req.Vote != "no" && req.Vote != "maybe" {
		http.Error(w, "Vote must be 'yes', 'no', or 'maybe'", http.StatusBadRequest)
		return
	}

	// Parse media ID
	mediaID, err := uuid.Parse(req.MediaID)
	if err != nil {
		http.Error(w, "Invalid media ID format", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Check if session exists and is active
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

	if session.Status != "active" {
		http.Error(w, "Session is not active", http.StatusBadRequest)
		return
	}

	// Cast the vote
	if err := h.voteRepo.CastVote(ctx, sessionID, userID, mediaID, req.Vote); err != nil {
		log.Printf("Error casting vote: %v", err)
		http.Error(w, "Failed to cast vote", http.StatusInternalServerError)
		return
	}

	// Check if this creates a match (optimization)
	isMatch := false
	if req.Vote == "yes" {
		isMatch, err = h.voteRepo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			log.Printf("Warning: Failed to check match: %v", err)
			// Don't fail the request, just log the error
		}
	}

	response := VoteResponse{
		Success: true,
		IsMatch: isMatch,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
