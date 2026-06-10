package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// InitDB initializes the database connection and runs migrations
func InitDB(connectionString string, adminEmail string, adminPassword string) (*sqlx.DB, error) {
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

	// Seed admin account
	if err := seedAdmin(conn.DB, adminEmail, adminPassword); err != nil {
		log.Printf("Warning: failed to seed admin account: %v", err)
	}

	return conn, nil
}

// runMigrations creates the necessary tables
func runMigrations(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create password_resets table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS password_resets (
		email VARCHAR(255) PRIMARY KEY,
		token VARCHAR(255) NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to create password_resets table: %w", err)
	}

	// Add status column if it doesn't exist (for existing databases)
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'`)

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`)
	if err != nil {
		log.Printf("Warning: failed to create index on users.email: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_id ON users(id)`)
	if err != nil {
		log.Printf("Warning: failed to create index on users.id: %v", err)
	}

	return nil
}

// seedAdmin creates a default admin account if none exists
func seedAdmin(db *sql.DB, adminEmail string, adminPassword string) error {
	// Check if any admin exists
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'ADMIN'`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check admin count: %w", err)
	}

	if count > 0 {
		log.Println("Admin account already exists, skipping seed")
		return nil
	}

	// Hash the admin password from config
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user
	now := time.Now()
	_, err = db.Exec(
		`INSERT INTO users (id, email, password, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"00000000-0000-0000-0000-000000000001",
		adminEmail,
		string(hashedPw),
		"ADMIN",
		"ACTIVE",
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("✅ Default admin account created: %s", adminEmail)
	return nil
}