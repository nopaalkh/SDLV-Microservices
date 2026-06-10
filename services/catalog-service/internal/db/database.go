package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

// InitDB initializes the database connection and runs migrations
func InitDB(connectionString string) (*sqlx.DB, error) {
	// Connect to database
	conn, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(conn.DB); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return conn, nil
}

// runMigrations creates the necessary tables
func runMigrations(db *sql.DB) error {
	// Create assets table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS assets (
		id VARCHAR(36) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10, 2) NOT NULL,
		category VARCHAR(100) NOT NULL DEFAULT 'Uncategorized',
		ai_tool_used VARCHAR(100),
		resolution VARCHAR(50),
		preview_url VARCHAR(255),
		storage_path VARCHAR(255) NOT NULL,
		sales_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_by VARCHAR(36) NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to create assets table: %w", err)
	}

	// Add category column if it doesn't exist (for existing databases)
	_, _ = db.Exec(`ALTER TABLE assets ADD COLUMN IF NOT EXISTS category VARCHAR(100) NOT NULL DEFAULT 'Uncategorized'`)

	// Add sales_count column if it doesn't exist
	_, _ = db.Exec(`ALTER TABLE assets ADD COLUMN IF NOT EXISTS sales_count INTEGER NOT NULL DEFAULT 0`)

	// Change preview_url to TEXT to support base64 images
	_, _ = db.Exec(`ALTER TABLE assets ALTER COLUMN preview_url TYPE TEXT`)

	// Create reviews table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS reviews (
		id VARCHAR(36) PRIMARY KEY,
		asset_id VARCHAR(36) NOT NULL,
		user_id VARCHAR(36) NOT NULL,
		rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
		comment TEXT,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to create assets table: %w", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_assets_title ON assets(title)`)
	if err != nil {
		log.Printf("Warning: failed to create index on assets.title: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_assets_created_by ON assets(created_by)`)
	if err != nil {
		log.Printf("Warning: failed to create index on assets.created_by: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_assets_category ON assets(category)`)
	if err != nil {
		log.Printf("Warning: failed to create index on assets.category: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_reviews_asset_id ON reviews(asset_id)`)
	if err != nil {
		log.Printf("Warning: failed to create index on reviews.asset_id: %v", err)
	}

	// Add unique constraint to prevent duplicate reviews from same user for same asset
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_reviews_user_asset ON reviews(user_id, asset_id)`)
	if err != nil {
		log.Printf("Warning: failed to create unique index on reviews(user_id, asset_id): %v", err)
	}

	return nil
}