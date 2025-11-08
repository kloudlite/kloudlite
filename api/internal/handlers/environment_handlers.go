package handlers

import (
	"fmt"
	"net/http"

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

	wm, err := h.workmachineRepo.GetByOwner(c, userEmail)
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
			Name: envName,
		},
		Spec: req.Spec,
	}

	// Set the OwnedBy field with the username (User's metadata.name) from JWT token
	// The webhook will handle adding ownership labels and metadata
	env.Spec.OwnedBy = username
	env.Spec.WorkMachineName = wm.Name
	env.Spec.NodeSelector = wm.Status.NodeLabels
	env.Spec.Tolerations = wm.Status.PodTolerations

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

	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
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

	c.JSON(http.StatusOK, env)
}

// ListEnvironments handles GET /api/v1/environments
func (h *EnvironmentHandlers) ListEnvironments(c *gin.Context) {
	// Parse query parameters for filtering
	labelSelector := c.Query("labelSelector")
	status := c.Query("status") // active, inactive, all

	var envList *environmentsv1.EnvironmentList
	var err error

	// Handle status-based filtering
	switch status {
	case "active":
		envList, err = h.envRepo.ListActive(c.Request.Context())
	case "inactive":
		envList, err = h.envRepo.ListInactive(c.Request.Context())
	default:
		// List all environments (cluster-scoped, so namespace is empty)
		if labelSelector != "" {
			envList, err = h.envRepo.List(c.Request.Context(), repository.WithLabelSelector(labelSelector))
		} else {
			envList, err = h.envRepo.List(c.Request.Context())
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

	c.JSON(http.StatusOK, gin.H{
		"environments": envList.Items,
		"count":        len(envList.Items),
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

	// Get existing environment
	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment for update",
			zap.String("name", name),
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

	// Get existing environment
	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment for patch",
			zap.String("name", name),
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

	// Check if force delete is requested
	force := c.Query("force") == "true"

	// Get the environment first to check if it's activated
	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment for deletion",
			zap.String("name", name),
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

	// Prevent deletion of activated environment unless forced
	if env.Spec.Activated && !force {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot delete an activated environment",
			"details": "Deactivate the environment first or use force=true query parameter",
		})
		return
	}

	// Delete the environment (cluster-scoped)
	if err := h.envRepo.Delete(c.Request.Context(), name); err != nil {
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

	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
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

	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
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

	env, err := h.envRepo.Get(c.Request.Context(), name) // cluster-scoped
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", name),
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
