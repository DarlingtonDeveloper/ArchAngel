package repository

import (
	"time"
)

// Analysis represents an analysis in the database
type Analysis struct {
	ID         string    `db:"id"`
	Language   string    `db:"language"`
	Code       string    `db:"code"`
	Context    string    `db:"context"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
	UserID     string    `db:"user_id"`
	ResultJSON string    `db:"result_json"`
}

// User represents a user in the database
type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	APIKey    string    `db:"api_key"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Active    bool      `db:"active"`
}

// Organization represents an organization in the database
type Organization struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	APIKey    string    `db:"api_key"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Active    bool      `db:"active"`
}

// TeamMember represents a user's membership in an organization
type TeamMember struct {
	UserID         string    `db:"user_id"`
	OrganizationID string    `db:"organization_id"`
	Role           string    `db:"role"` // "admin", "member", etc.
	JoinedAt       time.Time `db:"joined_at"`
}

// Rule represents a custom rule in the database
type Rule struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Pattern     string    `db:"pattern"`
	Severity    string    `db:"severity"`
	Language    string    `db:"language"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedBy   string    `db:"created_by"`
	IsActive    bool      `db:"is_active"`
}