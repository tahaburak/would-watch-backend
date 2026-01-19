package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
)

func TestVoteRepository_CastVote(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewVoteRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users, session, and media
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "voter1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "voter2")

	sessionID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)
	mediaID := testDB.SeedMediaItem(t, 123, "movie", "Test Movie")

	t.Run("successfully casts a yes vote", func(t *testing.T) {
		err := repo.CastVote(ctx, sessionID, user1ID, mediaID, "yes")
		if err != nil {
			t.Fatalf("CastVote failed: %v", err)
		}

		// Verify the vote was stored
		var vote string
		err = testDB.DB.QueryRow(
			"SELECT vote FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, user1ID, mediaID,
		).Scan(&vote)

		if err != nil {
			t.Fatalf("Failed to retrieve vote: %v", err)
		}

		if vote != "yes" {
			t.Errorf("Expected vote 'yes', got '%s'", vote)
		}
	})

	t.Run("successfully casts a no vote", func(t *testing.T) {
		media2ID := testDB.SeedMediaItem(t, 456, "movie", "Another Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, media2ID, "no")
		if err != nil {
			t.Fatalf("CastVote failed: %v", err)
		}

		var vote string
		err = testDB.DB.QueryRow(
			"SELECT vote FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, user1ID, media2ID,
		).Scan(&vote)

		if err != nil {
			t.Fatalf("Failed to retrieve vote: %v", err)
		}

		if vote != "no" {
			t.Errorf("Expected vote 'no', got '%s'", vote)
		}
	})

	t.Run("successfully casts a maybe vote", func(t *testing.T) {
		media3ID := testDB.SeedMediaItem(t, 789, "movie", "Maybe Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, media3ID, "maybe")
		if err != nil {
			t.Fatalf("CastVote failed: %v", err)
		}

		var vote string
		err = testDB.DB.QueryRow(
			"SELECT vote FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, user1ID, media3ID,
		).Scan(&vote)

		if err != nil {
			t.Fatalf("Failed to retrieve vote: %v", err)
		}

		if vote != "maybe" {
			t.Errorf("Expected vote 'maybe', got '%s'", vote)
		}
	})

	t.Run("updates existing vote on conflict", func(t *testing.T) {
		media4ID := testDB.SeedMediaItem(t, 111, "movie", "Update Movie")

		// Cast initial vote
		err := repo.CastVote(ctx, sessionID, user1ID, media4ID, "yes")
		if err != nil {
			t.Fatalf("First CastVote failed: %v", err)
		}

		// Change vote
		err = repo.CastVote(ctx, sessionID, user1ID, media4ID, "no")
		if err != nil {
			t.Fatalf("Second CastVote failed: %v", err)
		}

		// Verify vote was updated
		var vote string
		err = testDB.DB.QueryRow(
			"SELECT vote FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, user1ID, media4ID,
		).Scan(&vote)

		if err != nil {
			t.Fatalf("Failed to retrieve vote: %v", err)
		}

		if vote != "no" {
			t.Errorf("Expected updated vote 'no', got '%s'", vote)
		}

		// Verify only one vote exists
		var count int
		err = testDB.DB.QueryRow(
			"SELECT COUNT(*) FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, user1ID, media4ID,
		).Scan(&count)

		if err != nil {
			t.Fatalf("Failed to count votes: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 vote, got %d", count)
		}
	})

	t.Run("allows different users to vote on same media", func(t *testing.T) {
		media5ID := testDB.SeedMediaItem(t, 222, "movie", "Popular Movie")

		// User1 votes yes
		err := repo.CastVote(ctx, sessionID, user1ID, media5ID, "yes")
		if err != nil {
			t.Fatalf("User1 CastVote failed: %v", err)
		}

		// User2 votes yes
		err = repo.CastVote(ctx, sessionID, user2ID, media5ID, "yes")
		if err != nil {
			t.Fatalf("User2 CastVote failed: %v", err)
		}

		// Verify both votes exist
		var count int
		err = testDB.DB.QueryRow(
			"SELECT COUNT(*) FROM session_votes WHERE session_id = $1 AND media_id = $2",
			sessionID, media5ID,
		).Scan(&count)

		if err != nil {
			t.Fatalf("Failed to count votes: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected 2 votes, got %d", count)
		}
	})

	t.Run("fails with invalid vote type", func(t *testing.T) {
		media6ID := testDB.SeedMediaItem(t, 333, "movie", "Invalid Vote Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, media6ID, "invalid")
		if err == nil {
			t.Error("Expected CastVote to fail with invalid vote type")
		}
	})
}

func TestVoteRepository_CheckMatch(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewVoteRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users, session, and media
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	user3ID := uuid.New()
	testDB.SeedProfile(t, user3ID, "user3")

	sessionID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)

	t.Run("returns false when no votes", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 100, "movie", "No Votes Movie")

		isMatch, err := repo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			t.Fatalf("CheckMatch failed: %v", err)
		}

		if isMatch {
			t.Error("Expected no match with no votes")
		}
	})

	t.Run("returns false with only one yes vote", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 200, "movie", "One Vote Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, mediaID, "yes")
		if err != nil {
			t.Fatalf("CastVote failed: %v", err)
		}

		isMatch, err := repo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			t.Fatalf("CheckMatch failed: %v", err)
		}

		if isMatch {
			t.Error("Expected no match with only one yes vote")
		}
	})

	t.Run("returns true with two yes votes", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 300, "movie", "Match Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, mediaID, "yes")
		if err != nil {
			t.Fatalf("User1 CastVote failed: %v", err)
		}

		err = repo.CastVote(ctx, sessionID, user2ID, mediaID, "yes")
		if err != nil {
			t.Fatalf("User2 CastVote failed: %v", err)
		}

		isMatch, err := repo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			t.Fatalf("CheckMatch failed: %v", err)
		}

		if !isMatch {
			t.Error("Expected match with two yes votes")
		}
	})

	t.Run("returns true with three yes votes", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 400, "movie", "Three Votes Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, mediaID, "yes")
		if err != nil {
			t.Fatal(err)
		}

		err = repo.CastVote(ctx, sessionID, user2ID, mediaID, "yes")
		if err != nil {
			t.Fatal(err)
		}

		err = repo.CastVote(ctx, sessionID, user3ID, mediaID, "yes")
		if err != nil {
			t.Fatal(err)
		}

		isMatch, err := repo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			t.Fatalf("CheckMatch failed: %v", err)
		}

		if !isMatch {
			t.Error("Expected match with three yes votes")
		}
	})

	t.Run("returns false when no votes count as no match", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 500, "movie", "No Votes Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, mediaID, "no")
		if err != nil {
			t.Fatal(err)
		}

		err = repo.CastVote(ctx, sessionID, user2ID, mediaID, "no")
		if err != nil {
			t.Fatal(err)
		}

		isMatch, err := repo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			t.Fatalf("CheckMatch failed: %v", err)
		}

		if isMatch {
			t.Error("Expected no match with only no votes")
		}
	})

	t.Run("ignores maybe votes in match calculation", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 600, "movie", "Maybe Movie")

		err := repo.CastVote(ctx, sessionID, user1ID, mediaID, "yes")
		if err != nil {
			t.Fatal(err)
		}

		err = repo.CastVote(ctx, sessionID, user2ID, mediaID, "maybe")
		if err != nil {
			t.Fatal(err)
		}

		isMatch, err := repo.CheckMatch(ctx, sessionID, mediaID)
		if err != nil {
			t.Fatalf("CheckMatch failed: %v", err)
		}

		if isMatch {
			t.Error("Expected no match when only one yes and one maybe vote")
		}
	})
}

func TestVoteRepository_GetMatchesForSession(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewVoteRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users and session
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	sessionID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)

	t.Run("returns empty list when no matches", func(t *testing.T) {
		matches, err := repo.GetMatchesForSession(ctx, sessionID)
		if err != nil {
			t.Fatalf("GetMatchesForSession failed: %v", err)
		}

		if len(matches) != 0 {
			t.Errorf("Expected 0 matches, got %d", len(matches))
		}
	})

	t.Run("returns matched media items", func(t *testing.T) {
		// Create media and votes for matches
		media1ID := testDB.SeedMediaItem(t, 1001, "movie", "Matched Movie 1")
		testDB.SeedVote(t, sessionID, user1ID, media1ID, "yes")
		testDB.SeedVote(t, sessionID, user2ID, media1ID, "yes")

		media2ID := testDB.SeedMediaItem(t, 1002, "movie", "Matched Movie 2")
		testDB.SeedVote(t, sessionID, user1ID, media2ID, "yes")
		testDB.SeedVote(t, sessionID, user2ID, media2ID, "yes")

		// Create media without match (only one yes)
		media3ID := testDB.SeedMediaItem(t, 1003, "movie", "Not Matched Movie")
		testDB.SeedVote(t, sessionID, user1ID, media3ID, "yes")

		matches, err := repo.GetMatchesForSession(ctx, sessionID)
		if err != nil {
			t.Fatalf("GetMatchesForSession failed: %v", err)
		}

		if len(matches) != 2 {
			t.Fatalf("Expected 2 matches, got %d", len(matches))
		}

		// Verify matches are ordered by title
		if matches[0].Title != "Matched Movie 1" {
			t.Errorf("Expected first match 'Matched Movie 1', got '%s'", matches[0].Title)
		}

		if matches[1].Title != "Matched Movie 2" {
			t.Errorf("Expected second match 'Matched Movie 2', got '%s'", matches[1].Title)
		}
	})

	t.Run("excludes media with only no votes", func(t *testing.T) {
		session2ID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)

		media4ID := testDB.SeedMediaItem(t, 2001, "movie", "No Votes Movie")
		testDB.SeedVote(t, session2ID, user1ID, media4ID, "no")
		testDB.SeedVote(t, session2ID, user2ID, media4ID, "no")

		matches, err := repo.GetMatchesForSession(ctx, session2ID)
		if err != nil {
			t.Fatalf("GetMatchesForSession failed: %v", err)
		}

		if len(matches) != 0 {
			t.Errorf("Expected 0 matches for no votes, got %d", len(matches))
		}
	})
}

func TestVoteRepository_GetLikedMovies(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewVoteRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users and session
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	sessionID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)

	t.Run("returns empty list when no likes", func(t *testing.T) {
		titles, err := repo.GetLikedMovies(ctx, sessionID)
		if err != nil {
			t.Fatalf("GetLikedMovies failed: %v", err)
		}

		if len(titles) != 0 {
			t.Errorf("Expected 0 titles, got %d", len(titles))
		}
	})

	t.Run("returns titles of liked movies", func(t *testing.T) {
		// Create media and yes votes
		media1ID := testDB.SeedMediaItem(t, 3001, "movie", "Zulu Movie") // Starts with Z to test ordering
		testDB.SeedVote(t, sessionID, user1ID, media1ID, "yes")

		media2ID := testDB.SeedMediaItem(t, 3002, "movie", "Alpha Movie") // Starts with A
		testDB.SeedVote(t, sessionID, user2ID, media2ID, "yes")

		// Create media with no vote
		media3ID := testDB.SeedMediaItem(t, 3003, "movie", "Disliked Movie")
		testDB.SeedVote(t, sessionID, user1ID, media3ID, "no")

		titles, err := repo.GetLikedMovies(ctx, sessionID)
		if err != nil {
			t.Fatalf("GetLikedMovies failed: %v", err)
		}

		if len(titles) != 2 {
			t.Fatalf("Expected 2 titles, got %d", len(titles))
		}

		// Verify alphabetical ordering
		if titles[0] != "Alpha Movie" {
			t.Errorf("Expected first title 'Alpha Movie', got '%s'", titles[0])
		}

		if titles[1] != "Zulu Movie" {
			t.Errorf("Expected second title 'Zulu Movie', got '%s'", titles[1])
		}
	})

	t.Run("returns unique titles even with multiple yes votes", func(t *testing.T) {
		session2ID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)

		media4ID := testDB.SeedMediaItem(t, 4001, "movie", "Popular Movie")
		testDB.SeedVote(t, session2ID, user1ID, media4ID, "yes")
		testDB.SeedVote(t, session2ID, user2ID, media4ID, "yes")

		titles, err := repo.GetLikedMovies(ctx, session2ID)
		if err != nil {
			t.Fatalf("GetLikedMovies failed: %v", err)
		}

		if len(titles) != 1 {
			t.Errorf("Expected 1 unique title, got %d", len(titles))
		}

		if titles[0] != "Popular Movie" {
			t.Errorf("Expected 'Popular Movie', got '%s'", titles[0])
		}
	})

	t.Run("excludes maybe votes", func(t *testing.T) {
		session3ID := testDB.SeedWatchSession(t, user1ID, "Test Session", false)

		media5ID := testDB.SeedMediaItem(t, 5001, "movie", "Maybe Movie")
		testDB.SeedVote(t, session3ID, user1ID, media5ID, "maybe")

		titles, err := repo.GetLikedMovies(ctx, session3ID)
		if err != nil {
			t.Fatalf("GetLikedMovies failed: %v", err)
		}

		if len(titles) != 0 {
			t.Errorf("Expected 0 titles for maybe votes, got %d", len(titles))
		}
	})
}
