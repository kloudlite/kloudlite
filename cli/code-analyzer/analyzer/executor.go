package analyzer

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ScanResult represents the result of a scan
type ScanResult struct {
	ScanID   string        `json:"scanId"`
	ScanName string        `json:"scanName"`
	Category string        `json:"category"`
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

// Finding represents a scan finding
type Finding struct {
	ID             string   `json:"id"`
	Severity       string   `json:"severity"`
	Category       string   `json:"category"`
	File           string   `json:"file"`
	Line           int      `json:"line,omitempty"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Recommendation string   `json:"recommendation"`
	CWE            string   `json:"cwe,omitempty"`
	OWASP          string   `json:"owasp,omitempty"`
	Evidence       string   `json:"evidence,omitempty"`
	Tags           []string `json:"tags,omitempty"`
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

// Executor handles scan execution using Semgrep
type Executor struct {
	semgrep     *SemgrepRunner
	reportsPath string
	logger      *zap.Logger
}

// NewExecutor creates a new scan executor with Semgrep
func NewExecutor(reportsPath string, logger *zap.Logger) *Executor {
	return &Executor{
		semgrep:     NewSemgrepRunner(logger),
		reportsPath: reportsPath,
		logger:      logger,
	}
}

// NewExecutorWithConfig creates an executor with custom Semgrep config
func NewExecutorWithConfig(config SemgrepConfig, reportsPath string, logger *zap.Logger) *Executor {
	return &Executor{
		semgrep:     NewSemgrepRunnerWithConfig(config, logger),
		reportsPath: reportsPath,
		logger:      logger,
	}
}

// RunAllScans executes Semgrep scan on the workspace
func (e *Executor) RunAllScans(ctx context.Context, workspaceDir, workspaceName string, fileCount int) *AggregatedReport {
	startTime := time.Now()

	e.logger.Info("Starting Semgrep analysis",
		zap.String("workspace", workspaceName),
		zap.String("directory", workspaceDir),
	)

	// Run Semgrep scan
	result, err := e.semgrep.Scan(ctx, workspaceDir)
	if err != nil {
		e.logger.Error("Semgrep scan failed", zap.Error(err))
		return e.createErrorReport(workspaceName, startTime, err)
	}

	// Detect languages for report metadata
	languages := DetectLanguages(workspaceDir)

	// Categorize findings
	securityFindings, qualityFindings := e.categorizeFindings(result.Findings)

	// Build scan results
	scanResults := []ScanResult{
		{
			ScanID:   "SEMGREP-SECURITY",
			ScanName: "Semgrep Security Scan",
			Category: "security",
			Duration: result.Duration,
			Findings: securityFindings,
			Summary:  e.buildSummary(securityFindings),
		},
		{
			ScanID:   "SEMGREP-QUALITY",
			ScanName: "Semgrep Quality Scan",
			Category: "quality",
			Duration: result.Duration,
			Findings: qualityFindings,
			Summary:  e.buildSummary(qualityFindings),
		},
	}

	// Handle any error from Semgrep
	if result.Error != "" {
		scanResults[0].Error = result.Error
		scanResults[1].Error = result.Error
	}

	duration := time.Since(startTime)

	// Build aggregated report
	report := &AggregatedReport{
		Version:     "3.0.0", // New version for Semgrep-based analysis
		Workspace:   workspaceName,
		AnalyzedAt:  startTime,
		Duration:    duration.String(),
		Languages:   languages,
		FilesCount:  fileCount,
		ScansRun:    1, // Semgrep runs as single scan
		ScanResults: scanResults,
		Summary:     e.aggregateSummary(result.Findings),
	}

	if result.Error != "" {
		report.ScansFailed = 1
	}

	e.logger.Info("Completed Semgrep analysis",
		zap.String("workspace", workspaceName),
		zap.Duration("duration", duration),
		zap.Int("security_findings", len(securityFindings)),
		zap.Int("quality_findings", len(qualityFindings)),
	)

	return report
}

// categorizeFindings splits findings into security and quality categories
func (e *Executor) categorizeFindings(findings []Finding) (security []Finding, quality []Finding) {
	for _, f := range findings {
		if e.isSecurityFinding(f) {
			security = append(security, f)
		} else {
			quality = append(quality, f)
		}
	}
	return
}

// isSecurityFinding determines if a finding is security-related
func (e *Executor) isSecurityFinding(f Finding) bool {
	// Check category
	if f.Category == "security" {
		return true
	}

	// Check tags for security indicators
	for _, tag := range f.Tags {
		if strings.HasPrefix(tag, "SEC_") {
			return true
		}
	}

	// Check rule ID for security patterns
	ruleID := strings.ToLower(f.ID)
	securityPatterns := []string{
		"security", "injection", "xss", "ssrf", "secret",
		"credential", "password", "auth", "crypto", "traversal",
		"xxe", "deserial", "command", "sqli", "rce",
	}
	for _, pattern := range securityPatterns {
		if strings.Contains(ruleID, pattern) {
			return true
		}
	}

	return false
}

// buildSummary creates a summary from findings
func (e *Executor) buildSummary(findings []Finding) ScanSummary {
	summary := ScanSummary{
		TotalCount: len(findings),
	}

	for _, f := range findings {
		switch f.Severity {
		case "critical":
			summary.CriticalCount++
		case "high":
			summary.HighCount++
		case "medium":
			summary.MediumCount++
		case "low":
			summary.LowCount++
		}
	}

	return summary
}

// aggregateSummary builds the aggregated summary
func (e *Executor) aggregateSummary(findings []Finding) AggregatedSummary {
	summary := AggregatedSummary{
		Security: AggSecuritySummary{},
		Quality:  AggQualitySummary{Score: 100},
	}

	for _, f := range findings {
		isSecurity := e.isSecurityFinding(f)

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

	// Calculate quality score based on issues found
	deduction := summary.Quality.HighCount*10 + summary.Quality.MediumCount*5 + summary.Quality.LowCount*2
	summary.Quality.Score = max(0, 100-deduction)

	return summary
}

// createErrorReport creates an error report
func (e *Executor) createErrorReport(workspaceName string, startTime time.Time, err error) *AggregatedReport {
	return &AggregatedReport{
		Version:     "3.0.0",
		Workspace:   workspaceName,
		AnalyzedAt:  startTime,
		Duration:    time.Since(startTime).String(),
		ScansFailed: 1,
		Error:       err.Error(),
	}
}
