package repository

import (
	"context"
	"database/sql"
	"time"
)

// UserRole defines the role of a user in the system
type UserRole string

const (
	RoleBuyer   UserRole = "BUYER"
	RoleCreator UserRole = "CREATOR"
	RoleAdmin   UserRole = "ADMIN"
)

// UserStatus defines the account status
type UserStatus string

const (
	StatusActive    UserStatus = "ACTIVE"
	StatusSuspended UserStatus = "SUSPENDED"
)

// User represents a user in the system
type User struct {
	ID        string     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Email     string     `db:"email" json:"email"`
	Password  string     `db:"password" json:"-"`
	Role      UserRole   `db:"role" json:"role"`
	Status    UserStatus `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	GetAllUsers(ctx context.Context) ([]*User, error)
	UpdateUserStatus(ctx context.Context, id string, status UserStatus) error
	DeleteUser(ctx context.Context, id string) error
}

// PostgreSQLUserRepository implements UserRepository using PostgreSQL
type PostgreSQLUserRepository struct {
	db *sql.DB
}

// NewPostgreSQLUserRepository creates a new PostgreSQLUserRepository
func NewPostgreSQLUserRepository(db *sql.DB) *PostgreSQLUserRepository {
	return &PostgreSQLUserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *PostgreSQLUserRepository) CreateUser(ctx context.Context, user *User) error {
	query := `INSERT INTO users (id, name, email, password, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.Password, user.Role, user.Status, user.CreatedAt, user.UpdatedAt)
	return err
}

// FindUserByEmail finds a user by email
func (r *PostgreSQLUserRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, name, email, password, role, status, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// FindUserByID finds a user by ID
func (r *PostgreSQLUserRepository) FindUserByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, name, email, password, role, status, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user in the database
func (r *PostgreSQLUserRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `UPDATE users SET name = $1, email = $2, password = $3, role = $4, status = $5, updated_at = $6 WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.Password, user.Role, user.Status, user.UpdatedAt, user.ID)
	return err
}

// GetAllUsers retrieves all users from the database
func (r *PostgreSQLUserRepository) GetAllUsers(ctx context.Context) ([]*User, error) {
	query := `SELECT id, name, email, password, role, status, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// UpdateUserStatus updates a user's account status
func (r *PostgreSQLUserRepository) UpdateUserStatus(ctx context.Context, id string, status UserStatus) error {
	query := `UPDATE users SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// DeleteUser deletes a user by ID
func (r *PostgreSQLUserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}