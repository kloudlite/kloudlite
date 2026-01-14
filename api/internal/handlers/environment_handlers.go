package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnvironmentHandlers handles HTTP requests for Environment resources
type EnvironmentHandlers struct {
	envRepo         repository.EnvironmentRepository
	userRepo        repository.UserRepository
	workmachineRepo repository.WorkMachineRepository
	k8sClient       client.Client
	logger          *zap.Logger
}

// NewEnvironmentHandlers creates a new EnvironmentHandlers
func NewEnvironmentHandlers(envRepo repository.EnvironmentRepository, userRepo repository.UserRepository, workmachineRepo repository.WorkMachineRepository, k8sClient client.Client, logger *zap.Logger) *EnvironmentHandlers {
	return &EnvironmentHandlers{
		envRepo:         envRepo,
		userRepo:        userRepo,
		workmachineRepo: workmachineRepo,
		k8sClient:       k8sClient,
		logger:          logger,
	}
}

// getEnvNamespaceForUser gets the environment namespace for a user from their WorkMachine
func (h *EnvironmentHandlers) getEnvNamespaceForUser(ctx context.Context, username string) (string, error) {
	wm, err := h.workmachineRepo.GetByOwner(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to get workmachine for user %s: %w", username, err)
	}
	return wm.Spec.TargetNamespace, nil
}

// CreateEnvironment handles POST /api/v1/environments
func (h *EnvironmentHandlers) CreateEnvironment(c *gin.Context) {
	var req struct {
		Name string                         `json:"name" binding:"required"`
		Spec environmentsv1.EnvironmentSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create environment request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, userEmail, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Validate user exists by email from JWT token
	userList := &platformv1alpha1.UserList{}
	if err := h.k8sClient.List(c.Request.Context(), userList); err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate user",
		})
		return
	}

	userFound := false

	for _, u := range userList.Items {
		if u.Spec.Email == userEmail {
			userFound = true
			userEmail = u.Spec.Email
			break
		}
	}

	if !userFound {
		h.logger.Error("User not found by email", zap.String("email", userEmail))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "User not authorized",
			"details": fmt.Sprintf("User with email %s does not exist or is not authorized to create environments", userEmail),
		})
		return
	}

	// Use username (not email) to find workmachine since the label uses username
	wm, err := h.workmachineRepo.GetByOwner(c, username)
	if err != nil {
		c.Error(err)
		return
	}

	// Prefix environment name with username to avoid conflicts
	// Format: {username}--{envname}
	envName := fmt.Sprintf("%s--%s", username, req.Name)

	// Create Environment object with ownership
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envName,
			Namespace: wm.Namespace,
		},
		Spec: req.Spec,
	}

	// Set the OwnedBy field with the username (User's metadata.name) from JWT token
	// The webhook will handle adding ownership labels and metadata
	env.Spec.OwnedBy = username
	env.Spec.WorkMachineName = wm.Name

	// Extract node name from WorkMachine's NodeLabels
	if nodeName, ok := wm.Status.NodeLabels["kubernetes.io/hostname"]; ok {
		env.Spec.NodeName = nodeName
	}

	// Create the environment (cluster-scoped)
	if err := h.envRepo.Create(c.Request.Context(), env); err != nil {
		h.logger.Error("Failed to create environment",
			zap.String("name", req.Name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment created successfully",
		zap.String("name", req.Name),
		zap.String("namespace", req.Spec.TargetNamespace))

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Environment created successfully",
		"environment": env,
	})
}

// GetEnvironment handles GET /api/v1/environments/:name
func (h *EnvironmentHandlers) GetEnvironment(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	// Check if user has access to this environment
	if !UserHasAccessToEnvironment(username, env) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have access to this environment",
		})
		return
	}

	c.JSON(http.StatusOK, env)
}

// ListEnvironments handles GET /api/v1/environments
func (h *EnvironmentHandlers) ListEnvironments(c *gin.Context) {
	// Parse query parameters for filtering
	labelSelector := c.Query("labelSelector")
	status := c.Query("status") // active, inactive, all

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	var envList *environmentsv1.EnvironmentList

	// Handle status-based filtering
	switch status {
	case "active":
		envList, err = h.envRepo.ListActive(c.Request.Context(), namespace)
	case "inactive":
		envList, err = h.envRepo.ListInactive(c.Request.Context(), namespace)
	default:
		// List all environments in user's namespace
		if labelSelector != "" {
			envList, err = h.envRepo.List(c.Request.Context(), namespace, repository.WithLabelSelector(labelSelector))
		} else {
			envList, err = h.envRepo.List(c.Request.Context(), namespace)
		}
	}

	if err != nil {
		h.logger.Error("Failed to list environments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list environments",
			"details": err.Error(),
		})
		return
	}

	// Filter environments based on user access
	// User can see environments where:
	// 1. They are the owner
	// 2. Visibility is "shared" and they are in sharedWith list
	// 3. Visibility is "open"
	var accessibleEnvs []environmentsv1.Environment
	for _, env := range envList.Items {
		if UserHasAccessToEnvironment(username, &env) {
			accessibleEnvs = append(accessibleEnvs, env)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"environments": accessibleEnvs,
		"count":        len(accessibleEnvs),
	})
}

// UpdateEnvironment handles PUT /api/v1/environments/:name
func (h *EnvironmentHandlers) UpdateEnvironment(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var req struct {
		Spec environmentsv1.EnvironmentSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update environment request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	// Get existing environment
	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment for update",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	// Update the spec
	env.Spec = req.Spec

	// Update the environment
	if err := h.envRepo.Update(c.Request.Context(), env); err != nil {
		h.logger.Error("Failed to update environment",
			zap.String("name", name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment updated successfully",
		zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Environment updated successfully",
		"environment": env,
	})
}

// PatchEnvironment handles PATCH /api/v1/environments/:name
func (h *EnvironmentHandlers) PatchEnvironment(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		h.logger.Error("Failed to parse patch request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid patch data",
			"details": err.Error(),
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	// Get existing environment
	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment for patch",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	// Apply patches for specific fields
	if activated, ok := patch["activated"].(bool); ok {
		env.Spec.Activated = activated
	}

	if labels, ok := patch["labels"].(map[string]interface{}); ok {
		env.Spec.Labels = make(map[string]string)
		for k, v := range labels {
			if strVal, ok := v.(string); ok {
				env.Spec.Labels[k] = strVal
			}
		}
	}

	if annotations, ok := patch["annotations"].(map[string]interface{}); ok {
		env.Spec.Annotations = make(map[string]string)
		for k, v := range annotations {
			if strVal, ok := v.(string); ok {
				env.Spec.Annotations[k] = strVal
			}
		}
	}

	// Update the environment
	if err := h.envRepo.Update(c.Request.Context(), env); err != nil {
		h.logger.Error("Failed to patch environment",
			zap.String("name", name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to patch environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment patched successfully",
		zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Environment patched successfully",
		"environment": env,
	})
}

// DeleteEnvironment handles DELETE /api/v1/environments/:name
func (h *EnvironmentHandlers) DeleteEnvironment(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	// Verify environment exists before attempting deletion
	_, err = h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment for deletion",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	// Delete the environment
	// Environment can be deleted regardless of activation state
	if err := h.envRepo.Delete(c.Request.Context(), namespace, name); err != nil {
		h.logger.Error("Failed to delete environment",
			zap.String("name", name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment deleted successfully",
		zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message": "Environment deleted successfully",
		"name":    name,
	})
}

// ActivateEnvironment handles POST /api/v1/environments/:name/activate
func (h *EnvironmentHandlers) ActivateEnvironment(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	if env.Spec.Activated {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment is already activated",
		})
		return
	}

	env.Spec.Activated = true
	if err := h.envRepo.Update(c.Request.Context(), env); err != nil {
		h.logger.Error("Failed to activate environment",
			zap.String("name", name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to activate environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment activated successfully",
		zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Environment activated successfully",
		"environment": env,
	})
}

// DeactivateEnvironment handles POST /api/v1/environments/:name/deactivate
func (h *EnvironmentHandlers) DeactivateEnvironment(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	if !env.Spec.Activated {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment is already deactivated",
		})
		return
	}

	env.Spec.Activated = false
	if err := h.envRepo.Update(c.Request.Context(), env); err != nil {
		h.logger.Error("Failed to deactivate environment",
			zap.String("name", name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to deactivate environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment deactivated successfully",
		zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Environment deactivated successfully",
		"environment": env,
	})
}

// GetEnvironmentStatus handles GET /api/v1/environments/:name/status
func (h *EnvironmentHandlers) GetEnvironmentStatus(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":            env.Name,
		"namespace":       env.Spec.TargetNamespace,
		"activated":       env.Spec.Activated,
		"status":          env.Status,
		"resourceQuotas":  env.Spec.ResourceQuotas,
		"networkPolicies": env.Spec.NetworkPolicies,
	})
}

// EnvironmentStatusEvent represents a status event for WebSocket streaming
type EnvironmentStatusEvent struct {
	State                 string                                `json:"state"`
	Message               string                                `json:"message"`
	Activated             bool                                  `json:"activated"`
	SnapshotRestoreStatus *environmentsv1.SnapshotRestoreStatus `json:"snapshotRestoreStatus,omitempty"`
	Timestamp             time.Time                             `json:"timestamp"`
}

// GetEnvironmentCompose handles GET /api/v1/environments/:name/compose
func (h *EnvironmentHandlers) GetEnvironmentCompose(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	// Check if user has access to this environment
	if !UserHasAccessToEnvironment(username, env) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have access to this environment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":          env.Name,
		"compose":       env.Spec.Compose,
		"composeStatus": env.Status.ComposeStatus,
	})
}

// UpdateEnvironmentCompose handles PUT /api/v1/environments/:name/compose
func (h *EnvironmentHandlers) UpdateEnvironmentCompose(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var req struct {
		Compose *environmentsv1.CompositionSpec `json:"compose"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse compose request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	// Get existing environment
	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get environment for compose update",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("environment %s not found", name) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get environment",
			"details": err.Error(),
		})
		return
	}

	// Check if user has access to this environment
	if !UserHasAccessToEnvironment(username, env) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have access to this environment",
		})
		return
	}

	// Update the compose spec
	env.Spec.Compose = req.Compose

	// Update the environment
	if err := h.envRepo.Update(c.Request.Context(), env); err != nil {
		h.logger.Error("Failed to update environment compose",
			zap.String("name", name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update compose",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment compose updated successfully",
		zap.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message":       "Compose updated successfully",
		"compose":       env.Spec.Compose,
		"composeStatus": env.Status.ComposeStatus,
	})
}

// GetEnvironmentStatusWebSocket handles WebSocket connections for environment status streaming
// This endpoint provides real-time status updates via WebSocket (better Cloudflare support than SSE)
func (h *EnvironmentHandlers) GetEnvironmentStatusWebSocket(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment name is required"})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	// Verify environment exists and user has access
	env, err := h.envRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !UserHasAccessToEnvironment(username, env) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this environment"})
		return
	}

	// Upgrade to WebSocket
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established for environment status",
		zap.String("environment", name),
		zap.String("user", username))

	// Create a context that cancels when connection closes
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Handle incoming messages (for ping/pong and close)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				cancel()
				return
			}
		}
	}()

	// Helper function to send status event
	sendStatus := func(env *environmentsv1.Environment) error {
		event := EnvironmentStatusEvent{
			State:                 string(env.Status.State),
			Message:               env.Status.Message,
			Activated:             env.Spec.Activated,
			SnapshotRestoreStatus: env.Status.SnapshotRestoreStatus,
			Timestamp:             time.Now().UTC(),
		}
		return conn.WriteJSON(map[string]interface{}{
			"type": "status",
			"data": event,
		})
	}

	// Send initial status
	if err := sendStatus(env); err != nil {
		h.logger.Error("Failed to send initial status", zap.Error(err))
		return
	}

	// Start watching for changes
	watchChan, err := h.envRepo.Watch(ctx, namespace, repository.WithWatchFieldSelector(fmt.Sprintf("metadata.name=%s", name)))
	if err != nil {
		h.logger.Warn("Watch not available, using polling fallback", zap.Error(err))
		// Fall back to polling
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				env, err := h.envRepo.Get(ctx, namespace, name)
				if err != nil {
					continue
				}
				if err := sendStatus(env); err != nil {
					return
				}
			}
		}
	}

	// Stream watch events
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watchChan:
			if !ok {
				watchChan, err = h.envRepo.Watch(ctx, namespace, repository.WithWatchFieldSelector(fmt.Sprintf("metadata.name=%s", name)))
				if err != nil {
					h.logger.Error("Failed to restart environment watch", zap.Error(err))
					return
				}
				continue
			}

			if event.Error != nil {
				h.logger.Error("Watch error", zap.Error(event.Error))
				continue
			}

			if event.Type == repository.WatchEventDeleted {
				conn.WriteJSON(map[string]interface{}{
					"type": "deleted",
					"data": map[string]bool{"deleted": true},
				})
				return
			}

			if event.Object != nil {
				if err := sendStatus(event.Object); err != nil {
					return
				}
			}
		}
	}
}
