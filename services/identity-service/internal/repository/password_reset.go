package repository

import (
	"context"
	"database/sql"
	"time"
)

// PasswordReset represents a password reset token for a user
type PasswordReset struct {
	Email     string    `db:"email"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

// PasswordResetRepository defines the interface for password reset operations
type PasswordResetRepository interface {
	Create(ctx context.Context, email, token string, expiresAt time.Time) error
	GetByToken(ctx context.Context, token string) (*PasswordReset, error)
	DeleteByEmail(ctx context.Context, email string) error
}

// PostgreSQLPasswordResetRepository implements PasswordResetRepository
type PostgreSQLPasswordResetRepository struct {
	db *sql.DB
}

// NewPostgreSQLPasswordResetRepository creates a new repository
func NewPostgreSQLPasswordResetRepository(db *sql.DB) *PostgreSQLPasswordResetRepository {
	return &PostgreSQLPasswordResetRepository{db: db}
}

// Create stores or updates a password reset token for an email
func (r *PostgreSQLPasswordResetRepository) Create(ctx context.Context, email, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_resets (email, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (email)
		DO UPDATE SET token = $2, expires_at = $3
	`
	_, err := r.db.ExecContext(ctx, query, email, token, expiresAt)
	return err
}

// GetByToken retrieves a password reset entry by its token
func (r *PostgreSQLPasswordResetRepository) GetByToken(ctx context.Context, token string) (*PasswordReset, error) {
	query := `SELECT email, token, expires_at FROM password_resets WHERE token = $1`
	row := r.db.QueryRowContext(ctx, query, token)

	pr := &PasswordReset{}
	err := row.Scan(&pr.Email, &pr.Token, &pr.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return pr, nil
}

// DeleteByEmail removes any password reset token for the given email
func (r *PostgreSQLPasswordResetRepository) DeleteByEmail(ctx context.Context, email string) error {
	query := `DELETE FROM password_resets WHERE email = $1`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}
