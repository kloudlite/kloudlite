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
	appsv1 "k8s.io/api/apps/v1"
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

	// Scale down environment to stop all pods before snapshot
	if err := r.scaleEnvironment(ctx, namespace, 0, logger); err != nil {
		logger.Warn("Failed to scale down environment", zap.Error(err))
	}

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
			snapshot.Status.Message = "Preparing snapshot..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store snapshot path", zap.Error(err))
		}
	}

	// Wait for pods to terminate
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Stopping environment pods..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
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
			Metadata:        snapshotMetadata,
			MetadataPath:    snapshotPath,
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
			// Scale environment back up on failure
			if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
				logger.Warn("Failed to scale up environment after SnapshotRequest creation failure", zap.Error(scaleErr))
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

	// Scale environment back up only if no other snapshots are in progress
	if !r.hasOtherInProgressSnapshots(ctx, snapshot, envName, logger) {
		if err := r.scaleEnvironment(ctx, namespace, 1, logger); err != nil {
			logger.Warn("Failed to scale up environment after snapshot", zap.Error(err))
		}
	} else {
		logger.Info("Skipping scale up, other snapshots still in progress")
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

	// Scale down environment to stop all pods before restore
	if err := r.scaleEnvironment(ctx, namespace, 0, logger); err != nil {
		logger.Warn("Failed to scale down environment", zap.Error(err))
	}

	// Wait for pods to terminate
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate before restore")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Stopping environment pods before restore..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

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
			// Scale up and fail
			if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
				logger.Warn("Failed to scale up environment after restore failure", zap.Error(scaleErr))
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

			// Scale environment back up
			if !r.hasOtherInProgressSnapshots(ctx, snapshot, envName, logger) {
				if err := r.scaleEnvironment(ctx, namespace, 1, logger); err != nil {
					logger.Warn("Failed to scale up environment after restore", zap.Error(err))
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
			SourcePath:      snapshotPath,  // The btrfs snapshot
			SnapshotPath:    targetPath,    // The environment directory
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
			// Scale environment back up on failure
			if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
				logger.Warn("Failed to scale up environment after restore request creation failure", zap.Error(scaleErr))
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

// scaleEnvironment scales all deployments and statefulsets in a namespace to the specified replica count
func (r *SnapshotReconciler) scaleEnvironment(ctx context.Context, namespace string, replicas int32, logger *zap.Logger) error {
	// Scale deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}
	for _, deploy := range deployments.Items {
		if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas != replicas {
			deploy.Spec.Replicas = &replicas
			if err := r.Update(ctx, &deploy); err != nil {
				logger.Warn("Failed to scale deployment", zap.String("deployment", deploy.Name), zap.Error(err))
			} else {
				logger.Info("Scaled deployment", zap.String("deployment", deploy.Name), zap.Int32("replicas", replicas))
			}
		}
	}

	// Scale statefulsets
	statefulsets := &appsv1.StatefulSetList{}
	if err := r.List(ctx, statefulsets, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list statefulsets: %w", err)
	}
	for _, sts := range statefulsets.Items {
		if sts.Spec.Replicas != nil && *sts.Spec.Replicas != replicas {
			sts.Spec.Replicas = &replicas
			if err := r.Update(ctx, &sts); err != nil {
				logger.Warn("Failed to scale statefulset", zap.String("statefulset", sts.Name), zap.Error(err))
			} else {
				logger.Info("Scaled statefulset", zap.String("statefulset", sts.Name), zap.Int32("replicas", replicas))
			}
		}
	}

	return nil
}

// waitForPodsTerminated waits for all pods in a namespace to terminate
func (r *SnapshotReconciler) waitForPodsTerminated(ctx context.Context, namespace string, logger *zap.Logger) bool {
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(namespace)); err != nil {
		logger.Warn("Failed to list pods", zap.Error(err))
		return false
	}
	// Check if any non-terminated pods exist (excluding jobs/completed pods)
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			return false
		}
	}
	return true
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
