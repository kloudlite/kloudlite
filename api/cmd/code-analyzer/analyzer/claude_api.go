package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ClaudeAPI provides direct access to Claude API with prompt caching support
type ClaudeAPI struct {
	apiURL     string
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClaudeAPI creates a new Claude API client
func NewClaudeAPI(apiURL, apiKey string, logger *zap.Logger) *ClaudeAPI {
	return &ClaudeAPI{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  "claude-sonnet-4-20250514",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		logger: logger,
	}
}

// CacheControl for prompt caching
type CacheControl struct {
	Type string `json:"type"`
}

// SystemBlock represents a system message block with optional cache control
type SystemBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents a Claude API request
type ClaudeRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    []SystemBlock `json:"system,omitempty"`
	Messages  []Message     `json:"messages"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage represents token usage in the response
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// ClaudeResponse represents a Claude API response
type ClaudeResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

// ClaudeError represents an error response from Claude API
type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ClaudeErrorResponse wraps the error
type ClaudeErrorResponse struct {
	Type  string      `json:"type"`
	Error ClaudeError `json:"error"`
}

// ScanRequest represents a single scan to run
type ScanRequest struct {
	ScanID   string
	ScanName string
	Category ScanCategory
	Prompt   string
}

// ScanResponse represents the result of a single scan
type ScanResponse struct {
	ScanID   string
	ScanName string
	Category ScanCategory
	Findings []Finding
	Summary  ScanSummary
	Duration time.Duration
	Error    string
	Usage    Usage
}

// RunScansWithCache runs multiple scans against a cached codebase context
// This leverages Claude's prompt caching to avoid re-sending the codebase for each scan
func (c *ClaudeAPI) RunScansWithCache(ctx context.Context, codebaseContent string, scans []ScanRequest, maxConcurrent int) []ScanResponse {
	results := make([]ScanResponse, len(scans))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent)

	c.logger.Info("Starting cached scan execution",
		zap.Int("total_scans", len(scans)),
		zap.Int("max_concurrent", maxConcurrent),
		zap.Int("codebase_size", len(codebaseContent)),
	)

	for i, scan := range scans {
		wg.Add(1)
		go func(idx int, s ScanRequest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := c.runSingleScan(ctx, codebaseContent, s)
			results[idx] = result
		}(i, scan)
	}

	wg.Wait()
	return results
}

// runSingleScan executes a single scan against the cached codebase
func (c *ClaudeAPI) runSingleScan(ctx context.Context, codebaseContent string, scan ScanRequest) ScanResponse {
	startTime := time.Now()

	response := ScanResponse{
		ScanID:   scan.ScanID,
		ScanName: scan.ScanName,
		Category: scan.Category,
		Findings: []Finding{},
		Summary:  ScanSummary{},
	}

	// Build request with cached codebase in system message
	// Include strict analysis rules to ensure evidence-based findings
	systemPrompt := fmt.Sprintf(`<codebase>
%s
</codebase>

You are an EXTREMELY CONSERVATIVE code security analyzer. FALSE POSITIVES ARE UNACCEPTABLE.

BEFORE REPORTING ANY ISSUE, YOU MUST:
1. Find the EXACT vulnerable code pattern
2. Trace the data flow from source to sink
3. Search the ENTIRE codebase for mitigations (validation, sanitization, safe wrappers)
4. If ANY mitigation exists ANYWHERE in the flow, DO NOT REPORT

COMMON MITIGATIONS TO CHECK (if present, issue is NOT vulnerable):
- Input validation: isValid*, validate*, check*, regex checks
- Sanitization: html.EscapeString, escape*, sanitize*, clean*
- Safe APIs: subtle.ConstantTimeCompare, prepared statements, parameterized queries
- Synchronization: sync.Mutex, sync.RWMutex, sync.Once, channels
- Resource cleanup: defer Close(), defer cleanup patterns
- Allowlists: explicit field lists, whitelist checks
- Type safety: strong typing that prevents injection
- Hardcoded values: constants, config loaded at startup (not user input)

REPORT ONLY IF:
- Vulnerable pattern exists AND
- User-controlled input reaches it AND
- NO mitigation exists ANYWHERE in the codebase AND
- You can prove exploitability

WHEN IN DOUBT, DO NOT REPORT. Empty findings is better than false positives.

Output ONLY valid JSON. If no confirmed issues: {"findings":[],"summary":{"count":0}}`, codebaseContent)

	req := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 4096,
		System: []SystemBlock{
			{
				Type:         "text",
				Text:         systemPrompt,
				CacheControl: &CacheControl{Type: "ephemeral"},
			},
		},
		Messages: []Message{
			{
				Role:    "user",
				Content: scan.Prompt,
			},
		},
	}

	// Make API request
	respBody, err := c.post(ctx, req)
	response.Duration = time.Since(startTime)

	if err != nil {
		response.Error = err.Error()
		c.logger.Error("Scan failed",
			zap.String("scan_id", scan.ScanID),
			zap.Error(err),
		)
		return response
	}

	// Parse response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		response.Error = fmt.Sprintf("failed to parse response: %v", err)
		c.logger.Error("Failed to parse Claude response",
			zap.String("scan_id", scan.ScanID),
			zap.Error(err),
		)
		return response
	}

	response.Usage = claudeResp.Usage

	// Extract text content
	var textContent string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			textContent += block.Text
		}
	}

	// Parse JSON findings from text content
	findings, summary, err := c.parseFindings(textContent)
	if err != nil {
		response.Error = fmt.Sprintf("failed to parse findings: %v", err)
		c.logger.Warn("Failed to parse findings",
			zap.String("scan_id", scan.ScanID),
			zap.Error(err),
			zap.String("content_preview", truncateString(textContent, 200)),
		)
		return response
	}

	// Verify findings to eliminate false positives
	verifiedFindings := c.verifyFindings(ctx, codebaseContent, findings)

	response.Findings = verifiedFindings
	response.Summary = summary
	response.Summary.TotalCount = len(verifiedFindings)

	c.logger.Info("Scan completed",
		zap.String("scan_id", scan.ScanID),
		zap.Duration("duration", response.Duration),
		zap.Int("candidates", len(findings)),
		zap.Int("verified", len(verifiedFindings)),
		zap.Int("cache_read_tokens", claudeResp.Usage.CacheReadInputTokens),
		zap.Int("cache_creation_tokens", claudeResp.Usage.CacheCreationInputTokens),
	)

	return response
}

// verifyFindings performs a second pass to verify each finding is actually exploitable
func (c *ClaudeAPI) verifyFindings(ctx context.Context, codebaseContent string, findings []Finding) []Finding {
	if len(findings) == 0 {
		return findings
	}

	verified := make([]Finding, 0)

	for _, f := range findings {
		if c.isVerifiedVulnerability(ctx, codebaseContent, f) {
			verified = append(verified, f)
		} else {
			c.logger.Debug("Finding rejected by verification",
				zap.String("id", f.ID),
				zap.String("title", f.Title),
				zap.String("file", f.File),
				zap.Int("line", f.Line),
			)
		}
	}

	return verified
}

// isVerifiedVulnerability asks Claude to verify if a finding is actually exploitable
func (c *ClaudeAPI) isVerifiedVulnerability(ctx context.Context, codebaseContent string, f Finding) bool {
	verifyPrompt := fmt.Sprintf(`You are a security code reviewer. Your job is to VERIFY if a reported issue is a REAL vulnerability.

REPORTED ISSUE:
- Title: %s
- File: %s, Line: %d
- Description: %s

TASK: Search the codebase for MITIGATIONS that would make this issue NOT exploitable.

Check for:
1. Input validation before the vulnerable code (isValid*, validate*, check*)
2. Sanitization (html.EscapeString, escape*, sanitize*)
3. Parameterized queries (?, $1, :param placeholders)
4. Synchronization (sync.Mutex, sync.RWMutex, sync.Once)
5. Resource cleanup (defer Close(), connection pools)
6. Hardcoded values instead of user input
7. Any other mitigation that prevents exploitation

RESPOND WITH ONLY ONE WORD:
- "VULNERABLE" if the issue is real and exploitable with NO mitigations
- "MITIGATED" if ANY mitigation exists that prevents exploitation

Remember: When in doubt, say MITIGATED. False positives are worse than missing issues.`, f.Title, f.File, f.Line, f.Description)

	req := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 50,
		System: []SystemBlock{
			{
				Type:         "text",
				Text:         "<codebase>\n" + codebaseContent + "\n</codebase>",
				CacheControl: &CacheControl{Type: "ephemeral"},
			},
		},
		Messages: []Message{
			{Role: "user", Content: verifyPrompt},
		},
	}

	respBody, err := c.post(ctx, req)
	if err != nil {
		c.logger.Warn("Verification request failed, rejecting finding", zap.Error(err))
		return false // Fail closed - if we can't verify, reject the finding
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return false
	}

	// Extract response text
	var response string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			response = strings.TrimSpace(strings.ToUpper(block.Text))
			break
		}
	}

	// Only accept if explicitly confirmed as VULNERABLE
	return strings.Contains(response, "VULNERABLE") && !strings.Contains(response, "MITIGATED")
}

// post sends a POST request to the Claude API
func (c *ClaudeAPI) post(ctx context.Context, req ClaudeRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ClaudeErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// parseFindings extracts findings from Claude's response text
func (c *ClaudeAPI) parseFindings(text string) ([]Finding, ScanSummary, error) {
	// Extract JSON from response (might be wrapped in markdown code blocks)
	jsonStr := extractJSON(text)
	if jsonStr == "" {
		return []Finding{}, ScanSummary{}, nil
	}

	// Try to parse as findings response
	var result struct {
		Findings []Finding   `json:"findings"`
		Summary  ScanSummary `json:"summary"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, ScanSummary{}, fmt.Errorf("failed to parse findings JSON: %w", err)
	}

	// Calculate summary counts if not provided
	if result.Summary.TotalCount == 0 {
		for _, f := range result.Findings {
			switch f.Severity {
			case "critical":
				result.Summary.CriticalCount++
			case "high":
				result.Summary.HighCount++
			case "medium":
				result.Summary.MediumCount++
			case "low":
				result.Summary.LowCount++
			}
			result.Summary.TotalCount++
		}
	}

	return result.Findings, result.Summary, nil
}

// BuildCodebaseContent reads files and builds a single content string for the codebase
func BuildCodebaseContent(workspaceDir string, files []string, maxFileSize int64) (string, error) {
	var content bytes.Buffer

	for _, file := range files {
		fileContent, err := readFileContent(workspaceDir, file, maxFileSize)
		if err != nil {
			continue // Skip files that can't be read
		}

		content.WriteString(fmt.Sprintf("=== File: %s ===\n", file))
		content.WriteString(fileContent)
		content.WriteString("\n\n")
	}

	return content.String(), nil
}

// readFileContent reads a single file's content
func readFileContent(workspaceDir, relativePath string, maxSize int64) (string, error) {
	fullPath := filepath.Join(workspaceDir, relativePath)

	// Check file size first
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	if info.Size() > maxSize {
		return "", fmt.Errorf("file too large: %d > %d", info.Size(), maxSize)
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
