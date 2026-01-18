package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestDB wraps a database connection for testing
type TestDB struct {
	DB *sql.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return &TestDB{DB: db}
}

// Close closes the database connection
func (tdb *TestDB) Close() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// Cleanup cleans up test data from all tables
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	tables := []string{
		"session_votes",
		"room_participants",
		"watch_sessions",
		"media_items",
		"user_follows",
		"profiles",
		"auth.users",
	}

	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to cleanup %s: %v", table, err)
		}
	}
}

// SeedUser creates a test user in auth.users and returns the ID
// Note: In real tests, you'd typically mock auth.users or use Supabase's test helpers
func (tdb *TestDB) SeedUser(t *testing.T, email string) uuid.UUID {
	t.Helper()

	userID := uuid.New()

	// Note: This assumes auth.users table exists and is accessible
	// In production, Supabase manages this table
	query := `
		INSERT INTO auth.users (id, email, encrypted_password, email_confirmed_at, created_at, updated_at)
		VALUES ($1, $2, 'test_password', NOW(), NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`

	_, err := tdb.DB.Exec(query, userID, email)
	if err != nil {
		// If auth.users is not accessible, we'll just return the UUID
		// The handle_new_user trigger should create the profile
		t.Logf("Note: Could not insert into auth.users (expected in some test setups): %v", err)
	}

	return userID
}

// SeedProfile creates a test profile (and corresponding auth.users entry)
func (tdb *TestDB) SeedProfile(t *testing.T, userID uuid.UUID, username string) {
	t.Helper()

	// First, ensure auth.users entry exists
	authQuery := `
		INSERT INTO auth.users (id, email, encrypted_password, email_confirmed_at, created_at, updated_at)
		VALUES ($1, $2, 'test_password', NOW(), NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`

	email := username + "@test.com"
	_, err := tdb.DB.Exec(authQuery, userID, email)
	if err != nil {
		t.Fatalf("Failed to seed auth.users: %v", err)
	}

	// Then create/update the profile
	query := `
		INSERT INTO profiles (id, username, invite_preference, created_at, updated_at)
		VALUES ($1, $2, 'everyone', NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username
	`

	_, err = tdb.DB.Exec(query, userID, username)
	if err != nil {
		t.Fatalf("Failed to seed profile: %v", err)
	}
}

// SeedMediaItem creates a test media item and returns its ID
func (tdb *TestDB) SeedMediaItem(t *testing.T, tmdbID int, mediaType, title string) uuid.UUID {
	t.Helper()

	mediaID := uuid.New()

	query := `
		INSERT INTO media_items (id, tmdb_id, media_type, title, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '{}', NOW(), NOW())
		ON CONFLICT (tmdb_id, media_type) DO NOTHING
		RETURNING id
	`

	err := tdb.DB.QueryRow(query, mediaID, tmdbID, mediaType, title).Scan(&mediaID)
	if err != nil {
		// If conflict, try to get existing ID
		query = `SELECT id FROM media_items WHERE tmdb_id = $1 AND media_type = $2`
		err := tdb.DB.QueryRow(query, tmdbID, mediaType).Scan(&mediaID)
		if err != nil {
			t.Fatalf("Failed to seed media item: %v", err)
		}
	}

	return mediaID
}

// SeedWatchSession creates a test watch session and returns its ID
func (tdb *TestDB) SeedWatchSession(t *testing.T, creatorID uuid.UUID, name string, isPublic bool) uuid.UUID {
	t.Helper()

	sessionID := uuid.New()

	query := `
		INSERT INTO watch_sessions (id, creator_id, status, name, is_public, created_at, updated_at)
		VALUES ($1, $2, 'active', $3, $4, NOW(), NOW())
		RETURNING id
	`

	err := tdb.DB.QueryRow(query, sessionID, creatorID, name, isPublic).Scan(&sessionID)
	if err != nil {
		t.Fatalf("Failed to seed watch session: %v", err)
	}

	return sessionID
}

// SeedFollow creates a follow relationship
func (tdb *TestDB) SeedFollow(t *testing.T, followerID, followingID uuid.UUID) {
	t.Helper()

	query := `
		INSERT INTO user_follows (follower_id, following_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (follower_id, following_id) DO NOTHING
	`

	_, err := tdb.DB.Exec(query, followerID, followingID)
	if err != nil {
		t.Fatalf("Failed to seed follow: %v", err)
	}
}

// SeedRoomParticipant adds a participant to a room
func (tdb *TestDB) SeedRoomParticipant(t *testing.T, roomID, userID uuid.UUID, role, status string) {
	t.Helper()

	query := `
		INSERT INTO room_participants (room_id, user_id, role, status, joined_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (room_id, user_id) DO NOTHING
	`

	_, err := tdb.DB.Exec(query, roomID, userID, role, status)
	if err != nil {
		t.Fatalf("Failed to seed room participant: %v", err)
	}
}

// SeedVote creates a vote for a media item in a session
func (tdb *TestDB) SeedVote(t *testing.T, sessionID, userID, mediaID uuid.UUID, vote string) {
	t.Helper()

	query := `
		INSERT INTO session_votes (session_id, user_id, media_id, vote, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (session_id, user_id, media_id)
		DO UPDATE SET vote = EXCLUDED.vote, updated_at = NOW()
	`

	_, err := tdb.DB.Exec(query, sessionID, userID, mediaID, vote)
	if err != nil {
		t.Fatalf("Failed to seed vote: %v", err)
	}
}

// WithTransaction executes a function within a transaction and rolls it back
func (tdb *TestDB) WithTransaction(t *testing.T, fn func(*sql.Tx)) {
	t.Helper()

	tx, err := tdb.DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	defer tx.Rollback()

	fn(tx)
}

// ExecuteContext is a helper to execute queries with context
func (tdb *TestDB) ExecuteContext(ctx context.Context, query string, args ...interface{}) error {
	_, err := tdb.DB.ExecContext(ctx, query, args...)
	return err
}
