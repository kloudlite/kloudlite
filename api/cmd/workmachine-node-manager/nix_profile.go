package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	zap2 "go.uber.org/zap"
)

const (
	// profilesBaseDir is where all Kloudlite workspace profiles are stored
	profilesBaseDir = "/nix/profiles/kloudlite"
)

// BuildResult represents the result of building a Nix profile
type BuildResult struct {
	Success       bool
	StorePath     string
	PackagesPath  string
	FailedPackage string
	Error         string
}

// NixProfileManager handles declarative Nix profile generation and installation
type NixProfileManager struct {
	logger  *zap2.Logger
	cmdExec CommandExecutor
}

// NewNixProfileManager creates a new NixProfileManager
func NewNixProfileManager(logger *zap2.Logger, cmdExec CommandExecutor) *NixProfileManager {
	return &NixProfileManager{
		logger:  logger,
		cmdExec: cmdExec,
	}
}

// GetProfileDir returns the directory for a workspace's profile
func (m *NixProfileManager) GetProfileDir(workspace string) string {
	return filepath.Join(profilesBaseDir, workspace)
}

// GetProfileNixPath returns the path to the profile.nix file
func (m *NixProfileManager) GetProfileNixPath(workspace string) string {
	return filepath.Join(m.GetProfileDir(workspace), "profile.nix")
}

// GetPackagesPath returns the path to the packages symlink
func (m *NixProfileManager) GetPackagesPath(workspace string) string {
	return filepath.Join(m.GetProfileDir(workspace), "packages")
}

// ComputeSpecHash computes a deterministic hash of the package specifications
// This is used to detect if packages have changed
func (m *NixProfileManager) ComputeSpecHash(packages []packagesv1.PackageSpec) string {
	// Sort packages by name for deterministic ordering
	sorted := make([]packagesv1.PackageSpec, len(packages))
	copy(sorted, packages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	// Build a canonical string representation
	var parts []string
	for _, pkg := range sorted {
		part := pkg.Name
		if pkg.NixpkgsCommit != "" {
			part += "@" + pkg.NixpkgsCommit
		} else if pkg.Channel != "" {
			part += "#" + pkg.Channel
		}
		parts = append(parts, part)
	}

	hash := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return hex.EncodeToString(hash[:8]) // Short hash is sufficient
}

// branchCommitCache caches branch to commit SHA mappings
var branchCommitCache = struct {
	sync.RWMutex
	cache map[string]struct {
		commit    string
		expiresAt time.Time
	}
}{
	cache: make(map[string]struct {
		commit    string
		expiresAt time.Time
	}),
}

// resolveBranchToCommit resolves a GitHub branch name to its commit SHA
// This makes the URL cacheable by Nix since commit SHAs are immutable
func resolveBranchToCommit(branch string) (string, error) {
	// Check cache first (cache for 1 hour to avoid excessive API calls)
	branchCommitCache.RLock()
	if cached, ok := branchCommitCache.cache[branch]; ok && time.Now().Before(cached.expiresAt) {
		branchCommitCache.RUnlock()
		return cached.commit, nil
	}
	branchCommitCache.RUnlock()

	// Fetch from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/NixOS/nixpkgs/branches/%s", branch)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch branch info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d for branch %s", resp.StatusCode, branch)
	}

	var result struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode branch info: %w", err)
	}

	if result.Commit.SHA == "" {
		return "", fmt.Errorf("no commit SHA found for branch %s", branch)
	}

	// Cache the result for 1 hour
	branchCommitCache.Lock()
	branchCommitCache.cache[branch] = struct {
		commit    string
		expiresAt time.Time
	}{
		commit:    result.Commit.SHA,
		expiresAt: time.Now().Add(1 * time.Hour),
	}
	branchCommitCache.Unlock()

	return result.Commit.SHA, nil
}

// GenerateProfileNix creates a .nix file for the workspace's packages
func (m *NixProfileManager) GenerateProfileNix(workspace string, packages []packagesv1.PackageSpec) (string, error) {
	profileDir := m.GetProfileDir(workspace)
	nixPath := m.GetProfileNixPath(workspace)

	// Ensure profile directory exists
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create profile directory %s: %w", profileDir, err)
	}

	// Group packages by their nixpkgs source (commit or channel)
	type pkgSource struct {
		commit  string
		channel string
	}
	sourcePackages := make(map[pkgSource][]string)

	for _, pkg := range packages {
		src := pkgSource{
			commit:  pkg.NixpkgsCommit,
			channel: pkg.Channel,
		}
		sourcePackages[src] = append(sourcePackages[src], pkg.Name)
	}

	// Generate Nix expression
	var nixContent strings.Builder
	nixContent.WriteString("# Auto-generated by Kloudlite - DO NOT EDIT\n")
	nixContent.WriteString("# Workspace: " + workspace + "\n")
	nixContent.WriteString("# Generated: " + time.Now().UTC().Format(time.RFC3339) + "\n")
	nixContent.WriteString("let\n")

	// Generate nixpkgs imports for each source
	sourceVars := make(map[pkgSource]string)
	sourceIdx := 0
	for src := range sourcePackages {
		varName := fmt.Sprintf("pkgs_%d", sourceIdx)
		sourceVars[src] = varName

		if src.commit != "" {
			// Use specific commit
			nixContent.WriteString(fmt.Sprintf("  %s = import (fetchTarball {\n", varName))
			nixContent.WriteString(fmt.Sprintf("    url = \"https://github.com/nixos/nixpkgs/archive/%s.tar.gz\";\n", src.commit))
			nixContent.WriteString("  }) { config.allowUnfree = true; };\n")
		} else if src.channel != "" {
			// Resolve channel/branch to specific commit for cacheable URL
			commitSHA, err := resolveBranchToCommit(src.channel)
			if err != nil {
				m.logger.Warn("Failed to resolve branch to commit, using branch URL (uncacheable)",
					zap2.String("branch", src.channel),
					zap2.Error(err))
				// Fallback to branch URL if resolution fails
				nixContent.WriteString(fmt.Sprintf("  %s = import (fetchTarball {\n", varName))
				nixContent.WriteString(fmt.Sprintf("    url = \"https://github.com/nixos/nixpkgs/archive/refs/heads/%s.tar.gz\";\n", src.channel))
				nixContent.WriteString("  }) { config.allowUnfree = true; };\n")
			} else {
				// Use resolved commit SHA (cacheable)
				m.logger.Info("Resolved branch to commit",
					zap2.String("branch", src.channel),
					zap2.String("commit", commitSHA[:12]))
				nixContent.WriteString(fmt.Sprintf("  # Channel: %s (resolved to commit)\n", src.channel))
				nixContent.WriteString(fmt.Sprintf("  %s = import (fetchTarball {\n", varName))
				nixContent.WriteString(fmt.Sprintf("    url = \"https://github.com/nixos/nixpkgs/archive/%s.tar.gz\";\n", commitSHA))
				nixContent.WriteString("  }) { config.allowUnfree = true; };\n")
			}
		} else {
			// Use default nixpkgs (unstable)
			nixContent.WriteString(fmt.Sprintf("  %s = import <nixpkgs> { config.allowUnfree = true; };\n", varName))
		}
		sourceIdx++
	}

	// Use first source for buildEnv (they all have it)
	var firstVar string
	for _, v := range sourceVars {
		firstVar = v
		break
	}
	if firstVar == "" {
		firstVar = "pkgs_0"
		nixContent.WriteString("  pkgs_0 = import <nixpkgs> { config.allowUnfree = true; };\n")
	}

	nixContent.WriteString("in\n")
	nixContent.WriteString(fmt.Sprintf("%s.buildEnv {\n", firstVar))
	nixContent.WriteString(fmt.Sprintf("  name = \"%s-packages\";\n", workspace))
	nixContent.WriteString("  paths = [\n")

	// Add all packages
	for src, pkgNames := range sourcePackages {
		varName := sourceVars[src]
		for _, name := range pkgNames {
			// Use .out to bypass meta.outputsToInstall issues, with fallback
			nixContent.WriteString(fmt.Sprintf("    (%s.%s.out or %s.%s)\n", varName, name, varName, name))
		}
	}

	nixContent.WriteString("  ];\n")
	nixContent.WriteString("  extraOutputsToInstall = [ \"man\" \"doc\" \"info\" ];\n")
	nixContent.WriteString("  ignoreCollisions = true;\n") // Handle file conflicts gracefully
	nixContent.WriteString("}\n")

	// Write the file
	if err := os.WriteFile(nixPath, []byte(nixContent.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write profile.nix: %w", err)
	}

	m.logger.Info("Generated profile.nix",
		zap2.String("workspace", workspace),
		zap2.String("path", nixPath),
		zap2.Int("packageCount", len(packages)))

	return nixPath, nil
}

// BuildAndActivate builds the .nix file and atomically switches the packages symlink
// It streams output with [pkg:<workspace>] tags for CLI log filtering
func (m *NixProfileManager) BuildAndActivate(ctx context.Context, workspace string) (BuildResult, error) {
	nixPath := m.GetProfileNixPath(workspace)
	packagesPath := m.GetPackagesPath(workspace)

	// Check if profile.nix exists
	if _, err := os.Stat(nixPath); os.IsNotExist(err) {
		return BuildResult{}, fmt.Errorf("profile.nix does not exist at %s", nixPath)
	}

	m.logger.Info("Building Nix profile",
		zap2.String("workspace", workspace),
		zap2.String("nixPath", nixPath))

	// Build with nix-build using streaming output
	// Use --no-out-link to avoid creating result symlink, we'll manage our own
	buildScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-build %s --no-out-link 2>&1", nixPath)

	output, err := executeWithStreamingOutput(buildScript, workspace)
	if err != nil {
		// Parse the error to identify which package failed
		failedPkg := parseNixBuildError(string(output))
		m.logger.Error("Nix build failed",
			zap2.String("workspace", workspace),
			zap2.String("output", string(output)),
			zap2.Error(err))

		return BuildResult{
			Success:       false,
			FailedPackage: failedPkg,
			Error:         string(output),
		}, nil
	}

	storePath := strings.TrimSpace(string(output))
	// nix-build outputs the store path on success
	// It might have multiple lines, get the last non-empty line
	lines := strings.Split(storePath, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "/nix/store/") {
			storePath = line
			break
		}
	}

	if !strings.HasPrefix(storePath, "/nix/store/") {
		return BuildResult{
			Success: false,
			Error:   fmt.Sprintf("unexpected nix-build output: %s", string(output)),
		}, nil
	}

	m.logger.Info("Nix build succeeded",
		zap2.String("workspace", workspace),
		zap2.String("storePath", storePath))

	// Atomic symlink switch
	// Create new symlink, then rename (atomic on POSIX)
	tmpLink := packagesPath + ".new"
	os.Remove(tmpLink) // Remove if exists from failed previous attempt

	if err := os.Symlink(storePath, tmpLink); err != nil {
		return BuildResult{}, fmt.Errorf("failed to create symlink: %w", err)
	}

	if err := os.Rename(tmpLink, packagesPath); err != nil {
		os.Remove(tmpLink)
		return BuildResult{}, fmt.Errorf("failed to activate profile: %w", err)
	}

	m.logger.Info("Profile activated",
		zap2.String("workspace", workspace),
		zap2.String("packagesPath", packagesPath),
		zap2.String("storePath", storePath))

	return BuildResult{
		Success:      true,
		StorePath:    storePath,
		PackagesPath: packagesPath,
	}, nil
}

// executeWithStreamingOutput runs a command and prints each line with a tag prefix
// Tag format: [pkg:<tag>] <line>
// This allows the CLI to filter and display build progress
func executeWithStreamingOutput(script string, tag string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", script)
	cmd.Env = os.Environ()

	// Get pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Collect output while printing tagged lines
	var output []byte
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Function to read and print tagged output
	readAndPrint := func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			// Print tagged output for CLI to capture
			fmt.Printf("[pkg:%s] %s\n", tag, line)
			// Also collect the output
			mu.Lock()
			output = append(output, []byte(line+"\n")...)
			mu.Unlock()
		}
	}

	wg.Add(2)
	go readAndPrint(stdout)
	go readAndPrint(stderr)

	// Wait for all output to be read
	wg.Wait()

	// Wait for command to finish
	err = cmd.Wait()
	return output, err
}

// CleanupProfile removes the profile directory for a workspace
func (m *NixProfileManager) CleanupProfile(workspace string) error {
	profileDir := m.GetProfileDir(workspace)

	m.logger.Info("Cleaning up profile",
		zap2.String("workspace", workspace),
		zap2.String("profileDir", profileDir))

	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to remove profile directory: %w", err)
	}

	return nil
}

// parseNixBuildError attempts to extract the package name from a nix-build error
func parseNixBuildError(output string) string {
	// Common error patterns:
	// - "error: attribute 'packagename' missing"
	// - "error: undefined variable 'packagename'"
	// - "error: Package 'packagename' is marked as unfree"

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)

		// Look for "attribute 'X' missing" pattern
		if strings.Contains(line, "attribute") && strings.Contains(line, "missing") {
			// Extract the attribute name
			start := strings.Index(line, "'")
			if start != -1 {
				end := strings.Index(line[start+1:], "'")
				if end != -1 {
					return line[start+1 : start+1+end]
				}
			}
		}

		// Look for "undefined variable" pattern
		if strings.Contains(line, "undefined variable") {
			start := strings.Index(line, "'")
			if start != -1 {
				end := strings.Index(line[start+1:], "'")
				if end != -1 {
					return line[start+1 : start+1+end]
				}
			}
		}
	}

	return "" // Unknown which package failed
}
