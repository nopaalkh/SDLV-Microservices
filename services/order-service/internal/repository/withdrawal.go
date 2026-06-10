package repository

import (
	"context"
	"database/sql"
	"time"
)

// Status constants for withdrawals
const (
	WithdrawalStatusPending  = "PENDING"
	WithdrawalStatusApproved = "APPROVED"
	WithdrawalStatusRejected = "REJECTED"
)

// Withdrawal represents a creator's withdrawal request
type Withdrawal struct {
	ID            string    `db:"id" json:"id"`
	CreatorID     string    `db:"creator_id" json:"creator_id"`
	Amount        float64   `db:"amount" json:"amount"`
	Status        string    `db:"status" json:"status"`
	BankCode      string    `db:"bank_code" json:"bank_code"`
	AccountNumber string    `db:"account_number" json:"account_number"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// WithdrawalRepository defines the interface
type WithdrawalRepository interface {
	CreateWithdrawal(ctx context.Context, withdrawal *Withdrawal) error
	GetWithdrawalsByCreator(ctx context.Context, creatorID string) ([]*Withdrawal, error)
	GetAllWithdrawals(ctx context.Context) ([]*Withdrawal, error)
	UpdateWithdrawalStatus(ctx context.Context, id string, status string) error
	FindWithdrawalByID(ctx context.Context, id string) (*Withdrawal, error)
}

// PostgreSQLWithdrawalRepository implements WithdrawalRepository
type PostgreSQLWithdrawalRepository struct {
	db *sql.DB
}

// NewPostgreSQLWithdrawalRepository creates a new instance
func NewPostgreSQLWithdrawalRepository(db *sql.DB) *PostgreSQLWithdrawalRepository {
	return &PostgreSQLWithdrawalRepository{db: db}
}

// CreateWithdrawal creates a new withdrawal request
func (r *PostgreSQLWithdrawalRepository) CreateWithdrawal(ctx context.Context, w *Withdrawal) error {
	query := `INSERT INTO withdrawals (id, creator_id, amount, status, bank_code, account_number, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, w.ID, w.CreatorID, w.Amount, w.Status, w.BankCode, w.AccountNumber, w.CreatedAt, w.UpdatedAt)
	return err
}

// GetWithdrawalsByCreator retrieves all withdrawals for a creator
func (r *PostgreSQLWithdrawalRepository) GetWithdrawalsByCreator(ctx context.Context, creatorID string) ([]*Withdrawal, error) {
	query := `SELECT id, creator_id, amount, status, bank_code, account_number, created_at, updated_at FROM withdrawals WHERE creator_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*Withdrawal
	for rows.Next() {
		w := &Withdrawal{}
		if err := rows.Scan(&w.ID, &w.CreatorID, &w.Amount, &w.Status, &w.BankCode, &w.AccountNumber, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, rows.Err()
}

// GetAllWithdrawals retrieves all withdrawals (for admin)
func (r *PostgreSQLWithdrawalRepository) GetAllWithdrawals(ctx context.Context) ([]*Withdrawal, error) {
	query := `SELECT id, creator_id, amount, status, bank_code, account_number, created_at, updated_at FROM withdrawals ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*Withdrawal
	for rows.Next() {
		w := &Withdrawal{}
		if err := rows.Scan(&w.ID, &w.CreatorID, &w.Amount, &w.Status, &w.BankCode, &w.AccountNumber, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, rows.Err()
}

// UpdateWithdrawalStatus updates the status
func (r *PostgreSQLWithdrawalRepository) UpdateWithdrawalStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE withdrawals SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// FindWithdrawalByID finds a withdrawal by id
func (r *PostgreSQLWithdrawalRepository) FindWithdrawalByID(ctx context.Context, id string) (*Withdrawal, error) {
	query := `SELECT id, creator_id, amount, status, bank_code, account_number, created_at, updated_at FROM withdrawals WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	w := &Withdrawal{}
	err := row.Scan(&w.ID, &w.CreatorID, &w.Amount, &w.Status, &w.BankCode, &w.AccountNumber, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return w, nil
}
