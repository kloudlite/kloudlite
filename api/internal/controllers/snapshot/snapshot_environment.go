package snapshot

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handlePending starts the snapshot creation process
func (r *SnapshotReconciler) handlePending(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Snapshot is pending, starting creation")

	// Determine snapshot type
	if snapshot.Spec.WorkspaceRef != nil {
		return r.handleWorkspacePending(ctx, snapshot, logger)
	}

	// Environment snapshot
	if snapshot.Spec.EnvironmentRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Either environmentRef or workspaceRef must be set", logger)
	}

	envName := snapshot.Spec.EnvironmentRef.Name

	// Fetch the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		logger.Error("Failed to get environment", zap.Error(err), zap.String("environment", envName))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Environment not found: %s", envName), logger)
	}

	// Auto-detect parent snapshot from environment's lastRestoredSnapshot
	// This enables proper lineage when taking snapshots on cloned environments
	specUpdated := false
	if snapshot.Spec.ParentSnapshotRef == nil && env.Status.LastRestoredSnapshot != nil {
		logger.Info("Auto-detecting parent snapshot from environment's lastRestoredSnapshot",
			zap.String("parentSnapshot", env.Status.LastRestoredSnapshot.Name))
		restoredAt := env.Status.LastRestoredSnapshot.RestoredAt
		snapshot.Spec.ParentSnapshotRef = &snapshotv1.ParentSnapshotReference{
			Name:       env.Status.LastRestoredSnapshot.Name,
			RestoredAt: &restoredAt,
		}
		specUpdated = true
	}

	// Ensure labels are set for querying by environment
	labelsUpdated := false
	if snapshot.Labels == nil {
		snapshot.Labels = make(map[string]string)
	}
	if snapshot.Labels["snapshots.kloudlite.io/environment"] != envName {
		snapshot.Labels["snapshots.kloudlite.io/environment"] = envName
		labelsUpdated = true
	}
	if snapshot.Labels["kloudlite.io/owned-by"] != snapshot.Spec.OwnedBy {
		snapshot.Labels["kloudlite.io/owned-by"] = snapshot.Spec.OwnedBy
		labelsUpdated = true
	}
	if labelsUpdated || specUpdated {
		if err := r.Update(ctx, snapshot); err != nil {
			logger.Error("Failed to update snapshot", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Set state to Creating
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateCreating
		snapshot.Status.Message = "Preparing to create environment snapshot"
		snapshot.Status.WorkMachineName = env.Spec.WorkMachineName
		snapshot.Status.SnapshotType = snapshotv1.SnapshotTypeEnvironment
		snapshot.Status.TargetName = envName
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Creating", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleCreating manages the snapshot creation process
func (r *SnapshotReconciler) handleCreating(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	// Dispatch based on snapshot type
	if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
		return r.handleWorkspaceCreating(ctx, snapshot, logger)
	}

	// Environment snapshot
	if snapshot.Spec.EnvironmentRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Environment reference is required for environment snapshots", logger)
	}

	envName := snapshot.Spec.EnvironmentRef.Name

	// Fetch the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		logger.Error("Failed to get environment", zap.Error(err))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Environment not found: %s", envName), logger)
	}

	namespace := env.Spec.TargetNamespace
	if namespace == "" {
		return r.updateStatusFailed(ctx, snapshot, "Environment has no target namespace", logger)
	}

	// Store previous environment state if not already stored
	if snapshot.Status.PreviousEnvironmentActivated == nil {
		wasActivated := env.Spec.Activated
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.PreviousEnvironmentActivated = &wasActivated
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store previous environment state", zap.Error(err))
		}
	}

	// Deactivate environment properly using spec.activated field
	// This triggers the environment controller's proper shutdown sequence
	if env.Spec.Activated {
		logger.Info("Deactivating environment for snapshot", zap.String("environment", envName))
		env.Spec.Activated = false
		if err := r.Update(ctx, env); err != nil {
			logger.Error("Failed to deactivate environment", zap.Error(err))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Deactivating environment..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Wait for environment to be fully inactive
	if env.Status.State != environmentsv1.EnvironmentStateInactive {
		logger.Info("Waiting for environment to become inactive",
			zap.String("currentState", string(env.Status.State)))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = fmt.Sprintf("Waiting for environment shutdown (state: %s)...", env.Status.State)
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Wait for all pods to actually terminate (not just replicas=0)
	// This ensures databases have time to checkpoint and flush WAL
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Waiting for pods to terminate..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	logger.Info("Environment is inactive and all pods terminated, proceeding with snapshot")

	// Source path: the entire environment directory (btrfs subvolume)
	// All PVCs for this environment are stored under this directory
	sourcePath := filepath.Join("/var/lib/kloudlite/storage/environments", namespace)

	// Snapshot path: where the btrfs snapshot will be created
	snapshotPath := snapshot.Status.SnapshotPath
	if snapshotPath == "" {
		snapshotPath = filepath.Join(snapshotsBasePath, snapshot.Name)
		// Store the path immediately so subsequent reconciles use the same path
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.SnapshotPath = snapshotPath
			snapshot.Status.SourcePath = sourcePath
			snapshot.Status.Message = "Creating snapshot..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store snapshot path", zap.Error(err))
		}
	}

	// Collect K8s metadata
	var resourceMetadata *snapshotv1.ResourceMetadataInfo
	var snapshotMetadata *snapshotv1.SnapshotMetadata
	if snapshot.Spec.IncludeMetadata {
		exported, err := r.exportMetadata(ctx, namespace, logger)
		if err != nil {
			logger.Warn("Failed to collect metadata, continuing anyway", zap.Error(err))
		} else if exported != nil {
			resourceMetadata = exported.Info
			snapshotMetadata = exported.Metadata
		}
	}

	// Create a SINGLE SnapshotRequest for the entire environment directory
	snapshotReqName := fmt.Sprintf("%s-env", snapshot.Name)
	snapshotReqNamespace := fmt.Sprintf("wm-%s", env.Spec.OwnedBy)

	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotReqName,
			Namespace: snapshotReqNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot": snapshot.Name,
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:       snapshotv1.SnapshotOperationCreate,
			SourcePath:      sourcePath,
			SnapshotPath:    snapshotPath,
			SnapshotRef:     snapshot.Name,
			EnvironmentName: envName,
			ReadOnly:        true,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(snapshot, snapshotReq, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference", zap.Error(err))
	}

	// Create SnapshotRequest
	if err := r.Create(ctx, snapshotReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create SnapshotRequest", zap.Error(err))
			// Restore environment activation on failure
			if restoreErr := r.restoreEnvironmentActivation(ctx, snapshot, env, logger); restoreErr != nil {
				logger.Warn("Failed to restore environment activation after SnapshotRequest creation failure", zap.Error(restoreErr))
			}
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create SnapshotRequest: %v", err), logger)
		}
	}

	// Check if SnapshotRequest is complete (expectedCount = 1 for single environment snapshot)
	allComplete, err := r.checkSnapshotRequestsComplete(ctx, snapshot, 1, logger)
	if err != nil {
		logger.Error("Failed to check SnapshotRequest status", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !allComplete {
		// Update status with progress
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Creating btrfs snapshot..."
			snapshot.Status.ResourceMetadata = resourceMetadata
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status with progress", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Snapshot complete - get size from completed SnapshotRequest
	var totalSize int64
	completedReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotReqName, Namespace: snapshotReqNamespace}, completedReq); err == nil {
		totalSize = completedReq.Status.SizeBytes
	} else {
		logger.Warn("Failed to get SnapshotRequest for size", zap.Error(err))
	}

	// Update environment's lastRestoredSnapshot
	if err := r.updateEnvironmentLastRestored(ctx, envName, snapshot.Name, logger); err != nil {
		logger.Warn("Failed to update environment's lastRestoredSnapshot", zap.Error(err))
	}

	// Update status to Ready
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Snapshot created successfully"
		snapshot.Status.SnapshotPath = snapshotPath
		snapshot.Status.SourcePath = sourcePath
		snapshot.Status.SizeBytes = totalSize
		snapshot.Status.SizeHuman = formatSize(totalSize)
		snapshot.Status.CreatedAt = &now
		snapshot.Status.ResourceMetadata = resourceMetadata
		snapshot.Status.CollectedMetadata = snapshotMetadata
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Restore environment activation state if no other snapshots are in progress
	if !r.hasOtherInProgressSnapshots(ctx, snapshot, envName, logger) {
		if err := r.restoreEnvironmentActivation(ctx, snapshot, env, logger); err != nil {
			logger.Warn("Failed to restore environment activation", zap.Error(err))
		}
	} else {
		logger.Info("Skipping environment reactivation, other snapshots still in progress")
	}

	logger.Info("Snapshot created successfully",
		zap.String("path", snapshotPath),
		zap.String("source", sourcePath),
		zap.Int64("sizeBytes", totalSize))

	return reconcile.Result{}, nil
}

// handleRestoring handles the snapshot restore process
func (r *SnapshotReconciler) handleRestoring(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling snapshot restore",
		zap.String("snapshotType", string(snapshot.Status.SnapshotType)),
		zap.String("targetName", snapshot.Status.TargetName))

	// Dispatch based on snapshot type
	if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
		return r.handleWorkspaceRestoring(ctx, snapshot, logger)
	}

	// Environment restore
	if snapshot.Spec.EnvironmentRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Environment reference is required for environment restore", logger)
	}

	envName := snapshot.Spec.EnvironmentRef.Name

	// Fetch the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		logger.Error("Failed to get environment", zap.Error(err))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Environment not found: %s", envName), logger)
	}

	namespace := env.Spec.TargetNamespace
	if namespace == "" {
		return r.updateStatusFailed(ctx, snapshot, "Environment has no target namespace", logger)
	}

	// Store previous environment state if not already stored
	if snapshot.Status.PreviousEnvironmentActivated == nil {
		wasActivated := env.Spec.Activated
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.PreviousEnvironmentActivated = &wasActivated
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store previous environment state", zap.Error(err))
		}
	}

	// Deactivate environment properly using spec.activated field
	if env.Spec.Activated {
		logger.Info("Deactivating environment for restore", zap.String("environment", envName))
		env.Spec.Activated = false
		if err := r.Update(ctx, env); err != nil {
			logger.Error("Failed to deactivate environment", zap.Error(err))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Deactivating environment before restore..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Wait for environment to be fully inactive
	if env.Status.State != environmentsv1.EnvironmentStateInactive {
		logger.Info("Waiting for environment to become inactive before restore",
			zap.String("currentState", string(env.Status.State)))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = fmt.Sprintf("Waiting for environment shutdown (state: %s)...", env.Status.State)
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Wait for all pods to actually terminate before restore
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate before restore")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Waiting for pods to terminate..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	logger.Info("Environment is inactive and all pods terminated, proceeding with restore")

	// Source path: the btrfs snapshot containing the environment data
	snapshotPath := snapshot.Status.SnapshotPath
	if snapshotPath == "" {
		return r.updateStatusFailed(ctx, snapshot, "Snapshot path is not set", logger)
	}

	// Target path: the environment directory to restore to
	targetPath := filepath.Join("/var/lib/kloudlite/storage/environments", namespace)

	// Check existing restore request
	restoreReqName := fmt.Sprintf("%s-restore", snapshot.Name)
	restoreReqNamespace := fmt.Sprintf("wm-%s", env.Spec.OwnedBy)

	existingReq := &snapshotv1.SnapshotRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: restoreReqName, Namespace: restoreReqNamespace}, existingReq)
	if err == nil {
		// Restore request exists - check its status
		switch existingReq.Status.Phase {
		case snapshotv1.SnapshotRequestPhaseFailed:
			// Restore activation and fail
			if restoreErr := r.restoreEnvironmentActivation(ctx, snapshot, env, logger); restoreErr != nil {
				logger.Warn("Failed to restore environment activation after restore failure", zap.Error(restoreErr))
			}
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Restore failed: %s", existingReq.Status.Message), logger)

		case snapshotv1.SnapshotRequestPhaseCompleted:
			// Restore completed successfully
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
				snapshot.Status.State = snapshotv1.SnapshotStateReady
				snapshot.Status.Message = "Snapshot restored successfully"
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update status to Ready", zap.Error(err))
				return reconcile.Result{}, err
			}

			// Restore environment activation state
			if !r.hasOtherInProgressSnapshots(ctx, snapshot, envName, logger) {
				if err := r.restoreEnvironmentActivation(ctx, snapshot, env, logger); err != nil {
					logger.Warn("Failed to restore environment activation after restore", zap.Error(err))
				}
			}

			// Delete the completed restore request to allow future restores
			if err := r.Delete(ctx, existingReq); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete completed restore request", zap.Error(err))
			}

			// Track this restore on the environment for parent lineage
			if err := r.updateEnvironmentLastRestored(ctx, envName, snapshot.Name, logger); err != nil {
				logger.Warn("Failed to update environment's lastRestoredSnapshot", zap.Error(err))
			}

			logger.Info("Snapshot restored successfully", zap.String("environment", envName))
			return reconcile.Result{}, nil

		default:
			// Still in progress or pending
			logger.Info("Restore in progress, waiting...", zap.String("phase", string(existingReq.Status.Phase)))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
	} else if !apierrors.IsNotFound(err) {
		logger.Error("Failed to get restore request", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Create a SINGLE restore SnapshotRequest for the entire environment directory
	restoreReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      restoreReqName,
			Namespace: restoreReqNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/operation": "restore",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:       snapshotv1.SnapshotOperationRestore,
			SourcePath:      snapshotPath, // The btrfs snapshot
			SnapshotPath:    targetPath,   // The environment directory
			SnapshotRef:     snapshot.Name,
			EnvironmentName: envName,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(snapshot, restoreReq, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference", zap.Error(err))
	}

	if err := r.Create(ctx, restoreReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create restore SnapshotRequest", zap.Error(err))
			// Restore environment activation on failure
			if restoreErr := r.restoreEnvironmentActivation(ctx, snapshot, env, logger); restoreErr != nil {
				logger.Warn("Failed to restore environment activation after restore request creation failure", zap.Error(restoreErr))
			}
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create restore request: %v", err), logger)
		}
	}

	// Update status
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.Message = "Restoring environment data..."
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update status", zap.Error(err))
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

// hasOtherInProgressSnapshots checks if there are other snapshots for the same environment
// that are still in progress (Creating or Restoring state)
func (r *SnapshotReconciler) hasOtherInProgressSnapshots(ctx context.Context, currentSnapshot *snapshotv1.Snapshot, envName string, logger *zap.Logger) bool {
	snapshotList := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, snapshotList); err != nil {
		logger.Warn("Failed to list snapshots", zap.Error(err))
		return false
	}

	for _, snap := range snapshotList.Items {
		// Skip current snapshot
		if snap.Name == currentSnapshot.Name {
			continue
		}
		// Check if it's for the same environment
		if snap.Spec.EnvironmentRef != nil && snap.Spec.EnvironmentRef.Name == envName {
			// Check if it's in progress
			if snap.Status.State == snapshotv1.SnapshotStateCreating ||
				snap.Status.State == snapshotv1.SnapshotStateRestoring {
				logger.Info("Found other in-progress snapshot",
					zap.String("otherSnapshot", snap.Name),
					zap.String("state", string(snap.Status.State)))
				return true
			}
		}
	}
	return false
}

// waitForPodsTerminated waits for all pods in a namespace to terminate
// This is critical for databases that need time to checkpoint before snapshot
func (r *SnapshotReconciler) waitForPodsTerminated(ctx context.Context, namespace string, logger *zap.Logger) bool {
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(namespace)); err != nil {
		logger.Warn("Failed to list pods", zap.Error(err))
		return false
	}
	// Check if any non-terminated pods exist (excluding jobs/completed pods)
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			logger.Debug("Pod still running", zap.String("pod", pod.Name), zap.String("phase", string(pod.Status.Phase)))
			return false
		}
		// Also check if pod is terminating (has deletion timestamp)
		if pod.DeletionTimestamp != nil {
			logger.Debug("Pod still terminating", zap.String("pod", pod.Name))
			return false
		}
	}
	return true
}

// restoreEnvironmentActivation restores the environment's activation state after snapshot/restore
func (r *SnapshotReconciler) restoreEnvironmentActivation(ctx context.Context, snapshot *snapshotv1.Snapshot, env *environmentsv1.Environment, logger *zap.Logger) error {
	// Check if we have a stored previous state
	if snapshot.Status.PreviousEnvironmentActivated == nil {
		logger.Info("No previous environment activation state stored, skipping reactivation")
		return nil
	}

	previousActivated := *snapshot.Status.PreviousEnvironmentActivated
	if !previousActivated {
		logger.Info("Environment was not active before snapshot, skipping reactivation")
		return nil
	}

	// Re-fetch the environment to get the latest state
	currentEnv := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: env.Name}, currentEnv); err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// Only reactivate if it's currently inactive
	if !currentEnv.Spec.Activated {
		logger.Info("Reactivating environment after snapshot", zap.String("environment", env.Name))
		currentEnv.Spec.Activated = true
		if err := r.Update(ctx, currentEnv); err != nil {
			return fmt.Errorf("failed to reactivate environment: %w", err)
		}
	}

	return nil
}
