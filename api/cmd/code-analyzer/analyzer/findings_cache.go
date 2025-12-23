package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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

// DeduplicateFindings removes duplicate findings based on file, line, and normalized title
// This helps stabilize results across non-deterministic Claude outputs
// Also filters out:
// - Informational findings that indicate "no issues found"
// - Speculative/suggestive findings without concrete evidence
// - Best practice suggestions rather than actual issues
func DeduplicateFindings(findings []Finding) []Finding {
	if len(findings) == 0 {
		return findings
	}

	// Map to track unique findings by their signature
	seen := make(map[string]Finding)

	for _, f := range findings {
		// Skip informational findings - these are "no issues found" reports
		if strings.ToLower(f.Severity) == "informational" || strings.ToLower(f.Severity) == "info" {
			continue
		}

		// Skip speculative/suggestive findings
		if isSpeculativeFinding(f) {
			continue
		}

		sig := generateFindingSignature(f)

		// If we haven't seen this finding, or if this one has higher severity, keep it
		if existing, exists := seen[sig]; !exists {
			seen[sig] = f
		} else if severityRank(f.Severity) > severityRank(existing.Severity) {
			seen[sig] = f
		}
	}

	// Convert map to slice
	result := make([]Finding, 0, len(seen))
	for _, f := range seen {
		result = append(result, f)
	}

	// Sort by file, then line for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		if result[i].File != result[j].File {
			return result[i].File < result[j].File
		}
		if result[i].Line != result[j].Line {
			return result[i].Line < result[j].Line
		}
		return result[i].Title < result[j].Title
	})

	return result
}

// isSpeculativeFinding returns true if the finding is speculative/suggestive
// rather than a confirmed issue with evidence
func isSpeculativeFinding(f Finding) bool {
	// Combine all text for checking
	allText := strings.ToLower(f.Title + " " + f.Description + " " + f.Recommendation)

	// Phrases that indicate speculative/suggestive findings
	speculativePhrases := []string{
		"no action required",
		"no issues found",
		"no vulnerabilities detected",
		"no security issues",
		"appears to be secure",
		"looks secure",
		"properly implemented",
		"correctly implemented",
		"best practice suggests",
		"consider adding",
		"consider using",
		"consider implementing",
		"you may want to",
		"you might want to",
		"it would be better",
		"it might be better",
		"could be improved",
		"should consider",
		"recommend adding",
		"recommend implementing",
		"for better security",
		"for improved security",
		"as a best practice",
		"following best practices",
		"enhancement suggestion",
		"improvement suggestion",
		"no critical issues",
		"no major issues",
	}

	for _, phrase := range speculativePhrases {
		if strings.Contains(allText, phrase) {
			return true
		}
	}

	// Check for patterns that suggest speculation rather than confirmation
	// These are weaker signals, only use when combined with lack of evidence
	weakSpeculativePatterns := []string{
		"may be vulnerable",
		"might be vulnerable",
		"could be vulnerable",
		"potentially vulnerable",
		"possibly vulnerable",
		"may lead to",
		"might lead to",
		"could lead to",
		"may allow",
		"might allow",
		"could allow",
		"may expose",
		"might expose",
		"could expose",
	}

	for _, pattern := range weakSpeculativePatterns {
		if strings.Contains(allText, pattern) {
			// Check if there's concrete evidence in the description
			// If the description is vague (short or lacks file/line specifics), skip it
			if len(f.Description) < 100 && f.Line == 0 {
				return true
			}
		}
	}

	return false
}

// generateFindingSignature creates a unique signature for a finding
// based on file, line, and normalized title
func generateFindingSignature(f Finding) string {
	// Normalize the title by:
	// 1. Converting to lowercase
	// 2. Removing special characters
	// 3. Extracting key terms
	normalizedTitle := normalizeTitle(f.Title)

	// Create a composite key
	key := fmt.Sprintf("%s:%d:%s", f.File, f.Line, normalizedTitle)

	// Hash it to get a consistent signature
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter signature
}

// normalizeTitle normalizes a finding title for comparison
func normalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)

	// Remove common variations
	// E.g., "SQL Injection" vs "SQL Injection Vulnerability" vs "Potential SQL Injection"
	prefixesToRemove := []string{"potential ", "possible ", "suspected ", "likely "}
	for _, prefix := range prefixesToRemove {
		normalized = strings.TrimPrefix(normalized, prefix)
	}

	suffixesToRemove := []string{" vulnerability", " issue", " problem", " risk", " detected", " found"}
	for _, suffix := range suffixesToRemove {
		normalized = strings.TrimSuffix(normalized, suffix)
	}

	// Remove special characters and extra whitespace
	re := regexp.MustCompile(`[^a-z0-9\s]`)
	normalized = re.ReplaceAllString(normalized, " ")

	// Collapse multiple spaces
	spaceRe := regexp.MustCompile(`\s+`)
	normalized = spaceRe.ReplaceAllString(normalized, " ")

	// Trim and return
	return strings.TrimSpace(normalized)
}

// severityRank returns a numeric rank for severity (higher = more severe)
func severityRank(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
