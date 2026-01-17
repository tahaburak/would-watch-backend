package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Profile represents a user profile
type Profile struct {
	UserID           uuid.UUID `json:"user_id"`
	Username         *string   `json:"username,omitempty"`
	InvitePreference string    `json:"invite_preference"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

// SocialRepository handles social-related database operations
type SocialRepository struct {
	db *sql.DB
}

// NewSocialRepository creates a new social repository
func NewSocialRepository(db *sql.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// GetProfile retrieves a user's profile
func (r *SocialRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	query := `
		SELECT user_id, username, invite_preference, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`

	var profile Profile
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.InvitePreference,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return &profile, nil
}

// CreateOrUpdateProfile creates or updates a user's profile
func (r *SocialRepository) CreateOrUpdateProfile(ctx context.Context, userID uuid.UUID, username string, invitePreference string) error {
	query := `
		INSERT INTO profiles (user_id, username, invite_preference)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			invite_preference = EXCLUDED.invite_preference,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, userID, username, invitePreference)
	if err != nil {
		return fmt.Errorf("failed to create/update profile: %w", err)
	}

	return nil
}

// FollowUser creates a follow relationship
func (r *SocialRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `
		INSERT INTO user_follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, following_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}

	return nil
}

// UnfollowUser removes a follow relationship
func (r *SocialRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `
		DELETE FROM user_follows
		WHERE follower_id = $1 AND following_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}

	return nil
}

// GetFollowing retrieves users that a user is following
func (r *SocialRepository) GetFollowing(ctx context.Context, userID uuid.UUID) ([]Profile, error) {
	query := `
		SELECT p.user_id, p.username, p.invite_preference, p.created_at, p.updated_at
		FROM profiles p
		INNER JOIN user_follows uf ON p.user_id = uf.following_id
		WHERE uf.follower_id = $1
		ORDER BY p.username
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()

	var following []Profile
	for rows.Next() {
		var profile Profile
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.InvitePreference,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		following = append(following, profile)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating following: %w", err)
	}

	return following, nil
}

// IsFollowing checks if followerID is following followingID
func (r *SocialRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_follows
			WHERE follower_id = $1 AND following_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, followerID, followingID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check following status: %w", err)
	}

	return exists, nil
}

// SearchUsers searches for users by username or email
func (r *SocialRepository) SearchUsers(ctx context.Context, query string) ([]Profile, error) {
	searchQuery := `
		SELECT p.user_id, p.username, p.invite_preference, p.created_at, p.updated_at
		FROM profiles p
		WHERE p.username ILIKE $1
		LIMIT 20
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []Profile
	for rows.Next() {
		var profile Profile
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.InvitePreference,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		users = append(users, profile)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}
