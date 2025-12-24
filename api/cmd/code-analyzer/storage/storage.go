package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ReportType represents the type of analysis report
type ReportType string

const (
	ReportTypeSecurity   ReportType = "security"
	ReportTypeQuality    ReportType = "quality"
	ReportTypeAggregated ReportType = "aggregated"
)

// Severity represents finding severity level
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Finding represents a single analysis finding
type Finding struct {
	ID             string   `json:"id"`
	Severity       Severity `json:"severity"`
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

// Summary represents the report summary
type Summary struct {
	Score         int `json:"score,omitempty"`
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
}

// Metadata represents analysis metadata
type Metadata struct {
	Model            string `json:"model"`
	PromptTokens     int    `json:"promptTokens,omitempty"`
	CompletionTokens int    `json:"completionTokens,omitempty"`
}

// Report represents a complete analysis report
type Report struct {
	Version    string     `json:"version"`
	Type       ReportType `json:"type"`
	Workspace  string     `json:"workspace"`
	AnalyzedAt time.Time  `json:"analyzedAt"`
	Duration   string     `json:"duration"`
	FilesCount int        `json:"filesAnalyzed"`
	Summary    Summary    `json:"summary"`
	Findings   []Finding  `json:"findings"`
	Metadata   Metadata   `json:"metadata,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// ScanResult represents the result of a single scan
type ScanResult struct {
	ScanID   string      `json:"scanId"`
	ScanName string      `json:"scanName"`
	Category string      `json:"category"`
	Duration string      `json:"duration"`
	Findings []Finding   `json:"findings"`
	Summary  ScanSummary `json:"summary"`
	Error    string      `json:"error,omitempty"`
	Skipped  bool        `json:"skipped,omitempty"`
}

// ScanSummary represents summary for a single scan
type ScanSummary struct {
	CriticalCount int `json:"criticalCount,omitempty"`
	HighCount     int `json:"highCount,omitempty"`
	MediumCount   int `json:"mediumCount,omitempty"`
	LowCount      int `json:"lowCount,omitempty"`
	TotalCount    int `json:"totalCount"`
}

// AggregatedSummary represents combined summary across all scans
type AggregatedSummary struct {
	Security SecuritySummary `json:"security"`
	Quality  QualitySummary  `json:"quality"`
}

// SecuritySummary for aggregated report
type SecuritySummary struct {
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
	TotalCount    int `json:"totalCount"`
}

// QualitySummary for aggregated report
type QualitySummary struct {
	Score       int `json:"score"`
	HighCount   int `json:"highCount"`
	MediumCount int `json:"mediumCount"`
	LowCount    int `json:"lowCount"`
	TotalCount  int `json:"totalCount"`
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

// WorkspaceMetadata contains metadata about workspace analysis
type WorkspaceMetadata struct {
	LastSecurityAnalysis   time.Time `json:"lastSecurityAnalysis,omitempty"`
	LastQualityAnalysis    time.Time `json:"lastQualityAnalysis,omitempty"`
	LastAggregatedAnalysis time.Time `json:"lastAggregatedAnalysis,omitempty"`
	TotalFilesAnalyzed     int       `json:"totalFilesAnalyzed"`
	Languages              []string  `json:"languages,omitempty"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

// ReportInfo contains basic info about a stored report
type ReportInfo struct {
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
}

// Storage handles report storage operations
type Storage struct {
	basePath string
	logger   *zap.Logger
}

// NewStorage creates a new storage instance
func NewStorage(basePath string, logger *zap.Logger) *Storage {
	return &Storage{
		basePath: basePath,
		logger:   logger,
	}
}

// GetBasePath returns the base path for reports storage
func (s *Storage) GetBasePath() string {
	return s.basePath
}

// SaveReport saves a report to storage
func (s *Storage) SaveReport(workspace string, report *Report) error {
	// Create workspace directory if needed
	workspaceDir := filepath.Join(s.basePath, workspace)
	reportDir := filepath.Join(workspaceDir, string(report.Type))

	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Marshal report to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Save with timestamp filename
	timestamp := report.AnalyzedAt.Format("2006-01-02T15-04-05")
	timestampFile := filepath.Join(reportDir, fmt.Sprintf("%s.json", timestamp))
	if err := os.WriteFile(timestampFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write timestamp report: %w", err)
	}

	// Also save as latest.json
	latestFile := filepath.Join(reportDir, "latest.json")
	if err := os.WriteFile(latestFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write latest report: %w", err)
	}

	// Update workspace metadata
	if err := s.updateMetadata(workspace, report); err != nil {
		s.logger.Warn("Failed to update metadata", zap.Error(err))
	}

	// Cleanup old reports (keep last 10)
	s.cleanupOldReports(reportDir, 10)

	s.logger.Info("Saved report",
		zap.String("workspace", workspace),
		zap.String("type", string(report.Type)),
		zap.String("file", timestampFile),
	)

	return nil
}

// GetLatestReport retrieves the latest report for a workspace
func (s *Storage) GetLatestReport(workspace string, reportType ReportType) (*Report, error) {
	latestFile := filepath.Join(s.basePath, workspace, string(reportType), "latest.json")

	data, err := os.ReadFile(latestFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No report found
		}
		return nil, fmt.Errorf("failed to read report: %w", err)
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// GetReportHistory returns list of historical reports
func (s *Storage) GetReportHistory(workspace string, reportType ReportType) ([]ReportInfo, error) {
	reportDir := filepath.Join(s.basePath, workspace, string(reportType))

	entries, err := os.ReadDir(reportDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ReportInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read report directory: %w", err)
	}

	var reports []ReportInfo
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "latest.json" {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Parse timestamp from filename
		name := strings.TrimSuffix(entry.Name(), ".json")
		timestamp, err := time.Parse("2006-01-02T15-04-05", name)
		if err != nil {
			continue
		}

		reports = append(reports, ReportInfo{
			Timestamp: timestamp,
			Filename:  entry.Name(),
			Size:      info.Size(),
		})
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp.After(reports[j].Timestamp)
	})

	return reports, nil
}

// GetReport retrieves a specific report by filename
func (s *Storage) GetReport(workspace string, reportType ReportType, filename string) (*Report, error) {
	reportFile := filepath.Join(s.basePath, workspace, string(reportType), filename)

	data, err := os.ReadFile(reportFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read report: %w", err)
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// GetMetadata retrieves workspace metadata
func (s *Storage) GetMetadata(workspace string) (*WorkspaceMetadata, error) {
	metadataFile := filepath.Join(s.basePath, workspace, "metadata.json")

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &WorkspaceMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata WorkspaceMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// ListWorkspaces returns list of workspaces with reports
func (s *Storage) ListWorkspaces() ([]string, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var workspaces []string
	for _, entry := range entries {
		if entry.IsDir() {
			workspaces = append(workspaces, entry.Name())
		}
	}

	return workspaces, nil
}

// DeleteWorkspaceReports deletes all reports for a workspace
func (s *Storage) DeleteWorkspaceReports(workspace string) error {
	workspaceDir := filepath.Join(s.basePath, workspace)
	return os.RemoveAll(workspaceDir)
}

// SaveAggregatedReport saves an aggregated report to storage
func (s *Storage) SaveAggregatedReport(workspace string, report *AggregatedReport) error {
	// Create workspace directory if needed
	workspaceDir := filepath.Join(s.basePath, workspace)
	reportDir := filepath.Join(workspaceDir, string(ReportTypeAggregated))

	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Marshal report to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Save with timestamp filename
	timestamp := report.AnalyzedAt.Format("2006-01-02T15-04-05")
	timestampFile := filepath.Join(reportDir, fmt.Sprintf("%s.json", timestamp))
	if err := os.WriteFile(timestampFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write timestamp report: %w", err)
	}

	// Also save as latest.json
	latestFile := filepath.Join(reportDir, "latest.json")
	if err := os.WriteFile(latestFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write latest report: %w", err)
	}

	// Update workspace metadata
	if err := s.updateAggregatedMetadata(workspace, report); err != nil {
		s.logger.Warn("Failed to update metadata", zap.Error(err))
	}

	// Cleanup old reports (keep last 10)
	s.cleanupOldReports(reportDir, 10)

	s.logger.Info("Saved aggregated report",
		zap.String("workspace", workspace),
		zap.Int("scans_run", report.ScansRun),
		zap.String("file", timestampFile),
	)

	return nil
}

// GetLatestAggregatedReport retrieves the latest aggregated report for a workspace
func (s *Storage) GetLatestAggregatedReport(workspace string) (*AggregatedReport, error) {
	latestFile := filepath.Join(s.basePath, workspace, string(ReportTypeAggregated), "latest.json")

	data, err := os.ReadFile(latestFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No report found
		}
		return nil, fmt.Errorf("failed to read report: %w", err)
	}

	var report AggregatedReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// GetAggregatedReportHistory returns list of historical aggregated reports
func (s *Storage) GetAggregatedReportHistory(workspace string) ([]ReportInfo, error) {
	return s.GetReportHistory(workspace, ReportTypeAggregated)
}

func (s *Storage) updateAggregatedMetadata(workspace string, report *AggregatedReport) error {
	metadata, err := s.GetMetadata(workspace)
	if err != nil {
		metadata = &WorkspaceMetadata{}
	}

	metadata.LastAggregatedAnalysis = report.AnalyzedAt
	metadata.TotalFilesAnalyzed = report.FilesCount
	metadata.Languages = report.Languages
	metadata.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataFile := filepath.Join(s.basePath, workspace, "metadata.json")
	return os.WriteFile(metadataFile, data, 0644)
}

func (s *Storage) updateMetadata(workspace string, report *Report) error {
	metadata, err := s.GetMetadata(workspace)
	if err != nil {
		metadata = &WorkspaceMetadata{}
	}

	switch report.Type {
	case ReportTypeSecurity:
		metadata.LastSecurityAnalysis = report.AnalyzedAt
	case ReportTypeQuality:
		metadata.LastQualityAnalysis = report.AnalyzedAt
	}

	metadata.TotalFilesAnalyzed = report.FilesCount
	metadata.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataFile := filepath.Join(s.basePath, workspace, "metadata.json")
	return os.WriteFile(metadataFile, data, 0644)
}

func (s *Storage) cleanupOldReports(reportDir string, keep int) {
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		return
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "latest.json" {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, entry.Name())
		}
	}

	// Sort by name (which is timestamp-based)
	sort.Strings(files)

	// Delete oldest files if we have more than 'keep'
	if len(files) > keep {
		for _, file := range files[:len(files)-keep] {
			filePath := filepath.Join(reportDir, file)
			if err := os.Remove(filePath); err != nil {
				s.logger.Warn("Failed to cleanup old report", zap.String("file", filePath), zap.Error(err))
			}
		}
	}
}
