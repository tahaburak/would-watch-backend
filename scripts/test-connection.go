package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("ERROR: DATABASE_URL is not set")
		os.Exit(1)
	}

	// Mask password in output
	maskedURL := databaseURL
	if len(databaseURL) > 0 {
		// Simple masking - replace password part
		for i, char := range databaseURL {
			if char == ':' {
				// Find the @ after password
				if atIndex := findChar(databaseURL, '@', i); atIndex > 0 {
					maskedURL = databaseURL[:i+1] + "***" + databaseURL[atIndex:]
					break
				}
			}
		}
	}

	fmt.Printf("Testing connection with: %s\n", maskedURL)
	fmt.Println()

	// Try to connect
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		fmt.Printf("ERROR: Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test connection
	fmt.Println("Attempting to ping database...")
	if err := db.Ping(); err != nil {
		fmt.Printf("ERROR: Failed to ping database: %v\n", err)
		fmt.Println()
		fmt.Println("Troubleshooting tips:")
		fmt.Println("1. Verify your password is correct in the Supabase Dashboard")
		fmt.Println("2. Check if your password contains special characters that need URL encoding")
		fmt.Println("3. Try using the direct connection (port 5432) instead of pooler")
		fmt.Println("4. Ensure your Supabase project is active (not paused)")
		os.Exit(1)
	}

	fmt.Println("✓ Successfully connected to database!")
	
	// Try a simple query
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
		fmt.Printf("WARNING: Could not query database version: %v\n", err)
	} else {
		fmt.Printf("✓ Database version: %s\n", version)
	}
}

func findChar(s string, char rune, start int) int {
	for i := start; i < len(s); i++ {
		if rune(s[i]) == char {
			return i
		}
	}
	return -1
}
