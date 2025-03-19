package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// BaseAnalyzer provides common functionality for all language-specific analyzers
type BaseAnalyzer struct {
	Config map[string]string
}

// NewBaseAnalyzer creates a new BaseAnalyzer with the provided configuration
func NewBaseAnalyzer(config map[string]string) *BaseAnalyzer {
	if config == nil {
		config = make(map[string]string)
	}
	
	return &BaseAnalyzer{
		Config: config,
	}
}

// GetTimeout extracts the timeout from the configuration or returns a default value
func (b *BaseAnalyzer) GetTimeout() time.Duration {
	if timeoutStr, ok := b.Config["timeout"]; ok && timeoutStr != "" {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			return t
		}
	}
	return 10 * time.Second // Default timeout
}

// WrapError wraps an error with additional context
func (b *BaseAnalyzer) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// CreateTimeoutContext creates a context with timeout from the parent context
func (b *BaseAnalyzer) CreateTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, b.GetTimeout())
}

// CreateTempFile creates a temporary file with the provided content and extension
func (b *BaseAnalyzer) CreateTempFile(prefix, ext, content string) (string, error) {
	// Implementation would be here, but we'll likely use ioutil.TempFile
	// This is a stub for now
	return "", nil
}

// MapSeverity maps internal severity levels to standardized ones
func (b *BaseAnalyzer) MapSeverity(severityStr string) string {
	switch strings.ToLower(severityStr) {
	case "error", "critical", "fatal":
		return "error"
	case "warning", "warn":
		return "warning"
	case "info", "information":
		return "info"
	case "hint", "suggestion":
		return "suggestion"
	default:
		return "info"
	}
}

// CommonAnalyzeWrapper provides a template for the Analyze method
func (b *BaseAnalyzer) CommonAnalyzeWrapper(
	ctx context.Context,
	code string,
	options map[string]interface{},
	findIssuesFunc func(context.Context, string, map[string]interface{}) ([]Issue, error),
	suggestFixesFunc func(context.Context, string, []Issue) ([]Issue, error),
) (*AnalysisResult, error) {
	// Create timeout context
	ctxWithTimeout, cancel := b.CreateTimeoutContext(ctx)
	defer cancel()
	
	// Find issues
	issues, err := findIssuesFunc(ctxWithTimeout, code, options)
	if err != nil {
		return nil, b.WrapError(err, "error finding issues")
	}
	
	// Generate suggestions
	suggestions, err := suggestFixesFunc(ctxWithTimeout, code, issues)
	if err != nil {
		// Log error but continue with just the issues
		// In a real implementation, you'd use a proper logger
		fmt.Printf("Warning: error generating suggestions: %v\n", err)
	}
	
	return &AnalysisResult{
		Issues:      issues,
		Suggestions: suggestions,
	}, nil
}