package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// JavaScriptLinter implements the Linter interface for JavaScript and TypeScript
type JavaScriptLinter struct {
	*BaseAnalyzer
	eslintPath     string
	configPath     string
	typescriptMode bool
}

// ESLintResult represents the result from ESLint
type ESLintResult struct {
	FilePath string `json:"filePath"`
	Messages []struct {
		RuleID    string `json:"ruleId"`
		Severity  int    `json:"severity"`
		Message   string `json:"message"`
		Line      int    `json:"line"`
		Column    int    `json:"column"`
		NodeType  string `json:"nodeType"`
		Fix       *struct {
			Range   []int  `json:"range"`
			Text    string `json:"text"`
		} `json:"fix,omitempty"`
	} `json:"messages"`
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
}

// NewJavaScriptLinter creates a new JavaScript linter
func NewJavaScriptLinter(config map[string]string) *JavaScriptLinter {
	// Default ESLint path
	eslintPath := "eslint"
	if path, ok := config["eslintPath"]; ok && path != "" {
		eslintPath = path
	}
	
	// Default config path
	configPath := ""
	if path, ok := config["configPath"]; ok && path != "" {
		configPath = path
	}
	
	return &JavaScriptLinter{
		BaseAnalyzer:   NewBaseAnalyzer(config),
		eslintPath:     eslintPath,
		configPath:     configPath,
		typescriptMode: false,
	}
}

// NewTypeScriptLinter creates a new TypeScript linter
func NewTypeScriptLinter(config map[string]string) *JavaScriptLinter {
	linter := NewJavaScriptLinter(config)
	linter.typescriptMode = true
	return linter
}

// Language returns the identifier for the supported language
func (l *JavaScriptLinter) Language() string {
	if l.typescriptMode {
		return "typescript"
	}
	return "javascript"
}

// Analyze analyzes the provided code and returns issues found
func (l *JavaScriptLinter) Analyze(ctx context.Context, code string, options map[string]interface{}) (*AnalysisResult, error) {
	return l.BaseAnalyzer.CommonAnalyzeWrapper(
		ctx,
		code,
		options,
		l.findIssues,
		l.SuggestFixes,
	)
}

// findIssues analyzes the code and returns issues
func (l *JavaScriptLinter) findIssues(ctx context.Context, code string, options map[string]interface{}) ([]Issue, error) {
	// Create a temporary directory
	tmpDir, err := ioutil.TempDir("", "codehawk-eslint")
	if err != nil {
		return nil, l.WrapError(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a temporary file for the code
	ext := ".js"
	if l.typescriptMode {
		ext = ".ts"
	}
	
	tmpFile := filepath.Join(tmpDir, "code"+ext)
	if err := ioutil.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return nil, l.WrapError(err, "failed to write to temporary file")
	}
	
	// Prepare ESLint command
	args := []string{
		"--format=json",
		"--no-ignore",
	}
	
	if l.configPath != "" {
		args = append(args, "--config", l.configPath)
	} else {
		// Use a default configuration if none is provided
		configFile := filepath.Join(tmpDir, ".eslintrc.json")
		defaultConfig := l.getDefaultConfig()
		if err := ioutil.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
			return nil, l.WrapError(err, "failed to write default ESLint config")
		}
		args = append(args, "--config", configFile)
	}
	
	// Add the file to check
	args = append(args, tmpFile)
	
	// Run ESLint
	cmd := exec.CommandContext(
		ctx,
		l.eslintPath,
		args...,
	)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Execute ESLint
	err = cmd.Run()
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		// exit status 1 is normal for ESLint when it finds issues
		return nil, l.WrapError(fmt.Errorf("failed to run ESLint: %w, stderr: %s", err, stderr.String()), "eslint execution error")
	}
	
	// Parse the JSON output
	var eslintResults []ESLintResult
	if err := json.Unmarshal(stdout.Bytes(), &eslintResults); err != nil {
		return nil, l.WrapError(err, "failed to parse ESLint output")
	}
	
	if len(eslintResults) == 0 {
		return []Issue{}, nil
	}
	
	// Convert ESLint results to CodeHawk issues
	issues := make([]Issue, 0)
	for _, result := range eslintResults {
		for _, msg := range result.Messages {
			issue := l.convertESLintMessage(msg, code)
			issues = append(issues, issue)
		}
	}
	
	return issues, nil
}

// SuggestFixes attempts to generate fixes for the identified issues
func (l *JavaScriptLinter) SuggestFixes(ctx context.Context, code string, issues []Issue) ([]Issue, error) {
	// Create a temporary directory
	tmpDir, err := ioutil.TempDir("", "codehawk-eslint-fix")
	if err != nil {
		return nil, l.WrapError(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a temporary file for the code
	ext := ".js"
	if l.typescriptMode {
		ext = ".ts"
	}
	
	tmpFile := filepath.Join(tmpDir, "code"+ext)
	if err := ioutil.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return nil, l.WrapError(err, "failed to write to temporary file")
	}
	
	// Prepare ESLint command with --fix option
	args := []string{
		"--format=json",
		"--no-ignore",
		"--fix",
	}
	
	if l.configPath != "" {
		args = append(args, "--config", l.configPath)
	} else {
		// Use a default configuration if none is provided
		configFile := filepath.Join(tmpDir, ".eslintrc.json")
		defaultConfig := l.getDefaultConfig()
		if err := ioutil.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
			return nil, l.WrapError(err, "failed to write default ESLint config")
		}
		args = append(args, "--config", configFile)
	}
	
	// Add the file to check
	args = append(args, tmpFile)
	
	// Run ESLint with fix
	cmd := exec.CommandContext(
		ctx,
		l.eslintPath,
		args...,
	)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	// Execute ESLint
	err = cmd.Run()
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		// exit status 1 is normal for ESLint when it finds issues
		return nil, l.WrapError(fmt.Errorf("failed to run ESLint fix: %w, stderr: %s", err, stderr.String()), "eslint fix execution error")
	}
	
	// Read the fixed code
	fixedCode, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return nil, l.WrapError(err, "failed to read fixed code")
	}
	
	// Create suggestions for the issues that have a fix
	suggestions := make([]Issue, 0)
	lines := strings.Split(string(fixedCode), "\n")
	
	for _, issue := range issues {
		if issue.Line <= 0 || issue.Line > len(lines) {
			continue
		}
		
		lineIdx := issue.Line - 1
		fixedLine := lines[lineIdx]
		
		// Generate a suggestion if the line is different from the original
		originalLines := strings.Split(code, "\n")
		if lineIdx < len(originalLines) && fixedLine != originalLines[lineIdx] {
			suggestion := Issue{
				Line:     issue.Line,
				Column:   issue.Column,
				Message:  fmt.Sprintf("Suggested fix for: %s", issue.Message),
				Severity: "suggestion",
				RuleID:   issue.RuleID,
				Fix: &IssueFix{
					Description: "Auto-fix with ESLint",
					Replacement: fixedLine,
				},
			}
			
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions, nil
}

// convertESLintMessage converts an ESLint message to a CodeHawk issue
func (l *JavaScriptLinter) convertESLintMessage(msg struct {
	RuleID    string `json:"ruleId"`
	Severity  int    `json:"severity"`
	Message   string `json:"message"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	NodeType  string `json:"nodeType"`
	Fix       *struct {
		Range   []int  `json:"range"`
		Text    string `json:"text"`
	} `json:"fix,omitempty"`
}, code string) Issue {
	// Map ESLint severity to CodeHawk severity
	var severity string
	switch msg.Severity {
	case 2:
		severity = "error"
	case 1:
		severity = "warning"
	default:
		severity = "suggestion"
	}
	
	// Create the issue
	issue := Issue{
		Line:     msg.Line,
		Column:   &msg.Column,
		Message:  msg.Message,
		Severity: severity,
		RuleID:   msg.RuleID,
		Context:  msg.NodeType,
	}
	
	// Add fix if available
	if msg.Fix != nil {
		// Get the line containing the fix
		lines := strings.Split(code, "\n")
		if msg.Line > 0 && msg.Line <= len(lines) {
			issue.Fix = &IssueFix{
				Description: "Auto-fix with ESLint",
				Replacement: msg.Fix.Text,
			}
		}
	}
	
	return issue
}

// getDefaultConfig returns a default ESLint configuration
func (l *JavaScriptLinter) getDefaultConfig() string {
	if l.typescriptMode {
		return `{
			"parser": "@typescript-eslint/parser",
			"plugins": ["@typescript-eslint"],
			"extends": [
				"eslint:recommended",
				"plugin:@typescript-eslint/recommended"
			],
			"rules": {
				"semi": ["error", "always"],
				"quotes": ["warn", "single"],
				"indent": ["warn", 2],
				"no-unused-vars": "warn",
				"@typescript-eslint/explicit-function-return-type": "warn",
				"@typescript-eslint/no-explicit-any": "warn"
			}
		}`
	}
	
	return `{
		"env": {
			"browser": true,
			"node": true,
			"es6": true
		},
		"extends": "eslint:recommended",
		"parserOptions": {
			"ecmaVersion": 2018,
			"sourceType": "module"
		},
		"rules": {
			"semi": ["error", "always"],
			"quotes": ["warn", "single"],
			"indent": ["warn", 2],
			"no-unused-vars": "warn",
			"no-console": "warn"
		}
	}`
}