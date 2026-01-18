package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/tahaburak/would-watch-backend/internal/database"
	"github.com/tahaburak/would-watch-backend/internal/middleware"
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

// GetProfile handles GET /api/me/profile
func (h *SocialHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	profile, err := h.socialRepo.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("Error getting profile: %v", err)
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		// Return 404 or empty profile? Let's return 404 so frontend knows to create one
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

type UpdateProfileRequest struct {
	Username         string `json:"username"`
	InvitePreference string `json:"invite_preference"`
}

// UpdateProfile handles PUT /api/me/profile
func (h *SocialHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if req.InvitePreference != "everyone" && req.InvitePreference != "following" && req.InvitePreference != "none" {
		http.Error(w, "Invalid invite preference", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.socialRepo.CreateOrUpdateProfile(ctx, userID, req.Username, req.InvitePreference); err != nil {
		log.Printf("Error updating profile: %v", err)
		// Check for unique violation on username
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Fetch updated profile to return
	profile, err := h.socialRepo.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("Error returning updated profile: %v", err)
		w.WriteHeader(http.StatusOK) // Action succeeded anyway
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
