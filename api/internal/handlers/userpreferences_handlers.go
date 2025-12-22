package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
)

// UserPreferencesHandlers handles HTTP requests for UserPreferences resources
type UserPreferencesHandlers struct {
	repo   repository.UserPreferencesRepository
	logger *zap.Logger
}

// NewUserPreferencesHandlers creates a new UserPreferencesHandlers
func NewUserPreferencesHandlers(repo repository.UserPreferencesRepository, logger *zap.Logger) *UserPreferencesHandlers {
	return &UserPreferencesHandlers{
		repo:   repo,
		logger: logger,
	}
}

// GetMyPreferences handles GET /user-preferences
// Returns the current user's preferences
func (h *UserPreferencesHandlers) GetMyPreferences(c *gin.Context) {
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	prefs, err := h.repo.GetOrCreate(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get user preferences", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user preferences",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// PinWorkspaceRequest represents a request to pin a workspace
type PinWorkspaceRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// PinWorkspace handles POST /user-preferences/pinned-workspaces
func (h *UserPreferencesHandlers) PinWorkspace(c *gin.Context) {
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req PinWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	wsRef := platformv1alpha1.ResourceReference{
		Name:      req.Name,
		Namespace: req.Namespace,
	}

	if err := h.repo.AddPinnedWorkspace(c.Request.Context(), username, wsRef); err != nil {
		h.logger.Error("Failed to pin workspace", zap.Error(err), zap.String("username", username), zap.String("workspace", req.Name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to pin workspace",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Workspace pinned", zap.String("username", username), zap.String("workspace", req.Name), zap.String("namespace", req.Namespace))
	c.JSON(http.StatusOK, gin.H{"message": "Workspace pinned"})
}

// UnpinWorkspace handles DELETE /user-preferences/pinned-workspaces
func (h *UserPreferencesHandlers) UnpinWorkspace(c *gin.Context) {
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req PinWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	wsRef := platformv1alpha1.ResourceReference{
		Name:      req.Name,
		Namespace: req.Namespace,
	}

	if err := h.repo.RemovePinnedWorkspace(c.Request.Context(), username, wsRef); err != nil {
		h.logger.Error("Failed to unpin workspace", zap.Error(err), zap.String("username", username), zap.String("workspace", req.Name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to unpin workspace",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Workspace unpinned", zap.String("username", username), zap.String("workspace", req.Name), zap.String("namespace", req.Namespace))
	c.JSON(http.StatusOK, gin.H{"message": "Workspace unpinned"})
}

// PinEnvironmentRequest represents a request to pin an environment
type PinEnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
}

// PinEnvironment handles POST /user-preferences/pinned-environments
func (h *UserPreferencesHandlers) PinEnvironment(c *gin.Context) {
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req PinEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.AddPinnedEnvironment(c.Request.Context(), username, req.Name); err != nil {
		h.logger.Error("Failed to pin environment", zap.Error(err), zap.String("username", username), zap.String("environment", req.Name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to pin environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment pinned", zap.String("username", username), zap.String("environment", req.Name))
	c.JSON(http.StatusOK, gin.H{"message": "Environment pinned"})
}

// UnpinEnvironment handles DELETE /user-preferences/pinned-environments
func (h *UserPreferencesHandlers) UnpinEnvironment(c *gin.Context) {
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req PinEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.repo.RemovePinnedEnvironment(c.Request.Context(), username, req.Name); err != nil {
		h.logger.Error("Failed to unpin environment", zap.Error(err), zap.String("username", username), zap.String("environment", req.Name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to unpin environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment unpinned", zap.String("username", username), zap.String("environment", req.Name))
	c.JSON(http.StatusOK, gin.H{"message": "Environment unpinned"})
}
