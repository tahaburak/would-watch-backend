package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/burak/would_watch/backend/internal/middleware"
	"github.com/burak/would_watch/backend/internal/service"
	"github.com/google/uuid"
)

type RecommendationHandler struct {
	recService *service.RecommendationService
}

func NewRecommendationHandler(s *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recService: s}
}

// GetRecommendations handles GET /api/sessions/{id}/recommendations
func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	// Expected format: /api/sessions/{id}/recommendations
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[3] != "recommendations" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	sessionIDStr := parts[2]
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID format", http.StatusBadRequest)
		return
	}

	// Verify user is authenticated (redundant if middleware is used, but good for context extraction)
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	// Generate recommendations
	recommendations, err := h.recService.GenerateRecommendations(r.Context(), sessionID)
	if err != nil {
		log.Printf("Error generating recommendations: %v", err)
		http.Error(w, "Failed to generate recommendations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recommendations); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
