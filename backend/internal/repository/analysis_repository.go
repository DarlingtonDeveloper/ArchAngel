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
	// ErrAnalysisNotFound is returned when an analysis is not found
	ErrAnalysisNotFound = errors.New("analysis not found")
)

// AnalysisRepository defines methods for interacting with analyses
type AnalysisRepository interface {
	// StoreAnalysis stores an analysis in the repository
	StoreAnalysis(ctx context.Context, analysis *Analysis) error
	
	// GetAnalysis retrieves an analysis by ID
	GetAnalysis(ctx context.Context, id string) (*Analysis, error)
	
	// ListAnalyses lists analyses for a user
	ListAnalyses(ctx context.Context, userID string, limit, offset int) ([]*Analysis, error)
	
	// ListAnalysesForOrganization lists analyses for an organization
	ListAnalysesForOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*Analysis, error)
	
	// DeleteAnalysis deletes an analysis
	DeleteAnalysis(ctx context.Context, id string) error
	
	// CountAnalyses counts the total number of analyses for a user
	CountAnalyses(ctx context.Context, userID string) (int, error)
}

// PostgresAnalysisRepository is a PostgreSQL implementation of AnalysisRepository
type PostgresAnalysisRepository struct {
	db *sqlx.DB
}

// NewPostgresAnalysisRepository creates a new PostgresAnalysisRepository
func NewPostgresAnalysisRepository(db *sqlx.DB) *PostgresAnalysisRepository {
	return &PostgresAnalysisRepository{
		db: db,
	}
}

// StoreAnalysis stores an analysis in the database
func (r *PostgresAnalysisRepository) StoreAnalysis(ctx context.Context, analysis *Analysis) error {
	query := `
		INSERT INTO analyses (id, language, code, context, status, created_at, user_id, result_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET language = $2, code = $3, context = $4, status = $5, result_json = $8
	`
	
	if analysis.CreatedAt.IsZero() {
		analysis.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		analysis.ID,
		analysis.Language,
		analysis.Code,
		analysis.Context,
		analysis.Status,
		analysis.CreatedAt,
		analysis.UserID,
		analysis.ResultJSON,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store analysis: %w", err)
	}
	
	return nil
}

// GetAnalysis retrieves an analysis by ID
func (r *PostgresAnalysisRepository) GetAnalysis(ctx context.Context, id string) (*Analysis, error) {
	query := `
		SELECT id, language, code, context, status, created_at, user_id, result_json
		FROM analyses
		WHERE id = $1
	`
	
	var analysis Analysis
	err := r.db.GetContext(ctx, &analysis, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnalysisNotFound
		}
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}
	
	return &analysis, nil
}

// ListAnalyses lists analyses for a user
func (r *PostgresAnalysisRepository) ListAnalyses(ctx context.Context, userID string, limit, offset int) ([]*Analysis, error) {
	query := `
		SELECT id, language, context, status, created_at, user_id
		FROM analyses
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	if offset < 0 {
		offset = 0
	}
	
	var analyses []*Analysis
	err := r.db.SelectContext(ctx, &analyses, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list analyses: %w", err)
	}
	
	return analyses, nil
}

// ListAnalysesForOrganization lists analyses for an organization
func (r *PostgresAnalysisRepository) ListAnalysesForOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*Analysis, error) {
	query := `
		SELECT a.id, a.language, a.context, a.status, a.created_at, a.user_id
		FROM analyses a
		JOIN team_members tm ON a.user_id = tm.user_id
		WHERE tm.organization_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	if offset < 0 {
		offset = 0
	}
	
	var analyses []*Analysis
	err := r.db.SelectContext(ctx, &analyses, query, organizationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list analyses for organization: %w", err)
	}
	
	return analyses, nil
}

// DeleteAnalysis deletes an analysis
func (r *PostgresAnalysisRepository) DeleteAnalysis(ctx context.Context, id string) error {
	query := `DELETE FROM analyses WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ErrAnalysisNotFound
	}
	
	return nil
}

// CountAnalyses counts the total number of analyses for a user
func (r *PostgresAnalysisRepository) CountAnalyses(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM analyses WHERE user_id = $1`
	
	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count analyses: %w", err)
	}
	
	return count, nil
}