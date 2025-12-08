package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceHandlers handles HTTP requests for Workspace resources
type WorkspaceHandlers struct {
	wsRepo    repository.WorkspaceRepository
	userRepo  repository.UserRepository
	wmRepo    repository.WorkMachineRepository
	k8sClient client.Client
	logger    *zap.Logger
}

// NewWorkspaceHandlers creates a new WorkspaceHandlers
func NewWorkspaceHandlers(wsRepo repository.WorkspaceRepository, userRepo repository.UserRepository, wmRepo repository.WorkMachineRepository, k8sClient client.Client, logger *zap.Logger) *WorkspaceHandlers {
	return &WorkspaceHandlers{
		wsRepo:    wsRepo,
		userRepo:  userRepo,
		wmRepo:    wmRepo,
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// CreateWorkspace handles POST /api/v1/namespaces/:namespace/workspaces
func (h *WorkspaceHandlers) CreateWorkspace(c *gin.Context) {
	// The namespace parameter is ignored - we'll use the user's WorkMachine namespace
	_ = c.Param("namespace")

	var req struct {
		Name string                     `json:"name" binding:"required"`
		Spec workspacesv1.WorkspaceSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create workspace request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get the authenticated user from context
	username, userEmail, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Find the user's WorkMachine
	workMachine, err := h.wmRepo.GetByOwner(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to find user's WorkMachine", zap.Error(err), zap.String("user", userEmail))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User does not have a WorkMachine",
			"details": "Please ensure a WorkMachine is created for your user first",
		})
		return
	}

	// Use the WorkMachine's target namespace (which should be the WorkMachine name)
	workMachineNamespace := workMachine.Spec.TargetNamespace
	if workMachineNamespace == "" {
		// If TargetNamespace is not set, use the WorkMachine name as namespace
		workMachineNamespace = workMachine.Name
	}

	h.logger.Info("Creating workspace in user's WorkMachine namespace",
		zap.String("user", userEmail),
		zap.String("workMachine", workMachine.Name),
		zap.String("namespace", workMachineNamespace))

	// Use workspace name directly (no username prefix)
	// Workspaces are namespaced to WorkMachine, ensuring uniqueness
	workspaceName := req.Name

	// Create the workspace resource in the WorkMachine's namespace
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceName,
			Namespace: workMachineNamespace,
		},
		Spec: req.Spec,
	}

	// Ensure the owner is set to the authenticated user's username (metadata.name)
	workspace.Spec.OwnedBy = username
	// Set WorkmachineName to the actual WorkMachine name (not the namespace)
	workspace.Spec.WorkmachineName = workMachine.Name

	// Note: Default values are set by the admission webhook

	// Create the workspace
	err = h.wsRepo.Create(c.Request.Context(), workspace)
	if err != nil {
		h.logger.Error("Failed to create workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

// GetWorkspace handles GET /api/v1/namespaces/:namespace/workspaces/:name
func (h *WorkspaceHandlers) GetWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	workspace, err := h.wsRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to get workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// ListWorkspaces handles GET /api/v1/namespaces/:namespace/workspaces
func (h *WorkspaceHandlers) ListWorkspaces(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		namespace = "default"
	}

	// Check for query parameters
	owner := c.Query("owner")
	workMachine := c.Query("workMachine")
	status := c.Query("status")

	var workspaces *workspacesv1.WorkspaceList
	var err error

	// Filter based on query parameters
	if owner != "" {
		workspaces, err = h.wsRepo.GetByOwner(c.Request.Context(), owner, namespace)
	} else if workMachine != "" {
		workspaces, err = h.wsRepo.GetByWorkMachine(c.Request.Context(), workMachine, namespace)
	} else if status != "" {
		switch status {
		case "active":
			workspaces, err = h.wsRepo.ListActive(c.Request.Context(), namespace)
		case "suspended":
			workspaces, err = h.wsRepo.ListSuspended(c.Request.Context(), namespace)
		case "archived":
			workspaces, err = h.wsRepo.ListArchived(c.Request.Context(), namespace)
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status filter. Must be one of: active, suspended, archived",
			})
			return
		}
	} else {
		workspaces, err = h.wsRepo.List(c.Request.Context(), namespace)
	}

	if err != nil {
		h.logger.Error("Failed to list workspaces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list workspaces",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, workspaces)
}

// UpdateWorkspace handles PUT /api/v1/namespaces/:namespace/workspaces/:name
func (h *WorkspaceHandlers) UpdateWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	// Get existing workspace
	workspace, err := h.wsRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to get workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get workspace",
			"details": err.Error(),
		})
		return
	}

	// Parse update request
	var req struct {
		Spec workspacesv1.WorkspaceSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update workspace request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update workspace spec
	workspace.Spec = req.Spec

	// Update the workspace
	err = h.wsRepo.Update(c.Request.Context(), workspace)
	if err != nil {
		h.logger.Error("Failed to update workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// DeleteWorkspace handles DELETE /api/v1/namespaces/:namespace/workspaces/:name
func (h *WorkspaceHandlers) DeleteWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	err := h.wsRepo.Delete(c.Request.Context(), namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to delete workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SuspendWorkspace handles POST /api/v1/namespaces/:namespace/workspaces/:name/suspend
func (h *WorkspaceHandlers) SuspendWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	err := h.wsRepo.SuspendWorkspace(c.Request.Context(), name, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to suspend workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to suspend workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace suspended successfully",
	})
}

// ActivateWorkspace handles POST /api/v1/namespaces/:namespace/workspaces/:name/activate
func (h *WorkspaceHandlers) ActivateWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	err := h.wsRepo.ActivateWorkspace(c.Request.Context(), name, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to activate workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to activate workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace activated successfully",
	})
}

// ArchiveWorkspace handles POST /api/v1/namespaces/:namespace/workspaces/:name/archive
func (h *WorkspaceHandlers) ArchiveWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	err := h.wsRepo.ArchiveWorkspace(c.Request.Context(), name, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to archive workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to archive workspace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace archived successfully",
	})
}

// CloneWorkspace handles POST /api/v1/namespaces/:namespace/workspaces/:name/clone
func (h *WorkspaceHandlers) CloneWorkspace(c *gin.Context) {
	namespace := c.Param("namespace")
	sourceWorkspaceName := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	var req struct {
		Name string                     `json:"name" binding:"required"`
		Spec workspacesv1.WorkspaceSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse clone workspace request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify source workspace exists
	sourceWorkspace, err := h.wsRepo.Get(c.Request.Context(), namespace, sourceWorkspaceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source workspace not found"})
			return
		}
		h.logger.Error("Failed to get source workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get source workspace", "details": err.Error()})
		return
	}

	// Verify user owns the source workspace
	if sourceWorkspace.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only clone workspaces you own"})
		return
	}

	h.logger.Info("Cloning workspace",
		zap.String("source", sourceWorkspaceName),
		zap.String("target", req.Name),
		zap.String("namespace", namespace))

	// Create new workspace with CopyFrom set
	newWorkspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
		},
		Spec: req.Spec,
	}

	newWorkspace.Spec.CopyFrom = sourceWorkspaceName
	newWorkspace.Spec.OwnedBy = username
	newWorkspace.Spec.WorkmachineName = sourceWorkspace.Spec.WorkmachineName

	if err := h.wsRepo.Create(c.Request.Context(), newWorkspace); err != nil {
		h.logger.Error("Failed to create cloned workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cloned workspace", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newWorkspace)
}

// WorkspaceMetrics represents CPU and memory metrics for a workspace
type WorkspaceMetrics struct {
	CPU struct {
		Usage int64 `json:"usage"` // in millicores
	} `json:"cpu"`
	Memory struct {
		Usage int64 `json:"usage"` // in bytes
	} `json:"memory"`
	Timestamp string `json:"timestamp"`
}

// GetMetrics handles GET /api/v1/namespaces/:namespace/workspaces/:name/metrics
func (h *WorkspaceHandlers) GetMetrics(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	// Get the workspace to find its pod
	workspace, err := h.wsRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to get workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get workspace",
			"details": err.Error(),
		})
		return
	}

	// Get pod name from workspace status
	podName := workspace.Status.PodName
	if podName == "" {
		c.JSON(http.StatusOK, &WorkspaceMetrics{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Get pod metrics from Kubernetes metrics API
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var podMetrics metricsv1beta1.PodMetrics
	err = h.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      podName,
	}, &podMetrics)
	if err != nil {
		h.logger.Warn("Failed to get pod metrics", zap.Error(err), zap.String("pod", podName))
		c.JSON(http.StatusOK, &WorkspaceMetrics{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Get pod to read resource limits
	var pod corev1.Pod
	err = h.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      podName,
	}, &pod)
	if err != nil {
		h.logger.Warn("Failed to get pod", zap.Error(err), zap.String("pod", podName))
	}

	// Calculate metrics from all containers
	var totalCPU resource.Quantity
	var totalMemory resource.Quantity
	var cpuLimit resource.Quantity
	var memLimit resource.Quantity

	for _, container := range podMetrics.Containers {
		if cpu, ok := container.Usage[corev1.ResourceCPU]; ok {
			totalCPU.Add(cpu)
		}
		if mem, ok := container.Usage[corev1.ResourceMemory]; ok {
			totalMemory.Add(mem)
		}
	}

	// Get limits from pod spec
	if pod.Name != "" {
		for _, container := range pod.Spec.Containers {
			if limit, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				cpuLimit.Add(limit)
			}
			if limit, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
				memLimit.Add(limit)
			}
		}
	}

	metrics := &WorkspaceMetrics{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// CPU in millicores
	metrics.CPU.Usage = totalCPU.MilliValue()

	// Memory in bytes
	metrics.Memory.Usage = totalMemory.Value()

	c.JSON(http.StatusOK, metrics)
}

// GetPackageRequest handles GET /api/v1/namespaces/:namespace/workspaces/:name/packages
// Returns the PackageRequest status for a workspace (source of truth for package installation)
func (h *WorkspaceHandlers) GetPackageRequest(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	// Verify the workspace exists first
	_, err := h.wsRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		h.logger.Error("Failed to get workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get workspace",
			"details": err.Error(),
		})
		return
	}

	// PackageRequest is cluster-scoped with naming convention: {workspace-name}-packages
	packageRequestName := fmt.Sprintf("%s-packages", name)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var pkgReq packagesv1.PackageRequest
	err = h.k8sClient.Get(ctx, types.NamespacedName{
		Name: packageRequestName,
	}, &pkgReq)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// PackageRequest doesn't exist yet (workspace has no packages configured)
			c.JSON(http.StatusOK, nil)
			return
		}
		h.logger.Error("Failed to get package request", zap.Error(err), zap.String("name", packageRequestName))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get package request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &pkgReq)
}
