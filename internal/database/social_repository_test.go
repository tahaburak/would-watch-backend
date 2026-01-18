package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
)

func TestSocialRepository_FollowUser(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSocialRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	t.Run("successfully follows user", func(t *testing.T) {
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("FollowUser failed: %v", err)
		}

		// Verify follow relationship
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("IsFollowing failed: %v", err)
		}
		if !isFollowing {
			t.Error("Expected user1 to be following user2")
		}
	})

	t.Run("handles duplicate follow gracefully", func(t *testing.T) {
		// Try following again
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("FollowUser should handle duplicates: %v", err)
		}

		// Verify still only one follow relationship
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM user_follows WHERE follower_id = $1 AND following_id = $2", user1ID, user2ID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count follows: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 follow relationship, got %d", count)
		}
	})

	t.Run("fails when follower doesn't exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := repo.FollowUser(ctx, nonExistentID, user2ID)
		if err == nil {
			t.Error("Expected FollowUser to fail with non-existent follower")
		}
	})

	t.Run("fails when following user doesn't exist", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := repo.FollowUser(ctx, user1ID, nonExistentID)
		if err == nil {
			t.Error("Expected FollowUser to fail with non-existent following user")
		}
	})

	t.Run("prevents self-follow via CHECK constraint", func(t *testing.T) {
		err := repo.FollowUser(ctx, user1ID, user1ID)
		if err == nil {
			t.Error("Expected FollowUser to fail for self-follow")
		}
	})
}

func TestSocialRepository_UnfollowUser(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSocialRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users and follow relationship
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	testDB.SeedFollow(t, user1ID, user2ID)

	t.Run("successfully unfollows user", func(t *testing.T) {
		err := repo.UnfollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("UnfollowUser failed: %v", err)
		}

		// Verify follow relationship is gone
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("IsFollowing failed: %v", err)
		}
		if isFollowing {
			t.Error("Expected user1 to not be following user2 after unfollow")
		}
	})

	t.Run("handles unfollow when not following", func(t *testing.T) {
		// Try unfollowing again (already unfollowed in previous test)
		err := repo.UnfollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("UnfollowUser should handle non-existent relationship: %v", err)
		}
	})

	t.Run("doesn't affect other follow relationships", func(t *testing.T) {
		user3ID := uuid.New()
		testDB.SeedProfile(t, user3ID, "user3")

		// Create multiple follows
		testDB.SeedFollow(t, user1ID, user2ID)
		testDB.SeedFollow(t, user1ID, user3ID)
		testDB.SeedFollow(t, user2ID, user3ID)

		// Unfollow one relationship
		err := repo.UnfollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("UnfollowUser failed: %v", err)
		}

		// Verify other relationships still exist
		isFollowing, _ := repo.IsFollowing(ctx, user1ID, user3ID)
		if !isFollowing {
			t.Error("Expected user1 to still be following user3")
		}

		isFollowing, _ = repo.IsFollowing(ctx, user2ID, user3ID)
		if !isFollowing {
			t.Error("Expected user2 to still be following user3")
		}
	})
}

func TestSocialRepository_GetFollowing(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSocialRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "alice")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "bob")

	user3ID := uuid.New()
	testDB.SeedProfile(t, user3ID, "charlie")

	user4ID := uuid.New()
	testDB.SeedProfile(t, user4ID, "diana")

	t.Run("returns empty list when not following anyone", func(t *testing.T) {
		following, err := repo.GetFollowing(ctx, user1ID)
		if err != nil {
			t.Fatalf("GetFollowing failed: %v", err)
		}

		if len(following) != 0 {
			t.Errorf("Expected 0 following, got %d", len(following))
		}
	})

	t.Run("returns all users being followed", func(t *testing.T) {
		// User1 follows user2, user3, user4
		testDB.SeedFollow(t, user1ID, user2ID)
		testDB.SeedFollow(t, user1ID, user3ID)
		testDB.SeedFollow(t, user1ID, user4ID)

		following, err := repo.GetFollowing(ctx, user1ID)
		if err != nil {
			t.Fatalf("GetFollowing failed: %v", err)
		}

		if len(following) != 3 {
			t.Fatalf("Expected 3 following, got %d", len(following))
		}

		// Verify results are ordered by username (alphabetically)
		if following[0].Username == nil || *following[0].Username != "bob" {
			t.Error("Expected first user to be 'bob'")
		}
		if following[1].Username == nil || *following[1].Username != "charlie" {
			t.Error("Expected second user to be 'charlie'")
		}
		if following[2].Username == nil || *following[2].Username != "diana" {
			t.Error("Expected third user to be 'diana'")
		}
	})

	t.Run("doesn't return users not being followed", func(t *testing.T) {
		// User2 doesn't follow user1
		following, err := repo.GetFollowing(ctx, user2ID)
		if err != nil {
			t.Fatalf("GetFollowing failed: %v", err)
		}

		if len(following) != 0 {
			t.Errorf("Expected 0 following for user2, got %d", len(following))
		}
	})
}

func TestSocialRepository_IsFollowing(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSocialRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "user1")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "user2")

	user3ID := uuid.New()
	testDB.SeedProfile(t, user3ID, "user3")

	// User1 follows user2
	testDB.SeedFollow(t, user1ID, user2ID)

	t.Run("returns true when following", func(t *testing.T) {
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("IsFollowing failed: %v", err)
		}
		if !isFollowing {
			t.Error("Expected user1 to be following user2")
		}
	})

	t.Run("returns false when not following", func(t *testing.T) {
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user3ID)
		if err != nil {
			t.Fatalf("IsFollowing failed: %v", err)
		}
		if isFollowing {
			t.Error("Expected user1 to not be following user3")
		}
	})

	t.Run("follow relationship is directional", func(t *testing.T) {
		// User1 follows user2, but user2 doesn't follow user1
		isFollowing, err := repo.IsFollowing(ctx, user2ID, user1ID)
		if err != nil {
			t.Fatalf("IsFollowing failed: %v", err)
		}
		if isFollowing {
			t.Error("Expected user2 to not be following user1")
		}
	})

	t.Run("returns false for non-existent users", func(t *testing.T) {
		nonExistentID := uuid.New()
		isFollowing, err := repo.IsFollowing(ctx, nonExistentID, user2ID)
		if err != nil {
			t.Fatalf("IsFollowing failed: %v", err)
		}
		if isFollowing {
			t.Error("Expected false for non-existent follower")
		}
	})
}

func TestSocialRepository_SearchUsers(t *testing.T) {
	testDB := testutils.NewTestDB(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	repo := NewSocialRepository(testDB.DB)
	ctx := context.Background()

	// Setup: Create users with various usernames
	user1ID := uuid.New()
	testDB.SeedProfile(t, user1ID, "alice_wonder")

	user2ID := uuid.New()
	testDB.SeedProfile(t, user2ID, "bob_builder")

	user3ID := uuid.New()
	testDB.SeedProfile(t, user3ID, "alice_smith")

	user4ID := uuid.New()
	testDB.SeedProfile(t, user4ID, "charlie_brown")

	t.Run("finds users by partial username match", func(t *testing.T) {
		users, err := repo.SearchUsers(ctx, "alice")
		if err != nil {
			t.Fatalf("SearchUsers failed: %v", err)
		}

		if len(users) != 2 {
			t.Fatalf("Expected 2 users matching 'alice', got %d", len(users))
		}

		// Verify both alice users are in results
		usernames := make([]string, 0, 2)
		for _, user := range users {
			if user.Username != nil {
				usernames = append(usernames, *user.Username)
			}
		}

		hasAliceWonder := false
		hasAliceSmith := false
		for _, username := range usernames {
			if username == "alice_wonder" {
				hasAliceWonder = true
			}
			if username == "alice_smith" {
				hasAliceSmith = true
			}
		}

		if !hasAliceWonder || !hasAliceSmith {
			t.Error("Expected both alice users in search results")
		}
	})

	t.Run("search is case-insensitive", func(t *testing.T) {
		users, err := repo.SearchUsers(ctx, "ALICE")
		if err != nil {
			t.Fatalf("SearchUsers failed: %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected case-insensitive search to find 2 users, got %d", len(users))
		}
	})

	t.Run("returns empty list for no matches", func(t *testing.T) {
		users, err := repo.SearchUsers(ctx, "xyz_nonexistent")
		if err != nil {
			t.Fatalf("SearchUsers failed: %v", err)
		}

		if len(users) != 0 {
			t.Errorf("Expected 0 users for no match, got %d", len(users))
		}
	})

	t.Run("limits results to 20 users", func(t *testing.T) {
		// Create 25 users with similar names
		for i := 0; i < 25; i++ {
			userID := uuid.New()
			testDB.SeedProfile(t, userID, "test_user")
		}

		users, err := repo.SearchUsers(ctx, "test_user")
		if err != nil {
			t.Fatalf("SearchUsers failed: %v", err)
		}

		if len(users) > 20 {
			t.Errorf("Expected maximum 20 users, got %d", len(users))
		}
	})
}
