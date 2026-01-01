package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SnapshotHandlers handles HTTP requests for Snapshot resources
type SnapshotHandlers struct {
	snapshotRepo repository.SnapshotRepository
	envRepo      repository.EnvironmentRepository
	k8sClient    client.Client
	logger       *zap.Logger
}

// NewSnapshotHandlers creates a new SnapshotHandlers
func NewSnapshotHandlers(
	snapshotRepo repository.SnapshotRepository,
	envRepo repository.EnvironmentRepository,
	k8sClient client.Client,
	logger *zap.Logger,
) *SnapshotHandlers {
	return &SnapshotHandlers{
		snapshotRepo: snapshotRepo,
		envRepo:      envRepo,
		k8sClient:    k8sClient,
		logger:       logger,
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
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": envName,
				"kloudlite.io/owned-by":              username,
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: snapshotv1.EnvironmentReference{
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

	// List snapshots for this environment
	snapshots, err := h.snapshotRepo.ListByEnvironment(c.Request.Context(), envName)
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

	// Determine target environment
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
