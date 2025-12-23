package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FileManifest tracks the state of files in a workspace
type FileManifest struct {
	Version   string               `json:"version"`
	Workspace string               `json:"workspace"`
	Files     map[string]*FileInfo `json:"files"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

// FileInfo contains metadata about a single file
type FileInfo struct {
	Hash    string    `json:"hash"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

// DiffResult contains the results of comparing current state with manifest
type DiffResult struct {
	Added      []string // New files
	Modified   []string // Changed files (different hash)
	Deleted    []string // Files that no longer exist
	Unchanged  []string // Files that haven't changed
	AllFiles   []string // All current files
	HasChanges bool     // True if any changes detected
}

// ManifestManager handles manifest operations
type ManifestManager struct {
	basePath    string // Base path for storing manifests
	maxFileSize int64
	ignoredDirs map[string]bool
	ignoredExts map[string]bool
	logger      *zap.Logger
}

// NewManifestManager creates a new manifest manager
func NewManifestManager(basePath string, maxFileSize int64, logger *zap.Logger) *ManifestManager {
	return &ManifestManager{
		basePath:    basePath,
		maxFileSize: maxFileSize,
		ignoredDirs: map[string]bool{
			".git":         true,
			"node_modules": true,
			"vendor":       true,
			"__pycache__":  true,
			".nix-profile": true,
			".cache":       true,
			"dist":         true,
			"build":        true,
			".next":        true,
			"target":       true,
		},
		ignoredExts: map[string]bool{
			".exe":   true,
			".dll":   true,
			".so":    true,
			".dylib": true,
			".o":     true,
			".a":     true,
			".pyc":   true,
			".class": true,
			".jar":   true,
			".war":   true,
			".zip":   true,
			".tar":   true,
			".gz":    true,
			".png":   true,
			".jpg":   true,
			".jpeg":  true,
			".gif":   true,
			".ico":   true,
			".svg":   true,
			".woff":  true,
			".woff2": true,
			".ttf":   true,
			".eot":   true,
			".mp3":   true,
			".mp4":   true,
			".wav":   true,
			".pdf":   true,
		},
		logger: logger,
	}
}

// LoadManifest loads the manifest for a workspace
func (m *ManifestManager) LoadManifest(workspace string) (*FileManifest, error) {
	manifestPath := m.getManifestPath(workspace)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No manifest exists yet
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest FileManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves the manifest for a workspace
func (m *ManifestManager) SaveManifest(workspace string, manifest *FileManifest) error {
	manifestPath := m.getManifestPath(workspace)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	m.logger.Debug("Saved manifest",
		zap.String("workspace", workspace),
		zap.Int("files", len(manifest.Files)),
	)

	return nil
}

// ComputeCurrentManifest scans a workspace and builds a manifest of all files
func (m *ManifestManager) ComputeCurrentManifest(workspaceDir, workspaceName string) (*FileManifest, error) {
	manifest := &FileManifest{
		Version:   "1.0.0",
		Workspace: workspaceName,
		Files:     make(map[string]*FileInfo),
		UpdatedAt: time.Now(),
	}

	err := filepath.Walk(workspaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Get relative path
		relPath, err := filepath.Rel(workspaceDir, path)
		if err != nil {
			return nil
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if directory should be ignored
		if info.IsDir() {
			if m.shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip ignored file extensions
		if m.shouldIgnoreFile(info.Name()) {
			return nil
		}

		// Skip files that are too large
		if info.Size() > m.maxFileSize {
			return nil
		}

		// Compute file hash
		hash, err := m.computeFileHash(path)
		if err != nil {
			m.logger.Debug("Failed to hash file",
				zap.String("path", relPath),
				zap.Error(err),
			)
			return nil
		}

		manifest.Files[relPath] = &FileInfo{
			Hash:    hash,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk workspace: %w", err)
	}

	return manifest, nil
}

// ComputeDiff compares current workspace state with a previous manifest
func (m *ManifestManager) ComputeDiff(workspaceDir string, lastManifest *FileManifest) (*DiffResult, error) {
	// Get current state
	currentManifest, err := m.ComputeCurrentManifest(workspaceDir, "")
	if err != nil {
		return nil, err
	}

	result := &DiffResult{
		Added:     []string{},
		Modified:  []string{},
		Deleted:   []string{},
		Unchanged: []string{},
		AllFiles:  []string{},
	}

	// Collect all current files
	for path := range currentManifest.Files {
		result.AllFiles = append(result.AllFiles, path)
	}

	// If no previous manifest, everything is new
	if lastManifest == nil {
		result.Added = result.AllFiles
		result.HasChanges = len(result.Added) > 0
		return result, nil
	}

	// Check for added and modified files
	for path, currentInfo := range currentManifest.Files {
		lastInfo, exists := lastManifest.Files[path]
		if !exists {
			result.Added = append(result.Added, path)
		} else if currentInfo.Hash != lastInfo.Hash {
			result.Modified = append(result.Modified, path)
		} else {
			result.Unchanged = append(result.Unchanged, path)
		}
	}

	// Check for deleted files
	for path := range lastManifest.Files {
		if _, exists := currentManifest.Files[path]; !exists {
			result.Deleted = append(result.Deleted, path)
		}
	}

	result.HasChanges = len(result.Added) > 0 || len(result.Modified) > 0 || len(result.Deleted) > 0

	return result, nil
}

// ComputeDiffFromChangedFiles creates a diff result from a list of known changed files
// This is more efficient when we already know which files changed (from fsnotify)
func (m *ManifestManager) ComputeDiffFromChangedFiles(workspaceDir string, changedFiles []string, lastManifest *FileManifest) (*DiffResult, error) {
	result := &DiffResult{
		Added:     []string{},
		Modified:  []string{},
		Deleted:   []string{},
		Unchanged: []string{},
		AllFiles:  []string{},
	}

	// Get all current files for AllFiles list
	currentManifest, err := m.ComputeCurrentManifest(workspaceDir, "")
	if err != nil {
		return nil, err
	}

	for path := range currentManifest.Files {
		result.AllFiles = append(result.AllFiles, path)
	}

	// If no previous manifest, treat all changed files as added
	if lastManifest == nil {
		result.Added = changedFiles
		result.HasChanges = len(changedFiles) > 0
		return result, nil
	}

	// Categorize changed files
	for _, path := range changedFiles {
		// Get relative path if absolute
		relPath := path
		if filepath.IsAbs(path) {
			var err error
			relPath, err = filepath.Rel(workspaceDir, path)
			if err != nil {
				continue
			}
		}

		// Check if file exists now
		currentInfo, existsNow := currentManifest.Files[relPath]
		_, existedBefore := lastManifest.Files[relPath]

		if !existsNow && existedBefore {
			result.Deleted = append(result.Deleted, relPath)
		} else if existsNow && !existedBefore {
			result.Added = append(result.Added, relPath)
		} else if existsNow && existedBefore {
			// Check if actually modified (hash changed)
			if lastInfo, ok := lastManifest.Files[relPath]; ok {
				if currentInfo.Hash != lastInfo.Hash {
					result.Modified = append(result.Modified, relPath)
				}
			}
		}
	}

	// All files not in changed list are unchanged
	changedSet := make(map[string]bool)
	for _, f := range changedFiles {
		relPath := f
		if filepath.IsAbs(f) {
			relPath, _ = filepath.Rel(workspaceDir, f)
		}
		changedSet[relPath] = true
	}

	for path := range currentManifest.Files {
		if !changedSet[path] {
			result.Unchanged = append(result.Unchanged, path)
		}
	}

	result.HasChanges = len(result.Added) > 0 || len(result.Modified) > 0 || len(result.Deleted) > 0

	return result, nil
}

// getManifestPath returns the path to the manifest file for a workspace
func (m *ManifestManager) getManifestPath(workspace string) string {
	return filepath.Join(m.basePath, workspace, "manifest.json")
}

// computeFileHash computes SHA256 hash of a file
func (m *ManifestManager) computeFileHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// shouldIgnoreDir checks if a directory should be ignored
func (m *ManifestManager) shouldIgnoreDir(name string) bool {
	return m.ignoredDirs[name] || strings.HasPrefix(name, ".")
}

// shouldIgnoreFile checks if a file should be ignored based on extension
func (m *ManifestManager) shouldIgnoreFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return m.ignoredExts[ext]
}

// GetChangedFiles returns files that changed (added + modified)
func (d *DiffResult) GetChangedFiles() []string {
	files := make([]string, 0, len(d.Added)+len(d.Modified))
	files = append(files, d.Added...)
	files = append(files, d.Modified...)
	return files
}
