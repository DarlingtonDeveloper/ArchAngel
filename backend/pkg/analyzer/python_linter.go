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
)

// PythonLinter implements the Linter interface for Python
type PythonLinter struct {
	*BaseAnalyzer
	pylintPath string
}

// NewPythonLinter creates a new Python linter
func NewPythonLinter(config map[string]string) *PythonLinter {
	// Default pylint path
	pylintPath := "pylint"
	if path, ok := config["pylintPath"]; ok && path != "" {
		pylintPath = path
	}
	
	return &PythonLinter{
		BaseAnalyzer: NewBaseAnalyzer(config),
		pylintPath:   pylintPath,
	}
}

// Language returns the identifier for the supported language
func (l *PythonLinter) Language() string {
	return "python"
}

// Analyze analyzes the provided code and returns issues found
func (l *PythonLinter) Analyze(ctx context.Context, code string, options map[string]interface{}) (*AnalysisResult, error) {
	return l.BaseAnalyzer.CommonAnalyzeWrapper(
		ctx,
		code,
		options,
		l.findIssues,
		l.SuggestFixes,
	)
}

// findIssues analyzes the code and returns issues
func (l *PythonLinter) findIssues(ctx context.Context, code string, options map[string]interface{}) ([]Issue, error) {
	// Create a temporary file for the code
	tmpFile, err := ioutil.TempFile("", "codehawk-*.py")
	if err != nil {
		return nil, l.WrapError(err, "failed to create temporary file")
	}
	defer os.Remove(tmpFile.Name())
	
	// Write the code to the temporary file
	if _, err := tmpFile.Write([]byte(code)); err != nil {
		return nil, l.WrapError(err, "failed to write to temporary file")
	}
	
	if err := tmpFile.Close(); err != nil {
		return nil, l.WrapError(err, "failed to close temporary file")
	}
	
	// Run pylint on the temporary file
	cmd := exec.CommandContext(
		ctx,
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
		return nil, l.WrapError(fmt.Errorf("failed to run pylint: %w, stderr: %s", err, stderr.String()), "pylint execution error")
	}
	
	// Parse the JSON output
	var pylintResults []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &pylintResults); err != nil {
		return nil, l.WrapError(err, "failed to parse pylint output")
	}
	
	// Convert pylint results to CodeHawk issues
	issues := make([]Issue, 0, len(pylintResults))
	for _, result := range pylintResults {
		issue := l.convertPylintResult(result)
		issues = append(issues, issue)
	}
	
	return issues, nil
}

// SuggestFixes attempts to generate fixes for the identified issues
func (l *PythonLinter) SuggestFixes(ctx context.Context, code string, issues []Issue) ([]Issue, error) {
	// Create a temporary file for the code
	tmpFile, err := ioutil.TempFile("", "codehawk-*.py")
	if err != nil {
		return nil, l.WrapError(err, "failed to create temporary file")
	}
	defer os.Remove(tmpFile.Name())
	
	// Write the code to the temporary file
	if _, err := tmpFile.Write([]byte(code)); err != nil {
		return nil, l.WrapError(err, "failed to write to temporary file")
	}
	
	if err := tmpFile.Close(); err != nil {
		return nil, l.WrapError(err, "failed to close temporary file")
	}
	
	// Run pycodestyle for suggestions
	cmd := exec.CommandContext(
		ctx,
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
		return nil, l.WrapError(fmt.Errorf("failed to run pycodestyle: %w, stderr: %s", err, stderr.String()), "pycodestyle execution error")
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
	
	// Also generate suggestions for existing issues
	for _, issue := range issues {
		fix := l.generateFix(code, issue.Line, issue.Column != nil ? *issue.Column : 0, issue.RuleID)
		if fix != nil {
			suggestion := Issue{
				Line:     issue.Line,
				Column:   issue.Column,
				Message:  fmt.Sprintf("Suggested fix for: %s", issue.Message),
				Severity: "suggestion",
				RuleID:   issue.RuleID,
				Fix:      fix,
			}
			suggestions = append(suggestions, suggestion)
		}
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
	
	// Map pylint category to CodeHawk severity using BaseAnalyzer method
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
		
	case "C0111", "missing-docstring":
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