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
	"regexp"
	"strings"
)

// GoLinter implements the Linter interface for Go
type GoLinter struct {
	*BaseAnalyzer
	golangciLintPath string
	staticcheckPath  string
}

// NewGoLinter creates a new Go linter
func NewGoLinter(config map[string]string) *GoLinter {
	// Default paths
	golangciLintPath := "golangci-lint"
	if path, ok := config["golangciLintPath"]; ok && path != "" {
		golangciLintPath = path
	}
	
	staticcheckPath := "staticcheck"
	if path, ok := config["staticcheckPath"]; ok && path != "" {
		staticcheckPath = path
	}
	
	return &GoLinter{
		BaseAnalyzer:     NewBaseAnalyzer(config),
		golangciLintPath: golangciLintPath,
		staticcheckPath:  staticcheckPath,
	}
}

// Language returns the identifier for the supported language
func (l *GoLinter) Language() string {
	return "go"
}

// Analyze analyzes the provided code and returns issues found
func (l *GoLinter) Analyze(ctx context.Context, code string, options map[string]interface{}) (*AnalysisResult, error) {
	return l.BaseAnalyzer.CommonAnalyzeWrapper(
		ctx,
		code,
		options,
		l.findIssues,
		l.SuggestFixes,
	)
}

// findIssues analyzes the code and returns issues
func (l *GoLinter) findIssues(ctx context.Context, code string, options map[string]interface{}) ([]Issue, error) {
	// Create a temporary directory for Go module
	tmpDir, err := ioutil.TempDir("", "codehawk-go")
	if err != nil {
		return nil, l.WrapError(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tmpDir)
	
	// Set up a minimal Go module
	err = l.setupGoModule(tmpDir)
	if err != nil {
		return nil, l.WrapError(err, "failed to set up Go module")
	}
	
	// Create a temporary file for the code
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := ioutil.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return nil, l.WrapError(err, "failed to write to temporary file")
	}
	
	// Run both linters and combine results
	golangciIssues, err := l.runGolangciLint(ctx, tmpDir)
	if err != nil {
		// Log the error but continue with other linters
		fmt.Printf("Warning: golangci-lint failed: %v\n", err)
	}
	
	staticcheckIssues, err := l.runStaticcheck(ctx, tmpDir)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Warning: staticcheck failed: %v\n", err)
	}
	
	// Combine issues
	issues := append(golangciIssues, staticcheckIssues...)
	
	return issues, nil
}

// SuggestFixes attempts to generate fixes for the identified issues
func (l *GoLinter) SuggestFixes(ctx context.Context, code string, issues []Issue) ([]Issue, error) {
	// Create a temporary directory for Go module
	tmpDir, err := ioutil.TempDir("", "codehawk-go-fix")
	if err != nil {
		return nil, l.WrapError(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tmpDir)
	
	// Set up a minimal Go module
	err = l.setupGoModule(tmpDir)
	if err != nil {
		return nil, l.WrapError(err, "failed to set up Go module")
	}
	
	// Create a temporary file for the code
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := ioutil.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return nil, l.WrapError(err, "failed to write to temporary file")
	}
	
	// Run gofmt to get properly formatted code
	cmd := exec.CommandContext(
		ctx,
		"gofmt",
		"-w",
		tmpFile,
	)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return nil, l.WrapError(fmt.Errorf("failed to run gofmt: %w, stderr: %s", err, stderr.String()), "gofmt execution error")
	}
	
	// Read the formatted code
	formattedCode, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return nil, l.WrapError(err, "failed to read formatted code")
	}
	
	// Generate suggestions based on the formatted code and known issue patterns
	suggestions := make([]Issue, 0)
	formattedLines := strings.Split(string(formattedCode), "\n")
	originalLines := strings.Split(code, "\n")
	
	// First, add suggestions based on formatting differences
	for i := 0; i < len(originalLines) && i < len(formattedLines); i++ {
		if originalLines[i] != formattedLines[i] {
			lineNum := i + 1 // Convert to 1-based line numbering
			
			suggestion := Issue{
				Line:     lineNum,
				Message:  "Code formatting issue",
				Severity: "suggestion",
				RuleID:   "gofmt",
				Fix: &IssueFix{
					Description: "Format code according to gofmt",
					Replacement: formattedLines[i],
				},
			}
			
			suggestions = append(suggestions, suggestion)
		}
	}
	
	// Then, add specific suggestions for common issues
	for _, issue := range issues {
		suggestion := l.generateFixForIssue(issue, code)
		if suggestion.Fix != nil {
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions, nil
}

// setupGoModule sets up a minimal Go module for linting
func (l *GoLinter) setupGoModule(dir string) error {
	// Create go.mod file
	goModContent := `module codehawk.temp

go 1.17
`
	goModPath := filepath.Join(dir, "go.mod")
	if err := ioutil.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}
	
	return nil
}

// runGolangciLint runs golangci-lint and parses its output
func (l *GoLinter) runGolangciLint(ctx context.Context, dir string) ([]Issue, error) {
	// Prepare golangci-lint command
	cmd := exec.CommandContext(
		ctx,
		l.golangciLintPath,
		"run",
		"--out-format=json",
		"--config=none", // Use default config
		"./...",         // Analyze all packages
	)
	
	cmd.Dir = dir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Execute golangci-lint
	err := cmd.Run()
	// Exit code 1 is normal when golangci-lint finds issues
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return nil, l.WrapError(fmt.Errorf("failed to run golangci-lint: %w, stderr: %s", err, stderr.String()), "golangci-lint execution error")
	}
	
	// Parse the JSON output
	type GolangciLintIssue struct {
		FromLinter  string `json:"from_linter"`
		Text        string `json:"text"`
		SourceLines []string `json:"source_lines"`
		Pos         struct {
			Filename string `json:"filename"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		} `json:"pos"`
		Severity string `json:"severity,omitempty"`
	}
	
	type GolangciLintResult struct {
		Issues []GolangciLintIssue `json:"Issues"`
	}
	
	var result GolangciLintResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, l.WrapError(err, "failed to parse golangci-lint output")
	}
	
	// Convert golangci-lint issues to CodeHawk issues
	issues := make([]Issue, 0, len(result.Issues))
	for _, lintIssue := range result.Issues {
		// Skip issues from files other than main.go
		if filepath.Base(lintIssue.Pos.Filename) != "main.go" {
			continue
		}
		
		// Map golangci-lint severity to CodeHawk severity
		severity := l.BaseAnalyzer.MapSeverity(lintIssue.Severity)
		if severity == "info" {
			// Default to warning if not specified
			severity = "warning"
		}
		
		column := lintIssue.Pos.Column
		
		issue := Issue{
			Line:     lintIssue.Pos.Line,
			Column:   &column,
			Message:  lintIssue.Text,
			Severity: severity,
			RuleID:   lintIssue.FromLinter,
			Context:  strings.Join(lintIssue.SourceLines, "\n"),
		}
		
		issues = append(issues, issue)
	}
	
	return issues, nil
}

// runStaticcheck runs staticcheck and parses its output
func (l *GoLinter) runStaticcheck(ctx context.Context, dir string) ([]Issue, error) {
	// Prepare staticcheck command
	cmd := exec.CommandContext(
		ctx,
		l.staticcheckPath,
		"-f=json",
		"./...", // Analyze all packages
	)
	
	cmd.Dir = dir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Execute staticcheck
	err := cmd.Run()
	// Exit code 1 is normal when staticcheck finds issues
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return nil, l.WrapError(fmt.Errorf("failed to run staticcheck: %w, stderr: %s", err, stderr.String()), "staticcheck execution error")
	}
	
	// Parse the JSON output
	type StaticcheckIssue struct {
		Code     string `json:"code"`
		Severity string `json:"severity"`
		Message  string `json:"message"`
		Position struct {
			Filename string `json:"filename"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		} `json:"position"`
	}
	
	// Staticcheck outputs one JSON object per line
	issues := make([]Issue, 0)
	lines := strings.Split(stdout.String(), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		var scIssue StaticcheckIssue
		if err := json.Unmarshal([]byte(line), &scIssue); err != nil {
			return nil, l.WrapError(err, "failed to parse staticcheck output")
		}
		
		// Skip issues from files other than main.go
		if filepath.Base(scIssue.Position.Filename) != "main.go" {
			continue
		}
		
		// Map staticcheck severity to CodeHawk severity using BaseAnalyzer
		severity := l.BaseAnalyzer.MapSeverity(scIssue.Severity)
		
		column := scIssue.Position.Column
		
		issue := Issue{
			Line:     scIssue.Position.Line,
			Column:   &column,
			Message:  fmt.Sprintf("%s: %s", scIssue.Code, scIssue.Message),
			Severity: severity,
			RuleID:   scIssue.Code,
		}
		
		issues = append(issues, issue)
	}
	
	return issues, nil
}

// generateFixForIssue attempts to generate a fix for a specific issue
func (l *GoLinter) generateFixForIssue(issue Issue, code string) Issue {
	// Create a deep copy of the issue
	suggestion := Issue{
		Line:     issue.Line,
		Column:   issue.Column,
		Message:  fmt.Sprintf("Suggested fix for: %s", issue.Message),
		Severity: "suggestion",
		RuleID:   issue.RuleID,
	}
	
	// Get the problematic line
	lines := strings.Split(code, "\n")
	if issue.Line <= 0 || issue.Line > len(lines) {
		return suggestion
	}
	
	lineIdx := issue.Line - 1
	lineText := lines[lineIdx]
	
	// Generate fix based on rule ID
	ruleID := issue.RuleID
	
	switch {
	case ruleID == "gofmt":
		// Already handled by gofmt
		return suggestion
		
	case ruleID == "unused" || strings.Contains(issue.Message, "unused variable"):
		// Try to find the variable name
		varNameRegex := regexp.MustCompile(`'([^']+)'`)
		matches := varNameRegex.FindStringSubmatch(issue.Message)
		if len(matches) >= 2 {
			varName := matches[1]
			// Replace variable with blank identifier
			suggestion.Fix = &IssueFix{
				Description: "Replace unused variable with blank identifier",
				Replacement: strings.Replace(lineText, varName, "_", 1),
			}
		}
		
	case ruleID == "ineffassign" || strings.Contains(issue.Message, "ineffectual assignment"):
		// Try to find the variable name
		varNameRegex := regexp.MustCompile(`to ([^ ]+)`)
		matches := varNameRegex.FindStringSubmatch(issue.Message)
		if len(matches) >= 2 {
			varName := matches[1]
			// Comment out the ineffectual assignment
			suggestion.Fix = &IssueFix{
				Description: "Comment out ineffectual assignment",
				Replacement: "// " + lineText + " // Ineffectual assignment to " + varName,
			}
		}
		
	case strings.Contains(ruleID, "SA") || strings.Contains(issue.Message, "should check errors"):
		// Staticcheck error handling suggestion
		if strings.Contains(lineText, "=") && !strings.Contains(lineText, "if err") {
			// Try to add error checking
			suggestion.Fix = &IssueFix{
				Description: "Add error checking",
				Replacement: lineText + "\nif err != nil {\n\treturn err\n}",
			}
		}
		
	case strings.Contains(ruleID, "ST") || strings.Contains(issue.Message, "should use a simple channel send/receive"):
		// Staticcheck simplify channel operations
		if strings.Contains(lineText, "select {") {
			// Simplify select with single case
			suggestion.Fix = &IssueFix{
				Description: "Simplify channel operation",
				Replacement: "// Replace select with single case by using a simple channel operation",
			}
		}
		
	case strings.Contains(issue.Message, "exported") && strings.Contains(issue.Message, "comment"):
		// Add comment for exported declarations
		if strings.Contains(lineText, "func ") {
			funcName := strings.TrimSpace(strings.Split(strings.Split(lineText, "func ")[1], "(")[0])
			suggestion.Fix = &IssueFix{
				Description: "Add comment for exported function",
				Replacement: "// " + funcName + " is a function that does something\n" + lineText,
			}
		} else if strings.Contains(lineText, "type ") {
			typeName := strings.TrimSpace(strings.Split(strings.Split(lineText, "type ")[1], " ")[0])
			suggestion.Fix = &IssueFix{
				Description: "Add comment for exported type",
				Replacement: "// " + typeName + " represents something\n" + lineText,
			}
		}
		
	case strings.Contains(issue.Message, "should be"):
		// Variable naming suggestions
		words := strings.Split(issue.Message, " ")
		for i, word := range words {
			if word == "should" && i > 0 && i < len(words)-2 {
				varName := words[i-1]
				suggestion.Fix = &IssueFix{
					Description: "Rename variable to follow Go naming conventions",
					Replacement: strings.ReplaceAll(lineText, varName, words[i+2]),
				}
				break
			}
		}
	}
	
	return suggestion
}