package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Asset represents a digital asset in the system
type Asset struct {
	ID           string    `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Description  string    `db:"description" json:"description"`
	Price        float64   `db:"price" json:"price"`
	Category     string    `db:"category" json:"category"`
	AIToolUsed   string    `db:"ai_tool_used" json:"ai_tool_used"`
	Resolution   string    `db:"resolution" json:"resolution"`
	PreviewURL   string    `db:"preview_url" json:"preview_url"`
	StoragePath  string    `db:"storage_path" json:"storage_path"`
	SalesCount   int       `db:"sales_count" json:"sales_count"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	CreatedBy    string    `db:"created_by" json:"created_by"`
}

// AssetRepository defines the interface for asset repository operations
type AssetRepository interface {
	CreateAsset(ctx context.Context, asset *Asset) error
	FindAssetByID(ctx context.Context, id string) (*Asset, error)
	FindAssetsByCreator(ctx context.Context, creatorID string) ([]*Asset, error)
	SearchAssets(ctx context.Context, query string, category string, sortBy string, limit int, offset int) ([]*Asset, error)
	UpdateAsset(ctx context.Context, asset *Asset) error
	DeleteAsset(ctx context.Context, id string) error
	GetAllAssets(ctx context.Context) ([]*Asset, error)
	IncrementSalesCount(ctx context.Context, id string) error
}

// PostgreSQLAssetRepository implements AssetRepository using PostgreSQL
type PostgreSQLAssetRepository struct {
	db *sql.DB
}

// NewPostgreSQLAssetRepository creates a new PostgreSQLAssetRepository
func NewPostgreSQLAssetRepository(db *sql.DB) *PostgreSQLAssetRepository {
	return &PostgreSQLAssetRepository{db: db}
}

// CreateAsset creates a new asset in the database
func (r *PostgreSQLAssetRepository) CreateAsset(ctx context.Context, asset *Asset) error {
	query := `INSERT INTO assets (id, title, description, price, category, ai_tool_used, resolution, preview_url, storage_path, created_at, updated_at, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query, asset.ID, asset.Title, asset.Description, asset.Price, asset.Category, asset.AIToolUsed, asset.Resolution, asset.PreviewURL, asset.StoragePath, asset.CreatedAt, asset.UpdatedAt, asset.CreatedBy)
	return err
}

// FindAssetByID finds an asset by ID
func (r *PostgreSQLAssetRepository) FindAssetByID(ctx context.Context, id string) (*Asset, error) {
	query := `SELECT id, title, description, price, category, ai_tool_used, resolution, preview_url, storage_path, sales_count, created_at, updated_at, created_by FROM assets WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	asset := &Asset{}
	err := row.Scan(&asset.ID, &asset.Title, &asset.Description, &asset.Price, &asset.Category, &asset.AIToolUsed, &asset.Resolution, &asset.PreviewURL, &asset.StoragePath, &asset.SalesCount, &asset.CreatedAt, &asset.UpdatedAt, &asset.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return asset, nil
}

// FindAssetsByCreator finds all assets for a specific creator
func (r *PostgreSQLAssetRepository) FindAssetsByCreator(ctx context.Context, creatorID string) ([]*Asset, error) {
	query := `SELECT id, title, description, price, category, ai_tool_used, resolution, preview_url, storage_path, sales_count, created_at, updated_at, created_by FROM assets WHERE created_by = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*Asset
	for rows.Next() {
		asset := &Asset{}
		if err := rows.Scan(&asset.ID, &asset.Title, &asset.Description, &asset.Price, &asset.Category, &asset.AIToolUsed, &asset.Resolution, &asset.PreviewURL, &asset.StoragePath, &asset.SalesCount, &asset.CreatedAt, &asset.UpdatedAt, &asset.CreatedBy); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

// SearchAssets searches for assets with pagination and filters
func (r *PostgreSQLAssetRepository) SearchAssets(ctx context.Context, queryStr string, category string, sortBy string, limit int, offset int) ([]*Asset, error) {
	query := `SELECT id, title, description, price, category, ai_tool_used, resolution, preview_url, storage_path, sales_count, created_at, updated_at, created_by FROM assets WHERE 1=1`
	var args []interface{}
	argId := 1

	if queryStr != "" {
		query += fmt.Sprintf(` AND (title ILIKE $%d OR description ILIKE $%d)`, argId, argId)
		args = append(args, "%"+queryStr+"%")
		argId++
	}

	if category != "" {
		query += fmt.Sprintf(` AND category = $%d`, argId)
		args = append(args, category)
		argId++
	}

	switch sortBy {
	case "price_asc":
		query += ` ORDER BY price ASC`
	case "price_desc":
		query += ` ORDER BY price DESC`
	default:
		query += ` ORDER BY created_at DESC`
	}

	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argId, argId+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*Asset
	for rows.Next() {
		asset := &Asset{}
		if err := rows.Scan(&asset.ID, &asset.Title, &asset.Description, &asset.Price, &asset.Category, &asset.AIToolUsed, &asset.Resolution, &asset.PreviewURL, &asset.StoragePath, &asset.SalesCount, &asset.CreatedAt, &asset.UpdatedAt, &asset.CreatedBy); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

// UpdateAsset updates an existing asset
func (r *PostgreSQLAssetRepository) UpdateAsset(ctx context.Context, asset *Asset) error {
	query := `UPDATE assets SET title = $1, description = $2, price = $3, category = $4, ai_tool_used = $5, resolution = $6, preview_url = $7, storage_path = $8, updated_at = $9 WHERE id = $10`
	_, err := r.db.ExecContext(ctx, query, asset.Title, asset.Description, asset.Price, asset.Category, asset.AIToolUsed, asset.Resolution, asset.PreviewURL, asset.StoragePath, asset.UpdatedAt, asset.ID)
	return err
}

// DeleteAsset deletes an asset from the database
func (r *PostgreSQLAssetRepository) DeleteAsset(ctx context.Context, id string) error {
	query := `DELETE FROM assets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetAllAssets retrieves all assets
func (r *PostgreSQLAssetRepository) GetAllAssets(ctx context.Context) ([]*Asset, error) {
	query := `SELECT id, title, description, price, category, ai_tool_used, resolution, preview_url, storage_path, sales_count, created_at, updated_at, created_by FROM assets ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*Asset
	for rows.Next() {
		asset := &Asset{}
		if err := rows.Scan(&asset.ID, &asset.Title, &asset.Description, &asset.Price, &asset.Category, &asset.AIToolUsed, &asset.Resolution, &asset.PreviewURL, &asset.StoragePath, &asset.SalesCount, &asset.CreatedAt, &asset.UpdatedAt, &asset.CreatedBy); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

// IncrementSalesCount increments the sales count for an asset
func (r *PostgreSQLAssetRepository) IncrementSalesCount(ctx context.Context, id string) error {
	query := `UPDATE assets SET sales_count = sales_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}