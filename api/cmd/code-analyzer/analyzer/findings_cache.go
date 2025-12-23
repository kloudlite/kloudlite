package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// FindingsCache stores findings indexed by file path
// This allows incremental updates when only some files change
type FindingsCache struct {
	Version   string                        `json:"version"`
	Workspace string                        `json:"workspace"`
	UpdatedAt time.Time                     `json:"updatedAt"`
	Files     map[string]*FileFindingsEntry `json:"files"`
}

// FileFindings contains all findings for a specific file
type FileFindingsEntry struct {
	Findings   []Finding `json:"findings"`
	AnalyzedAt time.Time `json:"analyzedAt"`
	FileHash   string    `json:"fileHash"` // Hash of file when analyzed
}

// FindingsCacheManager handles findings cache operations
type FindingsCacheManager struct {
	basePath string
	logger   *zap.Logger
}

// NewFindingsCacheManager creates a new findings cache manager
func NewFindingsCacheManager(basePath string, logger *zap.Logger) *FindingsCacheManager {
	return &FindingsCacheManager{
		basePath: basePath,
		logger:   logger,
	}
}

// LoadCache loads the findings cache for a workspace
func (m *FindingsCacheManager) LoadCache(workspace string) (*FindingsCache, error) {
	cachePath := m.getCachePath(workspace)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No cache exists yet
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var cache FindingsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return &cache, nil
}

// SaveCache saves the findings cache for a workspace
func (m *FindingsCacheManager) SaveCache(workspace string, cache *FindingsCache) error {
	cachePath := m.getCachePath(workspace)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	m.logger.Debug("Saved findings cache",
		zap.String("workspace", workspace),
		zap.Int("files", len(cache.Files)),
	)

	return nil
}

// UpdateCache updates the cache with new findings for changed files
// It keeps findings for unchanged files and replaces findings for changed files
func (m *FindingsCacheManager) UpdateCache(
	workspace string,
	existingCache *FindingsCache,
	newFindings []Finding,
	changedFiles []string,
	manifest *FileManifest,
) *FindingsCache {
	now := time.Now()

	// Create new cache if none exists
	if existingCache == nil {
		existingCache = &FindingsCache{
			Version:   "1.0.0",
			Workspace: workspace,
			Files:     make(map[string]*FileFindingsEntry),
		}
	}

	cache := &FindingsCache{
		Version:   "1.0.0",
		Workspace: workspace,
		UpdatedAt: now,
		Files:     make(map[string]*FileFindingsEntry),
	}

	// Build a set of changed files for quick lookup
	changedSet := make(map[string]bool)
	for _, f := range changedFiles {
		changedSet[f] = true
	}

	// Copy unchanged file findings from existing cache
	for filePath, entry := range existingCache.Files {
		if !changedSet[filePath] {
			// File unchanged, keep existing findings
			cache.Files[filePath] = entry
		}
	}

	// Group new findings by file
	findingsByFile := make(map[string][]Finding)
	for _, f := range newFindings {
		findingsByFile[f.File] = append(findingsByFile[f.File], f)
	}

	// Add/update findings for changed files
	for _, filePath := range changedFiles {
		findings := findingsByFile[filePath]
		if findings == nil {
			findings = []Finding{} // No findings for this file
		}

		fileHash := ""
		if manifest != nil && manifest.Files[filePath] != nil {
			fileHash = manifest.Files[filePath].Hash
		}

		cache.Files[filePath] = &FileFindingsEntry{
			Findings:   findings,
			AnalyzedAt: now,
			FileHash:   fileHash,
		}
	}

	return cache
}

// GetAllFindings returns all findings from the cache
func (m *FindingsCacheManager) GetAllFindings(cache *FindingsCache) []Finding {
	if cache == nil {
		return []Finding{}
	}

	var findings []Finding
	for _, entry := range cache.Files {
		findings = append(findings, entry.Findings...)
	}

	return findings
}

// MergeFindings combines cached findings (for unchanged files) with new findings (for changed files)
func MergeFindings(
	existingCache *FindingsCache,
	newFindings []Finding,
	changedFiles []string,
	deletedFiles []string,
) []Finding {
	var result []Finding

	// Build sets for quick lookup
	changedSet := make(map[string]bool)
	for _, f := range changedFiles {
		changedSet[f] = true
	}

	deletedSet := make(map[string]bool)
	for _, f := range deletedFiles {
		deletedSet[f] = true
	}

	// Keep findings for unchanged files (not in changed or deleted)
	if existingCache != nil {
		for filePath, entry := range existingCache.Files {
			if !changedSet[filePath] && !deletedSet[filePath] {
				result = append(result, entry.Findings...)
			}
		}
	}

	// Add new findings for changed files
	result = append(result, newFindings...)

	return result
}

// getCachePath returns the path to the findings cache file for a workspace
func (m *FindingsCacheManager) getCachePath(workspace string) string {
	return filepath.Join(m.basePath, workspace, "findings-cache.json")
}

// InvalidateFiles removes findings for specific files from the cache
func (m *FindingsCacheManager) InvalidateFiles(cache *FindingsCache, files []string) {
	if cache == nil {
		return
	}

	for _, f := range files {
		delete(cache.Files, f)
	}
}

// GetFindingsForFiles returns findings only for specific files
func (m *FindingsCacheManager) GetFindingsForFiles(cache *FindingsCache, files []string) []Finding {
	if cache == nil {
		return []Finding{}
	}

	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}

	var findings []Finding
	for filePath, entry := range cache.Files {
		if fileSet[filePath] {
			findings = append(findings, entry.Findings...)
		}
	}

	return findings
}
