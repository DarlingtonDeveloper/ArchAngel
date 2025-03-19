package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/codehawk/backend/internal/repository"
	"github.com/yourusername/codehawk/backend/pkg/ai"
	"github.com/yourusername/codehawk/backend/pkg/analyzer"
)

// AnalysisService handles code analysis operations
type AnalysisService struct {
	linterRegistry *analyzer.LinterRegistry
	analysisRepo   repository.AnalysisRepository
	aiService      ai.AISuggestionService
	aiEnabled      bool
}

// AnalysisRequest represents a request to analyze code
type AnalysisRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	Context  string                 `json:"context"`
	UserID   string                 `json:"user_id,omitempty"`
	Options  map[string]interface{} `json:"options"`
}

// AnalysisResponse represents the response from a code analysis
type AnalysisResponse struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
	Language    string             `json:"language"`
	Context     string             `json:"context"`
	Timestamp   string             `json:"timestamp"`
	Issues      []analyzer.Issue   `json:"issues"`
	Suggestions []analyzer.Issue   `json:"suggestions"`
	AIEnhanced  bool               `json:"ai_enhanced,omitempty"`
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(
	linterRegistry *analyzer.LinterRegistry,
	analysisRepo repository.AnalysisRepository,
	aiService ai.AISuggestionService,
	aiEnabled bool,
) *AnalysisService {
	return &AnalysisService{
		linterRegistry: linterRegistry,
		analysisRepo:   analysisRepo,
		aiService:      aiService,
		aiEnabled:      aiEnabled,
	}
}

// AnalyzeCode analyzes code and returns the results
func (s *AnalysisService) AnalyzeCode(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// Check if we have a linter for this language
	linter, ok := s.linterRegistry.GetLinter(req.Language)
	if !ok {
		// Fall back to generic analysis
		return s.performGenericAnalysis(ctx, req)
	}

	// Analyze the code using the appropriate linter
	result, err := linter.Analyze(ctx, req.Code, req.Options)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// If AI suggestions are enabled, enhance with AI
	if s.aiEnabled && shouldUseAI(req.Options) {
		aiSuggestions, err := s.aiService.GetSuggestions(ctx, req.Code, req.Language, result.Issues)
		if err != nil {
			// Log error but continue without AI suggestions
			fmt.Printf("Error getting AI suggestions: %v\n", err)
		} else {
			result.Suggestions = append(result.Suggestions, aiSuggestions...)
		}
	}

	// Create response
	analysisID := generateAnalysisID()
	timestamp := time.Now().Format(time.RFC3339)
	
	response := &AnalysisResponse{
		ID:          analysisID,
		Status:      "success",
		Language:    req.Language,
		Context:     req.Context,
		Timestamp:   timestamp,
		Issues:      result.Issues,
		Suggestions: result.Suggestions,
		AIEnhanced:  s.aiEnabled && shouldUseAI(req.Options),
	}

	// Store the results if we have a repository
	if s.analysisRepo != nil && req.UserID != "" {
		if err := s.storeAnalysisResult(ctx, response, req); err != nil {
			// Log error but continue
			fmt.Printf("Error storing analysis: %v\n", err)
		}
	}

	return response, nil
}

// GetAnalysisById retrieves an analysis by ID
func (s *AnalysisService) GetAnalysisById(ctx context.Context, id string) (*AnalysisResponse, error) {
	if s.analysisRepo == nil {
		return nil, fmt.Errorf("no repository configured")
	}

	analysis, err := s.analysisRepo.GetAnalysis(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	// Parse the JSON result
	var response AnalysisResponse
	if err := json.Unmarshal([]byte(analysis.ResultJSON), &response); err != nil {
		return nil, fmt.Errorf("failed to parse analysis result: %w", err)
	}

	return &response, nil
}

// ListAnalysisForUser retrieves analyses for a user
func (s *AnalysisService) ListAnalysisForUser(ctx context.Context, userID string, limit, offset int) ([]*AnalysisResponse, error) {
	if s.analysisRepo == nil {
		return nil, fmt.Errorf("no repository configured")
	}

	analyses, err := s.analysisRepo.ListAnalyses(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list analyses: %w", err)
	}

	// Convert to response objects
	responses := make([]*AnalysisResponse, 0, len(analyses))
	for _, analysis := range analyses {
		var response AnalysisResponse
		if err := json.Unmarshal([]byte(analysis.ResultJSON), &response); err != nil {
			// Skip this one if parsing fails
			continue
		}
		responses = append(responses, &response)
	}

	return responses, nil
}

// storeAnalysisResult stores an analysis result in the repository
func (s *AnalysisService) storeAnalysisResult(ctx context.Context, response *AnalysisResponse, req AnalysisRequest) error {
	// Convert response to JSON
	resultJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis result: %w", err)
	}

	// Create analysis record
	analysis := &repository.Analysis{
		ID:         response.ID,
		Language:   response.Language,
		Code:       req.Code,
		Context:    response.Context,
		Status:     response.Status,
		CreatedAt:  time.Now(),
		UserID:     req.UserID,
		ResultJSON: string(resultJSON),
	}

	// Store in repository
	return s.analysisRepo.StoreAnalysis(ctx, analysis)
}

// performGenericAnalysis performs a generic analysis when no specific linter is available
func (s *AnalysisService) performGenericAnalysis(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// Create a simple generic analysis
	// In a real implementation, this would do more sophisticated analysis
	
	issues := []analyzer.Issue{
		{
			Line:     1,
			Message:  "Generic analysis performed - no specific linter available for " + req.Language,
			Severity: "info",
			RuleID:   "generic-analysis",
		},
	}

	// Create response
	analysisID := generateAnalysisID()
	timestamp := time.Now().Format(time.RFC3339)
	
	response := &AnalysisResponse{
		ID:          analysisID,
		Status:      "success",
		Language:    req.Language,
		Context:     req.Context,
		Timestamp:   timestamp,
		Issues:      issues,
		Suggestions: []analyzer.Issue{},
	}

	// Store the results if we have a repository
	if s.analysisRepo != nil && req.UserID != "" {
		if err := s.storeAnalysisResult(ctx, response, req); err != nil {
			// Log error but continue
			fmt.Printf("Error storing analysis: %v\n", err)
		}
	}

	return response, nil
}

// shouldUseAI determines if AI suggestions should be used based on options
func shouldUseAI(options map[string]interface{}) bool {
	if options == nil {
		return true // Default to enabled
	}
	
	if ai, ok := options["ai_suggestions"].(bool); ok {
		return ai
	}
	
	return true // Default to enabled
}

// generateAnalysisID generates a unique ID for an analysis
func generateAnalysisID() string {
	return "analysis-" + uuid.New().String()
}