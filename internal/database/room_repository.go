package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Room represents a watch room (extended watch session)
type Room struct {
	ID          uuid.UUID  `json:"id"`
	CreatorID   uuid.UUID  `json:"creator_id"`
	Name        *string    `json:"name,omitempty"`
	IsPublic    bool       `json:"is_public"`
	Status      string     `json:"status"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	CompletedAt *string    `json:"completed_at,omitempty"`
}

// RoomRepository handles room-related database operations
type RoomRepository struct {
	db *sql.DB
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// CreateRoom creates a new room with initial participants
func (r *RoomRepository) CreateRoom(ctx context.Context, creatorID uuid.UUID, name string, isPublic bool, initialMembers []uuid.UUID) (*Room, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the room
	query := `
		INSERT INTO watch_sessions (creator_id, name, is_public, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id, creator_id, name, is_public, status, created_at, updated_at, completed_at
	`

	var room Room
	err = tx.QueryRowContext(ctx, query, creatorID, name, isPublic).Scan(
		&room.ID,
		&room.CreatorID,
		&room.Name,
		&room.IsPublic,
		&room.Status,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Add creator as first participant
	participantQuery := `
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
	`
	_, err = tx.ExecContext(ctx, participantQuery, room.ID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to add creator to room: %w", err)
	}

	// Add initial members
	for _, memberID := range initialMembers {
		if memberID == creatorID {
			continue // Skip creator as already added
		}
		_, err = tx.ExecContext(ctx, participantQuery, room.ID, memberID)
		if err != nil {
			return nil, fmt.Errorf("failed to add member to room: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &room, nil
}

// AddParticipant adds a user to a room
func (r *RoomRepository) AddParticipant(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (room_id, user_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// GetRoomsByUser retrieves all rooms a user is part of
func (r *RoomRepository) GetRoomsByUser(ctx context.Context, userID uuid.UUID) ([]Room, error) {
	query := `
		SELECT DISTINCT ws.id, ws.creator_id, ws.name, ws.is_public, ws.status, ws.created_at, ws.updated_at, ws.completed_at
		FROM watch_sessions ws
		INNER JOIN room_participants rp ON ws.id = rp.room_id
		WHERE rp.user_id = $1
		ORDER BY ws.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(
			&room.ID,
			&room.CreatorID,
			&room.Name,
			&room.IsPublic,
			&room.Status,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rooms: %w", err)
	}

	return rooms, nil
}

// GetRoomByID retrieves a room by its ID
func (r *RoomRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*Room, error) {
	query := `
		SELECT id, creator_id, name, is_public, status, created_at, updated_at, completed_at
		FROM watch_sessions
		WHERE id = $1
	`

	var room Room
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&room.ID,
		&room.CreatorID,
		&room.Name,
		&room.IsPublic,
		&room.Status,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return &room, nil
}

// IsParticipant checks if a user is a participant in a room
func (r *RoomRepository) IsParticipant(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM room_participants
			WHERE room_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check participant status: %w", err)
	}

	return exists, nil
}
