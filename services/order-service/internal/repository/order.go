package repository

import (
	"context"
	"database/sql"
	"time"
)

// OrderStatus defines the status of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusFailed    OrderStatus = "FAILED"
	StatusCancelled OrderStatus = "CANCELLED"
)

// Order represents an order in the system
type Order struct {
	ID                  string      `db:"id" json:"id"`
	UserID              string      `db:"user_id" json:"user_id"`
	AssetID             string      `db:"asset_id" json:"asset_id"`
	CreatorID           string      `db:"creator_id" json:"creator_id"`
	Quantity            int         `db:"quantity" json:"quantity"`
	TotalPrice          float64     `db:"total_price" json:"total_price"`
	PlatformFee         float64     `db:"platform_fee" json:"platform_fee"`
	CreatorRevenue      float64     `db:"creator_revenue" json:"creator_revenue"`
	Status              OrderStatus `db:"status" json:"status"`
	SecureDownloadToken string      `db:"secure_download_token" json:"secure_download_token"`
	CreatedAt           time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time   `db:"updated_at" json:"updated_at"`
}

// OrderRepository defines the interface for order repository operations
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	FindOrderByID(ctx context.Context, id string) (*Order, error)
	FindOrdersByUserID(ctx context.Context, userID string) ([]*Order, error)
	FindOrdersByCreatorID(ctx context.Context, creatorID string) ([]*Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status OrderStatus) error
	UpdateOrderPayment(ctx context.Context, id string, status OrderStatus, downloadToken string) error
	GetAllOrders(ctx context.Context) ([]*Order, error)
	GetTotalPlatformRevenue(ctx context.Context) (float64, error)
}

// PostgreSQLOrderRepository implements OrderRepository using PostgreSQL
type PostgreSQLOrderRepository struct {
	db *sql.DB
}

// NewPostgreSQLOrderRepository creates a new PostgreSQLOrderRepository
func NewPostgreSQLOrderRepository(db *sql.DB) *PostgreSQLOrderRepository {
	return &PostgreSQLOrderRepository{db: db}
}

func (r *PostgreSQLOrderRepository) CreateOrder(ctx context.Context, order *Order) error {
	query := `INSERT INTO orders (id, user_id, asset_id, creator_id, quantity, total_price, platform_fee, creator_revenue, status, secure_download_token, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query, order.ID, order.UserID, order.AssetID, order.CreatorID, order.Quantity, order.TotalPrice, order.PlatformFee, order.CreatorRevenue, order.Status, order.SecureDownloadToken, order.CreatedAt, order.UpdatedAt)
	return err
}

// FindOrderByID finds an order by ID
func (r *PostgreSQLOrderRepository) FindOrderByID(ctx context.Context, id string) (*Order, error) {
	query := `SELECT id, user_id, asset_id, creator_id, quantity, total_price, platform_fee, creator_revenue, status, secure_download_token, created_at, updated_at FROM orders WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var order Order
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.AssetID,
		&order.CreatorID,
		&order.Quantity,
		&order.TotalPrice,
		&order.PlatformFee,
		&order.CreatorRevenue,
		&order.Status,
		&order.SecureDownloadToken,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// FindOrdersByUserID finds all orders for a specific user
func (r *PostgreSQLOrderRepository) FindOrdersByUserID(ctx context.Context, userID string) ([]*Order, error) {
	query := `SELECT id, user_id, asset_id, creator_id, quantity, total_price, platform_fee, creator_revenue, status, secure_download_token, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.AssetID,
			&order.CreatorID,
			&order.Quantity,
			&order.TotalPrice,
			&order.PlatformFee,
			&order.CreatorRevenue,
			&order.Status,
			&order.SecureDownloadToken,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetTotalPlatformRevenue returns the total platform fee from all PAID orders
func (r *PostgreSQLOrderRepository) GetTotalPlatformRevenue(ctx context.Context) (float64, error) {
	query := `SELECT COALESCE(SUM(platform_fee), 0) FROM orders WHERE status = $1`
	var total float64
	err := r.db.QueryRowContext(ctx, query, StatusPaid).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// FindOrdersByCreatorID finds all orders for a specific creator
func (r *PostgreSQLOrderRepository) FindOrdersByCreatorID(ctx context.Context, creatorID string) ([]*Order, error) {
	query := `SELECT id, user_id, asset_id, creator_id, quantity, total_price, platform_fee, creator_revenue, status, secure_download_token, created_at, updated_at FROM orders WHERE creator_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.AssetID,
			&order.CreatorID,
			&order.Quantity,
			&order.TotalPrice,
			&order.PlatformFee,
			&order.CreatorRevenue,
			&order.Status,
			&order.SecureDownloadToken,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// UpdateOrderStatus updates an order's status
func (r *PostgreSQLOrderRepository) UpdateOrderStatus(ctx context.Context, id string, status OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// UpdateOrderPayment updates an order's payment information
func (r *PostgreSQLOrderRepository) UpdateOrderPayment(ctx context.Context, id string, status OrderStatus, downloadToken string) error {
	query := `UPDATE orders SET status = $1, secure_download_token = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, status, downloadToken, time.Now(), id)
	return err
}

// GetAllOrders retrieves all orders
func (r *PostgreSQLOrderRepository) GetAllOrders(ctx context.Context) ([]*Order, error) {
	query := `SELECT id, user_id, asset_id, creator_id, quantity, total_price, platform_fee, creator_revenue, status, secure_download_token, created_at, updated_at FROM orders ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.AssetID,
			&order.CreatorID,
			&order.Quantity,
			&order.TotalPrice,
			&order.PlatformFee,
			&order.CreatorRevenue,
			&order.Status,
			&order.SecureDownloadToken,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, rows.Err()
}
