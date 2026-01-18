package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/tahaburak/would-watch-backend/internal/database"
	"github.com/tahaburak/would-watch-backend/internal/tmdb"
	"github.com/google/uuid"
)

// MediaHandler handles media-related API endpoints
type MediaHandler struct {
	tmdbClient *tmdb.Client
	mediaRepo  *database.MediaRepository
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(tmdbClient *tmdb.Client, mediaRepo *database.MediaRepository) *MediaHandler {
	return &MediaHandler{
		tmdbClient: tmdbClient,
		mediaRepo:  mediaRepo,
	}
}

// MovieSearchResult represents a movie in the search results with local UUID
type MovieSearchResult struct {
	ID               *uuid.UUID `json:"id,omitempty"`
	TMDBID           int        `json:"tmdb_id"`
	Title            string     `json:"title"`
	OriginalTitle    string     `json:"original_title"`
	Overview         string     `json:"overview"`
	PosterPath       string     `json:"poster_path"`
	BackdropPath     string     `json:"backdrop_path"`
	ReleaseDate      string     `json:"release_date"`
	VoteAverage      float64    `json:"vote_average"`
	VoteCount        int        `json:"vote_count"`
	Popularity       float64    `json:"popularity"`
	Adult            bool       `json:"adult"`
	OriginalLanguage string     `json:"original_language"`
	GenreIDs         []int      `json:"genre_ids"`
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Page         int                 `json:"page"`
	Results      []MovieSearchResult `json:"results"`
	TotalPages   int                 `json:"total_pages"`
	TotalResults int                 `json:"total_results"`
}

// SearchMovies handles GET /api/media/search?q=query
func (h *MediaHandler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Call TMDB API to search for movies
	tmdbResp, err := h.tmdbClient.SearchMovie(query)
	if err != nil {
		log.Printf("Error searching TMDB: %v", err)
		http.Error(w, "Failed to search movies", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	results := make([]MovieSearchResult, 0, len(tmdbResp.Results))

	// Cache each movie and get local UUID
	for _, movie := range tmdbResp.Results {
		// Cache the movie in our database
		localID, err := h.mediaRepo.CacheMovie(ctx, movie)
		if err != nil {
			log.Printf("Warning: Failed to cache movie %d: %v", movie.ID, err)
			// Continue even if caching fails - we can still return TMDB data
		}

		result := MovieSearchResult{
			ID:               localID,
			TMDBID:           movie.ID,
			Title:            movie.Title,
			OriginalTitle:    movie.OriginalTitle,
			Overview:         movie.Overview,
			PosterPath:       movie.PosterPath,
			BackdropPath:     movie.BackdropPath,
			ReleaseDate:      movie.ReleaseDate,
			VoteAverage:      movie.VoteAverage,
			VoteCount:        movie.VoteCount,
			Popularity:       movie.Popularity,
			Adult:            movie.Adult,
			OriginalLanguage: movie.OriginalLanguage,
			GenreIDs:         movie.GenreIDs,
		}
		results = append(results, result)
	}

	response := SearchResponse{
		Page:         tmdbResp.Page,
		Results:      results,
		TotalPages:   tmdbResp.TotalPages,
		TotalResults: tmdbResp.TotalResults,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
