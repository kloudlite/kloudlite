package devbox

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchPackages_Success(t *testing.T) {
	// Create mock response
	mockResponse := SearchResponse{
		NumResults: 2,
		Packages: []PackageSearchResult{
			{
				Name:        "vim",
				NumVersions: 52,
				Versions: []PackageVersion{
					{
						CommitHash:  "abc123",
						Version:     "9.0.1",
						LastUpdated: 1234567890,
						Platforms:   []string{"x86_64-linux"},
						Summary:     "Vim text editor",
						Homepage:    "https://vim.org",
						License:     "MIT",
						Name:        "vim",
					},
				},
			},
			{
				Name:        "vim-full",
				NumVersions: 32,
				Versions:    []PackageVersion{},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Errorf("Expected path /v1/search, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "vim" {
			t.Errorf("Expected query param q=vim, got %s", r.URL.Query().Get("q"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Override SearchAPI for testing
	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	// Test search
	resp, err := SearchPackages("vim")
	if err != nil {
		t.Fatalf("SearchPackages failed: %v", err)
	}

	if resp.NumResults != 2 {
		t.Errorf("Expected 2 results, got %d", resp.NumResults)
	}
	if len(resp.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(resp.Packages))
	}
	if resp.Packages[0].Name != "vim" {
		t.Errorf("Expected first package to be 'vim', got %s", resp.Packages[0].Name)
	}
}

func TestSearchPackages_EmptyQuery(t *testing.T) {
	_, err := SearchPackages("")
	if err == nil {
		t.Fatal("Expected error for empty query, got nil")
	}
	if err.Error() != "search query is required" {
		t.Errorf("Expected 'search query is required' error, got %s", err.Error())
	}
}

func TestSearchPackages_Non200Status(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	_, err := SearchPackages("test")
	if err == nil {
		t.Fatal("Expected error for 500 status, got nil")
	}
	if err.Error() != "API returned status 500" {
		t.Errorf("Expected 'API returned status 500', got %s", err.Error())
	}
}

func TestSearchPackages_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	_, err := SearchPackages("test")
	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}
}

func TestResolvePackageVersion_Success(t *testing.T) {
	mockResponse := ResolveResponse{
		Name:    "nodejs",
		Version: "20.0.0",
		Summary: "Node.js runtime",
		Systems: map[string]SystemInfo{
			"x86_64-linux": {
				FlakeInstallable: FlakeInstallable{
					Ref: FlakeRef{
						Type:  "github",
						Owner: "NixOS",
						Repo:  "nixpkgs",
						Rev:   "def456",
					},
					AttrPath: "nodejs_20",
				},
				LastUpdated: "2024-01-01",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/resolve" {
			t.Errorf("Expected path /v2/resolve, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("name") != "nodejs" {
			t.Errorf("Expected query param name=nodejs, got %s", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("version") != "20.0.0" {
			t.Errorf("Expected query param version=20.0.0, got %s", r.URL.Query().Get("version"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	resp, err := ResolvePackageVersion("nodejs", "20.0.0")
	if err != nil {
		t.Fatalf("ResolvePackageVersion failed: %v", err)
	}

	if resp.Name != "nodejs" {
		t.Errorf("Expected name 'nodejs', got %s", resp.Name)
	}
	if resp.Version != "20.0.0" {
		t.Errorf("Expected version '20.0.0', got %s", resp.Version)
	}
	if len(resp.Systems) != 1 {
		t.Errorf("Expected 1 system, got %d", len(resp.Systems))
	}
	if resp.Systems["x86_64-linux"].FlakeInstallable.AttrPath != "nodejs_20" {
		t.Errorf("Expected attrPath 'nodejs_20', got %s",
			resp.Systems["x86_64-linux"].FlakeInstallable.AttrPath)
	}
}

func TestResolvePackageVersion_EmptyParams(t *testing.T) {
	tests := []struct {
		name    string
		pkgName string
		version string
	}{
		{"empty name", "", "1.0.0"},
		{"empty version", "test", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolvePackageVersion(tt.pkgName, tt.version)
			if err == nil {
				t.Fatal("Expected error for empty params, got nil")
			}
			if err.Error() != "package name and version are required" {
				t.Errorf("Expected 'package name and version are required', got %s", err.Error())
			}
		})
	}
}

func TestResolvePackageVersion_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	_, err := ResolvePackageVersion("nonexistent", "1.0.0")
	if err == nil {
		t.Fatal("Expected error for 404, got nil")
	}
	expected := "version 1.0.0 not found for package nonexistent"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got %s", expected, err.Error())
	}
}

func TestResolvePackageVersion_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	_, err := ResolvePackageVersion("test", "1.0.0")
	if err == nil {
		t.Fatal("Expected error for 400 status, got nil")
	}
	if err.Error() != "API returned status 400" {
		t.Errorf("Expected 'API returned status 400', got %s", err.Error())
	}
}

func TestResolvePackageVersion_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	originalAPI := SearchAPI
	SearchAPI = server.URL
	defer func() { SearchAPI = originalAPI }()

	_, err := ResolvePackageVersion("test", "1.0.0")
	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}
}
