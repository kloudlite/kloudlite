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

	// Get the username from the authenticated request context
	username, exists := c.Get("user_username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username not found in context"})
		return
	}

	// Generate a token with specific repository pull access
	token, err := h.registryAuthHandler.GenerateRepositoryToken(username.(string), repo, []string{"pull"})
	if err != nil {
		h.logger.Error("Failed to generate repository token", zap.Error(err))
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

// DeleteTag deletes a specific tag from a repository
// Uses query params: ?repo=namespace/image&tag=v1.0
func (h *RegistryCatalogHandlers) DeleteTag(c *gin.Context) {
	// Get repository name from query param
	repo := c.Query("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository name is required (use ?repo=name)"})
		return
	}

	// Get tag from query param
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag is required (use ?tag=name)"})
		return
	}

	// Get the username from the authenticated request context
	username, exists := c.Get("user_username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username not found in context"})
		return
	}

	// Generate a token with specific repository delete access
	token, err := h.registryAuthHandler.GenerateRepositoryToken(username.(string), repo, []string{"pull", "delete"})
	if err != nil {
		h.logger.Error("Failed to generate repository token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate with registry"})
		return
	}

	// First, get the manifest digest for the tag
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", h.registryURL, repo, tag)
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodHead, manifestURL, nil)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to get manifest", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get manifest"})
		return
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Registry returned non-OK status", zap.Int("status", resp.StatusCode))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("registry returned status %d", resp.StatusCode)})
		return
	}

	// Get the digest from Docker-Content-Digest header
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		h.logger.Error("No digest in response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get manifest digest"})
		return
	}

	// Now delete the manifest by digest
	deleteURL := fmt.Sprintf("%s/v2/%s/manifests/%s", h.registryURL, repo, digest)
	deleteReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodDelete, deleteURL, nil)
	if err != nil {
		h.logger.Error("Failed to create delete request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create delete request"})
		return
	}

	deleteReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	deleteResp, err := h.httpClient.Do(deleteReq)
	if err != nil {
		h.logger.Error("Failed to delete manifest", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete manifest"})
		return
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusAccepted && deleteResp.StatusCode != http.StatusOK {
		h.logger.Error("Failed to delete tag", zap.Int("status", deleteResp.StatusCode))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to delete tag: registry returned status %d", deleteResp.StatusCode)})
		return
	}

	h.logger.Info("Tag deleted", zap.String("repo", repo), zap.String("tag", tag), zap.String("digest", digest))
	c.JSON(http.StatusOK, gin.H{"message": "tag deleted successfully"})
}
