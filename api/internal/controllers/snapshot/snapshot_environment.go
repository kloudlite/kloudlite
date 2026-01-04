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

	// Generate snapshot path once and store it in status (reuse if already set)
	// Use snapshot name (which is unique) to prevent path collisions with concurrent snapshots
	snapshotPath := snapshot.Status.SnapshotPath
	if snapshotPath == "" {
		snapshotPath = filepath.Join(snapshotsBasePath, snapshot.Name)
		// Store the path immediately so subsequent reconciles use the same path
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.SnapshotPath = snapshotPath
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

	// List PVCs in the environment namespace
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, client.InNamespace(namespace)); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		// Scale environment back up on failure
		if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
			logger.Warn("Failed to scale up environment after PVC list failure", zap.Error(scaleErr))
		}
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
	}

	// Create SnapshotRequest for each PVC
	var pvcSnapshots []snapshotv1.PVCSnapshotInfo
	for _, pvc := range pvcList.Items {
		// Get the actual PV path from the PersistentVolume
		pvName := pvc.Spec.VolumeName
		if pvName == "" {
			logger.Warn("PVC has no bound PV, skipping", zap.String("pvc", pvc.Name))
			continue
		}

		pv := &corev1.PersistentVolume{}
		if err := r.Get(ctx, client.ObjectKey{Name: pvName}, pv); err != nil {
			logger.Error("Failed to get PV", zap.Error(err), zap.String("pv", pvName))
			continue
		}

		// Get the actual host path from the PV (local-path-provisioner uses spec.local.path)
		var sourcePath string
		if pv.Spec.Local != nil && pv.Spec.Local.Path != "" {
			sourcePath = pv.Spec.Local.Path
		} else if pv.Spec.HostPath != nil && pv.Spec.HostPath.Path != "" {
			sourcePath = pv.Spec.HostPath.Path
		} else {
			logger.Warn("PV has no local or hostPath, skipping", zap.String("pv", pvName))
			continue
		}

		pvcSnapshotPath := filepath.Join(snapshotPath, "pvcs", pvc.Name)

		// Create SnapshotRequest
		snapshotReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", snapshot.Name, pvc.Name),
				Namespace: fmt.Sprintf("wm-%s", env.Spec.OwnedBy),
				Labels: map[string]string{
					"snapshots.kloudlite.io/snapshot": snapshot.Name,
					"snapshots.kloudlite.io/pvc":      pvc.Name,
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:       snapshotv1.SnapshotOperationCreate,
				SourcePath:      sourcePath,
				SnapshotPath:    pvcSnapshotPath,
				SnapshotRef:     snapshot.Name,
				EnvironmentName: envName,
				ReadOnly:        true,
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(snapshot, snapshotReq, r.Scheme); err != nil {
			logger.Error("Failed to set owner reference", zap.Error(err))
		}

		// Create or update SnapshotRequest
		if err := r.Create(ctx, snapshotReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create SnapshotRequest", zap.Error(err), zap.String("pvc", pvc.Name))
				// Scale environment back up on failure
				if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
					logger.Warn("Failed to scale up environment after SnapshotRequest creation failure", zap.Error(scaleErr))
				}
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create SnapshotRequest for PVC %s", pvc.Name), logger)
			}
		}

		pvcSnapshots = append(pvcSnapshots, snapshotv1.PVCSnapshotInfo{
			PVCName:      pvc.Name,
			SnapshotPath: pvcSnapshotPath,
			SourcePath:   sourcePath,
		})
	}

	// Collect K8s metadata if requested
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

	// Check if all SnapshotRequests are complete
	allComplete, err := r.checkSnapshotRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check SnapshotRequest status", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !allComplete {
		// Update status with progress
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Creating btrfs snapshots..."
			snapshot.Status.SnapshotPath = snapshotPath
			snapshot.Status.PVCSnapshots = pvcSnapshots
			snapshot.Status.ResourceMetadata = resourceMetadata
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status with progress", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// All snapshots complete - get sizes from completed SnapshotRequests
	snapshotReqNamespace := fmt.Sprintf("wm-%s", env.Spec.OwnedBy)
	for i, pvcInfo := range pvcSnapshots {
		snapshotReqName := fmt.Sprintf("%s-%s", snapshot.Name, pvcInfo.PVCName)
		snapshotReq := &snapshotv1.SnapshotRequest{}
		if err := r.Get(ctx, client.ObjectKey{Name: snapshotReqName, Namespace: snapshotReqNamespace}, snapshotReq); err == nil {
			pvcSnapshots[i].SizeBytes = snapshotReq.Status.SizeBytes
		} else {
			logger.Warn("Failed to get SnapshotRequest for size", zap.String("name", snapshotReqName), zap.Error(err))
		}
	}

	// Calculate total size
	var totalSize int64
	for _, pvcInfo := range pvcSnapshots {
		totalSize += pvcInfo.SizeBytes
	}

	// Update environment's lastRestoredSnapshot BEFORE setting snapshot to Ready
	// This ensures the frontend sees the current snapshot when it refreshes
	if err := r.updateEnvironmentLastRestored(ctx, envName, snapshot.Name, logger); err != nil {
		logger.Warn("Failed to update environment's lastRestoredSnapshot", zap.Error(err))
		// Continue - this is not a fatal error
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
		snapshot.Status.PVCSnapshots = pvcSnapshots
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

	// Check existing restore requests
	existingRestoreReqs := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, existingRestoreReqs, client.MatchingLabels{
		"snapshots.kloudlite.io/snapshot":  snapshot.Name,
		"snapshots.kloudlite.io/operation": "restore",
	}); err == nil && len(existingRestoreReqs.Items) > 0 {
		// Check for in-progress or pending requests - these are from THIS restore attempt
		anyInProgress := false
		anyPending := false
		allCompleted := true
		for _, req := range existingRestoreReqs.Items {
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
				// Scale up and fail
				if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
					logger.Warn("Failed to scale up environment after restore failure", zap.Error(scaleErr))
				}
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Restore failed: %s", req.Status.Message), logger)
			}
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseInProgress {
				anyInProgress = true
				allCompleted = false
			} else if req.Status.Phase == snapshotv1.SnapshotRequestPhasePending || req.Status.Phase == "" {
				anyPending = true
				allCompleted = false
			} else if req.Status.Phase != snapshotv1.SnapshotRequestPhaseCompleted {
				allCompleted = false
			}
		}

		if anyInProgress || anyPending {
			// Active restore in progress, wait for it
			logger.Info("Restore in progress, waiting...", zap.Bool("inProgress", anyInProgress), zap.Bool("pending", anyPending))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}

		if allCompleted {
			// All restore requests completed - mark snapshot as ready and scale up
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
				snapshot.Status.State = snapshotv1.SnapshotStateReady
				snapshot.Status.Message = "Snapshot restored successfully"
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update status to Ready", zap.Error(err))
				return reconcile.Result{}, err
			}
			if !r.hasOtherInProgressSnapshots(ctx, snapshot, envName, logger) {
				if err := r.scaleEnvironment(ctx, namespace, 1, logger); err != nil {
					logger.Warn("Failed to scale up environment after restore", zap.Error(err))
				}
			}
			// Delete the completed restore requests to allow future restores
			for _, req := range existingRestoreReqs.Items {
				if err := r.Delete(ctx, &req); err != nil && !apierrors.IsNotFound(err) {
					logger.Warn("Failed to delete completed restore request", zap.Error(err), zap.String("request", req.Name))
				}
			}
			// Track this restore on the environment for parent lineage
			if err := r.updateEnvironmentLastRestored(ctx, envName, snapshot.Name, logger); err != nil {
				logger.Warn("Failed to update environment's lastRestoredSnapshot", zap.Error(err))
				// Continue - this is not a fatal error
			}
			logger.Info("Snapshot restored successfully", zap.String("environment", envName))
			return reconcile.Result{}, nil
		}
	}

	// Create restore SnapshotRequests for each PVC snapshot
	for _, pvcInfo := range snapshot.Status.PVCSnapshots {
		// Target path is the original PV location (stored in SourcePath)
		targetPath := pvcInfo.SourcePath
		if targetPath == "" {
			// Fallback: look up PV path if not stored in snapshot
			pvcList := &corev1.PersistentVolumeClaimList{}
			if err := r.List(ctx, pvcList, client.InNamespace(namespace)); err == nil {
				for _, pvc := range pvcList.Items {
					if pvc.Name == pvcInfo.PVCName && pvc.Spec.VolumeName != "" {
						pv := &corev1.PersistentVolume{}
						if err := r.Get(ctx, client.ObjectKey{Name: pvc.Spec.VolumeName}, pv); err == nil {
							if pv.Spec.Local != nil && pv.Spec.Local.Path != "" {
								targetPath = pv.Spec.Local.Path
							} else if pv.Spec.HostPath != nil && pv.Spec.HostPath.Path != "" {
								targetPath = pv.Spec.HostPath.Path
							}
						}
						break
					}
				}
			}
		}
		if targetPath == "" {
			logger.Error("Could not determine target path for PVC", zap.String("pvc", pvcInfo.PVCName))
			// Scale environment back up on failure
			if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
				logger.Warn("Failed to scale up environment after target path lookup failure", zap.Error(scaleErr))
			}
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Could not determine target path for PVC %s", pvcInfo.PVCName), logger)
		}

		restoreReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-restore-%s", snapshot.Name, pvcInfo.PVCName),
				Namespace: snapshot.Status.WorkMachineName,
				Labels: map[string]string{
					"snapshots.kloudlite.io/snapshot":  snapshot.Name,
					"snapshots.kloudlite.io/operation": "restore",
					"snapshots.kloudlite.io/pvc":       pvcInfo.PVCName,
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:       snapshotv1.SnapshotOperationRestore,
				SourcePath:      pvcInfo.SnapshotPath,
				SnapshotPath:    targetPath,
				SnapshotRef:     snapshot.Name,
				EnvironmentName: envName,
			},
		}

		if err := r.Create(ctx, restoreReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create restore SnapshotRequest", zap.Error(err), zap.String("pvc", pvcInfo.PVCName))
				// Scale environment back up on failure
				if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
					logger.Warn("Failed to scale up environment after restore request creation failure", zap.Error(scaleErr))
				}
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create restore request for PVC %s", pvcInfo.PVCName), logger)
			}
		}
	}

	// Check if all restore requests are complete
	allComplete, err := r.checkRestoreRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check restore request status", zap.Error(err))
		// Scale environment back up even on failure
		if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
			logger.Warn("Failed to scale up environment after restore failure", zap.Error(scaleErr))
		}
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Restore failed: %v", err), logger)
	}

	if !allComplete {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Restore complete - update status back to Ready
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Snapshot restored successfully"
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Scale environment back up only if no other snapshots are in progress
	if !r.hasOtherInProgressSnapshots(ctx, snapshot, envName, logger) {
		if err := r.scaleEnvironment(ctx, namespace, 1, logger); err != nil {
			logger.Warn("Failed to scale up environment after restore", zap.Error(err))
		}
	} else {
		logger.Info("Skipping scale up after restore, other snapshots still in progress")
	}

	logger.Info("Snapshot restored successfully", zap.String("environment", envName))
	return reconcile.Result{}, nil
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
