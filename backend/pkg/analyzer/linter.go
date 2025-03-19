package analyzer

import (
	"context"
	"sync"
)

// Issue represents a code issue found by a linter
type Issue struct {
	Line        int         `json:"line"`
	Column      *int        `json:"column,omitempty"`
	Message     string      `json:"message"`
	Severity    string      `json:"severity"`
	RuleID      string      `json:"ruleId,omitempty"`
	Context     string      `json:"context,omitempty"`
	Fix         *IssueFix   `json:"fix,omitempty"`
	Suggestions []IssueFix  `json:"suggestions,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// IssueFix represents a suggested fix for an issue
type IssueFix struct {
	Description string `json:"description"`
	Replacement string `json:"replacement"`
}

// AnalysisResult represents the result of code analysis
type AnalysisResult struct {
	Issues      []Issue     `json:"issues"`
	Suggestions []Issue     `json:"suggestions,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// Linter defines the interface for language-specific linters
type Linter interface {
	// Language returns the identifier for the supported language
	Language() string
	
	// Analyze analyzes the provided code and returns issues found
	Analyze(ctx context.Context, code string, options map[string]interface{}) (*AnalysisResult, error)
	
	// SuggestFixes attempts to generate fixes for the identified issues
	SuggestFixes(ctx context.Context, code string, issues []Issue) ([]Issue, error)
}

// LinterRegistry manages the available linters
type LinterRegistry struct {
	linters map[string]Linter
	mu      sync.RWMutex
}

// NewLinterRegistry creates a new linter registry
func NewLinterRegistry() *LinterRegistry {
	return &LinterRegistry{
		linters: make(map[string]Linter),
	}
}

// Register adds a linter to the registry
func (r *LinterRegistry) Register(linter Linter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linters[linter.Language()] = linter
}

// GetLinter retrieves a linter for the specified language
func (r *LinterRegistry) GetLinter(language string) (Linter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	linter, ok := r.linters[language]
	return linter, ok
}

// GetSupportedLanguages returns a list of supported languages
func (r *LinterRegistry) GetSupportedLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	languages := make([]string, 0, len(r.linters))
	for lang := range r.linters {
		languages = append(languages, lang)
	}
	return languages
}