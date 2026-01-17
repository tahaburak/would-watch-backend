package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/burak/would_watch/backend/internal/database"
	"github.com/google/uuid"
)

// MatchHandler handles match-related API endpoints
type MatchHandler struct {
	voteRepo *database.VoteRepository
}

// NewMatchHandler creates a new match handler
func NewMatchHandler(voteRepo *database.VoteRepository) *MatchHandler {
	return &MatchHandler{
		voteRepo: voteRepo,
	}
}

// MatchesResponse represents the response for the matches endpoint
type MatchesResponse struct {
	Matches []database.MediaItem `json:"matches"`
	Count   int                  `json:"count"`
}

// GetMatches handles GET /api/sessions/{id}/matches
func (h *MatchHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	// Expected format: /api/sessions/{id}/matches
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[3] != "matches" {
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

	// Get matches for the session
	matches, err := h.voteRepo.GetMatchesForSession(ctx, sessionID)
	if err != nil {
		log.Printf("Error getting matches: %v", err)
		http.Error(w, "Failed to get matches", http.StatusInternalServerError)
		return
	}

	// If no matches found, return empty array
	if matches == nil {
		matches = []database.MediaItem{}
	}

	response := MatchesResponse{
		Matches: matches,
		Count:   len(matches),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
