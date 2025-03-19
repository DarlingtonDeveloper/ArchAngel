package ai

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/codehawk/backend/pkg/analyzer"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Value      []analyzer.Issue
	Expiration time.Time
}

// CachedAIService wraps an AIService with caching functionality
type CachedAIService struct {
	service AISuggestionService
	cache   map[string]CacheEntry
	mu      sync.RWMutex
	config  Config
}

// NewCachedAIService creates a new cached AI service
func NewCachedAIService(service AISuggestionService, config Config) AISuggestionService {
	if !config.EnableCaching {
		return service // Return the original service if caching is disabled
	}
	
	return &CachedAIService{
		service: service,
		cache:   make(map[string]CacheEntry),
		config:  config,
	}
}

// GetSuggestions generates AI-powered suggestions for code improvement
func (s *CachedAIService) GetSuggestions(ctx context.Context, code string, language string, issues []analyzer.Issue) ([]analyzer.Issue, error) {
	if !s.config.EnableCaching {
		return s.service.GetSuggestions(ctx, code, language, issues)
	}
	
	// Create a unique cache key for this request
	key := s.createCacheKey(code, language, issues)
	
	// Try to get from cache first
	s.mu.RLock()
	entry, found := s.cache[key]
	s.mu.RUnlock()
	
	if found && time.Now().Before(entry.Expiration) {
		return entry.Value, nil
	}
	
	// Call the underlying service if not found in cache
	suggestions, err := s.service.GetSuggestions(ctx, code, language, issues)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	s.mu.Lock()
	s.cache[key] = CacheEntry{
		Value:      suggestions,
		Expiration: time.Now().Add(s.config.CacheTTL),
	}
	s.mu.Unlock()
	
	return suggestions, nil
}

// GetCodeExplanation generates an explanation for code
func (s *CachedAIService) GetCodeExplanation(ctx context.Context, code string, language string) (string, error) {
	if !s.config.EnableCaching {
		return s.service.GetCodeExplanation(ctx, code, language)
	}
	
	// Not implementing caching for explanations yet, as they're stored differently
	// In a real implementation, we would need to add a separate cache for string responses
	return s.service.GetCodeExplanation(ctx, code, language)
}

// GetCodeRefactoring generates a refactored version of the code
func (s *CachedAIService) GetCodeRefactoring(ctx context.Context, code string, language string) (string, error) {
	if !s.config.EnableCaching {
		return s.service.GetCodeRefactoring(ctx, code, language)
	}
	
	// Not implementing caching for refactorings yet
	return s.service.GetCodeRefactoring(ctx, code, language)
}

// createCacheKey creates a unique key for the cache
func (s *CachedAIService) createCacheKey(code string, language string, issues []analyzer.Issue) string {
	// Create a hash of the inputs
	hasher := md5.New()
	
	// Add code and language to hash
	hasher.Write([]byte(code))
	hasher.Write([]byte(language))
	
	// Add issues to hash
	issuesJSON, _ := json.Marshal(issues)
	hasher.Write(issuesJSON)
	
	return hex.EncodeToString(hasher.Sum(nil))
}

// CleanExpiredEntries removes expired entries from the cache
func (s *CachedAIService) CleanExpiredEntries() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for key, entry := range s.cache {
		if now.After(entry.Expiration) {
			delete(s.cache, key)
		}
	}
}

// StartCacheCleanup starts a goroutine to periodically clean up expired cache entries
func (s *CachedAIService) StartCacheCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			s.CleanExpiredEntries()
		}
	}()
}