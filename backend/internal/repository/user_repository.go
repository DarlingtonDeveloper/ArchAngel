package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	
	// ErrDuplicateEmail is returned when a user with the same email already exists
	ErrDuplicateEmail = errors.New("user with this email already exists")
	
	// ErrInvalidAPIKey is returned when an API key is invalid
	ErrInvalidAPIKey = errors.New("invalid API key")
)

// UserRepository defines methods for interacting with users
type UserRepository interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *User) error
	
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id string) (*User, error)
	
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	
	// GetUserByAPIKey retrieves a user by API key
	GetUserByAPIKey(ctx context.Context, apiKey string) (*User, error)
	
	// UpdateUser updates a user
	UpdateUser(ctx context.Context, user *User) error
	
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id string) error
	
	// ListUsers lists all users
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
	
	// CountUsers counts the total number of users
	CountUsers(ctx context.Context) (int, error)
}

// PostgresUserRepository is a PostgreSQL implementation of UserRepository
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// CreateUser creates a new user
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	// Check if user with email already exists
	existingUser, err := r.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return ErrDuplicateEmail
	}
	
	query := `
		INSERT INTO users (id, email, name, api_key, created_at, updated_at, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}
	
	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.APIKey,
		user.CreatedAt,
		user.UpdatedAt,
		user.Active,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetUser retrieves a user by ID
func (r *PostgresUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, name, api_key, created_at, updated_at, active
		FROM users
		WHERE id = $1
	`
	
	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, name, api_key, created_at, updated_at, active
		FROM users
		WHERE email = $1
	`
	
	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return &user, nil
}

// GetUserByAPIKey retrieves a user by API key
func (r *PostgresUserRepository) GetUserByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	query := `
		SELECT id, email, name, api_key, created_at, updated_at, active
		FROM users
		WHERE api_key = $1 AND active = true
	`
	
	var user User
	err := r.db.GetContext(ctx, &user, query, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidAPIKey
		}
		return nil, fmt.Errorf("failed to get user by API key: %w", err)
	}
	
	return &user, nil
}

// UpdateUser updates a user
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = $2, name = $3, api_key = $4, updated_at = $5, active = $6
		WHERE id = $1
	`
	
	user.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.APIKey,
		user.UpdatedAt,
		user.Active,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

// DeleteUser deletes a user
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

// ListUsers lists all users
func (r *PostgresUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, email, name, created_at, updated_at, active
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	if offset < 0 {
		offset = 0
	}
	
	var users []*User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	
	return users, nil
}

// CountUsers counts the total number of users
func (r *PostgresUserRepository) CountUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	return count, nil
}