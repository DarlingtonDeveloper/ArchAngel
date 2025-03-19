package ai

import (
	"time"
)

// Config holds configuration for the AI service
type Config struct {
	// Endpoint is the URL of the AI service API
	Endpoint string
	
	// APIKey is the authentication key for the AI service
	APIKey string
	
	// OrgID is the organization ID for billing (used by some providers like OpenAI)
	OrgID string
	
	// Timeout is the maximum time to wait for AI responses
	Timeout time.Duration
	
	// Model is the AI model to use (e.g., "gpt-4", "claude-2")
	Model string
	
	// Temperature controls randomness in generation (0.0-1.0)
	Temperature float64
	
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int
	
	// EnableCaching determines if responses should be cached
	EnableCaching bool
	
	// CacheTTL is the time-to-live for cached responses
	CacheTTL time.Duration
}

// NewDefaultConfig creates a default configuration for the AI service
func NewDefaultConfig() Config {
	return Config{
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		Timeout:      30 * time.Second,
		Model:        "gpt-4",
		Temperature:  0.3,
		MaxTokens:    2048,
		EnableCaching: true,
		CacheTTL:     24 * time.Hour,
	}
}

// LLMProvider specifies which LLM provider to use
type LLMProvider string

const (
	// OpenAI provider
	OpenAI LLMProvider = "openai"
	
	// Anthropic provider
	Anthropic LLMProvider = "anthropic"
	
	// Custom provider
	Custom LLMProvider = "custom"
)

// GetProviderFromString converts a string to an LLMProvider
func GetProviderFromString(provider string) LLMProvider {
	switch provider {
	case "openai":
		return OpenAI
	case "anthropic":
		return Anthropic
	case "custom":
		return Custom
	default:
		return OpenAI // Default to OpenAI
	}
}