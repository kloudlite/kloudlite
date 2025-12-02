package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

// RegistryCatalogHandlers handles Docker Registry catalog operations
type RegistryCatalogHandlers struct {
	registryURL         string
	registryAuthHandler *RegistryAuthHandlers
	authService         services.AuthService
	logger              *zap.Logger
	httpClient          *http.Client
}

// NewRegistryCatalogHandlers creates a new RegistryCatalogHandlers
func NewRegistryCatalogHandlers(
	registryURL string,
	registryAuthHandler *RegistryAuthHandlers,
	authService services.AuthService,
	logger *zap.Logger,
) *RegistryCatalogHandlers {
	return &RegistryCatalogHandlers{
		registryURL:         strings.TrimSuffix(registryURL, "/"),
		registryAuthHandler: registryAuthHandler,
		authService:         authService,
		logger:              logger,
		httpClient:          &http.Client{},
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

// getRegistryToken generates a token for accessing the registry catalog
func (h *RegistryCatalogHandlers) getRegistryToken(c *gin.Context) (string, error) {
	// Get the username from the authenticated request context
	// The JWT middleware sets "user_username" in the context
	username, exists := c.Get("user_username")
	if !exists {
		return "", fmt.Errorf("username not found in context")
	}

	// Generate a Docker Registry token with catalog access
	// For catalog access, we use a special scope that allows reading all repos
	if h.registryAuthHandler == nil {
		return "", fmt.Errorf("registry auth handler not configured")
	}

	// Generate token using the registry auth handler's signing key
	token, err := h.registryAuthHandler.generateCatalogToken(username.(string))
	if err != nil {
		return "", fmt.Errorf("failed to generate registry token: %w", err)
	}

	return token, nil
}

// ListRepositories lists all repositories in the registry
func (h *RegistryCatalogHandlers) ListRepositories(c *gin.Context) {
	// Get registry token for authentication
	token, err := h.getRegistryToken(c)
	if err != nil {
		h.logger.Error("Failed to get registry token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate with registry"})
		return
	}

	// Call registry /_catalog endpoint
	url := fmt.Sprintf("%s/v2/_catalog", h.registryURL)

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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

	// Remove leading slash if present (from wildcard capture)
	repo = strings.TrimPrefix(repo, "/")

	// Get registry token for authentication
	token, err := h.getRegistryToken(c)
	if err != nil {
		h.logger.Error("Failed to get registry token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate with registry"})
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

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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
