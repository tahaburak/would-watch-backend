package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/burak/would_watch/backend/internal/tmdb"
	"github.com/google/uuid"
)

// MediaItem represents a media item stored in the database
type MediaItem struct {
	ID        uuid.UUID       `json:"id"`
	TMDBID    int             `json:"tmdb_id"`
	MediaType string          `json:"media_type"`
	Title     string          `json:"title"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// MediaRepository handles media-related database operations
type MediaRepository struct {
	db *sql.DB
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

// CacheMovie inserts or updates a movie in the database
// Uses INSERT ON CONFLICT DO NOTHING to avoid duplicates
func (r *MediaRepository) CacheMovie(ctx context.Context, movie tmdb.Movie) (*uuid.UUID, error) {
	// Prepare metadata as JSON
	metadata := map[string]interface{}{
		"original_title":    movie.OriginalTitle,
		"overview":          movie.Overview,
		"poster_path":       movie.PosterPath,
		"backdrop_path":     movie.BackdropPath,
		"release_date":      movie.ReleaseDate,
		"vote_average":      movie.VoteAverage,
		"vote_count":        movie.VoteCount,
		"popularity":        movie.Popularity,
		"adult":             movie.Adult,
		"video":             movie.Video,
		"original_language": movie.OriginalLanguage,
		"genre_ids":         movie.GenreIDs,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO media_items (tmdb_id, media_type, title, metadata)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tmdb_id, media_type) DO UPDATE
		SET title = EXCLUDED.title,
		    metadata = EXCLUDED.metadata,
		    updated_at = NOW()
		RETURNING id
	`

	var id uuid.UUID
	err = r.db.QueryRowContext(ctx, query,
		movie.ID,
		"movie",
		movie.Title,
		metadataJSON,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to cache movie: %w", err)
	}

	return &id, nil
}

// GetMediaByTMDBID retrieves a media item by its TMDB ID
func (r *MediaRepository) GetMediaByTMDBID(ctx context.Context, tmdbID int, mediaType string) (*MediaItem, error) {
	query := `
		SELECT id, tmdb_id, media_type, title, metadata, created_at, updated_at
		FROM media_items
		WHERE tmdb_id = $1 AND media_type = $2
	`

	var item MediaItem
	err := r.db.QueryRowContext(ctx, query, tmdbID, mediaType).Scan(
		&item.ID,
		&item.TMDBID,
		&item.MediaType,
		&item.Title,
		&item.Metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get media item: %w", err)
	}

	return &item, nil
}
