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
	"strconv"
	"strings"
	"time"
)

// PythonLinter implements the Linter interface for Python
type PythonLinter struct {
	// Configuration options
	pylintPath string
	timeout    time.Duration
}

// NewPythonLinter creates a new Python linter
func NewPythonLinter(options map[string]string) *PythonLinter {
	// Default pylint path
	pylintPath := "pylint"
	if path, ok := options["pylintPath"]; ok && path != "" {
		pylintPath = path
	}
	
	// Default timeout
	timeout := 10 * time.Second
	if timeoutStr, ok := options["timeout"]; ok && timeoutStr != "" {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}
	
	return &PythonLinter{
		pylintPath: pylintPath,
		timeout:    timeout,
	}
}

// Language returns the identifier for the supported language
func (l *PythonLinter) Language() string {
	return "python"
}

// Analyze analyzes the provided code and returns issues found
func (l *PythonLinter) Analyze(ctx context.Context, code string, options map[string]interface{}) (*AnalysisResult, error) {
	// Create a temporary file for the code
	tmpFile, err := ioutil.TempFile("", "codehawk-*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// Write the code to the temporary file
	if _, err := tmpFile.Write([]byte(code)); err != nil {
		return nil, fmt.Errorf("failed to write to temporary file: %w", err)
	}
	
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}
	
	// Create a timeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()
	
	// Run pylint on the temporary file
	cmd := exec.CommandContext(
		ctxWithTimeout,
		l.pylintPath,
		"--output-format=json",
		tmpFile.Name(),
	)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Execute pylint
	err = cmd.Run()
	if err != nil && err.Error() != "exit status 1" {
		// exit status 1 is normal for pylint when it finds issues
		return nil, fmt.Errorf("failed to run pylint: %w, stderr: %s", err, stderr.String())
	}
	
	// Parse the JSON output
	var pylintResults []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &pylintResults); err != nil {
		return nil, fmt.Errorf("failed to parse pylint output: %w", err)
	}
	
	// Convert pylint results to CodeHawk issues
	issues := make([]Issue, 0, len(pylintResults))
	for _, result := range pylintResults {
		issue := l.convertPylintResult(result)
		issues = append(issues, issue)
	}
	
	// Generate suggestions
	suggestions, err := l.SuggestFixes(ctx, code, issues)
	if err != nil {
		// Log the error but continue with the issues we have
		fmt.Printf("Failed to generate suggestions: %v\n", err)
	}
	
	return &AnalysisResult{
		Issues:      issues,
		Suggestions: suggestions,
	}, nil
}

// SuggestFixes attempts to generate fixes for the identified issues
func (l *PythonLinter) SuggestFixes(ctx context.Context, code string, issues []Issue) ([]Issue, error) {
	// Create a temporary file for the code
	tmpFile, err := ioutil.TempFile("", "codehawk-*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// Write the code to the temporary file
	if _, err := tmpFile.Write([]byte(code)); err != nil {
		return nil, fmt.Errorf("failed to write to temporary file: %w", err)
	}
	
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}
	
	// Create a timeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()
	
	// Run pycodestyle for suggestions
	cmd := exec.CommandContext(
		ctxWithTimeout,
		"pycodestyle",
		"--format=%(path)s:%(row)d:%(col)d: %(code)s %(text)s",
		tmpFile.Name(),
	)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Execute pycodestyle
	err = cmd.Run()
	if err != nil && err.Error() != "exit status 1" {
		// exit status 1 is normal when it finds issues
		return nil, fmt.Errorf("failed to run pycodestyle: %w, stderr: %s", err, stderr.String())
	}
	
	// Parse the output and generate suggestions
	suggestions := make([]Issue, 0)
	lines := strings.Split(stdout.String(), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, ": ")
		if len(parts) < 2 {
			continue
		}
		
		// Parse location
		locParts := strings.Split(parts[0], ":")
		if len(locParts) < 3 {
			continue
		}
		
		lineNum, err := strconv.Atoi(locParts[1])
		if err != nil {
			continue
		}
		
		colNum, err := strconv.Atoi(locParts[2])
		if err != nil {
			continue
		}
		
		// Parse code and message
		codeMsgParts := strings.SplitN(parts[1], " ", 2)
		if len(codeMsgParts) < 2 {
			continue
		}
		
		ruleID := codeMsgParts[0]
		message := codeMsgParts[1]
		
		// Generate a suggestion
		suggestion := Issue{
			Line:     lineNum,
			Column:   &colNum,
			Message:  fmt.Sprintf("%s: %s", ruleID, message),
			Severity: "suggestion",
			RuleID:   ruleID,
			Fix:      l.generateFix(code, lineNum, colNum, ruleID),
		}
		
		suggestions = append(suggestions, suggestion)
	}
	
	return suggestions, nil
}

// convertPylintResult converts a pylint result to a CodeHawk issue
func (l *PythonLinter) convertPylintResult(result map[string]interface{}) Issue {
	// Extract basic information
	line := int(result["line"].(float64))
	column := int(result["column"].(float64))
	messageID := result["message-id"].(string)
	message := result["message"].(string)
	symbol := result["symbol"].(string)
	
	// Map pylint category to CodeHawk severity
	var severity string
	switch result["type"] {
	case "error", "fatal":
		severity = "error"
	case "warning":
		severity = "warning"
	case "convention", "refactor":
		severity = "suggestion"
	default:
		severity = "info"
	}
	
	// Create the issue
	issue := Issue{
		Line:     line,
		Column:   &column,
		Message:  message,
		Severity: severity,
		RuleID:   messageID,
		Context:  symbol,
	}
	
	return issue
}

// generateFix attempts to generate a fix for a specific issue
func (l *PythonLinter) generateFix(code string, line, column int, ruleID string) *IssueFix {
	// Split the code into lines
	lines := strings.Split(code, "\n")
	if line <= 0 || line > len(lines) {
		return nil
	}
	
	// Get the problematic line (0-indexed)
	lineIdx := line - 1
	lineText := lines[lineIdx]
	
	// Generate fixes based on rule ID
	switch ruleID {
	case "E201": // Whitespace after '('
		if column <= 0 || column >= len(lineText) {
			return nil
		}
		replacement := lineText[:column] + lineText[column+1:]
		return &IssueFix{
			Description: "Remove whitespace after '('",
			Replacement: replacement,
		}
		
	case "E202": // Whitespace before ')'
		if column <= 1 || column >= len(lineText) {
			return nil
		}
		replacement := lineText[:column-1] + lineText[column:]
		return &IssueFix{
			Description: "Remove whitespace before ')'",
			Replacement: replacement,
		}
		
	case "E225": // Missing whitespace around operator
		if column <= 0 || column >= len(lineText)-1 {
			return nil
		}
		// Try to identify the operator
		op := lineText[column : column+1]
		replacement := lineText[:column] + " " + op + " " + lineText[column+1:]
		return &IssueFix{
			Description: "Add whitespace around operator",
			Replacement: replacement,
		}
		
	case "E302": // Expected 2 blank lines
		return &IssueFix{
			Description: "Add 2 blank lines",
			Replacement: lineText + "\n\n",
		}
		
	case "C0111": // Missing docstring
	case "missing-docstring":
		indent := getIndentation(lineText)
		
		if strings.Contains(lineText, "def ") {
			// Function docstring
			replacement := lineText + "\n" + indent + "    \"\"\"Add docstring here.\"\"\""
			return &IssueFix{
				Description: "Add function docstring",
				Replacement: replacement,
			}
		} else if strings.Contains(lineText, "class ") {
			// Class docstring
			replacement := lineText + "\n" + indent + "    \"\"\"Add class docstring here.\"\"\""
			return &IssueFix{
				Description: "Add class docstring",
				Replacement: replacement,
			}
		} else if lineIdx == 0 && !strings.HasPrefix(lineText, "\"\"\"") {
			// Module docstring
			replacement := "\"\"\"Add module docstring here.\"\"\"\n\n" + lineText
			return &IssueFix{
				Description: "Add module docstring",
				Replacement: replacement,
			}
		}
		
	case "E303": // Too many blank lines
		return &IssueFix{
			Description: "Remove extra blank lines",
			Replacement: lineText,
		}
		
	case "W0311": // Bad indentation
	case "bad-indentation":
		// Try to fix indentation (assuming 4 spaces)
		indent := getIndentation(lineText)
		properIndent := getProperIndentation(indent)
		replacement := properIndent + strings.TrimLeft(lineText, " \t")
		return &IssueFix{
			Description: "Fix indentation",
			Replacement: replacement,
		}
	}
	
	return nil
}

// getIndentation returns the indentation string from a line
func getIndentation(line string) string {
	for i, c := range line {
		if c != ' ' && c != '\t' {
			return line[:i]
		}
	}
	return ""
}

// getProperIndentation attempts to correct the indentation
func getProperIndentation(currentIndent string) string {
	// Count number of spaces
	count := 0
	for _, c := range currentIndent {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 4 // Assume a tab is 4 spaces
		}
	}
	
	// Round to nearest 4 spaces
	properCount := (count + 2) / 4 * 4
	return strings.Repeat(" ", properCount)
}