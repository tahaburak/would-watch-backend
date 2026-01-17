package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// WatchSession represents a watch session stored in the database
type WatchSession struct {
	ID          uuid.UUID  `json:"id"`
	CreatorID   uuid.UUID  `json:"creator_id"`
	Status      string     `json:"status"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	CompletedAt *string    `json:"completed_at,omitempty"`
}

// SessionRepository handles session-related database operations
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession creates a new watch session for a user
func (r *SessionRepository) CreateSession(ctx context.Context, creatorID uuid.UUID) (*WatchSession, error) {
	query := `
		INSERT INTO watch_sessions (creator_id, status)
		VALUES ($1, 'active')
		RETURNING id, creator_id, status, created_at, updated_at, completed_at
	`

	var session WatchSession
	err := r.db.QueryRowContext(ctx, query, creatorID).Scan(
		&session.ID,
		&session.CreatorID,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSessionByID retrieves a session by its ID
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*WatchSession, error) {
	query := `
		SELECT id, creator_id, status, created_at, updated_at, completed_at
		FROM watch_sessions
		WHERE id = $1
	`

	var session WatchSession
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.CreatorID,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// CompleteSession marks a session as completed
func (r *SessionRepository) CompleteSession(ctx context.Context, sessionID uuid.UUID) (*WatchSession, error) {
	query := `
		UPDATE watch_sessions
		SET status = 'completed'
		WHERE id = $1
		RETURNING id, creator_id, status, created_at, updated_at, completed_at
	`

	var session WatchSession
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.CreatorID,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to complete session: %w", err)
	}

	return &session, nil
}
