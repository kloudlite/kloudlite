package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// WebSocket upgrader with permissive origin check (auth is handled via JWT)
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins, auth is handled via JWT
	},
}

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

// sanitizeWorkspaceForNonOwner removes sensitive info from workspace response for non-owners
func (h *WorkspaceHandlers) sanitizeWorkspaceForNonOwner(ws workspacesv1.Workspace) workspacesv1.Workspace {
	// Keep only exposed URLs, remove IDE connections
	// accessUrls contains: code-server, ttyd, ssh, vscode-tunnel, etc.
	// These should be hidden for non-owners
	ws.Status.AccessURLs = make(map[string]string)

	// Keep exposedRoutes (user-defined port exposures)
	// ws.Status.ExposedRoutes - KEEP THIS

	// Hide pod details
	ws.Status.PodName = ""
	ws.Status.PodIP = ""
	ws.Status.NodeName = ""

	// Hide connection/idle info
	ws.Status.ActiveConnections = 0
	ws.Status.LastActivityTime = nil

	return ws
}

// requireOwnership checks if the authenticated user is the workspace owner
// Returns false and sends 403 response if not the owner
func (h *WorkspaceHandlers) requireOwnership(c *gin.Context, ws *workspacesv1.Workspace) bool {
	username, _, _, _ := middleware.GetUserFromContext(c)
	if ws.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Only the workspace owner can perform this action",
		})
		return false
	}
	return true
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

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
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

	// Check if user has access to this workspace
	if !UserHasAccessToWorkspace(username, workspace) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have access to this workspace",
		})
		return
	}

	// Sanitize response for non-owners
	if workspace.Spec.OwnedBy != username {
		*workspace = h.sanitizeWorkspaceForNonOwner(*workspace)
	}

	c.JSON(http.StatusOK, workspace)
}

// ListWorkspaces handles GET /api/v1/namespaces/:namespace/workspaces
func (h *WorkspaceHandlers) ListWorkspaces(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		namespace = "default"
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

	// Filter workspaces based on user access (visibility)
	var accessibleWorkspaces []workspacesv1.Workspace
	for _, ws := range workspaces.Items {
		if UserHasAccessToWorkspace(username, &ws) {
			// Sanitize response for non-owners
			if ws.Spec.OwnedBy != username {
				ws = h.sanitizeWorkspaceForNonOwner(ws)
			}
			accessibleWorkspaces = append(accessibleWorkspaces, ws)
		}
	}
	workspaces.Items = accessibleWorkspaces

	c.JSON(http.StatusOK, workspaces)
}

// ListAllWorkspaces handles GET /api/v1/workspaces
// Returns all workspaces the authenticated user has access to (owned + shared + open)
func (h *WorkspaceHandlers) ListAllWorkspaces(c *gin.Context) {
	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Error("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Check for query parameters
	status := c.Query("status")

	// Fetch ALL workspaces, then filter by visibility access
	workspaces, err := h.wsRepo.ListAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list all workspaces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list workspaces",
			"details": err.Error(),
		})
		return
	}

	// Filter by visibility access
	var accessibleWorkspaces []workspacesv1.Workspace
	for _, ws := range workspaces.Items {
		if UserHasAccessToWorkspace(username, &ws) {
			// Apply status filter if specified
			if status != "" {
				switch status {
				case "active", "suspended", "archived":
					if string(ws.Spec.Status) != status {
						continue
					}
				default:
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid status filter. Must be one of: active, suspended, archived",
					})
					return
				}
			}

			// Sanitize response for non-owners
			if ws.Spec.OwnedBy != username {
				ws = h.sanitizeWorkspaceForNonOwner(ws)
			}
			accessibleWorkspaces = append(accessibleWorkspaces, ws)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": accessibleWorkspaces,
		"count": len(accessibleWorkspaces),
	})
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

	// Only owner can update workspace
	if !h.requireOwnership(c, workspace) {
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

	// Get workspace to check ownership
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

	// Only owner can delete workspace
	if !h.requireOwnership(c, workspace) {
		return
	}

	err = h.wsRepo.Delete(c.Request.Context(), namespace, name)
	if err != nil {
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

	// Get workspace to check ownership
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

	// Only owner can suspend workspace
	if !h.requireOwnership(c, workspace) {
		return
	}

	err = h.wsRepo.SuspendWorkspace(c.Request.Context(), name, namespace)
	if err != nil {
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

	// Get workspace to check ownership
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

	// Only owner can activate workspace
	if !h.requireOwnership(c, workspace) {
		return
	}

	err = h.wsRepo.ActivateWorkspace(c.Request.Context(), name, namespace)
	if err != nil {
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

	// Get workspace to check ownership
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

	// Only owner can archive workspace
	if !h.requireOwnership(c, workspace) {
		return
	}

	err = h.wsRepo.ArchiveWorkspace(c.Request.Context(), name, namespace)
	if err != nil {
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

	// Check visibility-based access (not just ownership)
	if !UserHasAccessToWorkspace(username, sourceWorkspace) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to clone this workspace"})
		return
	}

	// Clone goes to the cloning user's WorkMachine namespace (not source namespace)
	clonerWorkMachine, err := h.wmRepo.GetByOwner(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get cloner's WorkMachine", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "You don't have a WorkMachine to clone to"})
		return
	}

	h.logger.Info("Cloning workspace",
		zap.String("source", fmt.Sprintf("%s/%s", namespace, sourceWorkspaceName)),
		zap.String("target", req.Name),
		zap.String("targetNamespace", clonerWorkMachine.Spec.TargetNamespace))

	// Create new workspace with CopyFrom set in user's own namespace
	newWorkspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: clonerWorkMachine.Spec.TargetNamespace, // User's own namespace
		},
		Spec: req.Spec,
	}

	newWorkspace.Spec.CopyFrom = fmt.Sprintf("%s/%s", namespace, sourceWorkspaceName) // Full reference
	newWorkspace.Spec.OwnedBy = username
	newWorkspace.Spec.WorkmachineName = clonerWorkMachine.Name

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

	// Only owner can access metrics
	if !h.requireOwnership(c, workspace) {
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

	// Only owner can access package request
	if !h.requireOwnership(c, workspace) {
		return
	}

	// PackageRequest is namespace-scoped with naming convention: {workspace-name}-packages
	packageRequestName := fmt.Sprintf("%s-packages", name)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var pkgReq packagesv1.PackageRequest
	err = h.k8sClient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      packageRequestName,
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

// UpdatePackageRequest handles PUT /api/v1/namespaces/:namespace/workspaces/:name/packages
// Creates or updates the PackageRequest for a workspace
func (h *WorkspaceHandlers) UpdatePackageRequest(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	// Parse request body
	var req struct {
		Packages []packagesv1.PackageSpec `json:"packages" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update package request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Verify the workspace exists and get its details
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

	// Only owner can update package request
	if !h.requireOwnership(c, workspace) {
		return
	}

	// PackageRequest is namespace-scoped with naming convention: {workspace-name}-packages
	packageRequestName := fmt.Sprintf("%s-packages", name)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Try to get existing PackageRequest
	var pkgReq packagesv1.PackageRequest
	err = h.k8sClient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      packageRequestName,
	}, &pkgReq)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new PackageRequest
			h.logger.Info("Creating new PackageRequest", zap.String("name", packageRequestName))

			pkgReq = packagesv1.PackageRequest{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      packageRequestName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "workspaces.kloudlite.io/v1",
							Kind:       "Workspace",
							Name:       workspace.Name,
							UID:        workspace.UID,
						},
					},
				},
				Spec: packagesv1.PackageRequestSpec{
					WorkspaceRef: name,
					Packages:     req.Packages,
					ProfileName:  packageRequestName,
				},
			}

			if err := h.k8sClient.Create(ctx, &pkgReq); err != nil {
				h.logger.Error("Failed to create package request", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to create package request",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusCreated, &pkgReq)
			return
		}

		h.logger.Error("Failed to get package request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get package request",
			"details": err.Error(),
		})
		return
	}

	// Update existing PackageRequest
	h.logger.Info("Updating PackageRequest", zap.String("name", packageRequestName))
	pkgReq.Spec.Packages = req.Packages

	if err := h.k8sClient.Update(ctx, &pkgReq); err != nil {
		h.logger.Error("Failed to update package request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update package request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &pkgReq)
}

// WorkspaceStatusEvent represents a status event for SSE streaming
type WorkspaceStatusEvent struct {
	Phase             string            `json:"phase"`
	Message           string            `json:"message"`
	Status            string            `json:"status"`
	ActiveConnections int               `json:"activeConnections"`
	IdleState         string            `json:"idleState"`
	AccessURLs        map[string]string `json:"accessUrls,omitempty"`
	Timestamp         time.Time         `json:"timestamp"`
}

// GetWorkspaceStatusStream handles GET /api/v1/namespaces/:namespace/workspaces/:name/status-stream
// This endpoint streams status updates via Server-Sent Events (SSE)
func (h *WorkspaceHandlers) GetWorkspaceStatusStream(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Workspace name is required",
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

	// Verify workspace exists and user has access
	ws, err := h.wsRepo.Get(ctx, namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check access
	if !UserHasAccessToWorkspace(username, ws) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have access to this workspace",
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Helper function to build and send status event
	sendStatusEvent := func(ws *workspacesv1.Workspace) {
		event := WorkspaceStatusEvent{
			Phase:             ws.Status.Phase,
			Message:           ws.Status.Message,
			Status:            string(ws.Spec.Status),
			ActiveConnections: ws.Status.ActiveConnections,
			IdleState:         ws.Status.IdleState,
			Timestamp:         time.Now().UTC(),
		}

		// Only include access URLs for owner
		if ws.Spec.OwnedBy == username {
			event.AccessURLs = ws.Status.AccessURLs
		}

		eventData, err := json.Marshal(event)
		if err != nil {
			h.logger.Error("Failed to marshal status event", zap.Error(err))
			return
		}

		c.Writer.Write([]byte("event: status\n"))
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(eventData)
		c.Writer.Write([]byte("\n\n"))
		c.Writer.Flush()
	}

	// Send initial status immediately
	sendStatusEvent(ws)

	// Start watching for changes
	watchChan, err := h.wsRepo.Watch(ctx, namespace, repository.WithWatchFieldSelector(fmt.Sprintf("metadata.name=%s", name)))
	if err != nil {
		h.logger.Warn("Watch not available, using polling fallback", zap.Error(err))
		// Fall back to polling if watch fails
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ws, err := h.wsRepo.Get(ctx, namespace, name)
				if err != nil {
					continue
				}
				sendStatusEvent(ws)
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
				// Watch channel closed, restart it
				watchChan, err = h.wsRepo.Watch(ctx, namespace, repository.WithWatchFieldSelector(fmt.Sprintf("metadata.name=%s", name)))
				if err != nil {
					h.logger.Error("Failed to restart workspace watch", zap.Error(err))
					return
				}
				continue
			}

			if event.Error != nil {
				h.logger.Error("Watch error", zap.Error(event.Error))
				continue
			}

			if event.Type == repository.WatchEventDeleted {
				// Workspace was deleted, send final event and close
				c.Writer.Write([]byte("event: deleted\n"))
				c.Writer.Write([]byte("data: {\"deleted\": true}\n\n"))
				c.Writer.Flush()
				return
			}

			if event.Object != nil {
				sendStatusEvent(event.Object)
			}
		}
	}
}

// GetWorkspaceStatusWebSocket handles WebSocket connections for workspace status streaming
// This endpoint provides real-time status updates via WebSocket (better Cloudflare support than SSE)
func (h *WorkspaceHandlers) GetWorkspaceStatusWebSocket(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace name is required"})
		return
	}

	// Get the authenticated user from JWT middleware context
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify workspace exists and user has access
	ws, err := h.wsRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !UserHasAccessToWorkspace(username, ws) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this workspace"})
		return
	}

	// Upgrade to WebSocket
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established for workspace status",
		zap.String("namespace", namespace),
		zap.String("workspace", name),
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
	sendStatus := func(ws *workspacesv1.Workspace) error {
		event := WorkspaceStatusEvent{
			Phase:             ws.Status.Phase,
			Message:           ws.Status.Message,
			Status:            string(ws.Spec.Status),
			ActiveConnections: ws.Status.ActiveConnections,
			IdleState:         ws.Status.IdleState,
			Timestamp:         time.Now().UTC(),
		}
		if ws.Spec.OwnedBy == username {
			event.AccessURLs = ws.Status.AccessURLs
		}
		return conn.WriteJSON(map[string]interface{}{
			"type": "status",
			"data": event,
		})
	}

	// Send initial status
	if err := sendStatus(ws); err != nil {
		h.logger.Error("Failed to send initial status", zap.Error(err))
		return
	}

	// Start watching for changes
	watchChan, err := h.wsRepo.Watch(ctx, namespace, repository.WithWatchFieldSelector(fmt.Sprintf("metadata.name=%s", name)))
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
				ws, err := h.wsRepo.Get(ctx, namespace, name)
				if err != nil {
					continue
				}
				if err := sendStatus(ws); err != nil {
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
				watchChan, err = h.wsRepo.Watch(ctx, namespace, repository.WithWatchFieldSelector(fmt.Sprintf("metadata.name=%s", name)))
				if err != nil {
					h.logger.Error("Failed to restart workspace watch", zap.Error(err))
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

// CodeAnalysisReport represents the response from code-analyzer service
type CodeAnalysisReport struct {
	Version    string    `json:"version"`
	Type       string    `json:"type"`
	Workspace  string    `json:"workspace"`
	AnalyzedAt time.Time `json:"analyzedAt"`
	Summary    struct {
		Score         int `json:"score"`
		CriticalCount int `json:"criticalCount"`
		HighCount     int `json:"highCount"`
		MediumCount   int `json:"mediumCount"`
		LowCount      int `json:"lowCount"`
	} `json:"summary"`
	Findings []struct {
		Severity       string `json:"severity"`
		Category       string `json:"category"`
		File           string `json:"file"`
		Line           int    `json:"line"`
		Title          string `json:"title"`
		Description    string `json:"description"`
		Recommendation string `json:"recommendation"`
	} `json:"findings"`
}

// CodeAnalysisResponse represents the combined code analysis response
type CodeAnalysisResponse struct {
	Security *CodeAnalysisReport `json:"security"`
	Quality  *CodeAnalysisReport `json:"quality"`
	Status   struct {
		Watching        bool      `json:"watching"`
		InProgress      bool      `json:"inProgress"`
		PendingAnalysis bool      `json:"pendingAnalysis"`
		LastAnalysis    time.Time `json:"lastAnalysis,omitempty"`
	} `json:"status"`
}

// GetCodeAnalysis handles GET /api/v1/namespaces/:namespace/workspaces/:name/code-analysis
// Proxies request to the code-analyzer service running on the WorkMachine
func (h *WorkspaceHandlers) GetCodeAnalysis(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	// Verify the workspace exists
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

	// Only owner can access code analysis
	if !h.requireOwnership(c, workspace) {
		return
	}

	// Build code-analyzer service URL
	// The code-analyzer service runs in the same namespace as the workspace
	codeAnalyzerURL := fmt.Sprintf("http://code-analyzer.%s.svc.cluster.local:8082", namespace)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response := CodeAnalysisResponse{}

	// Fetch security report
	securityURL := fmt.Sprintf("%s/reports/%s/security", codeAnalyzerURL, name)
	securityReq, err := http.NewRequestWithContext(ctx, "GET", securityURL, nil)
	if err == nil {
		securityResp, err := http.DefaultClient.Do(securityReq)
		if err == nil {
			defer securityResp.Body.Close()
			if securityResp.StatusCode == http.StatusOK {
				var report CodeAnalysisReport
				if err := json.NewDecoder(securityResp.Body).Decode(&report); err == nil {
					response.Security = &report
				}
			}
		}
	}

	// Fetch quality report
	qualityURL := fmt.Sprintf("%s/reports/%s/quality", codeAnalyzerURL, name)
	qualityReq, err := http.NewRequestWithContext(ctx, "GET", qualityURL, nil)
	if err == nil {
		qualityResp, err := http.DefaultClient.Do(qualityReq)
		if err == nil {
			defer qualityResp.Body.Close()
			if qualityResp.StatusCode == http.StatusOK {
				var report CodeAnalysisReport
				if err := json.NewDecoder(qualityResp.Body).Decode(&report); err == nil {
					response.Quality = &report
				}
			}
		}
	}

	// Fetch status
	statusURL := fmt.Sprintf("%s/status/%s", codeAnalyzerURL, name)
	statusReq, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err == nil {
		statusResp, err := http.DefaultClient.Do(statusReq)
		if err == nil {
			defer statusResp.Body.Close()
			if statusResp.StatusCode == http.StatusOK {
				json.NewDecoder(statusResp.Body).Decode(&response.Status)
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// TriggerCodeAnalysis handles POST /api/v1/namespaces/:namespace/workspaces/:name/code-analysis
// Triggers a manual code analysis for the workspace
func (h *WorkspaceHandlers) TriggerCodeAnalysis(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = "default"
	}

	// Verify the workspace exists
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

	// Only owner can trigger code analysis
	if !h.requireOwnership(c, workspace) {
		return
	}

	// Build code-analyzer service URL
	// Always use force=true for manual triggers to ensure fresh analysis
	codeAnalyzerURL := fmt.Sprintf("http://code-analyzer.%s.svc.cluster.local:8082", namespace)
	analyzeURL := fmt.Sprintf("%s/analyze/%s?force=true", codeAnalyzerURL, name)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", analyzeURL, nil)
	if err != nil {
		h.logger.Error("Failed to create analyze request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to trigger analysis",
		})
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.logger.Warn("Failed to reach code-analyzer service", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Code analyzer service unavailable",
			"details": "The code analyzer may not be running on this workspace's machine",
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	c.JSON(resp.StatusCode, result)
}
