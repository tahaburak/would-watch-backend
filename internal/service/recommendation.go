package service

import (
	"context"
	"fmt"
	"log"

	"github.com/tahaburak/would-watch-backend/internal/database"
	"github.com/tahaburak/would-watch-backend/internal/openai"
	"github.com/tahaburak/would-watch-backend/internal/tmdb"
	"github.com/google/uuid"
)

type RecommendationService struct {
	openaiClient *openai.Client
	tmdbClient   *tmdb.Client
	voteRepo     *database.VoteRepository
	mediaRepo    *database.MediaRepository
}

func NewRecommendationService(oid *openai.Client, t *tmdb.Client, v *database.VoteRepository, m *database.MediaRepository) *RecommendationService {
	return &RecommendationService{
		openaiClient: oid,
		tmdbClient:   t,
		voteRepo:     v,
		mediaRepo:    m,
	}
}

// GenerateRecommendations fetches liked movies, asks OpenAI, and caches results
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, sessionID uuid.UUID) ([]database.MediaItem, error) {
	// 1. Get liked movies from this session
	likedTitles, err := s.voteRepo.GetLikedMovies(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked movies: %w", err)
	}

	if len(likedTitles) == 0 {
		return []database.MediaItem{}, nil
	}

	// 2. Ask OpenAI for recommendations
	recommendedTMDBIDs, err := s.openaiClient.GetRecommendations(likedTitles)
	if err != nil {
		return nil, fmt.Errorf("failed to get openai recommendations: %w", err)
	}

	var recommendations []database.MediaItem

	// 3. Fetch details for each recommended movie + Cache in DB
	for _, tmdbID := range recommendedTMDBIDs {
		// Check if we already have it in DB?
		existing, err := s.mediaRepo.GetMediaByTMDBID(ctx, tmdbID, "movie")
		if err != nil {
			log.Printf("Warning: DB lookup failed for tmdb_id %d: %v", tmdbID, err)
		}

		if existing != nil {
			recommendations = append(recommendations, *existing)
			continue
		}

		// Not in DB, fetch from TMDB
		tmdbMovie, err := s.tmdbClient.GetMovieByID(tmdbID)
		if err != nil {
			log.Printf("Warning: TMDB fetch failed for tmdb_id %d: %v", tmdbID, err)
			continue
		}

		// Cache it
		_, err = s.mediaRepo.CacheMovie(ctx, *tmdbMovie)
		if err != nil {
			log.Printf("Warning: Failed to cache recommendation %d: %v", tmdbID, err)
			// Continue anyway, we can use the TMDB data (but we need MediaItem struct... for now let's re-fetch from DB or construct manually)
		}

		// Re-fetch from DB to get the UUID and consistent format
		saved, err := s.mediaRepo.GetMediaByTMDBID(ctx, tmdbID, "movie")
		if err == nil && saved != nil {
			recommendations = append(recommendations, *saved)
		}
	}

	return recommendations, nil
}
