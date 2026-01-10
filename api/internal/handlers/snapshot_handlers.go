package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	envv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

	// Check for existing in-progress snapshot requests for this environment
	existingRequests := &snapshotv1.SnapshotRequestList{}
	if err := h.k8sClient.List(c.Request.Context(), existingRequests,
		client.InNamespace(env.Spec.TargetNamespace),
		client.MatchingLabels{"snapshots.kloudlite.io/environment": envName}); err != nil {
		h.logger.Error("Failed to list existing snapshot requests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing requests"})
		return
	}

	// Check if any request is in progress
	for _, existingReq := range existingRequests.Items {
		state := existingReq.Status.State
		if state == "" || state == snapshotv1.SnapshotRequestStatePending ||
			state == snapshotv1.SnapshotRequestStateCreating ||
			state == snapshotv1.SnapshotRequestStateUploading {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "A snapshot request is already in progress for this environment",
				"request": existingReq.Name,
				"state":   string(state),
			})
			return
		}
	}

	// Get node name from the workmachine
	if env.Spec.WorkMachineName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment has no workmachine assigned"})
		return
	}
	nodeName, err := h.getNodeForWorkMachine(c.Request.Context(), env.Spec.WorkMachineName)
	if err != nil {
		h.logger.Error("Failed to get node for workmachine", zap.String("workmachine", env.Spec.WorkMachineName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "workmachine is not running"})
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

	labels := map[string]string{
		"kloudlite.io/owned-by":              env.Spec.OwnedBy,
		"snapshots.kloudlite.io/environment": envName,
		"snapshots.kloudlite.io/type":        "environment",
	}

	// Create the Snapshot object first with Pending state (so UI can see it immediately)
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:   snapshotName,
			Labels: labels,
		},
		Spec: snapshotv1.SnapshotSpec{
			Owner:          env.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to create snapshot", zap.String("name", snapshotName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot: %v", err)})
		return
	}

	// Capture K8s resources and create SnapshotArtifacts CR
	if err := h.createSnapshotArtifacts(c.Request.Context(), snapshotName, env.Spec.TargetNamespace); err != nil {
		h.logger.Warn("Failed to create snapshot artifacts", zap.Error(err))
		// Continue without artifacts - not a fatal error
	}

	// Create the SnapshotRequest CR (node-specific operation)
	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: env.Spec.TargetNamespace,
			Labels:    labels,
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			SnapshotName:   snapshotName,
			SourcePath:     fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", env.Spec.TargetNamespace),
			NodeName:       nodeName,
			Store:          "default",
			Owner:          env.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), snapshotReq); err != nil {
		h.logger.Error("Failed to create snapshot request", zap.String("name", requestName), zap.Error(err))
		// Clean up the snapshot we just created
		_ = h.k8sClient.Delete(c.Request.Context(), snapshot)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot request: %v", err)})
		return
	}

	h.logger.Info("Created environment snapshot request",
		zap.String("request", requestName),
		zap.String("snapshot", snapshotName),
		zap.String("environment", envName),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Snapshot creation started",
		"snapshot": snapshotToResponse(snapshot),
	})
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

	// Get environment to find the target namespace for SnapshotRequests
	env, err := h.environmentRepo.Get(c.Request.Context(), envName)
	if err != nil {
		h.logger.Error("Failed to get environment", zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get environment"})
		return
	}

	// List SnapshotRequests in the environment's namespace to get progress for pending snapshots
	snapshotRequests := &snapshotv1.SnapshotRequestList{}
	if err := h.k8sClient.List(c.Request.Context(), snapshotRequests, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		h.logger.Warn("Failed to list snapshot requests", zap.Error(err))
		// Continue without request info
	}

	// Build a map of snapshot name -> request for quick lookup
	requestBySnapshot := make(map[string]*snapshotv1.SnapshotRequest)
	for i := range snapshotRequests.Items {
		req := &snapshotRequests.Items[i]
		requestBySnapshot[req.Spec.SnapshotName] = req
	}

	response := make([]SnapshotResponse, len(snapshots.Items))
	for i, s := range snapshots.Items {
		resp := snapshotToResponse(&s)

		// If snapshot has no state, check the corresponding SnapshotRequest for progress
		if resp.State == "" {
			if req, ok := requestBySnapshot[s.Name]; ok {
				// Map SnapshotRequest state to a display state
				switch req.Status.State {
				case snapshotv1.SnapshotRequestStatePending, "":
					resp.State = snapshotv1.SnapshotState("Pending")
					resp.Message = "Waiting to start"
				case snapshotv1.SnapshotRequestStateCreating:
					resp.State = snapshotv1.SnapshotState("Creating")
					resp.Message = req.Status.Message
				case snapshotv1.SnapshotRequestStateUploading:
					resp.State = snapshotv1.SnapshotState("Uploading")
					resp.Message = req.Status.Message
				case snapshotv1.SnapshotRequestStateFailed:
					resp.State = snapshotv1.SnapshotStateFailed
					resp.Message = req.Status.Message
				}
			} else {
				resp.State = snapshotv1.SnapshotState("Pending")
				resp.Message = "Waiting for snapshot request"
			}
		}

		response[i] = resp
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": response,
		"count":     len(response),
	})
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

	workspaceLabel := fmt.Sprintf("%s--%s", namespace, workspaceName)

	// Check for existing in-progress snapshot requests for this workspace
	existingRequests := &snapshotv1.SnapshotRequestList{}
	if err := h.k8sClient.List(c.Request.Context(), existingRequests,
		client.InNamespace(namespace),
		client.MatchingLabels{"snapshots.kloudlite.io/workspace": workspaceLabel}); err != nil {
		h.logger.Error("Failed to list existing snapshot requests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing requests"})
		return
	}

	// Check if any request is in progress
	for _, existingReq := range existingRequests.Items {
		state := existingReq.Status.State
		if state == "" || state == snapshotv1.SnapshotRequestStatePending ||
			state == snapshotv1.SnapshotRequestStateCreating ||
			state == snapshotv1.SnapshotRequestStateUploading {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "A snapshot request is already in progress for this workspace",
				"request": existingReq.Name,
				"state":   string(state),
			})
			return
		}
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

	labels := map[string]string{
		"kloudlite.io/owned-by":            ws.Spec.OwnedBy,
		"snapshots.kloudlite.io/workspace": workspaceLabel,
		"snapshots.kloudlite.io/type":      "workspace",
	}

	// Create the Snapshot object first with Pending state (so UI can see it immediately)
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:   snapshotName,
			Labels: labels,
		},
		Spec: snapshotv1.SnapshotSpec{
			Owner:          ws.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to create snapshot", zap.String("name", snapshotName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot: %v", err)})
		return
	}

	// Create the SnapshotRequest CR (node-specific operation)
	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			SnapshotName:   snapshotName,
			SourcePath:     fmt.Sprintf("/var/lib/kloudlite/storage/workspaces/%s/%s", namespace, workspaceName),
			NodeName:       nodeName,
			Store:          "default",
			Owner:          ws.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), snapshotReq); err != nil {
		h.logger.Error("Failed to create snapshot request", zap.String("name", requestName), zap.Error(err))
		// Clean up the snapshot we just created
		_ = h.k8sClient.Delete(c.Request.Context(), snapshot)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot request: %v", err)})
		return
	}

	h.logger.Info("Created workspace snapshot request",
		zap.String("request", requestName),
		zap.String("snapshot", snapshotName),
		zap.String("workspace", fmt.Sprintf("%s/%s", namespace, workspaceName)),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Snapshot creation started",
		"snapshot": snapshotToResponse(snapshot),
	})
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

	// List SnapshotRequests in the workspace's namespace to get progress for pending snapshots
	snapshotRequests := &snapshotv1.SnapshotRequestList{}
	if err := h.k8sClient.List(c.Request.Context(), snapshotRequests, client.InNamespace(namespace)); err != nil {
		h.logger.Warn("Failed to list snapshot requests", zap.Error(err))
		// Continue without request info
	}

	// Build a map of snapshot name -> request for quick lookup
	requestBySnapshot := make(map[string]*snapshotv1.SnapshotRequest)
	for i := range snapshotRequests.Items {
		req := &snapshotRequests.Items[i]
		requestBySnapshot[req.Spec.SnapshotName] = req
	}

	response := make([]SnapshotResponse, len(snapshots.Items))
	for i, s := range snapshots.Items {
		resp := snapshotToResponse(&s)

		// If snapshot has no state, check the corresponding SnapshotRequest for progress
		if resp.State == "" {
			if req, ok := requestBySnapshot[s.Name]; ok {
				// Map SnapshotRequest state to a display state
				switch req.Status.State {
				case snapshotv1.SnapshotRequestStatePending, "":
					resp.State = snapshotv1.SnapshotState("Pending")
					resp.Message = "Waiting to start"
				case snapshotv1.SnapshotRequestStateCreating:
					resp.State = snapshotv1.SnapshotState("Creating")
					resp.Message = req.Status.Message
				case snapshotv1.SnapshotRequestStateUploading:
					resp.State = snapshotv1.SnapshotState("Uploading")
					resp.Message = req.Status.Message
				case snapshotv1.SnapshotRequestStateFailed:
					resp.State = snapshotv1.SnapshotStateFailed
					resp.Message = req.Status.Message
				}
			} else {
				resp.State = snapshotv1.SnapshotState("Pending")
				resp.Message = "Waiting for snapshot request"
			}
		}

		response[i] = resp
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": response,
		"count":     len(response),
	})
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

	c.JSON(http.StatusOK, gin.H{
		"snapshots": response,
		"count":     len(response),
	})
}

// ListReadySnapshots lists all ready snapshots available for forking
// GET /api/v1/snapshots/ready?type=environment&environment=env-name
func (h *SnapshotHandlers) ListReadySnapshots(c *gin.Context) {
	snapshotType := c.Query("type")       // "environment" or "workspace"
	environment := c.Query("environment") // filter by specific environment

	var snapshots *snapshotv1.SnapshotList
	var err error

	// Use environment-specific query if environment is specified
	if environment != "" {
		snapshots, err = h.snapshotRepo.ListByEnvironment(c.Request.Context(), environment)
	} else {
		// Fall back to listing by owner
		username := c.GetString("username")
		snapshots, err = h.snapshotRepo.ListByOwner(c.Request.Context(), username)
	}

	if err != nil {
		h.logger.Error("Failed to list snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}

	// Filter for Ready snapshots
	var filtered []SnapshotResponse
	for _, s := range snapshots.Items {
		// Only include Ready snapshots
		if s.Status.State != snapshotv1.SnapshotStateReady {
			continue
		}

		// Filter by type if specified
		if snapshotType != "" {
			typeLabel := s.Labels["snapshots.kloudlite.io/type"]
			if typeLabel != snapshotType {
				continue
			}
		}

		filtered = append(filtered, snapshotToResponse(&s))
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": filtered,
		"count":     len(filtered),
	})
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

// DeleteSnapshot decrements the refCount of a snapshot
// When refCount reaches 0, the snapshot will be automatically garbage collected
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

	// Decrement refCount (minimum 0)
	newRefCount := snapshot.Status.RefCount - 1
	if newRefCount < 0 {
		newRefCount = 0
	}

	snapshot.Status.RefCount = newRefCount
	if err := h.k8sClient.Status().Update(c.Request.Context(), snapshot); err != nil {
		h.logger.Error("Failed to update snapshot refCount", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update snapshot"})
		return
	}

	h.logger.Info("Decremented snapshot refCount",
		zap.String("snapshot", name),
		zap.Int32("newRefCount", newRefCount),
		zap.String("user", username),
	)

	message := "snapshot reference removed"
	if newRefCount == 0 {
		message = "snapshot marked for deletion (will be garbage collected)"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  message,
		"refCount": newRefCount,
	})
}

// CreateEnvironmentFromSnapshotRequest is the request body for creating an environment from a snapshot
type CreateEnvironmentFromSnapshotRequest struct {
	Name         string `json:"name" binding:"required"`
	SnapshotName string `json:"snapshotName" binding:"required"`
	Activated    bool   `json:"activated"`
}

// CreateEnvironmentFromSnapshot creates a new environment from a snapshot
// POST /api/v1/environments/from-snapshot
func (h *SnapshotHandlers) CreateEnvironmentFromSnapshot(c *gin.Context) {
	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CreateEnvironmentFromSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the snapshot to verify it exists and is ready
	snapshot, err := h.snapshotRepo.Get(c.Request.Context(), req.SnapshotName)
	if err != nil {
		h.logger.Error("Failed to get snapshot", zap.String("snapshot", req.SnapshotName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("snapshot %s not found", req.SnapshotName)})
		return
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("snapshot %s is not ready (state: %s)", req.SnapshotName, snapshot.Status.State)})
		return
	}

	// Get source environment from snapshot labels to find the workmachine
	sourceEnvName := snapshot.Labels["snapshots.kloudlite.io/environment"]
	if sourceEnvName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "snapshot does not have environment label"})
		return
	}

	// Get source environment to find the workmachine
	sourceEnv, err := h.environmentRepo.Get(c.Request.Context(), sourceEnvName)
	if err != nil {
		h.logger.Error("Failed to get source environment", zap.String("environment", sourceEnvName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("source environment %s not found", sourceEnvName)})
		return
	}

	// Create new environment name: {username}--{name}
	envName := fmt.Sprintf("%s--%s", username, req.Name)
	targetNamespace := envName

	// Create new environment
	newEnv := &envv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: envName,
			Labels: map[string]string{
				"kloudlite.io/owned-by": username,
			},
		},
		Spec: envv1.EnvironmentSpec{
			TargetNamespace: targetNamespace,
			Name:            req.Name,
			OwnedBy:         username,
			WorkMachineName: sourceEnv.Spec.WorkMachineName,
			Activated:       req.Activated,
			FromSnapshot: &envv1.FromSnapshotRef{
				SnapshotName: req.SnapshotName,
			},
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), newEnv); err != nil {
		h.logger.Error("Failed to create environment", zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create environment: %v", err)})
		return
	}

	// Increment snapshot refCount
	snapshot.Status.RefCount++
	if err := h.k8sClient.Status().Update(c.Request.Context(), snapshot); err != nil {
		h.logger.Warn("Failed to increment snapshot refCount", zap.String("snapshot", req.SnapshotName), zap.Error(err))
	}

	h.logger.Info("Created environment from snapshot",
		zap.String("environment", envName),
		zap.String("snapshot", req.SnapshotName),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "environment created from snapshot",
		"environment": envName,
		"snapshot":    req.SnapshotName,
	})
}

// createSnapshotArtifacts captures K8s resources and creates a SnapshotArtifacts CR
func (h *SnapshotHandlers) createSnapshotArtifacts(ctx context.Context, snapshotName, namespace string) error {
	artifacts := &snapshotv1.SnapshotArtifacts{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName, // Same name as the snapshot
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot": snapshotName,
			},
		},
		Spec: snapshotv1.SnapshotArtifactsSpec{
			SnapshotName: snapshotName,
		},
	}

	var compositionCount, configMapCount, secretCount int32

	// Capture Compositions
	compositions := &envv1.CompositionList{}
	if err := h.k8sClient.List(ctx, compositions, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list compositions: %w", err)
	}

	if len(compositions.Items) > 0 {
		cleanCompositions := make([]envv1.Composition, len(compositions.Items))
		for i, comp := range compositions.Items {
			cleanCompositions[i] = envv1.Composition{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "environments.kloudlite.io/v1",
					Kind:       "Composition",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        comp.Name,
					Labels:      comp.Labels,
					Annotations: comp.Annotations,
				},
				Spec: comp.Spec,
			}
		}

		data, err := yaml.Marshal(cleanCompositions)
		if err != nil {
			return fmt.Errorf("failed to marshal compositions: %w", err)
		}
		artifacts.Spec.Compositions = base64.StdEncoding.EncodeToString(data)
		compositionCount = int32(len(compositions.Items))
		h.logger.Info("Captured compositions", zap.Int("count", len(compositions.Items)))
	}

	// Capture ConfigMaps (excluding system ones)
	configMaps := &corev1.ConfigMapList{}
	if err := h.k8sClient.List(ctx, configMaps, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list configmaps: %w", err)
	}

	var userConfigMaps []corev1.ConfigMap
	for _, cm := range configMaps.Items {
		if cm.Name == "kube-root-ca.crt" {
			continue
		}
		userConfigMaps = append(userConfigMaps, corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        cm.Name,
				Labels:      cm.Labels,
				Annotations: cm.Annotations,
			},
			Data:       cm.Data,
			BinaryData: cm.BinaryData,
		})
	}

	if len(userConfigMaps) > 0 {
		data, err := yaml.Marshal(userConfigMaps)
		if err != nil {
			return fmt.Errorf("failed to marshal configmaps: %w", err)
		}
		artifacts.Spec.ConfigMaps = base64.StdEncoding.EncodeToString(data)
		configMapCount = int32(len(userConfigMaps))
		h.logger.Info("Captured configmaps", zap.Int("count", len(userConfigMaps)))
	}

	// Capture Secrets (excluding service account tokens and system secrets)
	secrets := &corev1.SecretList{}
	if err := h.k8sClient.List(ctx, secrets, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	var userSecrets []corev1.Secret
	for _, secret := range secrets.Items {
		if secret.Type == corev1.SecretTypeServiceAccountToken {
			continue
		}
		if secret.Name == "default-token" || secret.Name == "builder-dockercfg" {
			continue
		}
		userSecrets = append(userSecrets, corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.Name,
				Labels:      secret.Labels,
				Annotations: secret.Annotations,
			},
			Data: secret.Data,
			Type: secret.Type,
		})
	}

	if len(userSecrets) > 0 {
		data, err := yaml.Marshal(userSecrets)
		if err != nil {
			return fmt.Errorf("failed to marshal secrets: %w", err)
		}
		artifacts.Spec.Secrets = base64.StdEncoding.EncodeToString(data)
		secretCount = int32(len(userSecrets))
		h.logger.Info("Captured secrets", zap.Int("count", len(userSecrets)))
	}

	// Set status counts
	artifacts.Status = snapshotv1.SnapshotArtifactsStatus{
		CompositionCount: compositionCount,
		ConfigMapCount:   configMapCount,
		SecretCount:      secretCount,
	}

	// Create the SnapshotArtifacts CR
	if err := h.k8sClient.Create(ctx, artifacts); err != nil {
		return fmt.Errorf("failed to create SnapshotArtifacts: %w", err)
	}

	h.logger.Info("Created SnapshotArtifacts",
		zap.String("name", snapshotName),
		zap.Int32("compositions", compositionCount),
		zap.Int32("configMaps", configMapCount),
		zap.Int32("secrets", secretCount))

	return nil
}
