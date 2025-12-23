package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ScanResult represents the result of a single scan
type ScanResult struct {
	ScanID   string        `json:"scanId"`
	ScanName string        `json:"scanName"`
	Category ScanCategory  `json:"category"`
	Duration time.Duration `json:"duration"`
	Findings []Finding     `json:"findings"`
	Summary  ScanSummary   `json:"summary"`
	Error    string        `json:"error,omitempty"`
	Skipped  bool          `json:"skipped,omitempty"`
}

// ScanSummary represents summary for a single scan
type ScanSummary struct {
	CriticalCount int `json:"criticalCount,omitempty"`
	HighCount     int `json:"highCount,omitempty"`
	MediumCount   int `json:"mediumCount,omitempty"`
	LowCount      int `json:"lowCount,omitempty"`
	TotalCount    int `json:"totalCount"`
}

// Finding represents a scan finding (reusable across all scans)
type Finding struct {
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	File           string `json:"file"`
	Line           int    `json:"line,omitempty"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
	CWE            string `json:"cwe,omitempty"`
}

// AggregatedReport represents the combined results of all scans
type AggregatedReport struct {
	Version      string            `json:"version"`
	Workspace    string            `json:"workspace"`
	AnalyzedAt   time.Time         `json:"analyzedAt"`
	Duration     string            `json:"duration"`
	Languages    []string          `json:"languages"`
	FilesCount   int               `json:"filesAnalyzed"`
	ScansRun     int               `json:"scansRun"`
	ScansSkipped int               `json:"scansSkipped"`
	ScansFailed  int               `json:"scansFailed"`
	Summary      AggregatedSummary `json:"summary"`
	ScanResults  []ScanResult      `json:"scanResults"`
	Error        string            `json:"error,omitempty"`
}

// AggregatedSummary represents combined summary across all scans
type AggregatedSummary struct {
	Security AggSecuritySummary `json:"security"`
	Quality  AggQualitySummary  `json:"quality"`
}

// AggSecuritySummary for aggregated report
type AggSecuritySummary struct {
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
	TotalCount    int `json:"totalCount"`
}

// AggQualitySummary for aggregated report
type AggQualitySummary struct {
	Score       int `json:"score"`
	HighCount   int `json:"highCount"`
	MediumCount int `json:"mediumCount"`
	LowCount    int `json:"lowCount"`
	TotalCount  int `json:"totalCount"`
}

// Executor handles parallel scan execution
type Executor struct {
	claudeCode    *ClaudeCode
	logger        *zap.Logger
	maxConcurrent int
}

// NewExecutor creates a new scan executor
func NewExecutor(claudeCode *ClaudeCode, logger *zap.Logger, maxConcurrent int) *Executor {
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default to 10 concurrent scans
	}
	return &Executor{
		claudeCode:    claudeCode,
		logger:        logger,
		maxConcurrent: maxConcurrent,
	}
}

// RunAllScans executes all applicable scans in parallel
func (e *Executor) RunAllScans(ctx context.Context, workspaceDir, workspaceName string, fileCount int) *AggregatedReport {
	startTime := time.Now()

	// Detect languages in the workspace
	languages, err := DetectLanguages(workspaceDir)
	if err != nil {
		e.logger.Warn("Failed to detect languages", zap.Error(err))
		languages = []Language{}
	}

	e.logger.Info("Detected languages",
		zap.String("workspace", workspaceName),
		zap.Strings("languages", LanguagesToStrings(languages)),
	)

	// Filter scans based on detected languages
	applicableScans := FilterScansForLanguages(ScanRegistry, languages)
	skippedCount := len(ScanRegistry) - len(applicableScans)

	e.logger.Info("Running scans",
		zap.String("workspace", workspaceName),
		zap.Int("total_scans", len(ScanRegistry)),
		zap.Int("applicable_scans", len(applicableScans)),
		zap.Int("skipped_scans", skippedCount),
	)

	// Create channels for results and semaphore for concurrency control
	results := make(chan *ScanResult, len(applicableScans))
	semaphore := make(chan struct{}, e.maxConcurrent)
	var wg sync.WaitGroup

	// Launch all scans in parallel (with concurrency limit)
	for _, scan := range applicableScans {
		wg.Add(1)
		go func(s ScanDefinition) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := e.runSingleScan(ctx, workspaceDir, s)
			results <- result
		}(scan)
	}

	// Wait for all scans to complete and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	scanResults := make([]ScanResult, 0, len(applicableScans))
	failedCount := 0
	for result := range results {
		scanResults = append(scanResults, *result)
		if result.Error != "" {
			failedCount++
		}
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Aggregate results
	report := &AggregatedReport{
		Version:      "2.0.0",
		Workspace:    workspaceName,
		AnalyzedAt:   startTime,
		Duration:     duration.String(),
		Languages:    LanguagesToStrings(languages),
		FilesCount:   fileCount,
		ScansRun:     len(applicableScans),
		ScansSkipped: skippedCount,
		ScansFailed:  failedCount,
		ScanResults:  scanResults,
		Summary:      e.aggregateSummary(scanResults),
	}

	e.logger.Info("Completed all scans",
		zap.String("workspace", workspaceName),
		zap.Duration("duration", duration),
		zap.Int("scans_run", len(applicableScans)),
		zap.Int("scans_failed", failedCount),
	)

	return report
}

// runSingleScan executes a single scan and returns the result
func (e *Executor) runSingleScan(ctx context.Context, workspaceDir string, scan ScanDefinition) *ScanResult {
	startTime := time.Now()

	e.logger.Debug("Starting scan",
		zap.String("scan_id", scan.ID),
		zap.String("scan_name", scan.Name),
	)

	// Run Claude Code with the scan's prompt
	output, err := e.claudeCode.runClaudeCode(ctx, workspaceDir, scan.Prompt)
	duration := time.Since(startTime)

	result := &ScanResult{
		ScanID:   scan.ID,
		ScanName: scan.Name,
		Category: scan.Category,
		Duration: duration,
		Findings: []Finding{},
		Summary:  ScanSummary{},
	}

	if err != nil {
		e.logger.Warn("Scan failed",
			zap.String("scan_id", scan.ID),
			zap.Error(err),
		)
		result.Error = err.Error()
		return result
	}

	// Parse the JSON output
	findings, summary, err := e.parseOutput(output, scan)
	if err != nil {
		e.logger.Warn("Failed to parse scan output",
			zap.String("scan_id", scan.ID),
			zap.Error(err),
		)
		result.Error = fmt.Sprintf("parse error: %v", err)
		return result
	}

	result.Findings = findings
	result.Summary = summary

	e.logger.Debug("Scan completed",
		zap.String("scan_id", scan.ID),
		zap.Duration("duration", duration),
		zap.Int("findings", len(findings)),
	)

	return result
}

// parseOutput parses the JSON output from Claude Code
func (e *Executor) parseOutput(output string, scan ScanDefinition) ([]Finding, ScanSummary, error) {
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return []Finding{}, ScanSummary{}, fmt.Errorf("no JSON found in output")
	}

	// Try to parse the standard format
	var parsed struct {
		Findings []Finding `json:"findings"`
		Summary  struct {
			CriticalCount int `json:"criticalCount"`
			HighCount     int `json:"highCount"`
			MediumCount   int `json:"mediumCount"`
			LowCount      int `json:"lowCount"`
			Count         int `json:"count"`
			Score         int `json:"score"`
		} `json:"summary"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return []Finding{}, ScanSummary{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Add CWE references to findings if available
	for i := range parsed.Findings {
		if len(scan.CWE) > 0 && parsed.Findings[i].CWE == "" {
			parsed.Findings[i].CWE = scan.CWE[0]
		}
	}

	summary := ScanSummary{
		CriticalCount: parsed.Summary.CriticalCount,
		HighCount:     parsed.Summary.HighCount,
		MediumCount:   parsed.Summary.MediumCount,
		LowCount:      parsed.Summary.LowCount,
		TotalCount:    len(parsed.Findings),
	}

	// If count field was used instead of individual counts
	if parsed.Summary.Count > 0 && summary.TotalCount == 0 {
		summary.TotalCount = parsed.Summary.Count
	}

	return parsed.Findings, summary, nil
}

// aggregateSummary combines summaries from all scan results
func (e *Executor) aggregateSummary(results []ScanResult) AggregatedSummary {
	summary := AggregatedSummary{
		Security: AggSecuritySummary{},
		Quality:  AggQualitySummary{Score: 100}, // Start with perfect score
	}

	qualityIssueCount := 0

	for _, result := range results {
		if result.Error != "" || result.Skipped {
			continue
		}

		switch result.Category {
		case CategorySecurity:
			summary.Security.CriticalCount += result.Summary.CriticalCount
			summary.Security.HighCount += result.Summary.HighCount
			summary.Security.MediumCount += result.Summary.MediumCount
			summary.Security.LowCount += result.Summary.LowCount
			summary.Security.TotalCount += result.Summary.TotalCount

		case CategoryQuality, CategoryLanguage:
			summary.Quality.HighCount += result.Summary.HighCount
			summary.Quality.MediumCount += result.Summary.MediumCount
			summary.Quality.LowCount += result.Summary.LowCount
			summary.Quality.TotalCount += result.Summary.TotalCount
			qualityIssueCount += result.Summary.TotalCount
		}
	}

	// Calculate quality score based on issues found
	// Deduct points: High=-10, Medium=-5, Low=-2
	deduction := summary.Quality.HighCount*10 + summary.Quality.MediumCount*5 + summary.Quality.LowCount*2
	summary.Quality.Score = max(0, 100-deduction)

	return summary
}
