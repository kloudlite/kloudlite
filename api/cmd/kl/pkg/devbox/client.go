package devbox

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var SearchAPI = "https://search.devbox.sh"

// API response structures
type PackageVersion struct {
	CommitHash  string   `json:"commit_hash"`
	Version     string   `json:"version"`
	LastUpdated int64    `json:"last_updated"`
	Platforms   []string `json:"platforms"`
	Summary     string   `json:"summary"`
	Homepage    string   `json:"homepage"`
	License     string   `json:"license"`
	Name        string   `json:"name"`
}

type PackageSearchResult struct {
	Name        string           `json:"name"`
	NumVersions int              `json:"num_versions"`
	Versions    []PackageVersion `json:"versions"`
}

type SearchResponse struct {
	NumResults int                   `json:"num_results"`
	Packages   []PackageSearchResult `json:"packages"`
}

type FlakeRef struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Rev   string `json:"rev"`
}

type FlakeInstallable struct {
	Ref      FlakeRef `json:"ref"`
	AttrPath string   `json:"attr_path"`
}

type SystemInfo struct {
	FlakeInstallable FlakeInstallable `json:"flake_installable"`
	LastUpdated      string           `json:"last_updated"`
}

type ResolveResponse struct {
	Name    string                `json:"name"`
	Version string                `json:"version"`
	Summary string                `json:"summary"`
	Systems map[string]SystemInfo `json:"systems"`
}

// SearchPackages searches for packages using the Devbox API
func SearchPackages(query string) (*SearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	url := fmt.Sprintf("%s/v1/search?q=%s", SearchAPI, query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}

// ResolvePackageVersion resolves a package name and semantic version to a nixpkgs commit
func ResolvePackageVersion(name, version string) (*ResolveResponse, error) {
	if name == "" || version == "" {
		return nil, fmt.Errorf("package name and version are required")
	}

	url := fmt.Sprintf("%s/v2/resolve?name=%s&version=%s", SearchAPI, name, version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve package: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("version %s not found for package %s", version, name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var resolveResp ResolveResponse
	if err := json.NewDecoder(resp.Body).Decode(&resolveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resolveResp, nil
}
