package handlers

import (
	"context"
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
	registryURL string
	authService services.AuthService
	logger      *zap.Logger
	httpClient  *http.Client
}

// NewRegistryCatalogHandlers creates a new RegistryCatalogHandlers
func NewRegistryCatalogHandlers(
	registryURL string,
	authService services.AuthService,
	logger *zap.Logger,
) *RegistryCatalogHandlers {
	return &RegistryCatalogHandlers{
		registryURL: strings.TrimSuffix(registryURL, "/"),
		authService: authService,
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
	// Call registry /_catalog endpoint (no auth needed - registry runs without auth)
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

	// Filter out repositories with no tags (deleted repos remain in catalog until GC)
	var repos []RepositoryInfo
	for _, name := range catalogResp.Repositories {
		// Check if repository has any tags
		hasTags, err := h.repositoryHasTags(c.Request.Context(), name)
		if err != nil {
			h.logger.Debug("Failed to check tags for repository", zap.String("repo", name), zap.Error(err))
			// Include repo if we can't determine - better to show than hide
			repos = append(repos, RepositoryInfo{Name: name})
			continue
		}
		if hasTags {
			repos = append(repos, RepositoryInfo{Name: name})
		}
	}

	c.JSON(http.StatusOK, RepositoryListResponse{Repositories: repos})
}

// repositoryHasTags checks if a repository has any tags
func (h *RegistryCatalogHandlers) repositoryHasTags(ctx context.Context, repo string) (bool, error) {
	url := fmt.Sprintf("%s/v2/%s/tags/list", h.registryURL, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var tagResp TagListResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagResp); err != nil {
		return false, err
	}

	return len(tagResp.Tags) > 0, nil
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

	// Call registry tags/list endpoint (no auth needed)
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

	// Check if user owns this repository (repository format: namespace/image)
	if !strings.HasPrefix(repo, username.(string)+"/") {
		h.logger.Warn("Delete tag denied - not user's repository",
			zap.String("username", username.(string)),
			zap.String("repository", repo),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete tags from your own repositories"})
		return
	}

	// First, get the manifest digest for the tag (no auth needed)
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", h.registryURL, repo, tag)
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodHead, manifestURL, nil)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

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

	// Now delete the manifest by digest (no auth needed)
	deleteURL := fmt.Sprintf("%s/v2/%s/manifests/%s", h.registryURL, repo, digest)
	deleteReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodDelete, deleteURL, nil)
	if err != nil {
		h.logger.Error("Failed to create delete request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create delete request"})
		return
	}

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

// DeleteRepository deletes all tags from a repository, effectively deleting the repository
// Uses query param: ?repo=namespace/image
func (h *RegistryCatalogHandlers) DeleteRepository(c *gin.Context) {
	// Get repository name from query param
	repo := c.Query("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository name is required (use ?repo=name)"})
		return
	}

	// Get the username from the authenticated request context
	username, exists := c.Get("user_username")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username not found in context"})
		return
	}

	// Check if user owns this repository (repository format: namespace/image)
	if !strings.HasPrefix(repo, username.(string)+"/") {
		h.logger.Warn("Delete repository denied - not user's repository",
			zap.String("username", username.(string)),
			zap.String("repository", repo),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own repositories"})
		return
	}

	// First, list all tags for the repository (no auth needed)
	tagsURL := fmt.Sprintf("%s/v2/%s/tags/list", h.registryURL, repo)
	tagsReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, tagsURL, nil)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	tagsResp, err := h.httpClient.Do(tagsReq)
	if err != nil {
		h.logger.Error("Failed to fetch tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags"})
		return
	}
	defer tagsResp.Body.Close()

	if tagsResp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	if tagsResp.StatusCode != http.StatusOK {
		h.logger.Error("Registry returned non-OK status", zap.Int("status", tagsResp.StatusCode))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("registry returned status %d", tagsResp.StatusCode)})
		return
	}

	var tagList TagListResponse
	if err := json.NewDecoder(tagsResp.Body).Decode(&tagList); err != nil {
		h.logger.Error("Failed to decode tags response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode tags response"})
		return
	}

	if len(tagList.Tags) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "repository has no tags", "deleted": 0})
		return
	}

	// Delete each tag by getting its manifest digest and deleting the manifest
	deletedCount := 0
	failedTags := []string{}

	for _, tag := range tagList.Tags {
		// Get manifest digest (no auth needed)
		manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", h.registryURL, repo, tag)
		manifestReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodHead, manifestURL, nil)
		if err != nil {
			failedTags = append(failedTags, tag)
			continue
		}
		manifestReq.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
		manifestReq.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")

		manifestResp, err := h.httpClient.Do(manifestReq)
		if err != nil {
			failedTags = append(failedTags, tag)
			continue
		}
		manifestResp.Body.Close()

		if manifestResp.StatusCode != http.StatusOK {
			failedTags = append(failedTags, tag)
			continue
		}

		digest := manifestResp.Header.Get("Docker-Content-Digest")
		if digest == "" {
			failedTags = append(failedTags, tag)
			continue
		}

		// Delete manifest by digest (no auth needed)
		deleteURL := fmt.Sprintf("%s/v2/%s/manifests/%s", h.registryURL, repo, digest)
		deleteReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodDelete, deleteURL, nil)
		if err != nil {
			failedTags = append(failedTags, tag)
			continue
		}

		deleteResp, err := h.httpClient.Do(deleteReq)
		if err != nil {
			failedTags = append(failedTags, tag)
			continue
		}
		deleteResp.Body.Close()

		if deleteResp.StatusCode == http.StatusAccepted || deleteResp.StatusCode == http.StatusOK {
			deletedCount++
		} else {
			failedTags = append(failedTags, tag)
		}
	}

	h.logger.Info("Repository deletion completed",
		zap.String("repo", repo),
		zap.Int("deleted", deletedCount),
		zap.Int("failed", len(failedTags)),
	)

	if len(failedTags) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message":     "some tags could not be deleted",
			"deleted":     deletedCount,
			"failed":      len(failedTags),
			"failed_tags": failedTags,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "repository deleted successfully", "deleted": deletedCount})
}
