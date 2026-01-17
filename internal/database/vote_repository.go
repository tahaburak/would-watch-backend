package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Vote represents a vote in the database
type Vote struct {
	SessionID uuid.UUID `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	MediaID   uuid.UUID `json:"media_id"`
	Vote      string    `json:"vote"`
	CreatedAt string    `json:"created_at"`
}

// VoteRepository handles vote-related database operations
type VoteRepository struct {
	db *sql.DB
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db *sql.DB) *VoteRepository {
	return &VoteRepository{db: db}
}

// CastVote inserts or updates a user's vote for a media item in a session
func (r *VoteRepository) CastVote(ctx context.Context, sessionID, userID, mediaID uuid.UUID, vote string) error {
	query := `
		INSERT INTO session_votes (session_id, user_id, media_id, vote)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (session_id, user_id, media_id)
		DO UPDATE SET vote = EXCLUDED.vote, created_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, sessionID, userID, mediaID, vote)
	if err != nil {
		return fmt.Errorf("failed to cast vote: %w", err)
	}

	return nil
}

// CheckMatch checks if there's a match (2+ "yes" votes) for a media item in a session
func (r *VoteRepository) CheckMatch(ctx context.Context, sessionID, mediaID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM session_votes
		WHERE session_id = $1
		AND media_id = $2
		AND vote = 'yes'
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, sessionID, mediaID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check match: %w", err)
	}

	return count >= 2, nil
}

// GetMatchesForSession retrieves all media items with 2+ "yes" votes in a session
func (r *VoteRepository) GetMatchesForSession(ctx context.Context, sessionID uuid.UUID) ([]MediaItem, error) {
	query := `
		SELECT DISTINCT
			m.id,
			m.tmdb_id,
			m.media_type,
			m.title,
			m.metadata,
			m.created_at,
			m.updated_at
		FROM media_items m
		INNER JOIN session_votes sv ON m.id = sv.media_id
		WHERE sv.session_id = $1
		AND sv.vote = 'yes'
		GROUP BY m.id, m.tmdb_id, m.media_type, m.title, m.metadata, m.created_at, m.updated_at
		HAVING COUNT(sv.user_id) >= 2
		ORDER BY m.title
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}
	defer rows.Close()

	var matches []MediaItem
	for rows.Next() {
		var item MediaItem
		err := rows.Scan(
			&item.ID,
			&item.TMDBID,
			&item.MediaType,
			&item.Title,
			&item.Metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}
