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

	// Auto-detect parent snapshot for proper lineage tracking
	// Priority: 1) Latest snapshot for this environment, 2) LastRestoredSnapshot (for forked envs)
	specUpdated := false
	if snapshot.Spec.ParentSnapshotRef == nil {
		var parentSnapshotName string

		// First, try to find the latest existing snapshot for this environment
		latestSnapshot := r.findLatestSnapshotForEnvironment(ctx, envName, snapshot.Name, logger)
		if latestSnapshot != nil {
			parentSnapshotName = latestSnapshot.Name
			logger.Info("Auto-detecting parent from latest environment snapshot",
				zap.String("parentSnapshot", parentSnapshotName))
			snapshot.Spec.ParentSnapshotRef = &snapshotv1.ParentSnapshotReference{
				Name: parentSnapshotName,
			}
			specUpdated = true
		} else if env.Status.LastRestoredSnapshot != nil {
			// No existing snapshots - use LastRestoredSnapshot (fork scenario)
			parentSnapshotName = env.Status.LastRestoredSnapshot.Name
			logger.Info("Auto-detecting parent from environment's lastRestoredSnapshot (fork origin)",
				zap.String("parentSnapshot", parentSnapshotName))
			restoredAt := env.Status.LastRestoredSnapshot.RestoredAt
			snapshot.Spec.ParentSnapshotRef = &snapshotv1.ParentSnapshotReference{
				Name:       parentSnapshotName,
				RestoredAt: &restoredAt,
			}
			specUpdated = true
		}

		// Inherit description from parent snapshot if not set
		if parentSnapshotName != "" && snapshot.Spec.Description == "" {
			parentSnapshot := &snapshotv1.Snapshot{}
			if err := r.Get(ctx, client.ObjectKey{Name: parentSnapshotName}, parentSnapshot); err == nil {
				if parentSnapshot.Spec.Description != "" {
					logger.Info("Inheriting description from parent snapshot",
						zap.String("description", parentSnapshot.Spec.Description))
					snapshot.Spec.Description = parentSnapshot.Spec.Description
				}
			} else {
				logger.Warn("Failed to fetch parent snapshot for description inheritance", zap.Error(err))
			}
		}
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
	// Set parent label for lineage tracking (used during deletion re-linking)
	if snapshot.Spec.ParentSnapshotRef != nil {
		if snapshot.Labels["snapshots.kloudlite.io/parent"] != snapshot.Spec.ParentSnapshotRef.Name {
			snapshot.Labels["snapshots.kloudlite.io/parent"] = snapshot.Spec.ParentSnapshotRef.Name
			labelsUpdated = true
		}
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

	// Set environment to snapping state if not already
	// This triggers the environment controller to scale down deployments
	if env.Status.State != environmentsv1.EnvironmentStateSnapping {
		logger.Info("Setting environment to snapping state", zap.String("environment", envName))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, env, func() error {
			env.Status.State = environmentsv1.EnvironmentStateSnapping
			env.Status.Message = "Taking snapshot..."
			return nil
		}, logger); err != nil {
			logger.Error("Failed to set snapping state", zap.Error(err))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Snapping environment..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Wait for all pods to terminate (environment controller handles scaling down)
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate for snapshot")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Waiting for pods to terminate..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	logger.Info("All pods terminated, proceeding with snapshot")

	// Source path: the entire environment directory (btrfs subvolume)
	// All PVCs for this environment are stored under this directory
	sourcePath := filepath.Join(environmentsStorePath, namespace)

	// Snapshot path: where the btrfs snapshot will be created
	// Store in .snapshots/envs/{envName}/{snapshotName}/ for environment snapshots
	// Each environment has its own snapshot folder to avoid conflicts when forking
	snapshotPath := snapshot.Status.SnapshotPath
	if snapshotPath == "" {
		snapshotPath = filepath.Join(envSnapshotsBasePath, env.Name, snapshot.Name)
		// Store the path immediately so subsequent reconciles use the same path
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.SnapshotPath = snapshotPath
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

	// Update status to Ready
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Snapshot created successfully"
		snapshot.Status.SnapshotPath = snapshotPath
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

	// Set environment to snapping state if not already
	// This triggers the environment controller to scale down deployments
	if env.Status.State != environmentsv1.EnvironmentStateSnapping {
		logger.Info("Setting environment to snapping state for restore", zap.String("environment", envName))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, env, func() error {
			env.Status.State = environmentsv1.EnvironmentStateSnapping
			env.Status.Message = "Restoring snapshot..."
			return nil
		}, logger); err != nil {
			logger.Error("Failed to set snapping state", zap.Error(err))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Snapping environment for restore..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Wait for all pods to terminate (environment controller handles scaling down)
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate for restore")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Waiting for pods to terminate..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	logger.Info("All pods terminated, proceeding with restore")

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

// waitForPodsTerminated waits for all pods in a namespace to be fully deleted
// This is critical for databases that need time to checkpoint before snapshot
// We wait for ALL pods to be deleted (not just not Running) because:
// - Pods in Error/CrashLoopBackOff still have running processes
// - Database containers may still be writing to disk during shutdown
// - Only Succeeded pods (completed Jobs) are safe to ignore
func (r *SnapshotReconciler) waitForPodsTerminated(ctx context.Context, namespace string, logger *zap.Logger) bool {
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(namespace)); err != nil {
		logger.Warn("Failed to list pods", zap.Error(err))
		return false
	}
	// Wait for ALL pods to be deleted, except completed Job pods (Succeeded phase)
	// This ensures databases have fully shut down and checkpointed
	for _, pod := range pods.Items {
		// Skip completed Job pods - they're finished and won't write to disk
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}
		// Any other pod (Running, Pending, Failed, Unknown) means we should wait
		// Even Failed/Error pods may have just crashed and filesystem may be inconsistent
		logger.Debug("Pod still exists", zap.String("pod", pod.Name), zap.String("phase", string(pod.Status.Phase)))
		return false
	}
	return true
}

// restoreEnvironmentActivation restores the environment's state after snapshot/restore
func (r *SnapshotReconciler) restoreEnvironmentActivation(ctx context.Context, snapshot *snapshotv1.Snapshot, env *environmentsv1.Environment, logger *zap.Logger) error {
	// Re-fetch the environment to get the latest state
	currentEnv := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: env.Name}, currentEnv); err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// If environment should be active (Spec.Activated=true), set state back to active
	// This triggers the environment controller to scale up deployments
	if currentEnv.Spec.Activated && currentEnv.Status.State == environmentsv1.EnvironmentStateSnapping {
		logger.Info("Restoring environment to active state after snapshot", zap.String("environment", env.Name))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, currentEnv, func() error {
			currentEnv.Status.State = environmentsv1.EnvironmentStateActive
			currentEnv.Status.Message = "Environment is active"
			return nil
		}, logger); err != nil {
			return fmt.Errorf("failed to restore environment state: %w", err)
		}
	}

	return nil
}

// findLatestSnapshotForEnvironment finds the most recent Ready snapshot for an environment
// Excludes the current snapshot being created (by name)
func (r *SnapshotReconciler) findLatestSnapshotForEnvironment(
	ctx context.Context,
	envName string,
	excludeSnapshotName string,
	logger *zap.Logger,
) *snapshotv1.Snapshot {
	snapshotList := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, snapshotList, client.MatchingLabels{
		"snapshots.kloudlite.io/environment": envName,
	}); err != nil {
		logger.Warn("Failed to list snapshots for environment", zap.Error(err))
		return nil
	}

	var latestSnapshot *snapshotv1.Snapshot
	var latestTime *metav1.Time

	for i := range snapshotList.Items {
		snap := &snapshotList.Items[i]

		// Skip the current snapshot being created
		if snap.Name == excludeSnapshotName {
			continue
		}

		// Only consider Ready snapshots
		if snap.Status.State != snapshotv1.SnapshotStateReady {
			continue
		}

		// Find the most recent by CreatedAt
		if snap.Status.CreatedAt != nil {
			if latestTime == nil || snap.Status.CreatedAt.After(latestTime.Time) {
				latestSnapshot = snap
				latestTime = snap.Status.CreatedAt
			}
		}
	}

	return latestSnapshot
}
