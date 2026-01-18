package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
)

func TestRoomRepository_CreateRoom(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewRoomRepository(testDB.DB)
	ctx := context.Background()

	// Create test users
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "creator")

	memberID := uuid.New()
	testDB.SeedProfile(t, memberID, "member")

	t.Run("successfully creates room with creator only", func(t *testing.T) {
		room, err := repo.CreateRoom(ctx, creatorID, "Test Room", false, []uuid.UUID{})
		if err != nil {
			t.Fatalf("CreateRoom failed: %v", err)
		}

		if room.ID == uuid.Nil {
			t.Error("Expected room ID to be set")
		}

		if room.CreatorID != creatorID {
			t.Errorf("Expected creator ID %v, got %v", creatorID, room.CreatorID)
		}

		if room.Name == nil || *room.Name != "Test Room" {
			t.Errorf("Expected room name 'Test Room', got %v", room.Name)
		}

		if room.IsPublic {
			t.Error("Expected room to be private")
		}

		if room.Status != "active" {
			t.Errorf("Expected status 'active', got %s", room.Status)
		}

		// Verify creator is participant
		isParticipant, err := repo.IsParticipant(ctx, room.ID, creatorID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if !isParticipant {
			t.Error("Expected creator to be a participant")
		}
	})

	t.Run("successfully creates room with initial members", func(t *testing.T) {
		room, err := repo.CreateRoom(ctx, creatorID, "Team Room", true, []uuid.UUID{memberID})
		if err != nil {
			t.Fatalf("CreateRoom failed: %v", err)
		}

		// Verify creator is participant
		isParticipant, err := repo.IsParticipant(ctx, room.ID, creatorID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if !isParticipant {
			t.Error("Expected creator to be a participant")
		}

		// Verify member is participant
		isParticipant, err = repo.IsParticipant(ctx, room.ID, memberID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if !isParticipant {
			t.Error("Expected member to be a participant")
		}
	})

	t.Run("fails when creator profile doesn't exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.CreateRoom(ctx, nonExistentID, "Invalid Room", false, []uuid.UUID{})
		if err == nil {
			t.Error("Expected CreateRoom to fail with non-existent creator")
		}
	})

	t.Run("fails when initial member doesn't exist", func(t *testing.T) {
		nonExistentMemberID := uuid.New()
		_, err := repo.CreateRoom(ctx, creatorID, "Invalid Members Room", false, []uuid.UUID{nonExistentMemberID})
		if err == nil {
			t.Error("Expected CreateRoom to fail with non-existent member")
		}
	})

	t.Run("doesn't duplicate creator in participants", func(t *testing.T) {
		// Creator is also in initial members list
		room, err := repo.CreateRoom(ctx, creatorID, "Duplicate Check", false, []uuid.UUID{creatorID})
		if err != nil {
			t.Fatalf("CreateRoom failed: %v", err)
		}

		// Count participants
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM room_participants WHERE room_id = $1", room.ID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count participants: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 participant, got %d", count)
		}
	})
}

func TestRoomRepository_AddParticipant(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewRoomRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create room and users
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "creator")

	memberID := uuid.New()
	testDB.SeedProfile(t, memberID, "member")

	room, err := repo.CreateRoom(ctx, creatorID, "Test Room", false, []uuid.UUID{})
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	t.Run("successfully adds participant", func(t *testing.T) {
		err := repo.AddParticipant(ctx, room.ID, memberID)
		if err != nil {
			t.Fatalf("AddParticipant failed: %v", err)
		}

		// Verify participant was added
		isParticipant, err := repo.IsParticipant(ctx, room.ID, memberID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if !isParticipant {
			t.Error("Expected member to be a participant")
		}
	})

	t.Run("handles duplicate participant gracefully", func(t *testing.T) {
		// Try adding the same participant again
		err := repo.AddParticipant(ctx, room.ID, memberID)
		if err != nil {
			t.Errorf("AddParticipant should handle duplicates: %v", err)
		}
	})

	t.Run("fails when room doesn't exist", func(t *testing.T) {
		nonExistentRoomID := uuid.New()
		err := repo.AddParticipant(ctx, nonExistentRoomID, memberID)
		if err == nil {
			t.Error("Expected AddParticipant to fail with non-existent room")
		}
	})

	t.Run("fails when user doesn't exist", func(t *testing.T) {
		nonExistentUserID := uuid.New()
		err := repo.AddParticipant(ctx, room.ID, nonExistentUserID)
		if err == nil {
			t.Error("Expected AddParticipant to fail with non-existent user")
		}
	})
}

func TestRoomRepository_GetRoomsByUser(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewRoomRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	t.Run("returns empty list for user with no rooms", func(t *testing.T) {
		rooms, err := repo.GetRoomsByUser(ctx, user1ID)
		if err != nil {
			t.Fatalf("GetRoomsByUser failed: %v", err)
		}

		if len(rooms) != 0 {
			t.Errorf("Expected 0 rooms, got %d", len(rooms))
		}
	})

	t.Run("returns rooms user is participant in", func(t *testing.T) {
		// Create rooms
		room1, _ := repo.CreateRoom(ctx, user1ID, "User1's Room", false, []uuid.UUID{})
		room2, _ := repo.CreateRoom(ctx, user2ID, "User2's Room", true, []uuid.UUID{user1ID})

		// Get rooms for user1
		rooms, err := repo.GetRoomsByUser(ctx, user1ID)
		if err != nil {
			t.Fatalf("GetRoomsByUser failed: %v", err)
		}

		if len(rooms) != 2 {
			t.Fatalf("Expected 2 rooms, got %d", len(rooms))
		}

		// Verify rooms are returned in descending order by created_at
		if rooms[0].ID != room2.ID {
			t.Error("Expected rooms to be ordered by created_at DESC")
		}
		if rooms[1].ID != room1.ID {
			t.Error("Expected rooms to be ordered by created_at DESC")
		}
	})

	t.Run("doesn't return rooms user is not in", func(t *testing.T) {
		// Create room without user1
		repo.CreateRoom(ctx, user2ID, "Private Room", false, []uuid.UUID{})

		rooms, err := repo.GetRoomsByUser(ctx, user1ID)
		if err != nil {
			t.Fatalf("GetRoomsByUser failed: %v", err)
		}

		// Should still have 2 rooms from previous test
		if len(rooms) != 2 {
			t.Errorf("Expected 2 rooms, got %d", len(rooms))
		}
	})
}

func TestRoomRepository_GetRoomByID(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewRoomRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create user and room
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "creator")

	room, err := repo.CreateRoom(ctx, creatorID, "Test Room", true, []uuid.UUID{})
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	t.Run("successfully retrieves room by ID", func(t *testing.T) {
		retrieved, err := repo.GetRoomByID(ctx, room.ID)
		if err != nil {
			t.Fatalf("GetRoomByID failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected room to be found")
		}

		if retrieved.ID != room.ID {
			t.Errorf("Expected ID %v, got %v", room.ID, retrieved.ID)
		}

		if retrieved.CreatorID != creatorID {
			t.Errorf("Expected creator ID %v, got %v", creatorID, retrieved.CreatorID)
		}

		if *retrieved.Name != "Test Room" {
			t.Errorf("Expected name 'Test Room', got %s", *retrieved.Name)
		}

		if !retrieved.IsPublic {
			t.Error("Expected room to be public")
		}
	})

	t.Run("returns nil for non-existent room", func(t *testing.T) {
		nonExistentID := uuid.New()
		retrieved, err := repo.GetRoomByID(ctx, nonExistentID)
		if err != nil {
			t.Fatalf("GetRoomByID failed: %v", err)
		}

		if retrieved != nil {
			t.Error("Expected nil for non-existent room")
		}
	})
}

func TestRoomRepository_IsParticipant(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewRoomRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users and room
	creatorID := uuid.New()
	testDB.SeedProfile(t, creatorID, "creator")

	memberID := uuid.New()
	testDB.SeedProfile(t, memberID, "member")

	nonMemberID := uuid.New()
	testDB.SeedProfile(t, nonMemberID, "nonmember")

	room, err := repo.CreateRoom(ctx, creatorID, "Test Room", false, []uuid.UUID{memberID})
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	t.Run("returns true for creator", func(t *testing.T) {
		isParticipant, err := repo.IsParticipant(ctx, room.ID, creatorID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if !isParticipant {
			t.Error("Expected creator to be a participant")
		}
	})

	t.Run("returns true for member", func(t *testing.T) {
		isParticipant, err := repo.IsParticipant(ctx, room.ID, memberID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if !isParticipant {
			t.Error("Expected member to be a participant")
		}
	})

	t.Run("returns false for non-participant", func(t *testing.T) {
		isParticipant, err := repo.IsParticipant(ctx, room.ID, nonMemberID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if isParticipant {
			t.Error("Expected non-member to not be a participant")
		}
	})

	t.Run("returns false for non-existent room", func(t *testing.T) {
		nonExistentRoomID := uuid.New()
		isParticipant, err := repo.IsParticipant(ctx, nonExistentRoomID, creatorID)
		if err != nil {
			t.Fatalf("IsParticipant failed: %v", err)
		}
		if isParticipant {
			t.Error("Expected false for non-existent room")
		}
	})
}
