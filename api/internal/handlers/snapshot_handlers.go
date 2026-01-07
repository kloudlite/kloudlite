package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SnapshotHandlers handles HTTP requests for Snapshot resources
type SnapshotHandlers struct {
	snapshotRepo    repository.SnapshotRepository
	envRepo         repository.EnvironmentRepository
	workspaceRepo   repository.WorkspaceRepository
	workmachineRepo repository.WorkMachineRepository
	k8sClient       client.Client
	logger          *zap.Logger
}

// NewSnapshotHandlers creates a new SnapshotHandlers
func NewSnapshotHandlers(
	snapshotRepo repository.SnapshotRepository,
	envRepo repository.EnvironmentRepository,
	workspaceRepo repository.WorkspaceRepository,
	workmachineRepo repository.WorkMachineRepository,
	k8sClient client.Client,
	logger *zap.Logger,
) *SnapshotHandlers {
	return &SnapshotHandlers{
		snapshotRepo:    snapshotRepo,
		envRepo:         envRepo,
		workspaceRepo:   workspaceRepo,
		workmachineRepo: workmachineRepo,
		k8sClient:       k8sClient,
		logger:          logger,
	}
}

// CreateSnapshotRequest is the request body for creating a snapshot
type CreateSnapshotRequest struct {
	Description     string `json:"description,omitempty"`
	IncludeMetadata bool   `json:"includeMetadata"`
	KeepForDays     *int32 `json:"keepForDays,omitempty"`
}

// CreateSnapshot handles POST /api/v1/environments/:name/snapshots
func (h *SnapshotHandlers) CreateSnapshot(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment name is required"})
		return
	}

	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body - use defaults
		req = CreateSnapshotRequest{IncludeMetadata: true}
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify environment exists and user has access
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		h.logger.Error("Failed to get environment", zap.Error(err), zap.String("environment", envName))
		c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
		return
	}

	// Check ownership
	if env.Spec.OwnedBy != username {
		h.logger.Warn("User attempting to snapshot environment they don't own",
			zap.String("user", username),
			zap.String("environment", envName),
			zap.String("owner", env.Spec.OwnedBy))
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to snapshot this environment"})
		return
	}

	// Generate snapshot name
	timestamp := time.Now().UTC().Format("20060102-150405")
	snapshotName := fmt.Sprintf("%s-%s", envName, timestamp)

	// Build retention policy
	var retentionPolicy *snapshotv1.RetentionPolicy
	if req.KeepForDays != nil {
		retentionPolicy = &snapshotv1.RetentionPolicy{
			KeepForDays: req.KeepForDays,
		}
	}

	// Create snapshot
	// Note: ParentSnapshotRef is auto-detected by the controller from env.Status.LastRestoredSnapshot
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": envName,
				"kloudlite.io/owned-by":              username,
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: envName,
			},
			Description:     req.Description,
			OwnedBy:         username,
			IncludeMetadata: req.IncludeMetadata,
			RetentionPolicy: retentionPolicy,
		},
	}

	if err := h.snapshotRepo.Create(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to create snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create snapshot",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Snapshot created successfully",
		zap.String("snapshot", snapshotName),
		zap.String("environment", envName))

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Snapshot creation started",
		"snapshot": snapshot,
	})
}

// ListSnapshots handles GET /api/v1/environments/:name/snapshots
func (h *SnapshotHandlers) ListSnapshots(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment name is required"})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify environment exists and user has access
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
		return
	}

	// Check ownership
	if env.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view snapshots for this environment"})
		return
	}

	// If environment was cloned from a snapshot, list full lineage by walking parent refs
	// Otherwise, list snapshots for this specific environment
	var resultSnapshots []snapshotv1.Snapshot

	if env.Status.LastRestoredSnapshot != nil {
		// Build lineage by walking parentSnapshotRef chain
		lineage := h.buildSnapshotLineage(c.Request.Context(), env.Status.LastRestoredSnapshot.Name)
		resultSnapshots = lineage
	} else {
		// Not cloned - list snapshots for this environment only
		snapshots, err := h.snapshotRepo.ListByEnvironment(c.Request.Context(), envName)
		if err != nil {
			h.logger.Error("Failed to list snapshots", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list snapshots"})
			return
		}
		resultSnapshots = snapshots.Items
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": resultSnapshots,
		"count":     len(resultSnapshots),
	})
}

// buildSnapshotLineage builds the direct ancestor chain by walking parentSnapshotRef
// Returns only the direct lineage from root to the given snapshot (no branches)
// Example: If tree is A->B->C->D and C->E->F, starting from F returns [A,B,C,E,F]
func (h *SnapshotHandlers) buildSnapshotLineage(ctx context.Context, startSnapshotName string) []snapshotv1.Snapshot {
	var ancestors []snapshotv1.Snapshot

	// Walk UP the parent chain collecting all ancestors
	currentName := startSnapshotName
	for currentName != "" {
		snapshot, err := h.snapshotRepo.Get(ctx, currentName)
		if err != nil {
			break
		}
		// Prepend to maintain order from root to leaf
		ancestors = append([]snapshotv1.Snapshot{*snapshot}, ancestors...)

		if snapshot.Spec.ParentSnapshotRef == nil {
			break // Reached root
		}
		currentName = snapshot.Spec.ParentSnapshotRef.Name
	}

	return ancestors
}

// GetSnapshot handles GET /api/v1/snapshots/:name
func (h *SnapshotHandlers) GetSnapshot(c *gin.Context) {
	snapshotName := c.Param("name")
	if snapshotName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot name is required"})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get snapshot
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), snapshotName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Snapshot not found"})
		return
	}

	// Check ownership
	if snapshot.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view this snapshot"})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}

// RestoreSnapshotRequest is the request body for restoring a snapshot
type RestoreSnapshotRequest struct {
	TargetEnvironment string `json:"targetEnvironment,omitempty"` // If empty, restore to original environment
	TargetWorkspace   string `json:"targetWorkspace,omitempty"`   // If empty, restore to original workspace
}

// RestoreSnapshot handles POST /api/v1/snapshots/:name/restore
func (h *SnapshotHandlers) RestoreSnapshot(c *gin.Context) {
	snapshotName := c.Param("name")
	if snapshotName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot name is required"})
		return
	}

	var req RestoreSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = RestoreSnapshotRequest{}
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get snapshot
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), snapshotName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Snapshot not found"})
		return
	}

	// Check ownership
	if snapshot.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to restore this snapshot"})
		return
	}

	// Verify snapshot is ready
	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Snapshot is not ready for restore",
			"state":   snapshot.Status.State,
			"message": snapshot.Status.Message,
		})
		return
	}

	// Handle workspace snapshots
	if snapshot.Spec.WorkspaceRef != nil {
		targetWorkspaceName := req.TargetWorkspace
		if targetWorkspaceName == "" {
			targetWorkspaceName = snapshot.Spec.WorkspaceRef.Name
		}

		// Get the workspace namespace from the workmachine name
		wmNamespace := fmt.Sprintf("wm-%s", snapshot.Spec.OwnedBy)

		// Verify target workspace exists and user has access
		targetWorkspace, err := h.workspaceRepo.Get(c.Request.Context(), wmNamespace, targetWorkspaceName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Target workspace not found"})
			return
		}

		if targetWorkspace.Spec.OwnedBy != username {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to restore to this workspace"})
			return
		}

		// Update snapshot status to trigger restore
		snapshot.Status.State = snapshotv1.SnapshotStateRestoring
		snapshot.Status.Message = fmt.Sprintf("Restoring to workspace %s", targetWorkspaceName)

		if err := h.k8sClient.Status().Update(c.Request.Context(), snapshot); err != nil {
			h.logger.Error("Failed to update snapshot status for restore", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate restore"})
			return
		}

		h.logger.Info("Workspace snapshot restore initiated",
			zap.String("snapshot", snapshotName),
			zap.String("targetWorkspace", targetWorkspaceName))

		c.JSON(http.StatusOK, gin.H{
			"message":         "Restore initiated",
			"snapshot":        snapshotName,
			"targetWorkspace": targetWorkspaceName,
		})
		return
	}

	// Handle environment snapshots
	if snapshot.Spec.EnvironmentRef == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot has no environment or workspace reference"})
		return
	}

	targetEnvName := req.TargetEnvironment
	if targetEnvName == "" {
		targetEnvName = snapshot.Spec.EnvironmentRef.Name
	}

	// Verify target environment exists and user has access
	targetEnv, err := h.envRepo.Get(c.Request.Context(), targetEnvName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target environment not found"})
		return
	}

	if targetEnv.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to restore to this environment"})
		return
	}

	// Update snapshot status to trigger restore
	snapshot.Status.State = snapshotv1.SnapshotStateRestoring
	snapshot.Status.Message = fmt.Sprintf("Restoring to environment %s", targetEnvName)

	if err := h.k8sClient.Status().Update(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to update snapshot status for restore", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate restore"})
		return
	}

	h.logger.Info("Snapshot restore initiated",
		zap.String("snapshot", snapshotName),
		zap.String("targetEnvironment", targetEnvName))

	c.JSON(http.StatusOK, gin.H{
		"message":           "Restore initiated",
		"snapshot":          snapshotName,
		"targetEnvironment": targetEnvName,
	})
}

// DeleteSnapshot handles DELETE /api/v1/snapshots/:name
func (h *SnapshotHandlers) DeleteSnapshot(c *gin.Context) {
	snapshotName := c.Param("name")
	if snapshotName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot name is required"})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get snapshot
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), snapshotName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Snapshot not found"})
		return
	}

	// Check ownership
	if snapshot.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this snapshot"})
		return
	}

	// Delete snapshot
	if err := h.snapshotRepo.Delete(c.Request.Context(), snapshotName); err != nil {
		h.logger.Error("Failed to delete snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete snapshot"})
		return
	}

	h.logger.Info("Snapshot deleted successfully", zap.String("snapshot", snapshotName))

	c.JSON(http.StatusOK, gin.H{
		"message":  "Snapshot deletion initiated",
		"snapshot": snapshotName,
	})
}

// ListAllSnapshots handles GET /api/v1/snapshots (lists all snapshots for the authenticated user)
func (h *SnapshotHandlers) ListAllSnapshots(c *gin.Context) {
	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// List snapshots owned by user
	snapshots, err := h.snapshotRepo.ListByOwner(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to list snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots.Items,
		"count":     len(snapshots.Items),
	})
}

// CreateWorkspaceSnapshotRequest is the request body for creating a workspace snapshot
type CreateWorkspaceSnapshotRequest struct {
	Description     string `json:"description,omitempty"`
	IncludeMetadata bool   `json:"includeMetadata"`
	KeepForDays     *int32 `json:"keepForDays,omitempty"`
}

// CreateWorkspaceSnapshot handles POST /api/v1/namespaces/:namespace/workspaces/:name/snapshots
func (h *SnapshotHandlers) CreateWorkspaceSnapshot(c *gin.Context) {
	namespace := c.Param("namespace")
	workspaceName := c.Param("name")
	if namespace == "" || workspaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace and workspace name are required"})
		return
	}

	var req CreateWorkspaceSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body - use defaults
		req = CreateWorkspaceSnapshotRequest{IncludeMetadata: true}
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Use namespace from route parameter
	wmNamespace := namespace

	// Verify workspace exists and user has access
	workspace, err := h.workspaceRepo.Get(c.Request.Context(), wmNamespace, workspaceName)
	if err != nil {
		h.logger.Error("Failed to get workspace", zap.Error(err), zap.String("workspace", workspaceName))
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}

	// Check ownership
	if workspace.Spec.OwnedBy != username {
		h.logger.Warn("User attempting to snapshot workspace they don't own",
			zap.String("user", username),
			zap.String("workspace", workspaceName),
			zap.String("owner", workspace.Spec.OwnedBy))
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to snapshot this workspace"})
		return
	}

	// Generate snapshot name
	timestamp := time.Now().UTC().Format("20060102-150405")
	snapshotName := fmt.Sprintf("ws-%s-%s", workspaceName, timestamp)

	// Build retention policy
	var retentionPolicy *snapshotv1.RetentionPolicy
	if req.KeepForDays != nil {
		retentionPolicy = &snapshotv1.RetentionPolicy{
			KeepForDays: req.KeepForDays,
		}
	}

	// Create snapshot
	// Note: ParentSnapshotRef is auto-detected by the controller from workspace.Status.LastRestoredSnapshot
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
			Labels: map[string]string{
				"snapshots.kloudlite.io/workspace": workspaceName,
				"kloudlite.io/owned-by":            username,
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			WorkspaceRef: &snapshotv1.WorkspaceReference{
				Name:            workspaceName,
				WorkmachineName: workspace.Spec.WorkmachineName,
			},
			Description:     req.Description,
			OwnedBy:         username,
			IncludeMetadata: req.IncludeMetadata,
			RetentionPolicy: retentionPolicy,
		},
	}

	if err := h.snapshotRepo.Create(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to create workspace snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create snapshot",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Workspace snapshot created successfully",
		zap.String("snapshot", snapshotName),
		zap.String("workspace", workspaceName))

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Workspace snapshot creation started",
		"snapshot": snapshot,
	})
}

// ListWorkspaceSnapshots handles GET /api/v1/namespaces/:namespace/workspaces/:name/snapshots
func (h *SnapshotHandlers) ListWorkspaceSnapshots(c *gin.Context) {
	namespace := c.Param("namespace")
	workspaceName := c.Param("name")
	if namespace == "" || workspaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace and workspace name are required"})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Use namespace from route parameter
	wmNamespace := namespace

	// Verify workspace exists and user has access
	workspace, err := h.workspaceRepo.Get(c.Request.Context(), wmNamespace, workspaceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}

	// Check ownership
	if workspace.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view snapshots for this workspace"})
		return
	}

	// List snapshots for this workspace
	snapshots, err := h.snapshotRepo.ListByWorkspace(c.Request.Context(), workspaceName)
	if err != nil {
		h.logger.Error("Failed to list workspace snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots.Items,
		"count":     len(snapshots.Items),
	})
}

// PushSnapshotRequest is the request body for pushing a snapshot to registry
type PushSnapshotRequest struct {
	Repository string `json:"repository,omitempty"` // Optional: defaults to snapshots/{username}/ws/{ws} or /env/{env}
	Tag        string `json:"tag"`                  // Required: tag for the snapshot image
}

// PushSnapshot handles POST /api/v1/snapshots/:name/push
func (h *SnapshotHandlers) PushSnapshot(c *gin.Context) {
	snapshotName := c.Param("name")
	if snapshotName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot name is required"})
		return
	}

	var req PushSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Tag is required - snapshots are immutable
	if req.Tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag is required for push"})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get snapshot
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), snapshotName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Snapshot not found"})
		return
	}

	// Check ownership
	if snapshot.Spec.OwnedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to push this snapshot"})
		return
	}

	// Verify snapshot is ready
	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Snapshot is not ready for push",
			"state":   snapshot.Status.State,
			"message": snapshot.Status.Message,
		})
		return
	}

	// Check if already pushed - snapshots are immutable
	if snapshot.Status.RegistryStatus != nil && snapshot.Status.RegistryStatus.Pushed {
		c.JSON(http.StatusConflict, gin.H{
			"error":    "Snapshot already pushed",
			"snapshot": snapshotName,
			"imageRef": snapshot.Status.RegistryStatus.ImageRef,
		})
		return
	}

	// Set up registry reference - repository defaults based on snapshot type
	repository := req.Repository
	// Note: if empty, controller will set based on snapshot type (ws/env)
	tag := req.Tag // Already validated as required

	// Update snapshot spec to set registry reference
	snapshot.Spec.RegistryRef = &snapshotv1.SnapshotRegistryRef{
		Repository: repository,
		Tag:        tag,
	}

	if err := h.k8sClient.Update(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to update snapshot spec for push", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate push"})
		return
	}

	// Re-fetch snapshot to get updated resourceVersion before status update
	if err := h.k8sClient.Get(c.Request.Context(), client.ObjectKeyFromObject(snapshot), snapshot); err != nil {
		h.logger.Error("Failed to re-fetch snapshot after spec update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate push"})
		return
	}

	// Now update status to trigger the push
	snapshot.Status.State = snapshotv1.SnapshotStatePushing
	snapshot.Status.Message = "Pushing snapshot to registry..."

	if err := h.k8sClient.Status().Update(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to update snapshot status for push", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate push"})
		return
	}

	h.logger.Info("Snapshot push initiated",
		zap.String("snapshot", snapshotName),
		zap.String("tag", tag))

	c.JSON(http.StatusOK, gin.H{
		"message":  "Snapshot push initiated",
		"snapshot": snapshotName,
		"tag":      tag,
	})
}

// PullSnapshotRequest is the request body for pulling a snapshot from registry
type PullSnapshotRequest struct {
	Repository string `json:"repository"` // Required: repository path
	Tag        string `json:"tag"`        // Required: image tag
	Name       string `json:"name"`       // Optional: name for the new snapshot
}

// PullSnapshot handles POST /api/v1/snapshots/pull
func (h *SnapshotHandlers) PullSnapshot(c *gin.Context) {
	var req PullSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Repository == "" || req.Tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository and tag are required"})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Generate snapshot name if not provided
	snapshotName := req.Name
	if snapshotName == "" {
		timestamp := time.Now().UTC().Format("20060102-150405")
		snapshotName = fmt.Sprintf("clone-%s", timestamp)
	}

	// Create snapshot with pull state
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
			Labels: map[string]string{
				"kloudlite.io/owned-by":         username,
				"snapshots.kloudlite.io/cloned": "true",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			OwnedBy:         username,
			IncludeMetadata: true,
			RegistryRef: &snapshotv1.SnapshotRegistryRef{
				Repository: req.Repository,
				Tag:        req.Tag,
			},
		},
		Status: snapshotv1.SnapshotStatus{
			State:   snapshotv1.SnapshotStatePulling,
			Message: "Cloning snapshot from cloud...",
		},
	}

	if err := h.snapshotRepo.Create(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to create snapshot for clone", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create snapshot",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Snapshot clone from cloud initiated",
		zap.String("snapshot", snapshotName),
		zap.String("repository", req.Repository),
		zap.String("tag", req.Tag))

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Snapshot clone from cloud initiated",
		"snapshot": snapshot,
	})
}

// CreateWorkspaceFromSnapshotRequest is the request body for creating a workspace from snapshot
type CreateWorkspaceFromSnapshotRequest struct {
	Name          string `json:"name" binding:"required"`
	DisplayName   string `json:"displayName" binding:"required"`
	SnapshotName  string `json:"snapshotName" binding:"required"`
	Description   string `json:"description,omitempty"`
	VSCodeVersion string `json:"vscodeVersion,omitempty"`
}

// CreateWorkspaceFromSnapshot handles POST /api/v1/workspaces/from-snapshot
// This creates a new workspace from a pushed snapshot
func (h *SnapshotHandlers) CreateWorkspaceFromSnapshot(c *gin.Context) {
	var req CreateWorkspaceFromSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate snapshot exists and is pushed
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), req.SnapshotName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Snapshot '%s' not found", req.SnapshotName)})
		return
	}

	if snapshot.Status.RegistryStatus == nil || !snapshot.Status.RegistryStatus.Pushed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot is not pushed to registry. Only pushed snapshots can be used to create workspaces."})
		return
	}

	if snapshot.Status.SnapshotType != snapshotv1.SnapshotTypeWorkspace {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot is not a workspace snapshot"})
		return
	}

	// Get user's WorkMachine for namespace
	workMachine, err := h.workmachineRepo.GetByOwner(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get user's WorkMachine", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "You don't have a WorkMachine. Please create one first."})
		return
	}

	vscodeVersion := req.VSCodeVersion
	if vscodeVersion == "" {
		vscodeVersion = "latest"
	}

	// Create new workspace with FromSnapshot set
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: workMachine.Spec.TargetNamespace,
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName:     req.DisplayName,
			Description:     req.Description,
			OwnedBy:         username,
			WorkmachineName: workMachine.Name,
			VSCodeVersion:   vscodeVersion,
			FromSnapshot: &workspacesv1.FromSnapshotRef{
				SnapshotName: req.SnapshotName,
			},
		},
	}

	if err := h.workspaceRepo.Create(c.Request.Context(), workspace); err != nil {
		h.logger.Error("Failed to create workspace from snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create workspace",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Workspace created from snapshot",
		zap.String("workspace", req.Name),
		zap.String("snapshot", req.SnapshotName),
		zap.String("owner", username))

	c.JSON(http.StatusCreated, workspace)
}

// CreateEnvironmentFromSnapshotRequest is the request body for creating an environment from snapshot
type CreateEnvironmentFromSnapshotRequest struct {
	Name            string `json:"name" binding:"required"`
	SnapshotName    string `json:"snapshotName" binding:"required"`
	TargetNamespace string `json:"targetNamespace,omitempty"` // Optional, generated if not provided
	Activated       bool   `json:"activated"`                 // Whether to start active
}

// CreateEnvironmentFromSnapshot handles POST /api/v1/environments/from-snapshot
// This creates a new environment from a pushed snapshot
func (h *SnapshotHandlers) CreateEnvironmentFromSnapshot(c *gin.Context) {
	var req CreateEnvironmentFromSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get authenticated user
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate snapshot exists and is pushed
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), req.SnapshotName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Snapshot '%s' not found", req.SnapshotName)})
		return
	}

	if snapshot.Status.RegistryStatus == nil || !snapshot.Status.RegistryStatus.Pushed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot is not pushed to registry. Only pushed snapshots can be used to create environments."})
		return
	}

	if snapshot.Status.SnapshotType != snapshotv1.SnapshotTypeEnvironment {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Snapshot is not an environment snapshot"})
		return
	}

	// Get user's WorkMachine
	workMachine, err := h.workmachineRepo.GetByOwner(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to get user's WorkMachine", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "You don't have a WorkMachine. Please create one first."})
		return
	}

	// Generate target namespace if not provided
	targetNamespace := req.TargetNamespace
	if targetNamespace == "" {
		targetNamespace = fmt.Sprintf("env-%s", req.Name)
	}

	// Create new environment with FromSnapshot set
	environment := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
		Spec: environmentsv1.EnvironmentSpec{
			Name:            req.Name,
			TargetNamespace: targetNamespace,
			OwnedBy:         username,
			WorkMachineName: workMachine.Name,
			Activated:       req.Activated,
			FromSnapshot: &environmentsv1.FromSnapshotRef{
				SnapshotName: req.SnapshotName,
			},
		},
	}

	if err := h.envRepo.Create(c.Request.Context(), environment); err != nil {
		h.logger.Error("Failed to create environment from snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create environment",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment created from snapshot",
		zap.String("environment", req.Name),
		zap.String("snapshot", req.SnapshotName),
		zap.String("owner", username))

	c.JSON(http.StatusCreated, environment)
}

// ListPushedSnapshots handles GET /api/v1/snapshots/pushed
// Returns all pushed snapshots available for cloning (team-wide access)
func (h *SnapshotHandlers) ListPushedSnapshots(c *gin.Context) {
	// Get authenticated user
	_, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Filter by type if provided
	snapshotType := c.Query("type") // "workspace" or "environment"

	// List all snapshots (team-wide access)
	snapshots, err := h.snapshotRepo.List(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list snapshots"})
		return
	}

	// Filter to only pushed snapshots
	var pushedSnapshots []snapshotv1.Snapshot
	for _, snap := range snapshots.Items {
		// Only include pushed snapshots
		if snap.Status.RegistryStatus == nil || !snap.Status.RegistryStatus.Pushed {
			continue
		}

		// Filter by type if specified
		if snapshotType != "" {
			if snapshotType == "workspace" && snap.Status.SnapshotType != snapshotv1.SnapshotTypeWorkspace {
				continue
			}
			if snapshotType == "environment" && snap.Status.SnapshotType != snapshotv1.SnapshotTypeEnvironment {
				continue
			}
		}

		pushedSnapshots = append(pushedSnapshots, snap)
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": pushedSnapshots,
		"count":     len(pushedSnapshots),
	})
}
