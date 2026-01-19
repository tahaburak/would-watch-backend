package database

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
	"github.com/tahaburak/would-watch-backend/internal/tmdb"
)

func TestMediaRepository_CacheMovie(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewMediaRepository(testDB.DB)
	ctx := context.Background()

	t.Run("successfully caches a new movie", func(t *testing.T) {
		movie := tmdb.Movie{
			ID:               12345,
			Title:            "Test Movie",
			OriginalTitle:    "Test Movie Original",
			Overview:         "A test movie overview",
			PosterPath:       "/poster.jpg",
			BackdropPath:     "/backdrop.jpg",
			ReleaseDate:      "2024-01-01",
			VoteAverage:      7.5,
			VoteCount:        1000,
			Popularity:       85.5,
			Adult:            false,
			Video:            false,
			OriginalLanguage: "en",
			GenreIDs:         []int{28, 12},
		}

		id, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("CacheMovie failed: %v", err)
		}

		if id == nil {
			t.Fatal("Expected ID to be returned, got nil")
		}

		if *id == uuid.Nil {
			t.Error("Expected valid UUID, got nil UUID")
		}

		// Verify the movie was stored
		var storedTitle string
		var storedTMDBID int
		err = testDB.DB.QueryRow(
			"SELECT title, tmdb_id FROM media_items WHERE id = $1",
			id,
		).Scan(&storedTitle, &storedTMDBID)

		if err != nil {
			t.Fatalf("Failed to retrieve movie: %v", err)
		}

		if storedTitle != "Test Movie" {
			t.Errorf("Expected title 'Test Movie', got '%s'", storedTitle)
		}

		if storedTMDBID != 12345 {
			t.Errorf("Expected TMDB ID 12345, got %d", storedTMDBID)
		}
	})

	t.Run("stores metadata as JSON", func(t *testing.T) {
		movie := tmdb.Movie{
			ID:               54321,
			Title:            "Metadata Movie",
			OriginalTitle:    "Original Metadata Movie",
			Overview:         "Testing metadata storage",
			PosterPath:       "/meta_poster.jpg",
			ReleaseDate:      "2024-06-15",
			VoteAverage:      8.2,
			VoteCount:        5000,
			Popularity:       92.3,
			Adult:            false,
			Video:            false,
			OriginalLanguage: "en",
			GenreIDs:         []int{18, 80},
		}

		id, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("CacheMovie failed: %v", err)
		}

		// Retrieve and verify metadata
		var metadataJSON json.RawMessage
		err = testDB.DB.QueryRow(
			"SELECT metadata FROM media_items WHERE id = $1",
			id,
		).Scan(&metadataJSON)

		if err != nil {
			t.Fatalf("Failed to retrieve metadata: %v", err)
		}

		var metadata map[string]interface{}
		err = json.Unmarshal(metadataJSON, &metadata)
		if err != nil {
			t.Fatalf("Failed to unmarshal metadata: %v", err)
		}

		if metadata["overview"] != "Testing metadata storage" {
			t.Errorf("Expected overview 'Testing metadata storage', got '%v'", metadata["overview"])
		}

		if metadata["vote_average"] != 8.2 {
			t.Errorf("Expected vote_average 8.2, got %v", metadata["vote_average"])
		}

		if metadata["poster_path"] != "/meta_poster.jpg" {
			t.Errorf("Expected poster_path '/meta_poster.jpg', got '%v'", metadata["poster_path"])
		}
	})

	t.Run("updates existing movie on conflict", func(t *testing.T) {
		movie1 := tmdb.Movie{
			ID:          99999,
			Title:       "Original Title",
			Overview:    "Original overview",
			VoteAverage: 5.0,
		}

		// Cache the movie first time
		id1, err := repo.CacheMovie(ctx, movie1)
		if err != nil {
			t.Fatalf("First CacheMovie failed: %v", err)
		}

		// Cache again with updated data
		movie2 := tmdb.Movie{
			ID:          99999, // Same TMDB ID
			Title:       "Updated Title",
			Overview:    "Updated overview",
			VoteAverage: 8.0,
		}

		id2, err := repo.CacheMovie(ctx, movie2)
		if err != nil {
			t.Fatalf("Second CacheMovie failed: %v", err)
		}

		// Verify same UUID returned
		if *id1 != *id2 {
			t.Error("Expected same UUID for updated movie")
		}

		// Verify data was updated
		var title string
		var metadataJSON json.RawMessage
		err = testDB.DB.QueryRow(
			"SELECT title, metadata FROM media_items WHERE id = $1",
			id2,
		).Scan(&title, &metadataJSON)

		if err != nil {
			t.Fatalf("Failed to retrieve updated movie: %v", err)
		}

		if title != "Updated Title" {
			t.Errorf("Expected updated title 'Updated Title', got '%s'", title)
		}

		var metadata map[string]interface{}
		json.Unmarshal(metadataJSON, &metadata)

		if metadata["overview"] != "Updated overview" {
			t.Errorf("Expected updated overview, got '%v'", metadata["overview"])
		}

		if metadata["vote_average"] != 8.0 {
			t.Errorf("Expected updated vote_average 8.0, got %v", metadata["vote_average"])
		}

		// Verify only one record exists
		var count int
		err = testDB.DB.QueryRow(
			"SELECT COUNT(*) FROM media_items WHERE tmdb_id = $1 AND media_type = 'movie'",
			99999,
		).Scan(&count)

		if err != nil {
			t.Fatalf("Failed to count movies: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 movie, got %d", count)
		}
	})

	t.Run("handles movies with minimal data", func(t *testing.T) {
		movie := tmdb.Movie{
			ID:    11111,
			Title: "Minimal Movie",
			// All other fields are zero values
		}

		id, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("CacheMovie failed with minimal data: %v", err)
		}

		if id == nil {
			t.Fatal("Expected ID to be returned")
		}

		// Verify movie was stored
		var exists bool
		err = testDB.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM media_items WHERE id = $1)",
			id,
		).Scan(&exists)

		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}

		if !exists {
			t.Error("Expected movie to exist in database")
		}
	})

	t.Run("handles movies with special characters in title", func(t *testing.T) {
		movie := tmdb.Movie{
			ID:       22222,
			Title:    "Movie with 'quotes' and \"special\" & characters!",
			Overview: "Testing UTF-8: héllo wörld 日本語",
		}

		id, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("CacheMovie failed with special characters: %v", err)
		}

		// Verify title stored correctly
		var title string
		err = testDB.DB.QueryRow(
			"SELECT title FROM media_items WHERE id = $1",
			id,
		).Scan(&title)

		if err != nil {
			t.Fatalf("Failed to retrieve title: %v", err)
		}

		if title != movie.Title {
			t.Errorf("Title not stored correctly, expected '%s', got '%s'", movie.Title, title)
		}
	})
}

func TestMediaRepository_GetMediaByTMDBID(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewMediaRepository(testDB.DB)
	ctx := context.Background()

	t.Run("retrieves existing media item", func(t *testing.T) {
		// Cache a movie first
		movie := tmdb.Movie{
			ID:               67890,
			Title:            "Retrieve Movie",
			Overview:         "A movie to retrieve",
			VoteAverage:      7.8,
			OriginalLanguage: "en",
		}

		cachedID, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("Failed to cache movie: %v", err)
		}

		// Retrieve the movie
		item, err := repo.GetMediaByTMDBID(ctx, 67890, "movie")
		if err != nil {
			t.Fatalf("GetMediaByTMDBID failed: %v", err)
		}

		if item == nil {
			t.Fatal("Expected media item to be returned, got nil")
		}

		if item.ID != *cachedID {
			t.Errorf("Expected ID %s, got %s", cachedID, item.ID)
		}

		if item.TMDBID != 67890 {
			t.Errorf("Expected TMDB ID 67890, got %d", item.TMDBID)
		}

		if item.Title != "Retrieve Movie" {
			t.Errorf("Expected title 'Retrieve Movie', got '%s'", item.Title)
		}

		if item.MediaType != "movie" {
			t.Errorf("Expected media type 'movie', got '%s'", item.MediaType)
		}

		// Verify metadata is present
		if len(item.Metadata) == 0 {
			t.Error("Expected metadata to be present")
		}

		var metadata map[string]interface{}
		err = json.Unmarshal(item.Metadata, &metadata)
		if err != nil {
			t.Fatalf("Failed to unmarshal metadata: %v", err)
		}

		if metadata["overview"] != "A movie to retrieve" {
			t.Errorf("Expected overview 'A movie to retrieve', got '%v'", metadata["overview"])
		}
	})

	t.Run("returns nil for non-existent media", func(t *testing.T) {
		item, err := repo.GetMediaByTMDBID(ctx, 999999, "movie")
		if err != nil {
			t.Fatalf("GetMediaByTMDBID should not error for non-existent item: %v", err)
		}

		if item != nil {
			t.Error("Expected nil for non-existent media item")
		}
	})

	t.Run("distinguishes between media types", func(t *testing.T) {
		// Cache a movie
		movie := tmdb.Movie{
			ID:    12121,
			Title: "Type Test Movie",
		}

		_, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("Failed to cache movie: %v", err)
		}

		// Try to retrieve as TV show (should not find)
		item, err := repo.GetMediaByTMDBID(ctx, 12121, "tv")
		if err != nil {
			t.Fatalf("GetMediaByTMDBID failed: %v", err)
		}

		if item != nil {
			t.Error("Expected nil when searching for wrong media type")
		}

		// Retrieve as movie (should find)
		item, err = repo.GetMediaByTMDBID(ctx, 12121, "movie")
		if err != nil {
			t.Fatalf("GetMediaByTMDBID failed: %v", err)
		}

		if item == nil {
			t.Error("Expected to find movie with correct media type")
		}
	})
}

func TestMediaRepository_Integration(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewMediaRepository(testDB.DB)
	ctx := context.Background()

	t.Run("cache and retrieve workflow", func(t *testing.T) {
		// Step 1: Cache a movie
		movie := tmdb.Movie{
			ID:               55555,
			Title:            "Integration Movie",
			OriginalTitle:    "Integration Movie Original",
			Overview:         "Testing full workflow",
			PosterPath:       "/integration_poster.jpg",
			VoteAverage:      9.0,
			VoteCount:        10000,
			Popularity:       95.0,
			OriginalLanguage: "en",
			GenreIDs:         []int{28, 12, 14},
		}

		id, err := repo.CacheMovie(ctx, movie)
		if err != nil {
			t.Fatalf("Failed to cache movie: %v", err)
		}

		// Step 2: Retrieve by TMDB ID
		retrieved, err := repo.GetMediaByTMDBID(ctx, 55555, "movie")
		if err != nil {
			t.Fatalf("Failed to retrieve movie: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected to retrieve movie")
		}

		// Step 3: Verify all data matches
		if retrieved.ID != *id {
			t.Error("ID mismatch")
		}

		if retrieved.Title != movie.Title {
			t.Error("Title mismatch")
		}

		var metadata map[string]interface{}
		json.Unmarshal(retrieved.Metadata, &metadata)

		if metadata["vote_average"] != 9.0 {
			t.Error("Vote average mismatch")
		}

		genreIDs := metadata["genre_ids"].([]interface{})
		if len(genreIDs) != 3 {
			t.Errorf("Expected 3 genre IDs, got %d", len(genreIDs))
		}

		// Step 4: Update the movie
		updatedMovie := tmdb.Movie{
			ID:          55555,
			Title:       "Updated Integration Movie",
			Overview:    "Updated workflow",
			VoteAverage: 9.5,
		}

		_, err = repo.CacheMovie(ctx, updatedMovie)
		if err != nil {
			t.Fatalf("Failed to update movie: %v", err)
		}

		// Step 5: Retrieve and verify update
		final, err := repo.GetMediaByTMDBID(ctx, 55555, "movie")
		if err != nil {
			t.Fatalf("Failed to retrieve updated movie: %v", err)
		}

		if final.Title != "Updated Integration Movie" {
			t.Error("Title not updated")
		}

		var finalMetadata map[string]interface{}
		json.Unmarshal(final.Metadata, &finalMetadata)

		if finalMetadata["vote_average"] != 9.5 {
			t.Error("Vote average not updated")
		}

		// Verify still only one record
		var count int
		err = testDB.DB.QueryRow(
			"SELECT COUNT(*) FROM media_items WHERE tmdb_id = 55555",
		).Scan(&count)

		if err != nil {
			t.Fatalf("Failed to count records: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 record, got %d", count)
		}
	})
}
