package repository

import (
	"context"
	"database/sql"
	"time"
)

// Review represents a review for an asset
type Review struct {
	ID        string    `db:"id" json:"id"`
	AssetID   string    `db:"asset_id" json:"asset_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Rating    int       `db:"rating" json:"rating"`
	Comment   string    `db:"comment" json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ReviewRepository defines the interface for review operations
type ReviewRepository interface {
	CreateReview(ctx context.Context, review *Review) error
	GetReviewsByAssetID(ctx context.Context, assetID string) ([]*Review, error)
}

// PostgreSQLReviewRepository implements ReviewRepository
type PostgreSQLReviewRepository struct {
	db *sql.DB
}

// NewPostgreSQLReviewRepository creates a new instance
func NewPostgreSQLReviewRepository(db *sql.DB) *PostgreSQLReviewRepository {
	return &PostgreSQLReviewRepository{db: db}
}

// CreateReview creates a new review
func (r *PostgreSQLReviewRepository) CreateReview(ctx context.Context, review *Review) error {
	query := `INSERT INTO reviews (id, asset_id, user_id, rating, comment, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, review.ID, review.AssetID, review.UserID, review.Rating, review.Comment, review.CreatedAt)
	return err
}

// GetReviewsByAssetID gets reviews for an asset
func (r *PostgreSQLReviewRepository) GetReviewsByAssetID(ctx context.Context, assetID string) ([]*Review, error) {
	query := `SELECT id, asset_id, user_id, rating, comment, created_at FROM reviews WHERE asset_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		rev := &Review{}
		if err := rows.Scan(&rev.ID, &rev.AssetID, &rev.UserID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, rows.Err()
}
