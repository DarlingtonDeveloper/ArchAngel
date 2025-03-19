package analyzer

import (
	"fmt"
	"sync"
)

// LinterRegistry manages available linters
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
	
	language := linter.Language()
	r.linters[language] = linter
	
	// Log registration (in a real implementation you'd use a proper logger)
	fmt.Printf("Registered linter for %s\n", language)
}

// GetLinter retrieves a linter for a specific language
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

// RegisterDefaultLinters registers all default linters with their standard configurations
func (r *LinterRegistry) RegisterDefaultLinters() {
	// Python linter
	pythonLinter := NewPythonLinter(map[string]string{
		"pylintPath": "pylint",
		"timeout":    "15s",
	})
	r.Register(pythonLinter)
	
	// JavaScript linter
	jsLinter := NewJavaScriptLinter(map[string]string{
		"eslintPath": "eslint",
		"timeout":    "15s",
	})
	r.Register(jsLinter)
	
	// TypeScript linter
	tsLinter := NewTypeScriptLinter(map[string]string{
		"eslintPath": "eslint",
		"timeout":    "15s",
	})
	r.Register(tsLinter)
	
	// Go linter
	goLinter := NewGoLinter(map[string]string{
		"golangciLintPath": "golangci-lint",
		"staticcheckPath":  "staticcheck",
		"timeout":          "15s",
	})
	r.Register(goLinter)
}