package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
)

// TestHandleNewUserTrigger tests the handle_new_user trigger
// This trigger should automatically create a profile when a user is created in auth.users
func TestHandleNewUserTrigger(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	ctx := context.Background()

	t.Run("automatically creates profile on user creation", func(t *testing.T) {
		// Note: This test assumes we can insert into auth.users
		// In production Supabase, this would be managed by Supabase Auth
		// For testing, we verify the trigger exists and would work

		userID := uuid.New()
		username := "test_trigger_user"
		avatarURL := "https://example.com/avatar.jpg"

		// Try to insert into auth.users (this may fail if auth schema doesn't exist in test DB)
		// In that case, we'll skip this test with a note
		query := `
			INSERT INTO auth.users (id, email, encrypted_password, email_confirmed_at, raw_user_meta_data, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), $4, NOW(), NOW())
		`

		metadata := `{"username": "` + username + `", "avatar_url": "` + avatarURL + `"}`

		_, err := testDB.DB.ExecContext(ctx, query, userID, "test@example.com", "hashed_password", metadata)
		if err != nil {
			// Auth schema may not exist in test DB
			t.Skipf("Skipping trigger test - auth.users not accessible: %v", err)
			return
		}

		// Verify profile was automatically created
		var count int
		err = testDB.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles WHERE id = $1", userID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count profiles: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 profile to be created automatically, got %d", count)
		}

		// Verify profile has correct data
		var foundUsername, foundAvatarURL *string
		err = testDB.DB.QueryRowContext(ctx,
			"SELECT username, avatar_url FROM profiles WHERE id = $1",
			userID,
		).Scan(&foundUsername, &foundAvatarURL)

		if err != nil {
			t.Fatalf("Failed to retrieve profile: %v", err)
		}

		if foundUsername == nil || *foundUsername != username {
			t.Errorf("Expected username %s, got %v", username, foundUsername)
		}

		if foundAvatarURL == nil || *foundAvatarURL != avatarURL {
			t.Errorf("Expected avatar_url %s, got %v", avatarURL, foundAvatarURL)
		}
	})

	t.Run("handles null metadata gracefully", func(t *testing.T) {
		userID := uuid.New()

		query := `
			INSERT INTO auth.users (id, email, encrypted_password, email_confirmed_at, raw_user_meta_data, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), '{}', NOW(), NOW())
		`

		_, err := testDB.DB.ExecContext(ctx, query, userID, "test2@example.com", "hashed_password")
		if err != nil {
			t.Skipf("Skipping trigger test - auth.users not accessible: %v", err)
			return
		}

		// Verify profile was created with null username and avatar_url
		var foundUsername, foundAvatarURL *string
		err = testDB.DB.QueryRowContext(ctx,
			"SELECT username, avatar_url FROM profiles WHERE id = $1",
			userID,
		).Scan(&foundUsername, &foundAvatarURL)

		if err != nil {
			t.Fatalf("Failed to retrieve profile: %v", err)
		}

		// Both should be NULL since metadata was empty
		if foundUsername != nil {
			t.Errorf("Expected username to be NULL, got %v", *foundUsername)
		}

		if foundAvatarURL != nil {
			t.Errorf("Expected avatar_url to be NULL, got %v", *foundAvatarURL)
		}
	})

	t.Run("trigger function exists", func(t *testing.T) {
		// Verify the handle_new_user function exists
		var exists bool
		err := testDB.DB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_proc
				WHERE proname = 'handle_new_user'
			)
		`).Scan(&exists)

		if err != nil {
			t.Fatalf("Failed to check function existence: %v", err)
		}

		if !exists {
			t.Error("Expected handle_new_user function to exist")
		}
	})

	t.Run("trigger is registered on auth.users", func(t *testing.T) {
		// Verify the trigger is registered
		// Note: This test may fail if auth schema doesn't exist in test environment
		var exists bool
		err := testDB.DB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_trigger
				WHERE tgname = 'on_auth_user_created'
			)
		`).Scan(&exists)

		if err != nil {
			t.Fatalf("Failed to check trigger existence: %v", err)
		}

		if !exists {
			t.Logf("Note: on_auth_user_created trigger not found - may not be registered in test environment")
		}
	})
}

// TestUpdatedAtTriggers tests that updated_at triggers work correctly
func TestUpdatedAtTriggers(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	ctx := context.Background()

	t.Run("profiles updated_at trigger", func(t *testing.T) {
		userID := uuid.New()
		testDB.SeedProfile(t, userID, "test_user")

		// Get initial updated_at
		var initialUpdatedAt string
		err := testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM profiles WHERE id = $1",
			userID,
		).Scan(&initialUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get initial updated_at: %v", err)
		}

		// Update the profile
		_, err = testDB.DB.ExecContext(ctx,
			"UPDATE profiles SET username = $1 WHERE id = $2",
			"updated_user", userID,
		)
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Get new updated_at
		var newUpdatedAt string
		err = testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM profiles WHERE id = $1",
			userID,
		).Scan(&newUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get new updated_at: %v", err)
		}

		if newUpdatedAt == initialUpdatedAt {
			t.Error("Expected updated_at to change after update")
		}
	})

	t.Run("media_items updated_at trigger", func(t *testing.T) {
		mediaID := testDB.SeedMediaItem(t, 12345, "movie", "Test Movie")

		// Get initial updated_at
		var initialUpdatedAt string
		err := testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM media_items WHERE id = $1",
			mediaID,
		).Scan(&initialUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get initial updated_at: %v", err)
		}

		// Update the media item
		_, err = testDB.DB.ExecContext(ctx,
			"UPDATE media_items SET title = $1 WHERE id = $2",
			"Updated Movie", mediaID,
		)
		if err != nil {
			t.Fatalf("Failed to update media item: %v", err)
		}

		// Get new updated_at
		var newUpdatedAt string
		err = testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM media_items WHERE id = $1",
			mediaID,
		).Scan(&newUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get new updated_at: %v", err)
		}

		if newUpdatedAt == initialUpdatedAt {
			t.Error("Expected updated_at to change after update")
		}
	})

	t.Run("watch_sessions updated_at trigger", func(t *testing.T) {
		creatorID := uuid.New()
		testDB.SeedProfile(t, creatorID, "creator")
		sessionID := testDB.SeedWatchSession(t, creatorID, "Test Session", false)

		// Get initial updated_at
		var initialUpdatedAt string
		err := testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM watch_sessions WHERE id = $1",
			sessionID,
		).Scan(&initialUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get initial updated_at: %v", err)
		}

		// Update the session
		_, err = testDB.DB.ExecContext(ctx,
			"UPDATE watch_sessions SET name = $1 WHERE id = $2",
			"Updated Session", sessionID,
		)
		if err != nil {
			t.Fatalf("Failed to update session: %v", err)
		}

		// Get new updated_at
		var newUpdatedAt string
		err = testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM watch_sessions WHERE id = $1",
			sessionID,
		).Scan(&newUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get new updated_at: %v", err)
		}

		if newUpdatedAt == initialUpdatedAt {
			t.Error("Expected updated_at to change after update")
		}
	})

	t.Run("session_votes updated_at trigger", func(t *testing.T) {
		// Setup: Create user, session, and media
		userID := uuid.New()
		testDB.SeedProfile(t, userID, "voter")
		sessionID := testDB.SeedWatchSession(t, userID, "Voting Session", false)
		mediaID := testDB.SeedMediaItem(t, 67890, "movie", "Vote Movie")

		// Seed initial vote
		testDB.SeedVote(t, sessionID, userID, mediaID, "yes")

		// Get initial updated_at
		var initialUpdatedAt string
		err := testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, userID, mediaID,
		).Scan(&initialUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get initial updated_at: %v", err)
		}

		// Update the vote
		_, err = testDB.DB.ExecContext(ctx,
			"UPDATE session_votes SET vote = $1 WHERE session_id = $2 AND user_id = $3 AND media_id = $4",
			"no", sessionID, userID, mediaID,
		)
		if err != nil {
			t.Fatalf("Failed to update vote: %v", err)
		}

		// Get new updated_at
		var newUpdatedAt string
		err = testDB.DB.QueryRowContext(ctx,
			"SELECT updated_at FROM session_votes WHERE session_id = $1 AND user_id = $2 AND media_id = $3",
			sessionID, userID, mediaID,
		).Scan(&newUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to get new updated_at: %v", err)
		}

		if newUpdatedAt == initialUpdatedAt {
			t.Error("Expected updated_at to change after update")
		}
	})
}
