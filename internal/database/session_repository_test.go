package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
)

func TestSessionRepository_CreateSession(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create a user profile
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "session_creator")

	t.Run("successfully creates a session", func(t *testing.T) {
		session, err := repo.CreateSession(ctx, creatorID)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.ID == uuid.Nil {
			t.Error("Expected session ID to be set")
		}

		if session.CreatorID != creatorID {
			t.Errorf("Expected creator ID %s, got %s", creatorID, session.CreatorID)
		}

		if session.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", session.Status)
		}

		if session.CreatedAt == "" {
			t.Error("Expected created_at to be set")
		}

		if session.UpdatedAt == "" {
			t.Error("Expected updated_at to be set")
		}

		if session.CompletedAt != nil {
			t.Error("Expected completed_at to be nil for new session")
		}
	})

	t.Run("fails when creator doesn't exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.CreateSession(ctx, nonExistentID)
		if err == nil {
			t.Error("Expected CreateSession to fail with non-existent creator")
		}
	})

	t.Run("allows creating multiple sessions for same creator", func(t *testing.T) {
		session1, err := repo.CreateSession(ctx, creatorID)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		session2, err := repo.CreateSession(ctx, creatorID)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		if session1.ID == session2.ID {
			t.Error("Expected different session IDs for multiple sessions")
		}
	})
}

func TestSessionRepository_GetSessionByID(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create a user and session
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "session_creator")

	createdSession, err := repo.CreateSession(ctx, creatorID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("successfully retrieves existing session", func(t *testing.T) {
		session, err := repo.GetSessionByID(ctx, createdSession.ID)
		if err != nil {
			t.Fatalf("GetSessionByID failed: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.ID != createdSession.ID {
			t.Errorf("Expected session ID %s, got %s", createdSession.ID, session.ID)
		}

		if session.CreatorID != creatorID {
			t.Errorf("Expected creator ID %s, got %s", creatorID, session.CreatorID)
		}

		if session.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", session.Status)
		}
	})

	t.Run("returns nil for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()
		session, err := repo.GetSessionByID(ctx, nonExistentID)
		if err != nil {
			t.Fatalf("GetSessionByID should not error for non-existent ID: %v", err)
		}

		if session != nil {
			t.Error("Expected nil for non-existent session")
		}
	})
}

func TestSessionRepository_CompleteSession(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create a user and session
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "session_creator")

	createdSession, err := repo.CreateSession(ctx, creatorID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("successfully completes a session", func(t *testing.T) {
		session, err := repo.CompleteSession(ctx, createdSession.ID)
		if err != nil {
			t.Fatalf("CompleteSession failed: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned, got nil")
		}

		if session.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", session.Status)
		}

		// Note: completed_at is not automatically set by the current implementation
		// It would need a database trigger or explicit SQL to set it

		// Verify the session is actually updated in the database
		retrieved, err := repo.GetSessionByID(ctx, createdSession.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve completed session: %v", err)
		}

		if retrieved.Status != "completed" {
			t.Errorf("Expected retrieved session status 'completed', got '%s'", retrieved.Status)
		}
	})

	t.Run("returns nil for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()
		session, err := repo.CompleteSession(ctx, nonExistentID)
		if err != nil {
			t.Fatalf("CompleteSession should not error for non-existent ID: %v", err)
		}

		if session != nil {
			t.Error("Expected nil for non-existent session")
		}
	})

	t.Run("can complete already completed session", func(t *testing.T) {
		// Create another session
		newSession, err := repo.CreateSession(ctx, creatorID)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Complete it once
		_, err = repo.CompleteSession(ctx, newSession.ID)
		if err != nil {
			t.Fatalf("First CompleteSession failed: %v", err)
		}

		// Complete it again
		session, err := repo.CompleteSession(ctx, newSession.ID)
		if err != nil {
			t.Fatalf("Second CompleteSession failed: %v", err)
		}

		if session == nil {
			t.Fatal("Expected session to be returned")
		}

		if session.Status != "completed" {
			t.Errorf("Expected status to remain 'completed', got '%s'", session.Status)
		}
	})
}

func TestSessionRepository_SessionLifecycle(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create a user
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "lifecycle_creator")

	t.Run("full session lifecycle", func(t *testing.T) {
		// Step 1: Create session
		session, err := repo.CreateSession(ctx, creatorID)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		if session.Status != "active" {
			t.Errorf("New session should be active, got '%s'", session.Status)
		}

		sessionID := session.ID

		// Step 2: Retrieve session
		retrieved, err := repo.GetSessionByID(ctx, sessionID)
		if err != nil {
			t.Fatalf("Failed to retrieve session: %v", err)
		}

		if retrieved.ID != sessionID {
			t.Error("Retrieved wrong session")
		}

		if retrieved.Status != "active" {
			t.Error("Session status should still be active")
		}

		// Step 3: Complete session
		completed, err := repo.CompleteSession(ctx, sessionID)
		if err != nil {
			t.Fatalf("Failed to complete session: %v", err)
		}

		if completed.Status != "completed" {
			t.Error("Session should be completed")
		}

		// Step 4: Verify completion persisted
		final, err := repo.GetSessionByID(ctx, sessionID)
		if err != nil {
			t.Fatalf("Failed to retrieve final session: %v", err)
		}

		if final.Status != "completed" {
			t.Error("Final session status should be completed")
		}

		// Note: completed_at is not automatically set by the current implementation
	})
}
