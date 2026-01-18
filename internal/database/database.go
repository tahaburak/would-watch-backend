package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Client wraps the database connection
type Client struct {
	DB *sql.DB
}

// NewClient creates a new database client using PostgreSQL connection string
func NewClient(databaseURL string) (*Client, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Force simple protocol to avoid prepared statement issues with Supabase transaction poolers
	if !strings.Contains(databaseURL, "default_query_exec_mode") {
		separator := "?"
		if strings.Contains(databaseURL, "?") {
			separator = "&"
		}
		databaseURL += separator + "default_query_exec_mode=simple_protocol"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		// Provide helpful error message for common Supabase connection issues
		errStr := err.Error()
		if contains(errStr, "Unsupported or invalid secret format") {
			return nil, fmt.Errorf("failed to ping database: %w\n\nTROUBLESHOOTING: This usually means your DATABASE_URL has the wrong username format.\nFor Supabase pooler connections, use: postgresql://postgres.[project-ref]:[password]@...\nFor direct connections, use: postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres\nSee backend/TROUBLESHOOTING.md for more details", err)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{DB: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.DB.Close()
}
