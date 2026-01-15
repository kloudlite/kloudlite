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

// getEnvNamespaceForUser gets the environment namespace for a user from their WorkMachine
func (h *SnapshotHandlers) getEnvNamespaceForUser(ctx context.Context, username string) (string, error) {
	wm, err := h.workmachineRepo.GetByOwner(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to get workmachine for user %s: %w", username, err)
	}
	return wm.Spec.TargetNamespace, nil
}

// CreateSnapshotRequest is the request body for creating a snapshot
type CreateSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SnapshotResponse is the API response for a snapshot
type SnapshotResponse struct {
	Name        string                           `json:"name"`
	Namespace   string                           `json:"namespace,omitempty"`
	Description string                           `json:"description,omitempty"`
	State       snapshotv1.SnapshotState         `json:"state"`
	SizeHuman   string                           `json:"sizeHuman,omitempty"`
	SizeBytes   int64                            `json:"sizeBytes,omitempty"`
	CreatedAt   *metav1.Time                     `json:"createdAt,omitempty"`
	Registry    *snapshotv1.SnapshotRegistryInfo `json:"registry,omitempty"`
	Parent      string                           `json:"parent,omitempty"`
	StorageRefs []string                         `json:"storageRefs,omitempty"`
	Message     string                           `json:"message,omitempty"`
	IsLineage   bool                             `json:"isLineage,omitempty"`
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
		Namespace:   s.Namespace,
		Description: s.Spec.Description,
		State:       s.Status.State,
		SizeHuman:   s.Status.SizeHuman,
		SizeBytes:   s.Status.SizeBytes,
		CreatedAt:   s.Status.CreatedAt,
		Registry:    s.Status.Registry,
		Parent:      s.Spec.ParentSnapshot,
		StorageRefs: s.Status.StorageRefs,
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

// EnvironmentSnapshotRequestResponse is the API response for an environment snapshot request
type EnvironmentSnapshotRequestResponse struct {
	Name         string `json:"name"`
	SnapshotName string `json:"snapshotName"`
	Phase        string `json:"phase"`
	Message      string `json:"message,omitempty"`
}

// CreateEnvironmentSnapshot creates a new snapshot request for an environment
// This creates an EnvironmentSnapshotRequest CR which orchestrates the full snapshot workflow:
// 1. Stops workloads (scales down deployments)
// 2. Waits for pods to terminate
// 3. Creates Snapshot and SnapshotArtifacts CRs in the environment's namespace
// 4. Creates SnapshotRequest for the node to process
// 5. Restores environment state when done
// POST /api/v1/environments/:name/snapshots
func (h *SnapshotHandlers) CreateEnvironmentSnapshot(c *gin.Context) {
	envName := c.Param("name")
	username := c.GetString("user_username")

	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// Get environment to verify it exists and get details
	env, err := h.environmentRepo.Get(c.Request.Context(), namespace, envName)
	if err != nil {
		h.logger.Error("Failed to get environment", zap.String("namespace", namespace), zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("environment %s not found", envName)})
		return
	}

	// Check for existing in-progress EnvironmentSnapshotRequest for this environment
	existingRequests := &envv1.EnvironmentSnapshotRequestList{}
	if err := h.k8sClient.List(c.Request.Context(), existingRequests); err != nil {
		h.logger.Error("Failed to list existing snapshot requests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing requests"})
		return
	}

	// Check if any request is in progress for this environment
	for _, existingReq := range existingRequests.Items {
		if existingReq.Spec.EnvironmentName != envName {
			continue
		}
		phase := existingReq.Status.Phase
		if phase == "" || phase == envv1.EnvironmentSnapshotRequestPhasePending ||
			phase == envv1.EnvironmentSnapshotRequestPhaseStoppingWorkloads ||
			phase == envv1.EnvironmentSnapshotRequestPhaseWaitingForPods ||
			phase == envv1.EnvironmentSnapshotRequestPhaseCreatingSnapshot ||
			phase == envv1.EnvironmentSnapshotRequestPhaseUploadingSnapshot ||
			phase == envv1.EnvironmentSnapshotRequestPhaseRestoringEnvironment {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "A snapshot request is already in progress for this environment",
				"request": existingReq.Name,
				"phase":   string(phase),
			})
			return
		}
	}

	// Verify workmachine is assigned
	if env.Spec.WorkMachineName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment has no workmachine assigned"})
		return
	}

	// Verify workmachine is running
	_, err = h.getNodeForWorkMachine(c.Request.Context(), env.Spec.WorkMachineName)
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
	requestName := fmt.Sprintf("env-snap-req-%s", snapshotName)

	// Create the EnvironmentSnapshotRequest CR in the environment's target namespace
	envSnapshotReq := &envv1.EnvironmentSnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestName,
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":              env.Spec.OwnedBy,
				"snapshots.kloudlite.io/environment": envName,
			},
		},
		Spec: envv1.EnvironmentSnapshotRequestSpec{
			EnvironmentName:      envName,
			EnvironmentNamespace: namespace, // WorkMachine namespace where Environment lives
			SnapshotName:         snapshotName,
			Description:          req.Description,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), envSnapshotReq); err != nil {
		h.logger.Error("Failed to create EnvironmentSnapshotRequest", zap.String("name", requestName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot request: %v", err)})
		return
	}

	h.logger.Info("Created EnvironmentSnapshotRequest",
		zap.String("request", requestName),
		zap.String("snapshot", snapshotName),
		zap.String("environment", envName),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Snapshot creation started",
		"request": EnvironmentSnapshotRequestResponse{
			Name:         requestName,
			SnapshotName: snapshotName,
			Phase:        string(envv1.EnvironmentSnapshotRequestPhasePending),
			Message:      "Snapshot request created, waiting to start...",
		},
	})
}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

// ListEnvironmentSnapshots lists all snapshots for an environment
// Snapshots are namespaced in the environment's target namespace
// GET /api/v1/environments/:name/snapshots
func (h *SnapshotHandlers) ListEnvironmentSnapshots(c *gin.Context) {
	envName := c.Param("name")

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

	// Get environment to find the target namespace
	env, err := h.environmentRepo.Get(c.Request.Context(), namespace, envName)
	if err != nil {
		h.logger.Error("Failed to get environment", zap.String("namespace", namespace), zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get environment"})
		return
	}

	// List all snapshots in the environment's namespace (snapshots are now namespaced)
	snapshots := &snapshotv1.SnapshotList{}
	if err := h.k8sClient.List(c.Request.Context(), snapshots, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		h.logger.Error("Failed to list snapshots", zap.String("namespace", env.Spec.TargetNamespace), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
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

	// Convert to response slice
	response := make([]SnapshotResponse, 0, len(snapshots.Items))
	for _, s := range snapshots.Items {
		resp := snapshotToResponse(&s)

		// Check if this is a lineage snapshot (has a parent but was cloned from another env)
		resp.IsLineage = s.Spec.ParentSnapshot != "" && s.Labels["snapshots.kloudlite.io/environment"] != envName

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

		response = append(response, resp)
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
	username := c.GetString("user_username")

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
	// Snapshots are namespaced and owned by the workspace for automatic cleanup
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
			Namespace: namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "workspaces.kloudlite.io/v1",
				Kind:       "Workspace",
				Name:       ws.Name,
				UID:        ws.UID,
				Controller: boolPtr(true),
			}},
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
	username := c.GetString("user_username")

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
		username := c.GetString("user_username")
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

// DeleteSnapshot deletes a snapshot
// Snapshots are namespaced; the namespace query parameter is required
// DELETE /api/v1/snapshots/:name?namespace=...
func (h *SnapshotHandlers) DeleteSnapshot(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Query("namespace")
	username := c.GetString("user_username")
	ctx := c.Request.Context()

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace query parameter is required"})
		return
	}

	// Get snapshot with namespace (snapshots are namespaced)
	snapshot := &snapshotv1.Snapshot{}
	if err := h.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, snapshot); err != nil {
		h.logger.Error("Failed to get snapshot", zap.String("name", name), zap.String("namespace", namespace), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("snapshot %s not found in namespace %s", name, namespace)})
		return
	}

	h.logger.Info("Snapshot delete requested",
		zap.String("snapshot", name),
		zap.String("namespace", namespace),
		zap.String("user", username),
	)

	// Delete the snapshot - storage GC handled by snapshot controller
	if err := h.k8sClient.Delete(ctx, snapshot); err != nil {
		if client.IgnoreNotFound(err) != nil {
			h.logger.Error("Failed to delete snapshot", zap.String("name", name), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete snapshot: %v", err)})
			return
		}
	}

	h.logger.Info("Snapshot deletion initiated", zap.String("snapshot", name), zap.String("namespace", namespace))
	c.JSON(http.StatusOK, gin.H{
		"message":   "snapshot deletion initiated",
		"name":      name,
		"namespace": namespace,
	})
}

// CreateEnvironmentFromSnapshotRequest is the request body for creating an environment from a snapshot
type CreateEnvironmentFromSnapshotRequest struct {
	Name            string `json:"name" binding:"required"`
	SnapshotName    string `json:"snapshotName" binding:"required"`
	SourceNamespace string `json:"sourceNamespace" binding:"required"`
}

// CreateEnvironmentFromSnapshot creates a new environment from a snapshot by creating an EnvironmentForkRequest
// The EnvironmentForkRequest controller handles the actual environment creation and snapshot restoration
// POST /api/v1/environments/from-snapshot
//
// Deprecated: Use POST /api/v1/environments/:name/fork instead, which auto-selects the latest snapshot.
func (h *SnapshotHandlers) CreateEnvironmentFromSnapshot(c *gin.Context) {
	// Add deprecation header
	c.Header("Deprecation", "true")
	c.Header("Sunset", "2026-06-01")
	c.Header("Link", "</api/v1/environments/{name}/fork>; rel=\"successor-version\"")
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

	// Get the snapshot to verify it exists and is ready (snapshots are namespaced)
	snapshot := &snapshotv1.Snapshot{}
	if err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{Name: req.SnapshotName, Namespace: req.SourceNamespace}, snapshot); err != nil {
		h.logger.Error("Failed to get snapshot", zap.String("snapshot", req.SnapshotName), zap.String("namespace", req.SourceNamespace), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("snapshot %s not found in namespace %s", req.SnapshotName, req.SourceNamespace)})
		return
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("snapshot %s is not ready (state: %s)", req.SnapshotName, snapshot.Status.State)})
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

	// Use the provided name directly - namespace already provides user isolation
	envName := req.Name

	// Create EnvironmentForkRequest - the controller will handle the actual environment creation
	forkRequest := &envv1.EnvironmentForkRequest{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("fork-%s-", req.Name),
			Namespace:    namespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":        username,
				"kloudlite.io/new-environment": envName,
				"kloudlite.io/source-snapshot": req.SnapshotName,
			},
		},
		Spec: envv1.EnvironmentForkRequestSpec{
			NewEnvironmentName: envName,
			SourceSnapshot: envv1.SourceSnapshotRef{
				SnapshotName:    req.SnapshotName,
				SourceNamespace: req.SourceNamespace,
			},
			Overrides: &envv1.EnvironmentSpecOverrides{
				OwnedBy: username,
				Labels: map[string]string{
					"kloudlite.io/owned-by": username,
				},
			},
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), forkRequest); err != nil {
		h.logger.Error("Failed to create environment fork request", zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create environment fork request: %v", err)})
		return
	}

	h.logger.Info("Created environment fork request",
		zap.String("forkRequest", forkRequest.Name),
		zap.String("newEnvironment", envName),
		zap.String("snapshot", req.SnapshotName),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "environment fork request created",
		"forkRequest": forkRequest.Name,
		"environment": envName,
		"snapshot":    req.SnapshotName,
		"phase":       "Pending",
	})
}

// RestoreEnvironmentFromSnapshotRequest is the request body for restoring an environment from a snapshot
type RestoreEnvironmentFromSnapshotRequest struct {
	SnapshotName         string `json:"snapshotName" binding:"required"`
	SourceNamespace      string `json:"sourceNamespace"` // Optional: defaults to environment's target namespace
	ActivateAfterRestore bool   `json:"activateAfterRestore"`
}

// EnvironmentSnapshotRestoreResponse is the API response for an environment snapshot restore
type EnvironmentSnapshotRestoreResponse struct {
	Name         string `json:"name"`
	SnapshotName string `json:"snapshotName"`
	Phase        string `json:"phase"`
	Message      string `json:"message,omitempty"`
}

// RestoreEnvironmentFromSnapshot restores an existing environment from a snapshot
// This creates an EnvironmentSnapshotRestore CR which orchestrates the full restore workflow:
// 1. Stops workloads (scales down deployments)
// 2. Waits for pods to terminate
// 3. Downloads and restores snapshot data
// 4. Applies K8s artifacts (Compositions, ConfigMaps, Secrets)
// 5. Optionally activates the environment
// POST /api/v1/environments/:name/restore
func (h *SnapshotHandlers) RestoreEnvironmentFromSnapshot(c *gin.Context) {
	envName := c.Param("name")
	username := c.GetString("user_username")

	var req RestoreEnvironmentFromSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// Get environment to verify it exists
	env, err := h.environmentRepo.Get(c.Request.Context(), namespace, envName)
	if err != nil {
		h.logger.Error("Failed to get environment", zap.String("namespace", namespace), zap.String("name", envName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("environment %s not found", envName)})
		return
	}

	// Use sourceNamespace from request if provided, otherwise use environment's target namespace
	// This allows restoring from snapshots in different namespaces (for forking scenarios)
	snapshotNamespace := req.SourceNamespace
	if snapshotNamespace == "" {
		snapshotNamespace = env.Spec.TargetNamespace
	}

	// Get the snapshot to verify it exists and is ready (snapshots are namespaced)
	snapshot := &snapshotv1.Snapshot{}
	if err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{Name: req.SnapshotName, Namespace: snapshotNamespace}, snapshot); err != nil {
		h.logger.Error("Failed to get snapshot", zap.String("snapshot", req.SnapshotName), zap.String("namespace", snapshotNamespace), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("snapshot %s not found in namespace %s", req.SnapshotName, snapshotNamespace)})
		return
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("snapshot %s is not ready (state: %s)", req.SnapshotName, snapshot.Status.State)})
		return
	}

	// Check for existing in-progress EnvironmentSnapshotRestore for this environment
	existingRestores := &envv1.EnvironmentSnapshotRestoreList{}
	if err := h.k8sClient.List(c.Request.Context(), existingRestores); err != nil {
		h.logger.Error("Failed to list existing snapshot restores", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing restores"})
		return
	}

	// Check if any restore is in progress for this environment
	for _, existingRestore := range existingRestores.Items {
		if existingRestore.Spec.EnvironmentName != envName {
			continue
		}
		phase := existingRestore.Status.Phase
		if phase == "" || phase == envv1.EnvironmentSnapshotRestorePhasePending ||
			phase == envv1.EnvironmentSnapshotRestorePhaseStoppingWorkloads ||
			phase == envv1.EnvironmentSnapshotRestorePhaseWaitingForPods ||
			phase == envv1.EnvironmentSnapshotRestorePhaseDownloading ||
			phase == envv1.EnvironmentSnapshotRestorePhaseRestoringData ||
			phase == envv1.EnvironmentSnapshotRestorePhaseApplyingArtifacts ||
			phase == envv1.EnvironmentSnapshotRestorePhaseActivating {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "A snapshot restore is already in progress for this environment",
				"restore": existingRestore.Name,
				"phase":   string(phase),
			})
			return
		}
	}

	// Also check for in-progress EnvironmentSnapshotRequest (can't restore while snapshotting)
	existingRequests := &envv1.EnvironmentSnapshotRequestList{}
	if err := h.k8sClient.List(c.Request.Context(), existingRequests); err != nil {
		h.logger.Error("Failed to list existing snapshot requests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing requests"})
		return
	}

	for _, existingReq := range existingRequests.Items {
		if existingReq.Spec.EnvironmentName != envName {
			continue
		}
		phase := existingReq.Status.Phase
		if phase == "" || phase == envv1.EnvironmentSnapshotRequestPhasePending ||
			phase == envv1.EnvironmentSnapshotRequestPhaseStoppingWorkloads ||
			phase == envv1.EnvironmentSnapshotRequestPhaseWaitingForPods ||
			phase == envv1.EnvironmentSnapshotRequestPhaseCreatingSnapshot ||
			phase == envv1.EnvironmentSnapshotRequestPhaseUploadingSnapshot ||
			phase == envv1.EnvironmentSnapshotRequestPhaseRestoringEnvironment {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "A snapshot request is in progress for this environment, cannot restore",
				"request": existingReq.Name,
				"phase":   string(phase),
			})
			return
		}
	}

	// Verify workmachine is assigned
	if env.Spec.WorkMachineName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment has no workmachine assigned"})
		return
	}

	// Verify workmachine is running
	_, err = h.getNodeForWorkMachine(c.Request.Context(), env.Spec.WorkMachineName)
	if err != nil {
		h.logger.Error("Failed to get node for workmachine", zap.String("workmachine", env.Spec.WorkMachineName), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "workmachine is not running"})
		return
	}

	// Create restore name
	restoreName := fmt.Sprintf("env-restore-%s-%d", envName, time.Now().Unix())

	// Create the EnvironmentSnapshotRestore CR in the environment's target namespace
	envRestore := &envv1.EnvironmentSnapshotRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      restoreName,
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":              env.Spec.OwnedBy,
				"snapshots.kloudlite.io/environment": envName,
				"snapshots.kloudlite.io/snapshot":    req.SnapshotName,
			},
		},
		Spec: envv1.EnvironmentSnapshotRestoreSpec{
			EnvironmentName:      envName,
			EnvironmentNamespace: namespace, // WorkMachine namespace where Environment lives
			SnapshotName:         req.SnapshotName,
			SourceNamespace:      snapshotNamespace, // Use resolved namespace (defaults to env.Spec.TargetNamespace)
			ActivateAfterRestore: req.ActivateAfterRestore,
		},
	}

	if err := h.k8sClient.Create(c.Request.Context(), envRestore); err != nil {
		h.logger.Error("Failed to create EnvironmentSnapshotRestore", zap.String("name", restoreName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create snapshot restore: %v", err)})
		return
	}

	h.logger.Info("Created EnvironmentSnapshotRestore",
		zap.String("restore", restoreName),
		zap.String("snapshot", req.SnapshotName),
		zap.String("environment", envName),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Snapshot restore started",
		"restore": EnvironmentSnapshotRestoreResponse{
			Name:         restoreName,
			SnapshotName: req.SnapshotName,
			Phase:        string(envv1.EnvironmentSnapshotRestorePhasePending),
			Message:      "Snapshot restore created, waiting to start...",
		},
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

	var configMapCount, secretCount int32

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
		ConfigMapCount: configMapCount,
		SecretCount:    secretCount,
	}

	// Create the SnapshotArtifacts CR
	if err := h.k8sClient.Create(ctx, artifacts); err != nil {
		return fmt.Errorf("failed to create SnapshotArtifacts: %w", err)
	}

	h.logger.Info("Created SnapshotArtifacts",
		zap.String("name", snapshotName),
		zap.Int32("configMaps", configMapCount),
		zap.Int32("secrets", secretCount))

	return nil
}

// SnapshotOperationStatus represents the current snapshot operation status for an environment
type SnapshotOperationStatus struct {
	InProgress   bool   `json:"inProgress"`
	Operation    string `json:"operation,omitempty"`    // "creating" or "restoring"
	Name         string `json:"name,omitempty"`         // Name of the request/restore resource
	Phase        string `json:"phase,omitempty"`        // Current phase
	Message      string `json:"message,omitempty"`      // Status message
	SnapshotName string `json:"snapshotName,omitempty"` // Associated snapshot name
}

// GetEnvironmentSnapshotStatus returns the current snapshot operation status for an environment
// GET /api/v1/environments/:name/snapshots/status
func (h *SnapshotHandlers) GetEnvironmentSnapshotStatus(c *gin.Context) {
	envName := c.Param("name")
	ctx := c.Request.Context()

	status := SnapshotOperationStatus{
		InProgress: false,
	}

	// Check for in-progress EnvironmentSnapshotRequest
	requests := &envv1.EnvironmentSnapshotRequestList{}
	if err := h.k8sClient.List(ctx, requests); err != nil {
		h.logger.Error("Failed to list EnvironmentSnapshotRequests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check snapshot status"})
		return
	}

	for _, req := range requests.Items {
		if req.Spec.EnvironmentName != envName {
			continue
		}
		phase := req.Status.Phase
		if phase == "" || phase == envv1.EnvironmentSnapshotRequestPhasePending ||
			phase == envv1.EnvironmentSnapshotRequestPhaseStoppingWorkloads ||
			phase == envv1.EnvironmentSnapshotRequestPhaseWaitingForPods ||
			phase == envv1.EnvironmentSnapshotRequestPhaseCreatingSnapshot ||
			phase == envv1.EnvironmentSnapshotRequestPhaseUploadingSnapshot ||
			phase == envv1.EnvironmentSnapshotRequestPhaseRestoringEnvironment {
			status.InProgress = true
			status.Operation = "creating"
			status.Name = req.Name
			status.Phase = string(phase)
			status.Message = req.Status.Message
			status.SnapshotName = req.Spec.SnapshotName
			break
		}
	}

	// If no in-progress request, check for in-progress EnvironmentSnapshotRestore
	if !status.InProgress {
		restores := &envv1.EnvironmentSnapshotRestoreList{}
		if err := h.k8sClient.List(ctx, restores); err != nil {
			h.logger.Error("Failed to list EnvironmentSnapshotRestores", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check restore status"})
			return
		}

		for _, restore := range restores.Items {
			if restore.Spec.EnvironmentName != envName {
				continue
			}
			phase := restore.Status.Phase
			if phase == "" || phase == envv1.EnvironmentSnapshotRestorePhasePending ||
				phase == envv1.EnvironmentSnapshotRestorePhaseStoppingWorkloads ||
				phase == envv1.EnvironmentSnapshotRestorePhaseWaitingForPods ||
				phase == envv1.EnvironmentSnapshotRestorePhaseDownloading ||
				phase == envv1.EnvironmentSnapshotRestorePhaseRestoringData ||
				phase == envv1.EnvironmentSnapshotRestorePhaseApplyingArtifacts ||
				phase == envv1.EnvironmentSnapshotRestorePhaseActivating {
				status.InProgress = true
				status.Operation = "restoring"
				status.Name = restore.Name
				status.Phase = string(phase)
				status.Message = restore.Status.Message
				status.SnapshotName = restore.Spec.SnapshotName
				break
			}
		}
	}

	c.JSON(http.StatusOK, status)
}

// ForkEnvironmentRequest is the request body for forking an environment
type ForkEnvironmentRequest struct {
	Name string `json:"name" binding:"required"` // Name for the new forked environment
}

// ForkStatusResponse indicates whether an environment can be forked
type ForkStatusResponse struct {
	CanFork        bool   `json:"canFork"`
	LatestSnapshot string `json:"latestSnapshot,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	Message        string `json:"message,omitempty"`
}

// GetForkStatus checks if an environment can be forked (has ready snapshots)
// GET /api/v1/environments/:name/fork-status
func (h *SnapshotHandlers) GetForkStatus(c *gin.Context) {
	envName := c.Param("name")
	username := c.GetString("user_username")
	ctx := c.Request.Context()

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(ctx, username)
	if err != nil {
		c.JSON(http.StatusOK, ForkStatusResponse{
			CanFork: false,
			Message: "Failed to get user namespace",
		})
		return
	}

	// Get environment to find the target namespace
	env, err := h.environmentRepo.Get(ctx, namespace, envName)
	if err != nil {
		c.JSON(http.StatusOK, ForkStatusResponse{
			CanFork: false,
			Message: "Environment not found",
		})
		return
	}

	// Find the latest ready snapshot for this environment
	latestSnapshot, err := h.findLatestReadySnapshot(ctx, env.Spec.TargetNamespace, envName)
	if err != nil || latestSnapshot == nil {
		c.JSON(http.StatusOK, ForkStatusResponse{
			CanFork: false,
			Message: "No ready snapshots available. Create a snapshot first to enable forking.",
		})
		return
	}

	c.JSON(http.StatusOK, ForkStatusResponse{
		CanFork:        true,
		LatestSnapshot: latestSnapshot.Name,
		Namespace:      latestSnapshot.Namespace,
		Message:        fmt.Sprintf("Ready to fork from snapshot '%s'", latestSnapshot.Name),
	})
}

// ForkEnvironment creates a new environment from the latest snapshot of the source environment
// POST /api/v1/environments/:name/fork
func (h *SnapshotHandlers) ForkEnvironment(c *gin.Context) {
	sourceEnvName := c.Param("name")
	username := c.GetString("user_username")
	ctx := c.Request.Context()

	var req ForkEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the user's namespace from their WorkMachine
	namespace, err := h.getEnvNamespaceForUser(ctx, username)
	if err != nil {
		h.logger.Error("Failed to get namespace for user", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user namespace",
			"details": err.Error(),
		})
		return
	}

	// Get source environment to find the target namespace
	sourceEnv, err := h.environmentRepo.Get(ctx, namespace, sourceEnvName)
	if err != nil {
		h.logger.Error("Failed to get source environment", zap.String("namespace", namespace), zap.String("name", sourceEnvName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("environment %s not found", sourceEnvName)})
		return
	}

	// Find the latest ready snapshot for this environment
	latestSnapshot, err := h.findLatestReadySnapshot(ctx, sourceEnv.Spec.TargetNamespace, sourceEnvName)
	if err != nil {
		h.logger.Error("Failed to find snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find snapshots"})
		return
	}

	if latestSnapshot == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No ready snapshots available for this environment. Create a snapshot first to enable forking.",
			"code":  "NO_READY_SNAPSHOTS",
		})
		return
	}

	// Create EnvironmentForkRequest - the controller will handle the actual environment creation
	forkRequest := &envv1.EnvironmentForkRequest{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("fork-%s-", req.Name),
			Namespace:    namespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":           username,
				"kloudlite.io/new-environment":   req.Name,
				"kloudlite.io/source-environment": sourceEnvName,
				"kloudlite.io/source-snapshot":   latestSnapshot.Name,
			},
		},
		Spec: envv1.EnvironmentForkRequestSpec{
			NewEnvironmentName: req.Name,
			SourceSnapshot: envv1.SourceSnapshotRef{
				SnapshotName:    latestSnapshot.Name,
				SourceNamespace: latestSnapshot.Namespace,
			},
			Overrides: &envv1.EnvironmentSpecOverrides{
				OwnedBy: username,
				Labels: map[string]string{
					"kloudlite.io/owned-by": username,
				},
			},
		},
	}

	if err := h.k8sClient.Create(ctx, forkRequest); err != nil {
		h.logger.Error("Failed to create environment fork request", zap.String("name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create fork request: %v", err)})
		return
	}

	h.logger.Info("Created environment fork",
		zap.String("forkRequest", forkRequest.Name),
		zap.String("sourceEnvironment", sourceEnvName),
		zap.String("newEnvironment", req.Name),
		zap.String("snapshot", latestSnapshot.Name),
		zap.String("user", username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":           "Fork started",
		"forkRequest":       forkRequest.Name,
		"sourceEnvironment": sourceEnvName,
		"newEnvironment":    req.Name,
		"snapshot":          latestSnapshot.Name,
		"phase":             "Pending",
	})
}

// findLatestReadySnapshot finds the most recent ready snapshot for an environment
func (h *SnapshotHandlers) findLatestReadySnapshot(ctx context.Context, targetNamespace, envName string) (*snapshotv1.Snapshot, error) {
	// List all snapshots in the environment's namespace
	snapshots := &snapshotv1.SnapshotList{}
	if err := h.k8sClient.List(ctx, snapshots, client.InNamespace(targetNamespace)); err != nil {
		return nil, err
	}

	var latestSnapshot *snapshotv1.Snapshot
	var latestTime *metav1.Time

	for i := range snapshots.Items {
		s := &snapshots.Items[i]

		// Only consider ready snapshots
		if s.Status.State != snapshotv1.SnapshotStateReady {
			continue
		}

		// Only consider snapshots for this environment (not lineage snapshots from other envs)
		if envLabel := s.Labels["snapshots.kloudlite.io/environment"]; envLabel != envName {
			continue
		}

		// Find the most recent one by CreatedAt timestamp
		if s.Status.CreatedAt != nil {
			if latestTime == nil || s.Status.CreatedAt.After(latestTime.Time) {
				latestTime = s.Status.CreatedAt
				latestSnapshot = s
			}
		}
	}

	return latestSnapshot, nil
}
