package analyzer

import (
	"context"
	"os"
	"path/filepath"

	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/storage"
	"go.uber.org/zap"
)

const (
	ReportVersion     = "3.0.0"
	AggregatedVersion = "3.0.0"
	MaxFileSize       = 100 * 1024 // 100KB
)

// Analyzer orchestrates code analysis using Semgrep
type Analyzer struct {
	executor       *Executor
	storage        *storage.Storage
	workspacesPath string
	logger         *zap.Logger
}

// NewAnalyzer creates a new analyzer with default Semgrep configuration
func NewAnalyzer(
	storage *storage.Storage,
	workspacesPath string,
	logger *zap.Logger,
) *Analyzer {
	return &Analyzer{
		executor:       NewExecutor(storage.GetBasePath(), logger),
		storage:        storage,
		workspacesPath: workspacesPath,
		logger:         logger,
	}
}

// NewAnalyzerWithConfig creates a new analyzer with custom Semgrep configuration
func NewAnalyzerWithConfig(
	config SemgrepConfig,
	storage *storage.Storage,
	workspacesPath string,
	logger *zap.Logger,
) *Analyzer {
	return &Analyzer{
		executor:       NewExecutorWithConfig(config, storage.GetBasePath(), logger),
		storage:        storage,
		workspacesPath: workspacesPath,
		logger:         logger,
	}
}

// AnalyzeWorkspace runs Semgrep analysis on a workspace
func (a *Analyzer) AnalyzeWorkspace(ctx context.Context, workspaceName string) error {
	workspaceDir := filepath.Join(a.workspacesPath, workspaceName)

	a.logger.Info("Starting workspace analysis",
		zap.String("workspace", workspaceName),
		zap.String("directory", workspaceDir),
	)

	// Count files for metadata
	fileCount, err := CountFiles(workspaceDir, MaxFileSize)
	if err != nil {
		a.logger.Warn("Failed to count files", zap.Error(err))
		fileCount = 0
	}

	// Run Semgrep scan
	report := a.executor.RunAllScans(ctx, workspaceDir, workspaceName, fileCount)

	// Convert and save aggregated report
	aggregatedReport := a.convertToStorageReport(report)
	if err := a.storage.SaveAggregatedReport(workspaceName, aggregatedReport); err != nil {
		a.logger.Error("Failed to save aggregated report", zap.Error(err))
		return err
	}

	// Generate legacy security and quality reports for backward compatibility
	a.generateLegacyReports(workspaceName, report)

	a.logger.Info("Completed workspace analysis",
		zap.String("workspace", workspaceName),
		zap.String("duration", report.Duration),
		zap.Int("security_findings", report.Summary.Security.TotalCount),
		zap.Int("quality_findings", report.Summary.Quality.TotalCount),
	)

	return nil
}

// convertToStorageReport converts executor AggregatedReport to storage format
func (a *Analyzer) convertToStorageReport(report *AggregatedReport) *storage.AggregatedReport {
	scanResults := make([]storage.ScanResult, len(report.ScanResults))
	for i, sr := range report.ScanResults {
		findings := make([]storage.Finding, len(sr.Findings))
		for j, f := range sr.Findings {
			findings[j] = storage.Finding{
				ID:             f.ID,
				Severity:       storage.Severity(f.Severity),
				Category:       f.Category,
				File:           f.File,
				Line:           f.Line,
				Title:          f.Title,
				Description:    f.Description,
				Recommendation: f.Recommendation,
				CWE:            f.CWE,
				OWASP:          f.OWASP,
				Evidence:       f.Evidence,
				Tags:           f.Tags,
			}
		}
		scanResults[i] = storage.ScanResult{
			ScanID:   sr.ScanID,
			ScanName: sr.ScanName,
			Category: sr.Category,
			Duration: sr.Duration.String(),
			Findings: findings,
			Summary: storage.ScanSummary{
				CriticalCount: sr.Summary.CriticalCount,
				HighCount:     sr.Summary.HighCount,
				MediumCount:   sr.Summary.MediumCount,
				LowCount:      sr.Summary.LowCount,
				TotalCount:    sr.Summary.TotalCount,
			},
			Error:   sr.Error,
			Skipped: sr.Skipped,
		}
	}

	return &storage.AggregatedReport{
		Version:      report.Version,
		Workspace:    report.Workspace,
		AnalyzedAt:   report.AnalyzedAt,
		Duration:     report.Duration,
		Languages:    report.Languages,
		FilesCount:   report.FilesCount,
		ScansRun:     report.ScansRun,
		ScansSkipped: report.ScansSkipped,
		ScansFailed:  report.ScansFailed,
		Summary: storage.AggregatedSummary{
			Security: storage.SecuritySummary{
				CriticalCount: report.Summary.Security.CriticalCount,
				HighCount:     report.Summary.Security.HighCount,
				MediumCount:   report.Summary.Security.MediumCount,
				LowCount:      report.Summary.Security.LowCount,
				TotalCount:    report.Summary.Security.TotalCount,
			},
			Quality: storage.QualitySummary{
				Score:       report.Summary.Quality.Score,
				HighCount:   report.Summary.Quality.HighCount,
				MediumCount: report.Summary.Quality.MediumCount,
				LowCount:    report.Summary.Quality.LowCount,
				TotalCount:  report.Summary.Quality.TotalCount,
			},
		},
		ScanResults: scanResults,
		Error:       report.Error,
	}
}

// generateLegacyReports generates separate security and quality reports for backward compatibility
func (a *Analyzer) generateLegacyReports(workspaceName string, report *AggregatedReport) {
	// Collect all findings from scan results
	var allFindings []Finding
	for _, sr := range report.ScanResults {
		if sr.Error == "" {
			allFindings = append(allFindings, sr.Findings...)
		}
	}

	// Generate security report
	var securityFindings []storage.Finding
	for _, f := range allFindings {
		if f.Category == "security" || a.hasSecurityTag(f) {
			securityFindings = append(securityFindings, storage.Finding{
				ID:             f.ID,
				Severity:       storage.Severity(f.Severity),
				Category:       f.Category,
				File:           f.File,
				Line:           f.Line,
				Title:          f.Title,
				Description:    f.Description,
				Recommendation: f.Recommendation,
				CWE:            f.CWE,
				OWASP:          f.OWASP,
				Evidence:       f.Evidence,
				Tags:           f.Tags,
			})
		}
	}

	securityReport := &storage.Report{
		Version:    ReportVersion,
		Type:       storage.ReportTypeSecurity,
		Workspace:  workspaceName,
		AnalyzedAt: report.AnalyzedAt,
		Duration:   report.Duration,
		FilesCount: report.FilesCount,
		Summary: storage.Summary{
			CriticalCount: report.Summary.Security.CriticalCount,
			HighCount:     report.Summary.Security.HighCount,
			MediumCount:   report.Summary.Security.MediumCount,
			LowCount:      report.Summary.Security.LowCount,
		},
		Findings: securityFindings,
		Metadata: storage.Metadata{Model: "semgrep"},
	}
	a.storage.SaveReport(workspaceName, securityReport)

	// Generate quality report
	var qualityFindings []storage.Finding
	for _, f := range allFindings {
		if f.Category != "security" && !a.hasSecurityTag(f) {
			qualityFindings = append(qualityFindings, storage.Finding{
				ID:             f.ID,
				Severity:       storage.Severity(f.Severity),
				Category:       f.Category,
				File:           f.File,
				Line:           f.Line,
				Title:          f.Title,
				Description:    f.Description,
				Recommendation: f.Recommendation,
				CWE:            f.CWE,
				OWASP:          f.OWASP,
				Evidence:       f.Evidence,
				Tags:           f.Tags,
			})
		}
	}

	qualityReport := &storage.Report{
		Version:    ReportVersion,
		Type:       storage.ReportTypeQuality,
		Workspace:  workspaceName,
		AnalyzedAt: report.AnalyzedAt,
		Duration:   report.Duration,
		FilesCount: report.FilesCount,
		Summary: storage.Summary{
			Score:       report.Summary.Quality.Score,
			HighCount:   report.Summary.Quality.HighCount,
			MediumCount: report.Summary.Quality.MediumCount,
			LowCount:    report.Summary.Quality.LowCount,
		},
		Findings: qualityFindings,
		Metadata: storage.Metadata{Model: "semgrep"},
	}
	a.storage.SaveReport(workspaceName, qualityReport)
}

// hasSecurityTag checks if a finding has a security-related tag
func (a *Analyzer) hasSecurityTag(f Finding) bool {
	for _, tag := range f.Tags {
		if len(tag) > 4 && tag[:4] == "SEC_" {
			return true
		}
	}
	return false
}

// AnalyzeSecurity runs only security analysis (for backward compatibility)
func (a *Analyzer) AnalyzeSecurity(ctx context.Context, workspaceName string) error {
	return a.AnalyzeWorkspace(ctx, workspaceName)
}

// AnalyzeQuality runs only quality analysis (for backward compatibility)
func (a *Analyzer) AnalyzeQuality(ctx context.Context, workspaceName string) error {
	return a.AnalyzeWorkspace(ctx, workspaceName)
}

// CountFiles counts the number of files in a directory (excluding large files)
func CountFiles(dir string, maxSize int64) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		count++
		return nil
	})
	return count, err
}
