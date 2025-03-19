// CodeHawk API Client Example - Go
// This example shows how to use the CodeHawk API in a Go application

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// CodeHawkClient provides methods to interact with the CodeHawk API
type CodeHawkClient struct {
	apiKey      string
	apiURL      string
	httpClient  *http.Client
}

// NewCodeHawkClient creates a new CodeHawk API client
func NewCodeHawkClient(apiKey string, apiURL string) *CodeHawkClient {
	if apiURL == "" {
		apiURL = "https://api.codehawk.dev/api/v1"
	}
	
	return &CodeHawkClient{
		apiKey:      apiKey,
		apiURL:      apiURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// AnalysisRequest represents a code analysis request
type AnalysisRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	Context  string                 `json:"context,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// AnalysisResponse represents a code analysis response
type AnalysisResponse struct {
	ID          string        `json:"id"`
	Status      string        `json:"status"`
	Language    string        `json:"language"`
	Context     string        `json:"context,omitempty"`
	Timestamp   string        `json:"timestamp"`
	Issues      []Issue       `json:"issues"`
	Suggestions []Suggestion  `json:"suggestions,omitempty"`
}

// Issue represents a code issue
type Issue struct {
	Line     int     `json:"line"`
	Column   *int    `json:"column,omitempty"`
	Message  string  `json:"message"`
	Severity string  `json:"severity"`
	RuleID   string  `json:"rule_id,omitempty"`
	Context  string  `json:"context,omitempty"`
	Fix      *Fix    `json:"fix,omitempty"`
}

// Suggestion represents a code suggestion
type Suggestion struct {
	Line     int     `json:"line"`
	Column   *int    `json:"column,omitempty"`
	Message  string  `json:"message"`
	Severity string  `json:"severity"`
	RuleID   string  `json:"rule_id,omitempty"`
	Context  string  `json:"context,omitempty"`
	Fix      *Fix    `json:"fix,omitempty"`
}

// Fix represents a code fix
type Fix struct {
	Description string `json:"description"`
	Replacement string `json:"replacement"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// AnalyzeCode submits code for analysis
func (c *CodeHawkClient) AnalyzeCode(ctx context.Context, code, language, contextStr string, options map[string]interface{}) (*AnalysisResponse, error) {
	request := AnalysisRequest{
		Code:     code,
		Language: language,
		Context:  contextStr,
		Options:  options,
	}
	
	responseData, err := c.makeRequest(ctx, http.MethodPost, "/analyze", request)
	if err != nil {
		return nil, err
	}
	
	var response AnalysisResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &response, nil
}

// GetAnalysis gets an analysis by ID
func (c *CodeHawkClient) GetAnalysis(ctx context.Context, id string) (*AnalysisResponse, error) {
	responseData, err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/analysis/%s", id), nil)
	if err != nil {
		return nil, err
	}
	
	var response AnalysisResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &response, nil
}

// GetIssues gets issues for an analysis
func (c *CodeHawkClient) GetIssues(ctx context.Context, id, severity string) ([]Issue, error) {
	url := fmt.Sprintf("/analysis/%s/issues", id)
	if severity != "" {
		url = fmt.Sprintf("%s?severity=%s", url, severity)
	}
	
	responseData, err := c.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	var response struct {
		Issues []Issue `json:"issues"`
	}
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return response.Issues, nil
}

// GetSuggestions gets suggestions for an analysis
func (c *CodeHawkClient) GetSuggestions(ctx context.Context, id string) ([]Suggestion, error) {
	responseData, err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/analysis/%s/suggestions", id), nil)
	if err != nil {
		return nil, err
	}
	
	var response struct {
		Suggestions []Suggestion `json:"suggestions"`
	}
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return response.Suggestions, nil
}

// GetLanguages gets supported languages
func (c *CodeHawkClient) GetLanguages(ctx context.Context) ([]string, error) {
	responseData, err := c.makeRequest(ctx, http.MethodGet, "/languages", nil)
	if err != nil {
		return nil, err
	}
	
	var response struct {
		Languages []string `json:"languages"`
	}
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return response.Languages, nil
}

// makeRequest makes an API request
func (c *CodeHawkClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	url := c.apiURL + endpoint
	
	var req *http.Request
	var err error
	
	if body != nil {
		bodyData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Check for errors
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResp); err != nil {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(responseBody))
		}
		
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New("authentication failed, please check your API key")
		case http.StatusForbidden:
			return nil, errors.New("access denied, you do not have permission to perform this action")
		case http.StatusNotFound:
			return nil, errors.New("the requested resource was not found")
		case http.StatusTooManyRequests:
			return nil, errors.New("API rate limit exceeded, please try again later")
		default:
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}
	}
	
	return responseBody, nil
}

func main() {
	// Get API key from environment variable or use a default value
	apiKey := os.Getenv("CODEHAWK_API_KEY")
	if apiKey == "" {
		apiKey = "your-api-key" // Replace with your actual API key for testing
	}
	
	// Create client
	client := NewCodeHawkClient(apiKey, "")
	
	// Example code to analyze
	code := `package main

import "fmt"

func main() {
	var x = 10
	fmt.Println("Hello, World!")
}`
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Analyze code
	fmt.Println("Analyzing code...")
	analysis, err := client.AnalyzeCode(ctx, code, "go", "Example code", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Analysis ID: %s\n", analysis.ID)
	fmt.Printf("Found %d issues\n", len(analysis.Issues))
	
	// Print issues
	if len(analysis.Issues) > 0 {
		fmt.Println("\nIssues:")
		for i, issue := range analysis.Issues {
			fmt.Printf("%d. Line %d: [%s] %s\n", i+1, issue.Line, issue.Severity, issue.Message)
		}
	}
	
	// Get supported languages
	fmt.Println("\nGetting supported languages...")
	languages, err := client.GetLanguages(ctx)
	if err != nil {
		fmt.Printf("Error getting languages: %v\n", err)
		return
	}
	
	fmt.Printf("Supported languages: %v\n", languages)
}