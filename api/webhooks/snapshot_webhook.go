package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/pkg/logger"
	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	envv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CheckpointWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewCheckpointWebhook(logger logger.Logger, k8sClient client.Client) *CheckpointWebhook {
	return &CheckpointWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateCheckpoint handles validation webhook for Checkpoint CREATE operations
func (w *CheckpointWebhook) ValidateCheckpoint(c *gin.Context) {
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

	response := w.handleCheckpointValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *CheckpointWebhook) handleCheckpointValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	if req.Operation == admissionv1.Create {
		var cp checkpointv1.Checkpoint
		if err := json.Unmarshal(req.Object.Raw, &cp); err != nil {
			w.logger.Error("Failed to unmarshal Checkpoint: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result:  &metav1.Status{Message: "Failed to unmarshal Checkpoint object"},
			}
		}

		if err := w.validateNoConflictingOperations(cp.Spec.SourcePath, cp.Name); err != nil {
			w.logger.Warn("Checkpoint validation failed: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result:  &metav1.Status{Message: err.Error()},
			}
		}
	}

	if req.Operation == admissionv1.Delete {
		var cp checkpointv1.Checkpoint
		if err := json.Unmarshal(req.OldObject.Raw, &cp); err != nil {
			w.logger.Error("Failed to unmarshal Checkpoint: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result:  &metav1.Status{Message: "Failed to unmarshal Checkpoint object"},
			}
		}

		if err := w.validateCheckpointDelete(&cp); err != nil {
			w.logger.Warn("Checkpoint deletion validation failed: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result:  &metav1.Status{Message: err.Error()},
			}
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

// ValidateCheckpointRestore handles validation webhook for CheckpointRestore CRD
func (w *CheckpointWebhook) ValidateCheckpointRestore(c *gin.Context) {
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

	response := w.handleCheckpointRestoreValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *CheckpointWebhook) handleCheckpointRestoreValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var restore checkpointv1.CheckpointRestore
	if err := json.Unmarshal(req.Object.Raw, &restore); err != nil {
		w.logger.Error("Failed to unmarshal CheckpointRestore: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: "Failed to unmarshal CheckpointRestore object"},
		}
	}

	if err := w.validateNoConflictingOperations(restore.Spec.TargetPath, restore.Name); err != nil {
		w.logger.Warn("CheckpointRestore validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

// validateNoConflictingOperations checks that no other Checkpoint or CheckpointRestore
// is currently in progress for the same subvolume path
func (w *CheckpointWebhook) validateNoConflictingOperations(subvolumePath string, currentName string) error {
	ctx := context.Background()

	// Check for in-progress Checkpoints with the same sourcePath
	checkpoints := &checkpointv1.CheckpointList{}
	if err := w.k8sClient.List(ctx, checkpoints); err != nil {
		return fmt.Errorf("failed to list Checkpoints: %v", err)
	}

	for _, cp := range checkpoints.Items {
		if cp.Name == currentName {
			continue
		}
		if cp.Spec.SourcePath == subvolumePath {
			if isCheckpointInProgress(cp.Status.Phase) {
				return fmt.Errorf("another checkpoint operation is already in progress for path '%s': Checkpoint '%s' (phase: %s). Please wait for it to complete",
					subvolumePath, cp.Name, cp.Status.Phase)
			}
		}
	}

	// Check for in-progress CheckpointRestores with the same targetPath
	restores := &checkpointv1.CheckpointRestoreList{}
	if err := w.k8sClient.List(ctx, restores); err != nil {
		return fmt.Errorf("failed to list CheckpointRestores: %v", err)
	}

	for _, restore := range restores.Items {
		if restore.Name == currentName {
			continue
		}
		if restore.Spec.TargetPath == subvolumePath {
			if isCheckpointRestoreInProgress(restore.Status.Phase) {
				return fmt.Errorf("another checkpoint operation is already in progress for path '%s': CheckpointRestore '%s' (phase: %s). Please wait for it to complete",
					subvolumePath, restore.Name, restore.Status.Phase)
			}
		}
	}

	return nil
}

func isCheckpointInProgress(phase checkpointv1.CheckpointPhase) bool {
	switch phase {
	case checkpointv1.CheckpointPhaseReady, checkpointv1.CheckpointPhaseFailed:
		return false
	default:
		return true
	}
}

func isCheckpointRestoreInProgress(phase checkpointv1.CheckpointRestorePhase) bool {
	switch phase {
	case checkpointv1.CheckpointRestorePhaseCompleted, checkpointv1.CheckpointRestorePhaseFailed:
		return false
	default:
		return true
	}
}

func (w *CheckpointWebhook) validateCheckpointDelete(cp *checkpointv1.Checkpoint) error {
	ctx := context.Background()

	cpNamespace := cp.Namespace

	environments := &envv1.EnvironmentList{}
	if err := w.k8sClient.List(ctx, environments); err != nil {
		return fmt.Errorf("failed to list environments: %v", err)
	}

	for _, env := range environments.Items {
		if env.Spec.TargetNamespace == cpNamespace {
			if env.Status.LastRestoredCheckpoint != nil && env.Status.LastRestoredCheckpoint.Name == cp.Name {
				return fmt.Errorf("cannot delete checkpoint '%s': it is the current checkpoint for environment '%s' in namespace '%s'",
					cp.Name, env.Name, env.Namespace)
			}
		}
	}

	return nil
}

// ValidateEnvironmentCheckpoint handles validation webhook for EnvironmentCheckpoint CRD
func (w *CheckpointWebhook) ValidateEnvironmentCheckpoint(c *gin.Context) {
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

	response := w.handleEnvironmentCheckpointValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

// ValidateEnvironmentCheckpointRestore handles validation webhook for EnvironmentCheckpointRestore CRD
func (w *CheckpointWebhook) ValidateEnvironmentCheckpointRestore(c *gin.Context) {
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

	response := w.handleEnvironmentCheckpointRestoreValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *CheckpointWebhook) handleEnvironmentCheckpointValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var envCP envv1.EnvironmentCheckpoint
	if err := json.Unmarshal(req.Object.Raw, &envCP); err != nil {
		w.logger.Error("Failed to unmarshal EnvironmentCheckpoint: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: "Failed to unmarshal EnvironmentCheckpoint object"},
		}
	}

	if err := w.validateNoConflictingEnvironmentOperations(envCP.Spec.EnvironmentName, envCP.Name); err != nil {
		w.logger.Warn("EnvironmentCheckpoint validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

func (w *CheckpointWebhook) handleEnvironmentCheckpointRestoreValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var envRestore envv1.EnvironmentCheckpointRestore
	if err := json.Unmarshal(req.Object.Raw, &envRestore); err != nil {
		w.logger.Error("Failed to unmarshal EnvironmentCheckpointRestore: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: "Failed to unmarshal EnvironmentCheckpointRestore object"},
		}
	}

	if err := w.validateNoConflictingEnvironmentOperations(envRestore.Spec.EnvironmentName, envRestore.Name); err != nil {
		w.logger.Warn("EnvironmentCheckpointRestore validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

// validateNoConflictingEnvironmentOperations checks that no other EnvironmentCheckpoint or EnvironmentCheckpointRestore
// is currently in progress for the same environment
func (w *CheckpointWebhook) validateNoConflictingEnvironmentOperations(envName string, currentName string) error {
	ctx := context.Background()

	envCheckpoints := &envv1.EnvironmentCheckpointList{}
	if err := w.k8sClient.List(ctx, envCheckpoints); err != nil {
		return fmt.Errorf("failed to list EnvironmentCheckpoints: %v", err)
	}

	for _, cp := range envCheckpoints.Items {
		if cp.Name == currentName {
			continue
		}
		if cp.Spec.EnvironmentName == envName {
			if isEnvironmentCheckpointInProgress(cp.Status.Phase) {
				return fmt.Errorf("another checkpoint operation is already in progress for environment '%s': EnvironmentCheckpoint '%s' (phase: %s). Please wait for it to complete",
					envName, cp.Name, cp.Status.Phase)
			}
		}
	}

	envRestores := &envv1.EnvironmentCheckpointRestoreList{}
	if err := w.k8sClient.List(ctx, envRestores); err != nil {
		return fmt.Errorf("failed to list EnvironmentCheckpointRestores: %v", err)
	}

	for _, restore := range envRestores.Items {
		if restore.Name == currentName {
			continue
		}
		if restore.Spec.EnvironmentName == envName {
			if isEnvironmentCheckpointRestoreInProgress(restore.Status.Phase) {
				return fmt.Errorf("another checkpoint operation is already in progress for environment '%s': EnvironmentCheckpointRestore '%s' (phase: %s). Please wait for it to complete",
					envName, restore.Name, restore.Status.Phase)
			}
		}
	}

	return nil
}

func isEnvironmentCheckpointInProgress(phase envv1.EnvironmentCheckpointPhase) bool {
	switch phase {
	case envv1.EnvironmentCheckpointPhaseCompleted, envv1.EnvironmentCheckpointPhaseFailed:
		return false
	default:
		return true
	}
}

func isEnvironmentCheckpointRestoreInProgress(phase envv1.EnvironmentCheckpointRestorePhase) bool {
	switch phase {
	case envv1.EnvironmentCheckpointRestorePhaseCompleted, envv1.EnvironmentCheckpointRestorePhaseFailed:
		return false
	default:
		return true
	}
}
