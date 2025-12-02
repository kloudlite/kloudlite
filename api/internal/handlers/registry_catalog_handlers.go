package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegistryCatalogHandlers handles Docker Registry catalog operations
type RegistryCatalogHandlers struct {
	registryURL string
	logger      *zap.Logger
	httpClient  *http.Client
}

// NewRegistryCatalogHandlers creates a new RegistryCatalogHandlers
func NewRegistryCatalogHandlers(registryURL string, logger *zap.Logger) *RegistryCatalogHandlers {
	return &RegistryCatalogHandlers{
		registryURL: strings.TrimSuffix(registryURL, "/"),
		logger:      logger,
		httpClient:  &http.Client{},
	}
}

// RegistryCatalogResponse represents the response from registry /_catalog endpoint
type RegistryCatalogResponse struct {
	Repositories []string `json:"repositories"`
}

// RepositoryInfo represents a repository with metadata
type RepositoryInfo struct {
	Name string `json:"name"`
}

// RepositoryListResponse is the API response for listing repositories
type RepositoryListResponse struct {
	Repositories []RepositoryInfo `json:"repositories"`
}

// TagListResponse represents the response from registry /tags/list endpoint
type TagListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// ListRepositories lists all repositories in the registry
func (h *RegistryCatalogHandlers) ListRepositories(c *gin.Context) {
	// Call registry /_catalog endpoint
	url := fmt.Sprintf("%s/v2/_catalog", h.registryURL)

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to fetch catalog from registry", zap.Error(err), zap.String("url", url))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch catalog from registry"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Registry returned non-OK status", zap.Int("status", resp.StatusCode))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("registry returned status %d", resp.StatusCode)})
		return
	}

	var catalogResp RegistryCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&catalogResp); err != nil {
		h.logger.Error("Failed to decode registry response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode registry response"})
		return
	}

	// Convert to our response format
	repos := make([]RepositoryInfo, len(catalogResp.Repositories))
	for i, name := range catalogResp.Repositories {
		repos[i] = RepositoryInfo{Name: name}
	}

	c.JSON(http.StatusOK, RepositoryListResponse{Repositories: repos})
}

// ListTags lists all tags for a specific repository
func (h *RegistryCatalogHandlers) ListTags(c *gin.Context) {
	// Get repository name from path - handle nested paths like "username/image"
	repo := c.Param("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository name is required"})
		return
	}

	// URL encode the repo name for the registry API
	url := fmt.Sprintf("%s/v2/%s/tags/list", h.registryURL, repo)

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to fetch tags from registry", zap.Error(err), zap.String("url", url))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags from registry"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Registry returned non-OK status", zap.Int("status", resp.StatusCode))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("registry returned status %d", resp.StatusCode)})
		return
	}

	var tagResp TagListResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagResp); err != nil {
		h.logger.Error("Failed to decode registry response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode registry response"})
		return
	}

	c.JSON(http.StatusOK, tagResp)
}
