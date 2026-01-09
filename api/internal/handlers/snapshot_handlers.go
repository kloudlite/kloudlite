package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SnapshotHandlers handles snapshot-related HTTP requests
type SnapshotHandlers struct {
	snapshotRepo    repository.SnapshotRepository
	environmentRepo repository.EnvironmentRepository
	workspaceRepo   repository.WorkspaceRepository
	workmachineRepo repository.WorkMachineRepository
	k8sClient       client.Client
	logger          *zap.Logger
}

// NewSnapshotHandlers creates a new SnapshotHandlers instance
func NewSnapshotHandlers(
	snapshotRepo repository.SnapshotRepository,
	environmentRepo repository.EnvironmentRepository,
	workspaceRepo repository.WorkspaceRepository,
	workmachineRepo repository.WorkMachineRepository,
	k8sClient client.Client,
	logger *zap.Logger,
) *SnapshotHandlers {
	return &SnapshotHandlers{
		snapshotRepo:    snapshotRepo,
		environmentRepo: environmentRepo,
		workspaceRepo:   workspaceRepo,
		workmachineRepo: workmachineRepo,
		k8sClient:       k8sClient,
		logger:          logger,
	}
}

// CreateSnapshotRequest is the request body for creating a snapshot
type CreateSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SnapshotResponse is the API response for a snapshot
type SnapshotResponse struct {
	Name        string                           `json:"name"`
	Description string                           `json:"description,omitempty"`
	State       snapshotv1.SnapshotState         `json:"state"`
	SizeHuman   string                           `json:"sizeHuman,omitempty"`
	SizeBytes   int64                            `json:"sizeBytes,omitempty"`
	CreatedAt   *metav1.Time                     `json:"createdAt,omitempty"`
	Registry    *snapshotv1.SnapshotRegistryInfo `json:"registry,omitempty"`
	Parent      string                           `json:"parent,omitempty"`
	RefCount    int32                            `json:"refCount"`
	Message     string                           `json:"message,omitempty"`
}

// SnapshotRequestResponse is the API response for a snapshot request
type SnapshotRequestResponse struct {
	Name            string                          `json:"name"`
	SnapshotName    string                          `json:"snapshotName"`
	State           snapshotv1.SnapshotRequestState `json:"state"`
	Message         string                          `json:"message,omitempty"`
	StartedAt       *metav1.Time                    `json:"startedAt,omitempty"`
	CompletedAt     *metav1.Time                    `json:"completedAt,omitempty"`
	CreatedSnapshot string                          `json:"createdSnapshot,omitempty"`
}

// getNodeForWorkMachine finds the k8s node for a workmachine by label
func (h *SnapshotHandlers) getNodeForWorkMachine(ctx context.Context, workmachineName string) (string, error) {
	var nodes corev1.NodeList
	if err := h.k8sClient.List(ctx, &nodes, client.MatchingLabels{
		"kloudlite.io/workmachine": workmachineName,
	}); err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no node found for workmachine %s", workmachineName)
	}
	return nodes.Items[0].Name, nil
}

func snapshotToResponse(s *snapshotv1.Snapshot) SnapshotResponse {
	return SnapshotResponse{
		Name:        s.Name,
		Description: s.Spec.Description,
		State:       s.Status.State,
		SizeHuman:   s.Status.SizeHuman,
		SizeBytes:   s.Status.SizeBytes,
		CreatedAt:   s.Status.CreatedAt,
		Registry:    s.Status.Registry,
		Parent:      s.Spec.ParentSnapshot,
		RefCount:    s.Status.RefCount,
		Message:     s.Status.Message,
	}
}

func snapshotRequestToResponse(r *snapshotv1.SnapshotRequest) SnapshotRequestResponse {
	return SnapshotRequestResponse{
		Name:            r.Name,
		SnapshotName:    r.Spec.SnapshotName,
		State:           r.Status.State,
		Message:         r.Status.Message,
		StartedAt:       r.Status.StartedAt,
		CompletedAt:     r.Status.CompletedAt,
		CreatedSnapshot: r.Status.CreatedSnapshot,
	}
}

// CreateEnvironmentSnapshot creates a new snapshot request for an environment
// POST /api/v1/environments/:name/snapshots
func (h *SnapshotHandlers) CreateEnvironmentSnapshot(c *gin.Context) {
	envName := c.Param("name")
	username := c.GetString("username")

	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get environment to verify it exists and get details
	env, err := h.environmentRepo.Get(c.Request.Context(), envName)
	if err != nil {
		h.logger.Error("Failed to get environment", zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("environment %s not found", envName)})
		return
	}

	// Get node name from environment spec
	if env.Spec.NodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment has no node assigned"})
		return
	}

	// Build snapshot name: {envName}-{snapshotName}-{timestamp}
	// If no name provided, use "snapshot" as default
	namepart := req.Name
	if namepart == "" {
		namepart = "snapshot"
	}
	snapshotName := fmt.Sprintf("%s-%s-%d", envName, namepart, time.Now().Unix())
	requestName := fmt.Sprintf("req-%s", snapshotName)

	// Determine parent snapshot from environment's lastRestoredSnapshot
	var parentSnapshot string
	if env.Status.LastRestoredSnapshot != nil {
		parentSnapshot = env.Status.LastRestoredSnapshot.Name
	}

	// Create the SnapshotRequest CR (node-specific operation)
	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":              env.Spec.OwnedBy,
				"snapshots.kloudlite.io/environment": envName,
				"snapshots.kloudlite.io/type":        "environment",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			SnapshotName:   snapshotName,
			SourcePath:     fmt.Sprintf("/data/environments/%s", env.Spec.TargetNamespace),
			NodeName:       env.Spec.NodeName,
			Store:          "default",
			Owner:          env.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), snapshotReq); err != nil {
		h.logger.Error("Failed to create snapshot request", zap.String("name", requestName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot request: %v", err)})
		return
	}

	h.logger.Info("Created environment snapshot request",
		zap.String("request", requestName),
		zap.String("snapshot", snapshotName),
		zap.String("environment", envName),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, snapshotRequestToResponse(snapshotReq))
}

// ListEnvironmentSnapshots lists all snapshots for an environment
// GET /api/v1/environments/:name/snapshots
func (h *SnapshotHandlers) ListEnvironmentSnapshots(c *gin.Context) {
	envName := c.Param("name")

	snapshots, err := h.snapshotRepo.ListByEnvironment(c.Request.Context(), envName)
	if err != nil {
		h.logger.Error("Failed to list snapshots", zap.String("environment", envName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}

	response := make([]SnapshotResponse, len(snapshots.Items))
	for i, s := range snapshots.Items {
		response[i] = snapshotToResponse(&s)
	}

	c.JSON(http.StatusOK, response)
}

// CreateWorkspaceSnapshot creates a new snapshot request for a workspace
// POST /api/v1/namespaces/:namespace/workspaces/:name/snapshots
func (h *SnapshotHandlers) CreateWorkspaceSnapshot(c *gin.Context) {
	namespace := c.Param("namespace")
	workspaceName := c.Param("name")
	username := c.GetString("username")

	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get workspace to verify it exists
	ws, err := h.workspaceRepo.Get(c.Request.Context(), namespace, workspaceName)
	if err != nil {
		h.logger.Error("Failed to get workspace", zap.String("namespace", namespace), zap.String("name", workspaceName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("workspace %s/%s not found", namespace, workspaceName)})
		return
	}

	// Get node name for the workmachine
	nodeName, err := h.getNodeForWorkMachine(c.Request.Context(), ws.Spec.WorkmachineName)
	if err != nil {
		h.logger.Error("Failed to get node for workmachine", zap.String("workmachine", ws.Spec.WorkmachineName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "workmachine is not running"})
		return
	}

	// Build snapshot name
	// If no name provided, use "snapshot" as default
	namepart := req.Name
	if namepart == "" {
		namepart = "snapshot"
	}
	snapshotName := fmt.Sprintf("%s-%s-%s-%d", namespace, workspaceName, namepart, time.Now().Unix())
	requestName := fmt.Sprintf("req-%s", snapshotName)

	// Determine parent snapshot from workspace's lastRestoredSnapshot
	var parentSnapshot string
	if ws.Status.LastRestoredSnapshot != nil {
		parentSnapshot = ws.Status.LastRestoredSnapshot.Name
	}

	// Create the SnapshotRequest CR (node-specific operation)
	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":            ws.Spec.OwnedBy,
				"snapshots.kloudlite.io/workspace": fmt.Sprintf("%s--%s", namespace, workspaceName),
				"snapshots.kloudlite.io/type":      "workspace",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			SnapshotName:   snapshotName,
			SourcePath:     fmt.Sprintf("/data/workspaces/%s/%s", namespace, workspaceName),
			NodeName:       nodeName,
			Store:          "default",
			Owner:          ws.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), snapshotReq); err != nil {
		h.logger.Error("Failed to create snapshot request", zap.String("name", requestName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot request: %v", err)})
		return
	}

	h.logger.Info("Created workspace snapshot request",
		zap.String("request", requestName),
		zap.String("snapshot", snapshotName),
		zap.String("workspace", fmt.Sprintf("%s/%s", namespace, workspaceName)),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, snapshotRequestToResponse(snapshotReq))
}

// ListWorkspaceSnapshots lists all snapshots for a workspace
// GET /api/v1/namespaces/:namespace/workspaces/:name/snapshots
func (h *SnapshotHandlers) ListWorkspaceSnapshots(c *gin.Context) {
	namespace := c.Param("namespace")
	workspaceName := c.Param("name")

	workspaceLabel := fmt.Sprintf("%s--%s", namespace, workspaceName)
	snapshots, err := h.snapshotRepo.ListByWorkspace(c.Request.Context(), workspaceLabel)
	if err != nil {
		h.logger.Error("Failed to list snapshots", zap.String("workspace", workspaceLabel), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}

	response := make([]SnapshotResponse, len(snapshots.Items))
	for i, s := range snapshots.Items {
		response[i] = snapshotToResponse(&s)
	}

	c.JSON(http.StatusOK, response)
}

// ListAllSnapshots lists all snapshots accessible to the user
// GET /api/v1/snapshots
func (h *SnapshotHandlers) ListAllSnapshots(c *gin.Context) {
	username := c.GetString("username")

	snapshots, err := h.snapshotRepo.ListByOwner(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("Failed to list snapshots", zap.String("owner", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}

	response := make([]SnapshotResponse, len(snapshots.Items))
	for i, s := range snapshots.Items {
		response[i] = snapshotToResponse(&s)
	}

	c.JSON(http.StatusOK, response)
}

// GetSnapshot gets a snapshot by name
// GET /api/v1/snapshots/:name
func (h *SnapshotHandlers) GetSnapshot(c *gin.Context) {
	name := c.Param("name")

	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get snapshot", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("snapshot %s not found", name)})
		return
	}

	c.JSON(http.StatusOK, snapshotToResponse(snapshot))
}

// DeleteSnapshot deletes a snapshot
// DELETE /api/v1/snapshots/:name
func (h *SnapshotHandlers) DeleteSnapshot(c *gin.Context) {
	name := c.Param("name")
	username := c.GetString("username")

	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get snapshot", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("snapshot %s not found", name)})
		return
	}

	// Check if snapshot has references
	if snapshot.Status.RefCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("snapshot has %d active references", snapshot.Status.RefCount)})
		return
	}

	if err := h.snapshotRepo.Delete(c.Request.Context(), name); err != nil {
		h.logger.Error("Failed to delete snapshot", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete snapshot"})
		return
	}

	h.logger.Info("Deleted snapshot",
		zap.String("snapshot", name),
		zap.String("user", username),
	)

	c.JSON(http.StatusOK, gin.H{"message": "snapshot deleted"})
}
