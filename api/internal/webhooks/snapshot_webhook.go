package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	envv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SnapshotWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewSnapshotWebhook(logger logger.Logger, k8sClient client.Client) *SnapshotWebhook {
	return &SnapshotWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateSnapshotRequest handles validation webhook for SnapshotRequest CRD
func (w *SnapshotWebhook) ValidateSnapshotRequest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleSnapshotRequestValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

// ValidateSnapshotRestore handles validation webhook for SnapshotRestore CRD
func (w *SnapshotWebhook) ValidateSnapshotRestore(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleSnapshotRestoreValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *SnapshotWebhook) handleSnapshotRequestValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only validate CREATE operations
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var snapshotReq snapshotv1.SnapshotRequest
	if err := json.Unmarshal(req.Object.Raw, &snapshotReq); err != nil {
		w.logger.Error("Failed to unmarshal SnapshotRequest: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal SnapshotRequest object",
			},
		}
	}

	if err := w.validateNoConflictingOperations(snapshotReq.Spec.SourcePath, snapshotReq.Name); err != nil {
		w.logger.Warn("SnapshotRequest validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

func (w *SnapshotWebhook) handleSnapshotRestoreValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only validate CREATE operations
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var snapshotRestore snapshotv1.SnapshotRestore
	if err := json.Unmarshal(req.Object.Raw, &snapshotRestore); err != nil {
		w.logger.Error("Failed to unmarshal SnapshotRestore: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal SnapshotRestore object",
			},
		}
	}

	if err := w.validateNoConflictingOperations(snapshotRestore.Spec.TargetPath, snapshotRestore.Name); err != nil {
		w.logger.Warn("SnapshotRestore validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

// validateNoConflictingOperations checks that no other SnapshotRequest or SnapshotRestore
// is currently in progress for the same subvolume path
func (w *SnapshotWebhook) validateNoConflictingOperations(subvolumePath string, currentName string) error {
	ctx := context.Background()

	// Check for in-progress SnapshotRequests with the same sourcePath
	snapshotRequests := &snapshotv1.SnapshotRequestList{}
	if err := w.k8sClient.List(ctx, snapshotRequests); err != nil {
		return fmt.Errorf("failed to list SnapshotRequests: %v", err)
	}

	for _, req := range snapshotRequests.Items {
		// Skip the current resource being created
		if req.Name == currentName {
			continue
		}

		// Check if this request targets the same subvolume
		if req.Spec.SourcePath == subvolumePath {
			// Check if it's in progress (not completed or failed)
			if isSnapshotRequestInProgress(req.Status.State) {
				return fmt.Errorf("another snapshot operation is already in progress for path '%s': SnapshotRequest '%s' (state: %s). Please wait for it to complete",
					subvolumePath, req.Name, req.Status.State)
			}
		}
	}

	// Check for in-progress SnapshotRestores with the same targetPath
	snapshotRestores := &snapshotv1.SnapshotRestoreList{}
	if err := w.k8sClient.List(ctx, snapshotRestores); err != nil {
		return fmt.Errorf("failed to list SnapshotRestores: %v", err)
	}

	for _, restore := range snapshotRestores.Items {
		// Skip the current resource being created
		if restore.Name == currentName {
			continue
		}

		// Check if this restore targets the same subvolume
		if restore.Spec.TargetPath == subvolumePath {
			// Check if it's in progress (not completed or failed)
			if isSnapshotRestoreInProgress(restore.Status.State) {
				return fmt.Errorf("another snapshot operation is already in progress for path '%s': SnapshotRestore '%s' (state: %s). Please wait for it to complete",
					subvolumePath, restore.Name, restore.Status.State)
			}
		}
	}

	return nil
}

// isSnapshotRequestInProgress returns true if the snapshot request is still processing
func isSnapshotRequestInProgress(state snapshotv1.SnapshotRequestState) bool {
	switch state {
	case snapshotv1.SnapshotRequestStateCompleted, snapshotv1.SnapshotRequestStateFailed:
		return false
	default:
		// Pending, Creating, Uploading are all in-progress states
		return true
	}
}

// isSnapshotRestoreInProgress returns true if the snapshot restore is still processing
func isSnapshotRestoreInProgress(state snapshotv1.SnapshotRestoreState) bool {
	switch state {
	case snapshotv1.SnapshotRestoreStateCompleted, snapshotv1.SnapshotRestoreStateFailed:
		return false
	default:
		// Pending, Downloading, Restoring are all in-progress states
		return true
	}
}

// ValidateEnvironmentSnapshotRequest handles validation webhook for EnvironmentSnapshotRequest CRD
func (w *SnapshotWebhook) ValidateEnvironmentSnapshotRequest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleEnvironmentSnapshotRequestValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

// ValidateEnvironmentSnapshotRestore handles validation webhook for EnvironmentSnapshotRestore CRD
func (w *SnapshotWebhook) ValidateEnvironmentSnapshotRestore(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleEnvironmentSnapshotRestoreValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *SnapshotWebhook) handleEnvironmentSnapshotRequestValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only validate CREATE operations
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var envSnapReq envv1.EnvironmentSnapshotRequest
	if err := json.Unmarshal(req.Object.Raw, &envSnapReq); err != nil {
		w.logger.Error("Failed to unmarshal EnvironmentSnapshotRequest: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal EnvironmentSnapshotRequest object",
			},
		}
	}

	envName := envSnapReq.Spec.EnvironmentName
	if err := w.validateNoConflictingEnvironmentOperations(envName, envSnapReq.Name); err != nil {
		w.logger.Warn("EnvironmentSnapshotRequest validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

func (w *SnapshotWebhook) handleEnvironmentSnapshotRestoreValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only validate CREATE operations
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var envSnapRestore envv1.EnvironmentSnapshotRestore
	if err := json.Unmarshal(req.Object.Raw, &envSnapRestore); err != nil {
		w.logger.Error("Failed to unmarshal EnvironmentSnapshotRestore: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal EnvironmentSnapshotRestore object",
			},
		}
	}

	envName := envSnapRestore.Spec.EnvironmentName
	if err := w.validateNoConflictingEnvironmentOperations(envName, envSnapRestore.Name); err != nil {
		w.logger.Warn("EnvironmentSnapshotRestore validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

// validateNoConflictingEnvironmentOperations checks that no other EnvironmentSnapshotRequest or EnvironmentSnapshotRestore
// is currently in progress for the same environment
func (w *SnapshotWebhook) validateNoConflictingEnvironmentOperations(envName string, currentName string) error {
	ctx := context.Background()

	// Check for in-progress EnvironmentSnapshotRequests for the same environment
	envSnapRequests := &envv1.EnvironmentSnapshotRequestList{}
	if err := w.k8sClient.List(ctx, envSnapRequests); err != nil {
		return fmt.Errorf("failed to list EnvironmentSnapshotRequests: %v", err)
	}

	for _, req := range envSnapRequests.Items {
		// Skip the current resource being created
		if req.Name == currentName {
			continue
		}

		// Check if this request targets the same environment
		if req.Spec.EnvironmentName == envName {
			// Check if it's in progress (not completed or failed)
			if isEnvironmentSnapshotRequestInProgress(req.Status.Phase) {
				return fmt.Errorf("another snapshot operation is already in progress for environment '%s': EnvironmentSnapshotRequest '%s' (phase: %s). Please wait for it to complete",
					envName, req.Name, req.Status.Phase)
			}
		}
	}

	// Check for in-progress EnvironmentSnapshotRestores for the same environment
	envSnapRestores := &envv1.EnvironmentSnapshotRestoreList{}
	if err := w.k8sClient.List(ctx, envSnapRestores); err != nil {
		return fmt.Errorf("failed to list EnvironmentSnapshotRestores: %v", err)
	}

	for _, restore := range envSnapRestores.Items {
		// Skip the current resource being created
		if restore.Name == currentName {
			continue
		}

		// Check if this restore targets the same environment
		if restore.Spec.EnvironmentName == envName {
			// Check if it's in progress (not completed or failed)
			if isEnvironmentSnapshotRestoreInProgress(restore.Status.Phase) {
				return fmt.Errorf("another snapshot operation is already in progress for environment '%s': EnvironmentSnapshotRestore '%s' (phase: %s). Please wait for it to complete",
					envName, restore.Name, restore.Status.Phase)
			}
		}
	}

	return nil
}

// isEnvironmentSnapshotRequestInProgress returns true if the request is still processing
func isEnvironmentSnapshotRequestInProgress(phase envv1.EnvironmentSnapshotRequestPhase) bool {
	switch phase {
	case envv1.EnvironmentSnapshotRequestPhaseCompleted, envv1.EnvironmentSnapshotRequestPhaseFailed:
		return false
	default:
		// Pending, StoppingWorkloads, WaitingForPods, CreatingSnapshot, UploadingSnapshot, RestoringEnvironment are all in-progress
		return true
	}
}

// isEnvironmentSnapshotRestoreInProgress returns true if the restore is still processing
func isEnvironmentSnapshotRestoreInProgress(phase envv1.EnvironmentSnapshotRestorePhase) bool {
	switch phase {
	case envv1.EnvironmentSnapshotRestorePhaseCompleted, envv1.EnvironmentSnapshotRestorePhaseFailed:
		return false
	default:
		// Pending, StoppingWorkloads, WaitingForPods, Downloading, RestoringData, ApplyingArtifacts, Activating are all in-progress
		return true
	}
}

// ValidateSnapshot handles validation webhook for Snapshot DELETE operations
func (w *SnapshotWebhook) ValidateSnapshot(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleSnapshotValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *SnapshotWebhook) handleSnapshotValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only validate DELETE operations
	if req.Operation != admissionv1.Delete {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var snapshot snapshotv1.Snapshot
	if err := json.Unmarshal(req.OldObject.Raw, &snapshot); err != nil {
		w.logger.Error("Failed to unmarshal Snapshot: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal Snapshot object",
			},
		}
	}

	if err := w.validateSnapshotDelete(&snapshot); err != nil {
		w.logger.Warn("Snapshot deletion validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

func (w *SnapshotWebhook) validateSnapshotDelete(snapshot *snapshotv1.Snapshot) error {
	ctx := context.Background()

	// Get snapshot's namespace
	snapshotNamespace := snapshot.Namespace

	// List ALL environments to find the one with targetNamespace == snapshot.namespace
	environments := &envv1.EnvironmentList{}
	if err := w.k8sClient.List(ctx, environments); err != nil {
		return fmt.Errorf("failed to list environments: %v", err)
	}

	// Find environment that targets this namespace
	for _, env := range environments.Items {
		if env.Spec.TargetNamespace == snapshotNamespace {
			// Check if this snapshot is the current snapshot for the environment
			if env.Status.LastRestoredSnapshot != nil && env.Status.LastRestoredSnapshot.Name == snapshot.Name {
				return fmt.Errorf("cannot delete snapshot '%s': it is the current snapshot for environment '%s' in namespace '%s'",
					snapshot.Name, env.Name, env.Namespace)
			}
		}
	}

	return nil
}
