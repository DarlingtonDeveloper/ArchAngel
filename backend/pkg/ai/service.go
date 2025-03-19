package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/yourusername/codehawk/backend/pkg/analyzer"
)

// AISuggestionService provides AI-powered code suggestions
type AISuggestionService interface {
	// GetSuggestions generates AI-powered suggestions for code improvement
	GetSuggestions(ctx context.Context, code string, language string, issues []analyzer.Issue) ([]analyzer.Issue, error)
	
	// GetCodeExplanation generates an explanation for a piece of code
	GetCodeExplanation(ctx context.Context, code string, language string) (string, error)
	
	// GetCodeRefactoring generates a refactored version of the code
	GetCodeRefactoring(ctx context.Context, code string, language string) (string, error)
}

// Config holds the configuration for the AI service
type Config struct {
	// LLM API endpoint
	Endpoint string
	// API key for authentication
	APIKey string
	// Organization ID for billing
	OrgID string
	// Request timeout
	Timeout time.Duration
	// Model to use (e.g., "gpt-4", "codegen-350m", "claude-2", etc.)
	Model string
	// Temperature for generation (0.0-1.0)
	Temperature float64
	// Maximum tokens to generate
	MaxTokens int
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

// DefaultAIService implements AISuggestionService using an LLM API
type DefaultAIService struct {
	config  Config
	client  *http.Client
	provider LLMProvider
}

// NewAIService creates a new AI suggestion service
func NewAIService(config Config, provider LLMProvider) AISuggestionService {
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}
	
	// Set default values if not provided
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	if config.Temperature == 0 {
		config.Temperature = 0.3
	}
	
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}
	
	// Set default model based on provider
	if config.Model == "" {
		switch provider {
		case OpenAI:
			config.Model = "gpt-4"
		case Anthropic:
			config.Model = "claude-2"
		default:
			config.Model = "gpt-4"
		}
	}
	
	return &DefaultAIService{
		config:  config,
		client:  client,
		provider: provider,
	}
}

// GetSuggestions generates AI-powered suggestions for code improvement
func (s *DefaultAIService) GetSuggestions(ctx context.Context, code string, language string, issues []analyzer.Issue) ([]analyzer.Issue, error) {
	// Create prompt for code suggestions
	prompt := createSuggestionPrompt(code, language, issues)
	
	// Get LLM response
	response, err := s.callLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI suggestions: %w", err)
	}
	
	// Parse suggestions from response
	suggestions, err := parseSuggestions(response, language)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI suggestions: %w", err)
	}
	
	return suggestions, nil
}

// GetCodeExplanation generates an explanation for a piece of code
func (s *DefaultAIService) GetCodeExplanation(ctx context.Context, code string, language string) (string, error) {
	// Create prompt for code explanation
	prompt := fmt.Sprintf("Please explain the following %s code in detail, focusing on its purpose, functionality, and any potential issues or improvements:\n\n```%s\n%s\n```", 
		language, language, code)
	
	// Get LLM response
	response, err := s.callLLM(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get code explanation: %w", err)
	}
	
	return response, nil
}

// GetCodeRefactoring generates a refactored version of the code
func (s *DefaultAIService) GetCodeRefactoring(ctx context.Context, code string, language string) (string, error) {
	// Create prompt for code refactoring
	prompt := fmt.Sprintf("Please refactor the following %s code to improve its readability, efficiency, and adherence to best practices. Provide only the refactored code without explanations:\n\n```%s\n%s\n```", 
		language, language, code)
	
	// Get LLM response
	response, err := s.callLLM(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get code refactoring: %w", err)
	}
	
	// Extract code from markdown code blocks if present
	refactoredCode := extractCodeFromMarkdown(response, language)
	if refactoredCode != "" {
		return refactoredCode, nil
	}
	
	return response, nil
}

// callLLM makes a call to the language model API
func (s *DefaultAIService) callLLM(ctx context.Context, prompt string) (string, error) {
	switch s.provider {
	case OpenAI:
		return s.callOpenAI(ctx, prompt)
	case Anthropic:
		return s.callAnthropic(ctx, prompt)
	default:
		return s.callCustomLLM(ctx, prompt)
	}
}

// callOpenAI makes a call to the OpenAI API
func (s *DefaultAIService) callOpenAI(ctx context.Context, prompt string) (string, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"model":       s.config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are CodeHawk, an AI assistant specialized in code analysis, suggestions, and improvements. You provide detailed, actionable advice to help developers write better code.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": s.config.Temperature,
		"max_tokens":  s.config.MaxTokens,
	}
	
	// Convert request body to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.config.Endpoint,
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	if s.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", s.config.OrgID)
	}
	
	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", body)
	}
	
	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	
	return response.Choices[0].Message.Content, nil
}

// callAnthropic makes a call to the Anthropic API
func (s *DefaultAIService) callAnthropic(ctx context.Context, prompt string) (string, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"model":         s.config.Model,
		"prompt":        fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt),
		"max_tokens_to_sample": s.config.MaxTokens,
		"temperature":   s.config.Temperature,
	}
	
	// Convert request body to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.config.Endpoint,
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.config.APIKey)
	
	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", body)
	}
	
	// Parse response
	var response struct {
		Completion string `json:"completion"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	return response.Completion, nil
}

// callCustomLLM makes a call to a custom LLM API
func (s *DefaultAIService) callCustomLLM(ctx context.Context, prompt string) (string, error) {
	// Prepare request body - this would be customized for your specific LLM API
	requestBody := map[string]interface{}{
		"prompt":      prompt,
		"temperature": s.config.Temperature,
		"max_length":  s.config.MaxTokens,
		"model":       s.config.Model,
	}
	
	// Convert request body to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.config.Endpoint,
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	
	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", body)
	}
	
	// Parse response - adjust this based on your API's response format
	var response struct {
		Result string `json:"result"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	return response.Result, nil
}

// Helper functions

// createSuggestionPrompt creates a prompt for code suggestions
func createSuggestionPrompt(code string, language string, issues []analyzer.Issue) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("As a code improvement assistant, analyze the following %s code and provide specific, actionable suggestions to improve it:\n\n", language))
	sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", language, code))
	
	if len(issues) > 0 {
		sb.WriteString("The following issues have already been identified:\n\n")
		
		for i, issue := range issues {
			sb.WriteString(fmt.Sprintf("%d. Line %d: %s (%s)\n", i+1, issue.Line, issue.Message, issue.Severity))
		}
		
		sb.WriteString("\nBased on these issues and your own analysis, provide additional suggestions that focus on:\n")
	} else {
		sb.WriteString("Provide suggestions that focus on:\n")
	}
	
	sb.WriteString("1. Code structure and organization\n")
	sb.WriteString("2. Efficiency and performance\n")
	sb.WriteString("3. Readability and maintainability\n")
	sb.WriteString("4. Best practices for " + language + "\n\n")
	sb.WriteString("For each suggestion, include:\n")
	sb.WriteString("- The line number\n")
	sb.WriteString("- A clear description of the issue\n")
	sb.WriteString("- A specific code sample showing how to fix it\n")
	sb.WriteString("- The severity (error, warning, or suggestion)\n\n")
	sb.WriteString("Format your response as JSON with the following structure for each suggestion:\n")
	sb.WriteString("```json\n[{\"line\": 5, \"message\": \"Description\", \"severity\": \"suggestion\", \"replacement\": \"fixed code\"}, ...]\n```\n")
	
	return sb.String()
}

// parseSuggestions parses suggestions from LLM response
func parseSuggestions(response string, language string) ([]analyzer.Issue, error) {
	// Extract JSON from the response
	jsonStr := extractJSONFromResponse(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}
	
	// Parse JSON
	var rawSuggestions []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawSuggestions); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions JSON: %w", err)
	}
	
	// Convert to Issue structs
	suggestions := make([]analyzer.Issue, 0, len(rawSuggestions))
	
	for _, raw := range rawSuggestions {
		lineF, ok := raw["line"].(float64)
		if !ok {
			continue
		}
		
		message, ok := raw["message"].(string)
		if !ok {
			continue
		}
		
		severity, ok := raw["severity"].(string)
		if !ok {
			severity = "suggestion"
		}
		
		// Create issue
		line := int(lineF)
		suggestion := analyzer.Issue{
			Line:     line,
			Message:  message,
			Severity: severity,
			RuleID:   "ai-suggestion",
		}
		
		// Add fix if available
		if replacement, ok := raw["replacement"].(string); ok {
			suggestion.Fix = &analyzer.IssueFix{
				Description: "AI suggested fix",
				Replacement: replacement,
			}
		}
		
		suggestions = append(suggestions, suggestion)
	}
	
	return suggestions, nil
}

// extractJSONFromResponse extracts JSON from LLM response
func extractJSONFromResponse(response string) string {
	// Look for JSON in code blocks
	jsonPattern := "```json\n(.*?)```"
	jsonRegex := regexp.MustCompile(jsonPattern)
	matches := jsonRegex.FindStringSubmatch(response)
	
	if len(matches) >= 2 {
		return matches[1]
	}
	
	// If no JSON code block, try to find array directly
	if strings.HasPrefix(strings.TrimSpace(response), "[") && strings.Contains(response, "{") {
		// Try to extract the full array
		depth := 0
		inString := false
		escape := false
		
		for i := 0; i < len(response); i++ {
			c := response[i]
			
			if escape {
				escape = false
				continue
			}
			
			if c == '\\' {
				escape = true
				continue
			}
			
			if c == '"' && !escape {
				inString = !inString
				continue
			}
			
			if !inString {
				if c == '[' {
					depth++
				} else if c == ']' {
					depth--
					if depth == 0 {
						return response[:i+1]
					}
				}
			}
		}
	}
	
	// As a last resort, treat the entire response as JSON
	return response
}

// extractCodeFromMarkdown extracts code from markdown code blocks
func extractCodeFromMarkdown(markdown string, language string) string {
	// Look for code blocks with the specified language
	pattern := fmt.Sprintf("```%s\n(.*?)```", language)
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(markdown)
	
	if len(matches) >= 2 {
		return matches[1]
	}
	
	// If not found, look for any code blocks
	pattern = "```.*?\n(.*?)```"
	regex = regexp.MustCompile(pattern)
	matches = regex.FindStringSubmatch(markdown)
	
	if len(matches) >= 2 {
		return matches[1]
	}
	
	return ""
}