package analyzer

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/storage"
	"go.uber.org/zap"
)

const (
	ReportVersion        = "1.0.0"
	AggregatedVersion    = "2.0.0"
	MaxFileSize          = 100 * 1024 // 100KB
	DefaultMaxConcurrent = 10         // Run up to 10 scans concurrently
)

// Analyzer orchestrates code analysis
type Analyzer struct {
	claudeCode     *ClaudeCode
	claudeAPI      *ClaudeAPI
	executor       *Executor
	storage        *storage.Storage
	workspacesPath string
	logger         *zap.Logger
	maxConcurrent  int
	useDirectAPI   bool
}

// NewAnalyzer creates a new analyzer using Claude CLI
func NewAnalyzer(
	claudeCode *ClaudeCode,
	storage *storage.Storage,
	workspacesPath string,
	logger *zap.Logger,
) *Analyzer {
	return &Analyzer{
		claudeCode:     claudeCode,
		executor:       NewExecutor(claudeCode, logger, DefaultMaxConcurrent),
		storage:        storage,
		workspacesPath: workspacesPath,
		logger:         logger,
		maxConcurrent:  DefaultMaxConcurrent,
		useDirectAPI:   false,
	}
}

// NewAnalyzerWithAPI creates a new analyzer using direct Claude API with prompt caching
func NewAnalyzerWithAPI(
	claudeAPI *ClaudeAPI,
	storage *storage.Storage,
	workspacesPath string,
	reportsPath string,
	logger *zap.Logger,
) *Analyzer {
	return &Analyzer{
		claudeAPI:      claudeAPI,
		executor:       NewExecutorWithAPI(claudeAPI, reportsPath, logger, DefaultMaxConcurrent),
		storage:        storage,
		workspacesPath: workspacesPath,
		logger:         logger,
		maxConcurrent:  DefaultMaxConcurrent,
		useDirectAPI:   true,
	}
}

// SetMaxConcurrent sets the maximum concurrent scans
func (a *Analyzer) SetMaxConcurrent(max int) {
	a.maxConcurrent = max
	if a.useDirectAPI {
		a.executor = NewExecutorWithAPI(a.claudeAPI, a.storage.GetBasePath(), a.logger, max)
	} else {
		a.executor = NewExecutor(a.claudeCode, a.logger, max)
	}
}

// AnalyzeWorkspace runs all applicable scans on a workspace in parallel
func (a *Analyzer) AnalyzeWorkspace(ctx context.Context, workspaceName string) error {
	workspaceDir := filepath.Join(a.workspacesPath, workspaceName)

	a.logger.Info("Starting multi-scan workspace analysis", zap.String("workspace", workspaceName))

	// Count files for metadata
	fileCount, err := CountFiles(workspaceDir, MaxFileSize)
	if err != nil {
		a.logger.Warn("Failed to count files", zap.Error(err))
		fileCount = 0
	}

	// Run all scans in parallel using the executor
	report := a.executor.RunAllScans(ctx, workspaceDir, workspaceName, fileCount)

	// Convert executor report to storage format and save
	aggregatedReport := a.convertToStorageReport(report)
	if err := a.storage.SaveAggregatedReport(workspaceName, aggregatedReport); err != nil {
		a.logger.Error("Failed to save aggregated report", zap.Error(err))
		return err
	}

	// Also generate legacy security and quality reports for backward compatibility
	a.generateLegacyReports(workspaceName, report)

	a.logger.Info("Completed multi-scan workspace analysis",
		zap.String("workspace", workspaceName),
		zap.String("duration", report.Duration),
		zap.Int("scans_run", report.ScansRun),
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
			}
		}
		scanResults[i] = storage.ScanResult{
			ScanID:   sr.ScanID,
			ScanName: sr.ScanName,
			Category: string(sr.Category),
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
	// Generate security report
	securityFindings := make([]storage.Finding, 0)
	for _, sr := range report.ScanResults {
		if sr.Category == CategorySecurity && sr.Error == "" {
			for _, f := range sr.Findings {
				securityFindings = append(securityFindings, storage.Finding{
					ID:             f.ID,
					Severity:       storage.Severity(f.Severity),
					Category:       f.Category,
					File:           f.File,
					Line:           f.Line,
					Title:          f.Title,
					Description:    f.Description,
					Recommendation: f.Recommendation,
				})
			}
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
		Metadata: storage.Metadata{Model: "claude-code-multi-scan"},
	}
	a.storage.SaveReport(workspaceName, securityReport)

	// Generate quality report
	qualityFindings := make([]storage.Finding, 0)
	for _, sr := range report.ScanResults {
		if (sr.Category == CategoryQuality || sr.Category == CategoryLanguage) && sr.Error == "" {
			for _, f := range sr.Findings {
				qualityFindings = append(qualityFindings, storage.Finding{
					ID:             f.ID,
					Severity:       storage.Severity(f.Severity),
					Category:       f.Category,
					File:           f.File,
					Line:           f.Line,
					Title:          f.Title,
					Description:    f.Description,
					Recommendation: f.Recommendation,
				})
			}
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
		Metadata: storage.Metadata{Model: "claude-code-multi-scan"},
	}
	a.storage.SaveReport(workspaceName, qualityReport)
}

// AnalyzeWorkspaceLegacy runs the old-style security and quality analysis in parallel
func (a *Analyzer) AnalyzeWorkspaceLegacy(ctx context.Context, workspaceName string) error {
	workspaceDir := filepath.Join(a.workspacesPath, workspaceName)

	a.logger.Info("Starting legacy workspace analysis (parallel)", zap.String("workspace", workspaceName))

	// Count files for metadata
	fileCount, err := CountFiles(workspaceDir, MaxFileSize)
	if err != nil {
		a.logger.Warn("Failed to count files", zap.Error(err))
		fileCount = 0
	}

	// Run security and quality analysis in parallel
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := a.runSecurityAnalysis(ctx, workspaceName, workspaceDir, fileCount); err != nil {
			a.logger.Error("Security analysis failed", zap.String("workspace", workspaceName), zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.runQualityAnalysis(ctx, workspaceName, workspaceDir, fileCount); err != nil {
			a.logger.Error("Quality analysis failed", zap.String("workspace", workspaceName), zap.Error(err))
		}
	}()

	wg.Wait()

	a.logger.Info("Completed legacy workspace analysis", zap.String("workspace", workspaceName))
	return nil
}

// AnalyzeSecurity runs only security analysis
func (a *Analyzer) AnalyzeSecurity(ctx context.Context, workspaceName string) error {
	workspaceDir := filepath.Join(a.workspacesPath, workspaceName)
	fileCount, _ := CountFiles(workspaceDir, MaxFileSize)
	return a.runSecurityAnalysis(ctx, workspaceName, workspaceDir, fileCount)
}

// AnalyzeQuality runs only quality analysis
func (a *Analyzer) AnalyzeQuality(ctx context.Context, workspaceName string) error {
	workspaceDir := filepath.Join(a.workspacesPath, workspaceName)
	fileCount, _ := CountFiles(workspaceDir, MaxFileSize)
	return a.runQualityAnalysis(ctx, workspaceName, workspaceDir, fileCount)
}

func (a *Analyzer) runSecurityAnalysis(ctx context.Context, workspaceName, workspaceDir string, fileCount int) error {
	a.logger.Info("Running security analysis", zap.String("workspace", workspaceName))

	result, duration, err := a.claudeCode.AnalyzeSecurity(ctx, workspaceDir)
	if err != nil {
		// Save error report
		errorReport := &storage.Report{
			Version:    ReportVersion,
			Type:       storage.ReportTypeSecurity,
			Workspace:  workspaceName,
			AnalyzedAt: time.Now(),
			Duration:   duration.String(),
			FilesCount: fileCount,
			Summary:    storage.Summary{},
			Findings:   []storage.Finding{},
			Error:      err.Error(),
		}
		a.storage.SaveReport(workspaceName, errorReport)
		return err
	}

	// Convert result to storage format
	report := &storage.Report{
		Version:    ReportVersion,
		Type:       storage.ReportTypeSecurity,
		Workspace:  workspaceName,
		AnalyzedAt: time.Now(),
		Duration:   duration.String(),
		FilesCount: fileCount,
		Summary: storage.Summary{
			CriticalCount: result.Summary.CriticalCount,
			HighCount:     result.Summary.HighCount,
			MediumCount:   result.Summary.MediumCount,
			LowCount:      result.Summary.LowCount,
		},
		Findings: convertSecurityFindings(result.Findings),
		Metadata: storage.Metadata{
			Model: "claude-code",
		},
	}

	if err := a.storage.SaveReport(workspaceName, report); err != nil {
		a.logger.Error("Failed to save security report", zap.Error(err))
		return err
	}

	a.logger.Info("Security analysis complete",
		zap.String("workspace", workspaceName),
		zap.Duration("duration", duration),
		zap.Int("findings", len(result.Findings)),
	)

	return nil
}

func (a *Analyzer) runQualityAnalysis(ctx context.Context, workspaceName, workspaceDir string, fileCount int) error {
	a.logger.Info("Running quality analysis", zap.String("workspace", workspaceName))

	result, duration, err := a.claudeCode.AnalyzeQuality(ctx, workspaceDir)
	if err != nil {
		// Save error report
		errorReport := &storage.Report{
			Version:    ReportVersion,
			Type:       storage.ReportTypeQuality,
			Workspace:  workspaceName,
			AnalyzedAt: time.Now(),
			Duration:   duration.String(),
			FilesCount: fileCount,
			Summary:    storage.Summary{Score: 0},
			Findings:   []storage.Finding{},
			Error:      err.Error(),
		}
		a.storage.SaveReport(workspaceName, errorReport)
		return err
	}

	// Convert result to storage format
	report := &storage.Report{
		Version:    ReportVersion,
		Type:       storage.ReportTypeQuality,
		Workspace:  workspaceName,
		AnalyzedAt: time.Now(),
		Duration:   duration.String(),
		FilesCount: fileCount,
		Summary: storage.Summary{
			Score:       result.Summary.Score,
			HighCount:   result.Summary.HighCount,
			MediumCount: result.Summary.MediumCount,
			LowCount:    result.Summary.LowCount,
		},
		Findings: convertQualityFindings(result.Findings),
		Metadata: storage.Metadata{
			Model: "claude-code",
		},
	}

	if err := a.storage.SaveReport(workspaceName, report); err != nil {
		a.logger.Error("Failed to save quality report", zap.Error(err))
		return err
	}

	a.logger.Info("Quality analysis complete",
		zap.String("workspace", workspaceName),
		zap.Duration("duration", duration),
		zap.Int("findings", len(result.Findings)),
		zap.Int("score", result.Summary.Score),
	)

	return nil
}

func convertSecurityFindings(findings []SecurityFinding) []storage.Finding {
	result := make([]storage.Finding, len(findings))
	for i, f := range findings {
		result[i] = storage.Finding{
			ID:             f.ID,
			Severity:       storage.Severity(f.Severity),
			Category:       f.Category,
			File:           f.File,
			Line:           f.Line,
			Title:          f.Title,
			Description:    f.Description,
			Recommendation: f.Recommendation,
		}
	}
	return result
}

func convertQualityFindings(findings []QualityFinding) []storage.Finding {
	result := make([]storage.Finding, len(findings))
	for i, f := range findings {
		result[i] = storage.Finding{
			ID:             f.ID,
			Severity:       storage.Severity(f.Severity),
			Category:       f.Category,
			File:           f.File,
			Line:           f.Line,
			Title:          f.Title,
			Description:    f.Description,
			Recommendation: f.Recommendation,
		}
	}
	return result
}
