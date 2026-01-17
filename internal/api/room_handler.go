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

// RoomHandler handles room management endpoints
type RoomHandler struct {
	roomRepo   *database.RoomRepository
	socialRepo *database.SocialRepository
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomRepo *database.RoomRepository, socialRepo *database.SocialRepository) *RoomHandler {
	return &RoomHandler{
		roomRepo:   roomRepo,
		socialRepo: socialRepo,
	}
}

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	Name           string   `json:"name"`
	IsPublic       bool     `json:"is_public"`
	InitialMembers []string `json:"initial_members"`
}

// InviteRequest represents the request to invite a user to a room
type InviteRequest struct {
	UserID string `json:"user_id"`
}

// CreateRoom handles POST /api/rooms
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
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

	creatorID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse initial members
	var memberIDs []uuid.UUID
	for _, memberStr := range req.InitialMembers {
		memberID, err := uuid.Parse(memberStr)
		if err != nil {
			http.Error(w, "Invalid member ID format", http.StatusBadRequest)
			return
		}
		memberIDs = append(memberIDs, memberID)
	}

	ctx := context.Background()

	// Create room
	room, err := h.roomRepo.CreateRoom(ctx, creatorID, req.Name, req.IsPublic, memberIDs)
	if err != nil {
		log.Printf("Error creating room: %v", err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

// InviteToRoom handles POST /api/rooms/{id}/invite
func (h *RoomHandler) InviteToRoom(w http.ResponseWriter, r *http.Request) {
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

	inviterID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract room ID from URL
	// Expected format: /api/rooms/{id}/invite
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[3] != "invite" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(parts[2])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Check if room exists and inviter is creator
	room, err := h.roomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error getting room: %v", err)
		http.Error(w, "Failed to get room", http.StatusInternalServerError)
		return
	}

	if room == nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if room.CreatorID != inviterID {
		http.Error(w, "Only room creator can invite users", http.StatusForbidden)
		return
	}

	// Get target user's profile to check invite preferences
	profile, err := h.socialRepo.GetProfile(ctx, targetUserID)
	if err != nil {
		log.Printf("Error getting profile: %v", err)
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check invite preference
	if profile.InvitePreference == "none" {
		http.Error(w, "User does not accept invitations", http.StatusForbidden)
		return
	}

	if profile.InvitePreference == "following" {
		// Check if target user is following the inviter
		isFollowing, err := h.socialRepo.IsFollowing(ctx, targetUserID, inviterID)
		if err != nil {
			log.Printf("Error checking following status: %v", err)
			http.Error(w, "Failed to check following status", http.StatusInternalServerError)
			return
		}

		if !isFollowing {
			http.Error(w, "User only accepts invites from people they follow", http.StatusForbidden)
			return
		}
	}

	// Add user to room
	if err := h.roomRepo.AddParticipant(ctx, roomID, targetUserID); err != nil {
		log.Printf("Error adding participant: %v", err)
		http.Error(w, "Failed to add user to room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User invited successfully",
	})
}

// GetRooms handles GET /api/rooms
func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
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

	// Get rooms for user
	rooms, err := h.roomRepo.GetRoomsByUser(ctx, userID)
	if err != nil {
		log.Printf("Error getting rooms: %v", err)
		http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
		return
	}

	if rooms == nil {
		rooms = []database.Room{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": rooms,
		"count": len(rooms),
	})
}
