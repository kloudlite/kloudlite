package environment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// generateHash generates an 8-character hash from the input string
func generateHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])[:8]
}

// updateHashAndSubdomain computes and sets the hash and subdomain in environment status
func (r *EnvironmentReconciler) updateHashAndSubdomain(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Compute hash from envName-owner
	hash := generateHash(fmt.Sprintf("%s-%s", environment.Spec.Name, environment.Spec.OwnedBy))

	// Get subdomain from HOSTED_SUBDOMAIN env var (shared across all environments)
	subdomain := os.Getenv("HOSTED_SUBDOMAIN")
	if subdomain == "" {
		logger.Debug("HOSTED_SUBDOMAIN env var not set, subdomain will be empty")
	}

	// Only update if values changed
	if environment.Status.Hash == hash && environment.Status.Subdomain == subdomain {
		return nil
	}

	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.Hash = hash
		environment.Status.Subdomain = subdomain
		return nil
	}, logger)
}

// updateEnvironmentStatus safely updates environment status with retry logic
func (r *EnvironmentReconciler) updateEnvironmentStatus(ctx context.Context, environment *environmentsv1.Environment, state environmentsv1.EnvironmentState, message string, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = state
		environment.Status.Message = message

		now := metav1.Now()
		if state == environmentsv1.EnvironmentStateActive {
			environment.Status.LastActivatedTime = &now
		} else if state == environmentsv1.EnvironmentStateInactive {
			environment.Status.LastDeactivatedTime = &now
		}

		return nil
	}, logger)
}

// addOrUpdateCondition adds or updates a condition in the environment status
func (r *EnvironmentReconciler) addOrUpdateCondition(environment *environmentsv1.Environment, conditionType environmentsv1.EnvironmentConditionType, status metav1.ConditionStatus, reason, message string) {
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	now := metav1.Now()
	newCondition := environmentsv1.EnvironmentCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: &now,
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or add new one
	found := false
	for i, condition := range environment.Status.Conditions {
		if condition.Type == conditionType {
			environment.Status.Conditions[i] = newCondition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, newCondition)
	}
}

// handleSnapshotRestore handles environment creation from a snapshot
// This creates a SnapshotRestore resource and waits for it to complete
func (r *EnvironmentReconciler) handleSnapshotRestore(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	snapshotName := environment.Spec.FromSnapshot.SnapshotName

	// Get the snapshot to verify it exists and is ready
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("Snapshot not found", zap.String("snapshot", snapshotName))
			return r.failSnapshotRestore(ctx, environment, fmt.Sprintf("Snapshot %s not found", snapshotName), logger)
		}
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		logger.Info("Snapshot not ready, waiting", zap.String("state", string(snapshot.Status.State)))
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePending,
			fmt.Sprintf("Waiting for snapshot to be ready (state: %s)", snapshot.Status.State), logger)
	}

	// Get the node name from the workmachine
	if environment.Spec.WorkMachineName == "" {
		return r.failSnapshotRestore(ctx, environment, "Environment has no workmachine assigned", logger)
	}

	nodeName, err := r.getNodeForWorkMachine(ctx, environment.Spec.WorkMachineName)
	if err != nil {
		logger.Warn("WorkMachine not ready, waiting", zap.Error(err))
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePending,
			"Waiting for workmachine to be ready", logger)
	}

	// Ensure namespace exists for the SnapshotRestore resource
	namespaceExists, err := r.ensureNamespaceExists(ctx, environment, logger)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !namespaceExists {
		// Namespace was just created, requeue to continue
		return reconcile.Result{Requeue: true}, nil
	}

	// Define the restore name
	restoreName := fmt.Sprintf("env-restore-%s", environment.Name)

	// Check if SnapshotRestore already exists
	restore := &snapshotv1.SnapshotRestore{}
	err = r.Get(ctx, client.ObjectKey{Name: restoreName, Namespace: environment.Spec.TargetNamespace}, restore)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error("Failed to get SnapshotRestore", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Create the SnapshotRestore resource
		logger.Info("Creating SnapshotRestore", zap.String("restore", restoreName), zap.String("snapshot", snapshotName))

		targetPath := fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", environment.Spec.TargetNamespace)

		restore = &snapshotv1.SnapshotRestore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      restoreName,
				Namespace: environment.Spec.TargetNamespace,
				Labels: map[string]string{
					"kloudlite.io/owned-by":              environment.Spec.OwnedBy,
					"snapshots.kloudlite.io/environment": environment.Name,
					"snapshots.kloudlite.io/source":      snapshotName,
				},
			},
			Spec: snapshotv1.SnapshotRestoreSpec{
				SnapshotName: snapshotName,
				TargetPath:   targetPath,
				NodeName:     nodeName,
			},
		}

		if err := r.Create(ctx, restore); err != nil {
			logger.Error("Failed to create SnapshotRestore", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Update status to show we're starting the restore
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePulling,
			"Downloading snapshot from registry", logger)
	}

	// SnapshotRestore exists, check its status
	switch restore.Status.State {
	case snapshotv1.SnapshotRestoreStatePending:
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePending,
			"Waiting to start restore", logger)

	case snapshotv1.SnapshotRestoreStateDownloading:
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePulling,
			"Downloading snapshot from registry", logger)

	case snapshotv1.SnapshotRestoreStateRestoring:
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhaseDataRestoring,
			"Restoring snapshot data", logger)

	case snapshotv1.SnapshotRestoreStateCompleted:
		// Restore completed! Update environment and clear FromSnapshot
		logger.Info("Snapshot restore completed successfully", zap.String("snapshot", snapshotName))

		now := metav1.Now()

		// Update status with completed restore and LastRestoredSnapshot
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
				Phase:          environmentsv1.SnapshotRestorePhaseCompleted,
				Message:        "Snapshot restored successfully",
				SourceSnapshot: snapshotName,
				CompletionTime: &now,
			}
			environment.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
				Name:       snapshotName,
				RestoredAt: now,
			}
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Clear FromSnapshot to proceed with normal reconciliation
		environment.Spec.FromSnapshot = nil
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to clear fromSnapshot", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Cleared fromSnapshot, proceeding with normal environment reconciliation")
		return reconcile.Result{Requeue: true}, nil

	case snapshotv1.SnapshotRestoreStateFailed:
		return r.failSnapshotRestore(ctx, environment,
			fmt.Sprintf("Snapshot restore failed: %s", restore.Status.Message), logger)

	default:
		logger.Warn("Unknown restore state", zap.String("state", string(restore.Status.State)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
}

// getNodeForWorkMachine finds the k8s node for a workmachine by label
func (r *EnvironmentReconciler) getNodeForWorkMachine(ctx context.Context, workmachineName string) (string, error) {
	var nodes corev1.NodeList
	if err := r.List(ctx, &nodes, client.MatchingLabels{
		"kloudlite.io/workmachine": workmachineName,
	}); err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no node found for workmachine %s", workmachineName)
	}
	return nodes.Items[0].Name, nil
}

// updateSnapshotRestoreStatus updates the environment's snapshot restore status
func (r *EnvironmentReconciler) updateSnapshotRestoreStatus(ctx context.Context, environment *environmentsv1.Environment, phase environmentsv1.SnapshotRestorePhase, message string, logger *zap.Logger) (reconcile.Result, error) {
	now := metav1.Now()

	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		if environment.Status.SnapshotRestoreStatus == nil {
			environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
				StartTime:      &now,
				SourceSnapshot: environment.Spec.FromSnapshot.SnapshotName,
			}
		}
		environment.Status.SnapshotRestoreStatus.Phase = phase
		environment.Status.SnapshotRestoreStatus.Message = message
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update snapshot restore status", zap.Error(err))
	}

	// Requeue to check progress
	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// failSnapshotRestore marks the snapshot restore as failed and clears FromSnapshot
func (r *EnvironmentReconciler) failSnapshotRestore(ctx context.Context, environment *environmentsv1.Environment, errorMessage string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot restore failed", zap.String("error", errorMessage))

	now := metav1.Now()

	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
			Phase:          environmentsv1.SnapshotRestorePhaseFailed,
			Message:        errorMessage,
			ErrorMessage:   errorMessage,
			SourceSnapshot: environment.Spec.FromSnapshot.SnapshotName,
			CompletionTime: &now,
		}
		environment.Status.State = environmentsv1.EnvironmentStateError
		environment.Status.Message = errorMessage
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update status", zap.Error(err))
	}

	// Clear FromSnapshot even on failure to avoid infinite loops
	environment.Spec.FromSnapshot = nil
	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear fromSnapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
