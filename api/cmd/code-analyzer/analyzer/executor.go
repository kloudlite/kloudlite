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
	claudeCode       *ClaudeCode
	claudeAPI        *ClaudeAPI
	manifestMgr      *ManifestManager
	findingsCacheMgr *FindingsCacheManager
	logger           *zap.Logger
	maxConcurrent    int
	useDirectAPI     bool
	reportsPath      string
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
		useDirectAPI:  false, // Default to CLI mode for backward compatibility
	}
}

// NewExecutorWithAPI creates an executor that uses direct API calls with caching
func NewExecutorWithAPI(claudeAPI *ClaudeAPI, reportsPath string, logger *zap.Logger, maxConcurrent int) *Executor {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	return &Executor{
		claudeAPI:        claudeAPI,
		manifestMgr:      NewManifestManager(reportsPath, MaxFileSize, logger),
		findingsCacheMgr: NewFindingsCacheManager(reportsPath, logger),
		logger:           logger,
		maxConcurrent:    maxConcurrent,
		useDirectAPI:     true,
		reportsPath:      reportsPath,
	}
}

// RunAllScans executes all applicable scans in parallel
func (e *Executor) RunAllScans(ctx context.Context, workspaceDir, workspaceName string, fileCount int) *AggregatedReport {
	// Use incremental analysis if direct API is enabled
	if e.useDirectAPI {
		return e.RunIncrementalScans(ctx, workspaceDir, workspaceName, nil)
	}

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

// RunIncrementalScans runs scans using direct API with incremental analysis
// If changedFiles is nil, it computes the diff from the manifest
// If changedFiles is provided (from watcher), it uses those files directly
func (e *Executor) RunIncrementalScans(ctx context.Context, workspaceDir, workspaceName string, changedFiles []string) *AggregatedReport {
	startTime := time.Now()

	e.logger.Info("Starting incremental scan analysis",
		zap.String("workspace", workspaceName),
		zap.Int("changed_files_hint", len(changedFiles)),
	)

	// Load previous manifest
	lastManifest, err := e.manifestMgr.LoadManifest(workspaceName)
	if err != nil {
		e.logger.Warn("Failed to load manifest, running full analysis", zap.Error(err))
	}

	// Compute diff
	var diff *DiffResult
	if changedFiles != nil && len(changedFiles) > 0 {
		// Use provided changed files (from watcher)
		diff, err = e.manifestMgr.ComputeDiffFromChangedFiles(workspaceDir, changedFiles, lastManifest)
	} else {
		// Compute diff from manifest
		diff, err = e.manifestMgr.ComputeDiff(workspaceDir, lastManifest)
	}
	if err != nil {
		e.logger.Error("Failed to compute diff", zap.Error(err))
		return e.createErrorReport(workspaceName, startTime, err)
	}

	// Log diff summary
	e.logger.Info("Computed file diff",
		zap.Int("added", len(diff.Added)),
		zap.Int("modified", len(diff.Modified)),
		zap.Int("deleted", len(diff.Deleted)),
		zap.Int("unchanged", len(diff.Unchanged)),
		zap.Bool("has_changes", diff.HasChanges),
	)

	// Load existing findings cache
	existingCache, err := e.findingsCacheMgr.LoadCache(workspaceName)
	if err != nil {
		e.logger.Warn("Failed to load findings cache", zap.Error(err))
	}

	// If no changes, return cached report
	if !diff.HasChanges && existingCache != nil {
		e.logger.Info("No changes detected, returning cached report")
		return e.buildReportFromCache(workspaceName, startTime, existingCache, diff)
	}

	// Detect languages
	languages, err := DetectLanguages(workspaceDir)
	if err != nil {
		e.logger.Warn("Failed to detect languages", zap.Error(err))
		languages = []Language{}
	}

	// Filter applicable scans
	applicableScans := FilterScansForLanguages(ScanRegistry, languages)
	skippedCount := len(ScanRegistry) - len(applicableScans)

	// Determine files to analyze
	filesToAnalyze := diff.GetChangedFiles()
	if lastManifest == nil {
		// First analysis: analyze all files
		filesToAnalyze = diff.AllFiles
	}

	e.logger.Info("Running incremental scans",
		zap.Int("files_to_analyze", len(filesToAnalyze)),
		zap.Int("applicable_scans", len(applicableScans)),
	)

	// Build codebase content for changed files
	codebaseContent, err := BuildCodebaseContent(workspaceDir, filesToAnalyze, MaxFileSize)
	if err != nil {
		e.logger.Error("Failed to build codebase content", zap.Error(err))
		return e.createErrorReport(workspaceName, startTime, err)
	}

	// Build scan requests
	scanRequests := make([]ScanRequest, len(applicableScans))
	for i, scan := range applicableScans {
		scanRequests[i] = ScanRequest{
			ScanID:   scan.ID,
			ScanName: scan.Name,
			Category: scan.Category,
			Prompt:   scan.Prompt,
		}
	}

	// Run all scans with cached context
	scanResponses := e.claudeAPI.RunScansWithCache(ctx, codebaseContent, scanRequests, e.maxConcurrent)

	// Convert responses to results and collect findings
	scanResults := make([]ScanResult, len(scanResponses))
	var newFindings []Finding
	failedCount := 0

	for i, resp := range scanResponses {
		scanResults[i] = ScanResult{
			ScanID:   resp.ScanID,
			ScanName: resp.ScanName,
			Category: resp.Category,
			Duration: resp.Duration,
			Findings: resp.Findings,
			Summary:  resp.Summary,
			Error:    resp.Error,
		}

		if resp.Error != "" {
			failedCount++
		} else {
			newFindings = append(newFindings, resp.Findings...)
		}
	}

	// Merge findings with cached findings for unchanged files
	allFindings := MergeFindings(existingCache, newFindings, filesToAnalyze, diff.Deleted)

	// Deduplicate findings to stabilize results across non-deterministic Claude outputs
	allFindings = DeduplicateFindings(allFindings)

	// Compute new manifest
	newManifest, err := e.manifestMgr.ComputeCurrentManifest(workspaceDir, workspaceName)
	if err != nil {
		e.logger.Warn("Failed to compute new manifest", zap.Error(err))
	}

	// Deduplicate new findings before caching
	deduplicatedNewFindings := DeduplicateFindings(newFindings)

	// Update findings cache
	newCache := e.findingsCacheMgr.UpdateCache(workspaceName, existingCache, deduplicatedNewFindings, filesToAnalyze, newManifest)

	// Save manifest and cache
	if newManifest != nil {
		if err := e.manifestMgr.SaveManifest(workspaceName, newManifest); err != nil {
			e.logger.Warn("Failed to save manifest", zap.Error(err))
		}
	}
	if newCache != nil {
		if err := e.findingsCacheMgr.SaveCache(workspaceName, newCache); err != nil {
			e.logger.Warn("Failed to save findings cache", zap.Error(err))
		}
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Build report
	report := &AggregatedReport{
		Version:      "2.1.0", // New version for incremental analysis
		Workspace:    workspaceName,
		AnalyzedAt:   startTime,
		Duration:     duration.String(),
		Languages:    LanguagesToStrings(languages),
		FilesCount:   len(diff.AllFiles),
		ScansRun:     len(applicableScans),
		ScansSkipped: skippedCount,
		ScansFailed:  failedCount,
		ScanResults:  scanResults,
		Summary:      e.aggregateSummaryFromFindings(allFindings),
	}

	e.logger.Info("Completed incremental scan analysis",
		zap.String("workspace", workspaceName),
		zap.Duration("duration", duration),
		zap.Int("scans_run", len(applicableScans)),
		zap.Int("scans_failed", failedCount),
		zap.Int("total_findings", len(allFindings)),
	)

	return report
}

// createErrorReport creates an error report
func (e *Executor) createErrorReport(workspaceName string, startTime time.Time, err error) *AggregatedReport {
	return &AggregatedReport{
		Version:    "2.1.0",
		Workspace:  workspaceName,
		AnalyzedAt: startTime,
		Duration:   time.Since(startTime).String(),
		Error:      err.Error(),
	}
}

// buildReportFromCache builds a report from cached findings
func (e *Executor) buildReportFromCache(workspaceName string, startTime time.Time, cache *FindingsCache, diff *DiffResult) *AggregatedReport {
	allFindings := e.findingsCacheMgr.GetAllFindings(cache)

	// Deduplicate findings from cache
	allFindings = DeduplicateFindings(allFindings)

	return &AggregatedReport{
		Version:    "2.1.0",
		Workspace:  workspaceName,
		AnalyzedAt: startTime,
		Duration:   time.Since(startTime).String(),
		FilesCount: len(diff.AllFiles),
		Summary:    e.aggregateSummaryFromFindings(allFindings),
	}
}

// aggregateSummaryFromFindings calculates summary from a list of findings
func (e *Executor) aggregateSummaryFromFindings(findings []Finding) AggregatedSummary {
	summary := AggregatedSummary{
		Security: AggSecuritySummary{},
		Quality:  AggQualitySummary{Score: 100},
	}

	for _, f := range findings {
		// Determine if security or quality based on finding ID prefix
		isSecurity := len(f.ID) > 3 && f.ID[:3] == "SEC"

		switch f.Severity {
		case "critical":
			if isSecurity {
				summary.Security.CriticalCount++
			}
		case "high":
			if isSecurity {
				summary.Security.HighCount++
			} else {
				summary.Quality.HighCount++
			}
		case "medium":
			if isSecurity {
				summary.Security.MediumCount++
			} else {
				summary.Quality.MediumCount++
			}
		case "low":
			if isSecurity {
				summary.Security.LowCount++
			} else {
				summary.Quality.LowCount++
			}
		}

		if isSecurity {
			summary.Security.TotalCount++
		} else {
			summary.Quality.TotalCount++
		}
	}

	// Calculate quality score
	deduction := summary.Quality.HighCount*10 + summary.Quality.MediumCount*5 + summary.Quality.LowCount*2
	summary.Quality.Score = max(0, 100-deduction)

	return summary
}
