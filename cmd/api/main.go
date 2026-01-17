package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/burak/would_watch/backend/internal/api"
	"github.com/burak/would_watch/backend/internal/config"
	"github.com/burak/would_watch/backend/internal/database"
	"github.com/burak/would_watch/backend/internal/middleware"
	"github.com/burak/would_watch/backend/internal/openai"
	"github.com/burak/would_watch/backend/internal/service"
	"github.com/burak/would_watch/backend/internal/tmdb"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize Database Client
	dbClient, err := database.NewClient(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()
	log.Printf("Database client initialized")

	// Initialize TMDB Client
	tmdbClient := tmdb.NewClient(cfg.TMDBAPIKey)
	log.Printf("TMDB client initialized")

	// Initialize Repositories
	mediaRepo := database.NewMediaRepository(dbClient.DB)
	sessionRepo := database.NewSessionRepository(dbClient.DB)
	voteRepo := database.NewVoteRepository(dbClient.DB)

	// Initialize Handlers
	// Initialize Handlers
	mediaHandler := api.NewMediaHandler(tmdbClient, mediaRepo)
	sessionHandler := api.NewSessionHandler(sessionRepo)
	voteHandler := api.NewVoteHandler(voteRepo, sessionRepo)
	matchHandler := api.NewMatchHandler(voteRepo)

	// Initialize AI & Recommendations
	openAIClient := openai.NewClient(cfg.OpenAIAPIKey)
	recService := service.NewRecommendationService(openAIClient, tmdbClient, voteRepo, mediaRepo)
	recHandler := api.NewRecommendationHandler(recService)

	// Initialize Router
	mux := http.NewServeMux()

	// Apply CORS
	handler := middleware.CORSMiddleware(mux)

	// Public endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(cfg.SupabaseJWTSecret)

	// Protected endpoints - Media
	mux.Handle("/api/media/search", authMiddleware(http.HandlerFunc(mediaHandler.SearchMovies)))

	// Protected endpoints - Sessions
	mux.Handle("/api/sessions", authMiddleware(http.HandlerFunc(sessionHandler.CreateSession)))
	mux.Handle("/api/sessions/", authMiddleware(http.HandlerFunc(sessionHandler.GetSession)))

	// Protected endpoints - Voting
	mux.Handle("/api/sessions/{id}/vote", authMiddleware(http.HandlerFunc(voteHandler.CastVote)))
	mux.Handle("/api/sessions/{id}/complete", authMiddleware(http.HandlerFunc(sessionHandler.CompleteSession)))
	mux.Handle("/api/sessions/{id}/matches", authMiddleware(http.HandlerFunc(matchHandler.GetMatches)))
	mux.Handle("/api/sessions/{id}/recommendations", authMiddleware(http.HandlerFunc(recHandler.GetRecommendations)))

	// Protected endpoints - User info (example)
	mux.Handle("/api/me", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User ID not found", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": userID,
			"message": "This is a protected endpoint",
		})
	})))

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Registered routes:")
	log.Printf("  GET  /health")
	log.Printf("  GET  /api/me (protected)")
	log.Printf("  GET  /api/media/search (protected)")
	log.Printf("  POST /api/sessions (protected)")
	log.Printf("  GET  /api/sessions/{id} (protected)")
	log.Printf("  POST /api/sessions/{id}/vote (protected)")
	log.Printf("  POST /api/sessions/{id}/complete (protected)")
	log.Printf("  GET  /api/sessions/{id}/matches (protected)")

	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
