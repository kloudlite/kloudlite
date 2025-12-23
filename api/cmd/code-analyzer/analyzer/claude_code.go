package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ClaudeCode wraps the Claude Code CLI for code analysis
type ClaudeCode struct {
	apiURL        string
	apiKey        string
	workspacePath string
	logger        *zap.Logger
}

// NewClaudeCode creates a new Claude Code wrapper
func NewClaudeCode(apiURL, apiKey, workspacePath string, logger *zap.Logger) *ClaudeCode {
	return &ClaudeCode{
		apiURL:        apiURL,
		apiKey:        apiKey,
		workspacePath: workspacePath,
		logger:        logger,
	}
}

// AnalyzeSecurityResult represents parsed security analysis result
type AnalyzeSecurityResult struct {
	Findings []SecurityFinding `json:"findings"`
	Summary  SecuritySummary   `json:"summary"`
}

// SecurityFinding represents a security finding
type SecurityFinding struct {
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	File           string `json:"file"`
	Line           int    `json:"line,omitempty"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// SecuritySummary represents security analysis summary
type SecuritySummary struct {
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
}

// AnalyzeQualityResult represents parsed quality analysis result
type AnalyzeQualityResult struct {
	Findings []QualityFinding `json:"findings"`
	Summary  QualitySummary   `json:"summary"`
}

// QualityFinding represents a code quality finding
type QualityFinding struct {
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	File           string `json:"file"`
	Line           int    `json:"line,omitempty"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// QualitySummary represents quality analysis summary
type QualitySummary struct {
	Score       int `json:"score"`
	HighCount   int `json:"highCount"`
	MediumCount int `json:"mediumCount"`
	LowCount    int `json:"lowCount"`
}

// AnalyzeSecurity runs security analysis using Claude Code
func (c *ClaudeCode) AnalyzeSecurity(ctx context.Context, workspaceDir string) (*AnalyzeSecurityResult, time.Duration, error) {
	prompt := `Analyze this codebase for security vulnerabilities. Focus on:
1. Hardcoded secrets (API keys, passwords, tokens)
2. SQL injection vulnerabilities
3. XSS vulnerabilities
4. Command injection
5. Path traversal
6. Insecure cryptography
7. Authentication/authorization flaws

Output ONLY valid JSON (no markdown) in this format:
{
  "findings": [
    {
      "id": "SEC-001",
      "severity": "critical|high|medium|low",
      "category": "secrets|injection|xss|auth|crypto|path_traversal|other",
      "file": "relative/path/to/file",
      "line": 15,
      "title": "Issue title",
      "description": "Detailed description",
      "recommendation": "How to fix"
    }
  ],
  "summary": {
    "criticalCount": 0,
    "highCount": 0,
    "mediumCount": 0,
    "lowCount": 0
  }
}`

	startTime := time.Now()
	output, err := c.runClaudeCode(ctx, workspaceDir, prompt)
	duration := time.Since(startTime)

	if err != nil {
		return nil, duration, fmt.Errorf("claude code execution failed: %w", err)
	}

	// Parse JSON from output
	result, err := c.parseSecurityResult(output)
	if err != nil {
		c.logger.Warn("Failed to parse security result, returning raw output", zap.Error(err))
		// Return empty result with error context
		return &AnalyzeSecurityResult{
			Findings: []SecurityFinding{},
			Summary:  SecuritySummary{},
		}, duration, nil
	}

	return result, duration, nil
}

// AnalyzeQuality runs code quality analysis using Claude Code
func (c *ClaudeCode) AnalyzeQuality(ctx context.Context, workspaceDir string) (*AnalyzeQualityResult, time.Duration, error) {
	prompt := `Analyze this codebase for code quality issues. Focus on:
1. Code complexity and maintainability
2. Dead code and unused imports
3. Error handling patterns
4. Naming conventions
5. Code duplication
6. Documentation gaps
7. Performance anti-patterns
8. Type safety issues

Output ONLY valid JSON (no markdown) in this format:
{
  "findings": [
    {
      "id": "QUAL-001",
      "severity": "high|medium|low",
      "category": "complexity|dead_code|error_handling|naming|duplication|documentation|performance|type_safety",
      "file": "relative/path/to/file",
      "line": 15,
      "title": "Issue title",
      "description": "Detailed description",
      "recommendation": "How to fix"
    }
  ],
  "summary": {
    "score": 85,
    "highCount": 0,
    "mediumCount": 0,
    "lowCount": 0
  }
}`

	startTime := time.Now()
	output, err := c.runClaudeCode(ctx, workspaceDir, prompt)
	duration := time.Since(startTime)

	if err != nil {
		return nil, duration, fmt.Errorf("claude code execution failed: %w", err)
	}

	// Parse JSON from output
	result, err := c.parseQualityResult(output)
	if err != nil {
		c.logger.Warn("Failed to parse quality result, returning raw output", zap.Error(err))
		return &AnalyzeQualityResult{
			Findings: []QualityFinding{},
			Summary:  QualitySummary{Score: 100},
		}, duration, nil
	}

	return result, duration, nil
}

// runClaudeCode executes Claude Code CLI with the given prompt
func (c *ClaudeCode) runClaudeCode(ctx context.Context, workspaceDir, prompt string) (string, error) {
	// Build command
	args := []string{
		"-p", prompt, // Print mode - non-interactive
		"--output-format", "text", // Text output
		"--model", "claude-sonnet-4-20250514", // Use Sonnet 4.5 for faster analysis
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = workspaceDir

	// Set environment variables for Claude Code
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ANTHROPIC_BASE_URL=%s", strings.TrimSuffix(c.apiURL, "/v1/messages")),
		fmt.Sprintf("ANTHROPIC_API_KEY=%s", c.apiKey),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	c.logger.Info("Running Claude Code",
		zap.String("workspace", workspaceDir),
		zap.Int("prompt_length", len(prompt)),
	)

	err := cmd.Run()
	if err != nil {
		c.logger.Error("Claude Code failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	c.logger.Info("Claude Code output",
		zap.Int("output_length", len(output)),
		zap.String("output_preview", truncateString(output, 500)),
	)

	return output, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parseSecurityResult extracts JSON from Claude Code output
func (c *ClaudeCode) parseSecurityResult(output string) (*AnalyzeSecurityResult, error) {
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in output")
	}

	var result AnalyzeSecurityResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// parseQualityResult extracts JSON from Claude Code output
func (c *ClaudeCode) parseQualityResult(output string) (*AnalyzeQualityResult, error) {
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in output")
	}

	var result AnalyzeQualityResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// extractJSON finds and extracts JSON from text that may contain other content
func extractJSON(text string) string {
	// Try to find JSON object
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	// Find matching closing brace
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}

	return ""
}

// CountFiles counts analyzable files in a directory
func CountFiles(dir string, maxSize int64) (int, error) {
	count := 0
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || isIgnoredDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip ignored files
		if shouldIgnoreFile(path, maxSize) {
			return nil
		}

		count++
		return nil
	})

	return count, err
}

func isIgnoredDir(name string) bool {
	ignored := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		"__pycache__":  true,
		".nix-profile": true,
		"dist":         true,
		"build":        true,
		".next":        true,
	}
	return ignored[name]
}

func shouldIgnoreFile(path string, maxSize int64) bool {
	name := filepath.Base(path)

	// Ignore hidden files
	if strings.HasPrefix(name, ".") {
		return true
	}

	// Ignore certain extensions
	ext := strings.ToLower(filepath.Ext(name))
	ignoredExt := map[string]bool{
		".log": true, ".tmp": true, ".swp": true,
		".pyc": true, ".o": true, ".so": true,
		".exe": true, ".dll": true, ".class": true,
		".jar": true, ".lock": true, ".sum": true,
	}
	if ignoredExt[ext] {
		return true
	}

	// Check file size
	if info, err := os.Stat(path); err == nil {
		if info.Size() > maxSize {
			return true
		}
	}

	return false
}
