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
	// Create orders table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		asset_id VARCHAR(36) NOT NULL,
		creator_id VARCHAR(36) NOT NULL DEFAULT '',
		quantity INTEGER NOT NULL,
		total_price DECIMAL(10, 2) NOT NULL,
		platform_fee DECIMAL(10, 2) NOT NULL DEFAULT 0,
		creator_revenue DECIMAL(10, 2) NOT NULL DEFAULT 0,
		status VARCHAR(20) NOT NULL,
		secure_download_token VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// Add new columns to existing databases
	_, _ = db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS creator_id VARCHAR(36) NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS platform_fee DECIMAL(10, 2) NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS creator_revenue DECIMAL(10, 2) NOT NULL DEFAULT 0`)

	// Create withdrawals table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS withdrawals (
		id VARCHAR(36) PRIMARY KEY,
		creator_id VARCHAR(36) NOT NULL,
		amount DECIMAL(10, 2) NOT NULL,
		status VARCHAR(20) NOT NULL,
		bank_code VARCHAR(50) NOT NULL,
		account_number VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to create withdrawals table: %w", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`)
	if err != nil {
		log.Printf("Warning: failed to create index on orders.user_id: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_creator_id ON orders(creator_id)`)
	if err != nil {
		log.Printf("Warning: failed to create index on orders.user_id: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_asset_id ON orders(asset_id)`)
	if err != nil {
		log.Printf("Warning: failed to create index on orders.asset_id: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`)
	if err != nil {
		log.Printf("Warning: failed to create index on orders.status: %v", err)
	}

	return nil
}