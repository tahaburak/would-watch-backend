package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Client wraps the database connection
type Client struct {
	DB *sql.DB
}

// NewClient creates a new database client using PostgreSQL connection string
func NewClient(databaseURL string) (*Client, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{DB: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.DB.Close()
}
