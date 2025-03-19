package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/codehawk/backend/pkg/analyzer"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	
	// InitialDelay is the initial delay before the first retry
	InitialDelay time.Duration
	
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	
	// BackoffFactor is the factor by which the delay increases after each retry
	BackoffFactor float64
}

// NewDefaultRetryConfig creates a default retry configuration
func NewDefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryableAIService wraps an AIService with retry functionality
type RetryableAIService struct {
	service AISuggestionService
	config  RetryConfig
}

// NewRetryableAIService creates a new retryable AI service
func NewRetryableAIService(service AISuggestionService, config RetryConfig) AISuggestionService {
	return &RetryableAIService{
		service: service,
		config:  config,
	}
}

// GetSuggestions generates AI-powered suggestions for code improvement
func (s *RetryableAIService) GetSuggestions(ctx context.Context, code string, language string, issues []analyzer.Issue) ([]analyzer.Issue, error) {
	return s.withRetry(ctx, func() ([]analyzer.Issue, error) {
		return s.service.GetSuggestions(ctx, code, language, issues)
	})
}

// GetCodeExplanation generates an explanation for code
func (s *RetryableAIService) GetCodeExplanation(ctx context.Context, code string, language string) (string, error) {
	result, err := s.withRetryString(ctx, func() (string, error) {
		return s.service.GetCodeExplanation(ctx, code, language)
	})
	
	return result, err
}

// GetCodeRefactoring generates a refactored version of the code
func (s *RetryableAIService) GetCodeRefactoring(ctx context.Context, code string, language string) (string, error) {
	result, err := s.withRetryString(ctx, func() (string, error) {
		return s.service.GetCodeRefactoring(ctx, code, language)
	})
	
	return result, err
}

// withRetry retries a function that returns issues with exponential backoff
func (s *RetryableAIService) withRetry(ctx context.Context, fn func() ([]analyzer.Issue, error)) ([]analyzer.Issue, error) {
	var lastErr error
	
	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		// First attempt is not a retry
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := s.calculateDelay(attempt)
			
			// Create a timer
			timer := time.NewTimer(delay)
			
			// Wait for the timer or context cancellation
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
				// Continue with retry
			}
		}
		
		// Call the function
		result, err := fn()
		if err == nil {
			return result, nil
		}
		
		// Check if error is retryable
		if !s.isRetryableError(err) {
			return nil, err
		}
		
		lastErr = err
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// withRetryString retries a function that returns a string with exponential backoff
func (s *RetryableAIService) withRetryString(ctx context.Context, fn func() (string, error)) (string, error) {
	var lastErr error
	
	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		// First attempt is not a retry
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := s.calculateDelay(attempt)
			
			// Create a timer
			timer := time.NewTimer(delay)
			
			// Wait for the timer or context cancellation
			select {
			case <-ctx.Done():
				timer.Stop()
				return "", ctx.Err()
			case <-timer.C:
				// Continue with retry
			}
		}
		
		// Call the function
		result, err := fn()
		if err == nil {
			return result, nil
		}
		
		// Check if error is retryable
		if !s.isRetryableError(err) {
			return "", err
		}
		
		lastErr = err
	}
	
	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateDelay calculates the delay for a retry attempt
func (s *RetryableAIService) calculateDelay(attempt int) time.Duration {
	// Calculate delay with exponential backoff
	delay := s.config.InitialDelay * time.Duration(float64(attempt)*s.config.BackoffFactor)
	
	// Cap at max delay
	if delay > s.config.MaxDelay {
		delay = s.config.MaxDelay
	}
	
	return delay
}

// isRetryableError determines if an error is retryable
func (s *RetryableAIService) isRetryableError(err error) bool {
	// Retryable errors generally include network issues and server errors
	
	// Check for HTTP errors
	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		// Retry on server errors (5xx) and rate limiting (429)
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == http.StatusTooManyRequests
	}
	
	// Check for context errors (don't retry)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	
	// Retry on network errors and timeouts
	return true
}

// HttpError represents an HTTP error
type HttpError struct {
	StatusCode int
	Message    string
}

// Error returns the error message
func (e *HttpError) Error() string {
	return fmt.Sprintf("HTTP error: %d %s", e.StatusCode, e.Message)
}