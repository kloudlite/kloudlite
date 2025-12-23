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
	ReportVersion = "1.0.0"
	MaxFileSize   = 100 * 1024 // 100KB
)

// Analyzer orchestrates code analysis
type Analyzer struct {
	claudeCode     *ClaudeCode
	storage        *storage.Storage
	workspacesPath string
	logger         *zap.Logger
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(
	claudeCode *ClaudeCode,
	storage *storage.Storage,
	workspacesPath string,
	logger *zap.Logger,
) *Analyzer {
	return &Analyzer{
		claudeCode:     claudeCode,
		storage:        storage,
		workspacesPath: workspacesPath,
		logger:         logger,
	}
}

// AnalyzeWorkspace runs both security and quality analysis on a workspace in parallel
func (a *Analyzer) AnalyzeWorkspace(ctx context.Context, workspaceName string) error {
	workspaceDir := filepath.Join(a.workspacesPath, workspaceName)

	a.logger.Info("Starting workspace analysis (parallel)", zap.String("workspace", workspaceName))

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

	a.logger.Info("Completed workspace analysis", zap.String("workspace", workspaceName))
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
