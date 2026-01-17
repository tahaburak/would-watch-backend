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

// SocialHandler handles social-related API endpoints
type SocialHandler struct {
	socialRepo *database.SocialRepository
}

// NewSocialHandler creates a new social handler
func NewSocialHandler(socialRepo *database.SocialRepository) *SocialHandler {
	return &SocialHandler{
		socialRepo: socialRepo,
	}
}

// FollowUser handles POST /api/follows/{id}
func (h *SocialHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user ID
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	followerID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract target user ID from URL
	// Expected format: /api/follows/{id}
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	followingID, err := uuid.Parse(parts[2])
	if err != nil {
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}

	// Prevent self-follow
	if followerID == followingID {
		http.Error(w, "Cannot follow yourself", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Follow the user
	if err := h.socialRepo.FollowUser(ctx, followerID, followingID); err != nil {
		log.Printf("Error following user: %v", err)
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User followed successfully",
	})
}

// UnfollowUser handles DELETE /api/follows/{id}
func (h *SocialHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user ID
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	followerID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract target user ID from URL
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	followingID, err := uuid.Parse(parts[2])
	if err != nil {
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Unfollow the user
	if err := h.socialRepo.UnfollowUser(ctx, followerID, followingID); err != nil {
		log.Printf("Error unfollowing user: %v", err)
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User unfollowed successfully",
	})
}

// GetFollowing handles GET /api/me/following
func (h *SocialHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user ID
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get following list
	following, err := h.socialRepo.GetFollowing(ctx, userID)
	if err != nil {
		log.Printf("Error getting following: %v", err)
		http.Error(w, "Failed to get following", http.StatusInternalServerError)
		return
	}

	if following == nil {
		following = []database.Profile{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"following": following,
		"count":     len(following),
	})
}

// SearchUsers handles GET /api/users/search?q=
func (h *SocialHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Search for users
	users, err := h.socialRepo.SearchUsers(ctx, query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []database.Profile{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}
