package snapshot

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleWorkspacePending starts the workspace snapshot creation process
func (r *SnapshotReconciler) handleWorkspacePending(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	wsRef := snapshot.Spec.WorkspaceRef
	wmNamespace := fmt.Sprintf("wm-%s", snapshot.Spec.OwnedBy)

	logger.Info("Starting workspace snapshot",
		zap.String("workspace", wsRef.Name),
		zap.String("workmachine", wsRef.WorkmachineName))

	// Fetch the workspace
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err != nil {
		logger.Error("Failed to get workspace", zap.Error(err), zap.String("workspace", wsRef.Name))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Workspace not found: %s", wsRef.Name), logger)
	}

	// Validate ownership
	if workspace.Spec.OwnedBy != snapshot.Spec.OwnedBy {
		return r.updateStatusFailed(ctx, snapshot, "Workspace is not owned by the snapshot creator", logger)
	}

	// Ensure labels are set for querying by workspace
	labelsUpdated := false
	if snapshot.Labels == nil {
		snapshot.Labels = make(map[string]string)
	}
	if snapshot.Labels["snapshots.kloudlite.io/workspace"] != wsRef.Name {
		snapshot.Labels["snapshots.kloudlite.io/workspace"] = wsRef.Name
		labelsUpdated = true
	}
	if snapshot.Labels["kloudlite.io/owned-by"] != snapshot.Spec.OwnedBy {
		snapshot.Labels["kloudlite.io/owned-by"] = snapshot.Spec.OwnedBy
		labelsUpdated = true
	}
	if labelsUpdated {
		if err := r.Update(ctx, snapshot); err != nil {
			logger.Error("Failed to update snapshot labels", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Store previous workspace status and suspend the workspace
	previousStatus := workspace.Spec.Status
	wasSuspended := previousStatus == "suspended"

	if !wasSuspended {
		// Suspend the workspace
		logger.Info("Suspending workspace for snapshot", zap.String("workspace", wsRef.Name))
		workspace.Spec.Status = "suspended"
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to suspend workspace", zap.Error(err))
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to suspend workspace: %v", err), logger)
		}
	}

	// Set state to Creating
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateCreating
		snapshot.Status.Message = "Suspending workspace and preparing to create snapshot"
		snapshot.Status.WorkMachineName = wmNamespace
		snapshot.Status.SnapshotType = snapshotv1.SnapshotTypeWorkspace
		snapshot.Status.TargetName = wsRef.Name
		snapshot.Status.WorkspaceName = wsRef.Name
		snapshot.Status.WorkspaceWasSuspended = wasSuspended
		snapshot.Status.PreviousWorkspaceStatus = previousStatus
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Creating", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleWorkspaceCreating manages the workspace snapshot creation process
func (r *SnapshotReconciler) handleWorkspaceCreating(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	wsRef := snapshot.Spec.WorkspaceRef
	wmNamespace := snapshot.Status.WorkMachineName

	logger.Info("Creating workspace snapshot",
		zap.String("workspace", wsRef.Name),
		zap.String("namespace", wmNamespace))

	// Fetch the workspace to check pod status
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		return r.handleWorkspaceSnapshotFailure(ctx, snapshot, fmt.Sprintf("Workspace not found: %s", wsRef.Name), logger)
	}

	// Wait for workspace pod to terminate
	podName := fmt.Sprintf("workspace-%s", wsRef.Name)
	pod := &corev1.Pod{}
	podErr := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: wmNamespace}, pod)

	if podErr == nil {
		// Pod still exists, wait for it to terminate
		logger.Info("Waiting for workspace pod to terminate", zap.String("pod", podName))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Waiting for workspace pod to terminate..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status message", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if !apierrors.IsNotFound(podErr) {
		logger.Error("Failed to check workspace pod", zap.Error(podErr))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Pod is terminated, proceed with snapshot

	// Generate snapshot path once and store it in status (reuse if already set)
	// Use snapshot name (which is unique) to prevent path collisions with concurrent snapshots
	snapshotPath := snapshot.Status.SnapshotPath
	if snapshotPath == "" {
		snapshotPath = filepath.Join(snapshotsBasePath, snapshot.Name)
		// Store the path immediately so subsequent reconciles use the same path
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.SnapshotPath = snapshotPath
			snapshot.Status.Message = "Preparing workspace snapshot..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store snapshot path", zap.Error(err))
		}
	}

	// Workspace home directory path
	sourcePath := filepath.Join(workspaceHomePath, wsRef.Name)
	workspaceSnapshotPath := filepath.Join(snapshotPath, "home")

	// Create SnapshotRequest for workspace home directory
	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-home", snapshot.Name),
			Namespace: wmNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/workspace": wsRef.Name,
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:     snapshotv1.SnapshotOperationCreate,
			SourcePath:    sourcePath,
			SnapshotPath:  workspaceSnapshotPath,
			SnapshotRef:   snapshot.Name,
			WorkspaceName: wsRef.Name,
			ReadOnly:      true,
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
			return r.handleWorkspaceSnapshotFailure(ctx, snapshot, "Failed to create SnapshotRequest for workspace home", logger)
		}
	}

	// Export workspace metadata if requested
	var packageRequestsPath string
	if snapshot.Spec.IncludeMetadata {
		var err error
		packageRequestsPath, err = r.exportWorkspaceMetadata(ctx, wsRef.Name, wmNamespace, snapshotPath, logger)
		if err != nil {
			logger.Warn("Failed to export workspace metadata, continuing anyway", zap.Error(err))
		}
	}

	// Check if SnapshotRequest is complete (workspace has 1 SnapshotRequest for home directory)
	allComplete, err := r.checkSnapshotRequestsComplete(ctx, snapshot, 1, logger)
	if err != nil {
		logger.Error("Failed to check SnapshotRequest status", zap.Error(err))
		return r.handleWorkspaceSnapshotFailure(ctx, snapshot, fmt.Sprintf("SnapshotRequest failed: %v", err), logger)
	}

	if !allComplete {
		// Update status with progress
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Creating workspace snapshot..."
			snapshot.Status.SnapshotPath = snapshotPath
			snapshot.Status.PackageRequestsPath = packageRequestsPath
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status with progress", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Snapshot complete - get size from SnapshotRequest
	var totalSize int64
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{"snapshots.kloudlite.io/snapshot": snapshot.Name}); err == nil {
		for _, req := range snapshotReqList.Items {
			totalSize += req.Status.SizeBytes
		}
	}

	// Resume workspace if it wasn't already suspended
	if !snapshot.Status.WorkspaceWasSuspended {
		logger.Info("Resuming workspace after snapshot", zap.String("workspace", wsRef.Name))

		// Refetch workspace to get latest version
		if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err == nil {
			originalStatus := snapshot.Status.PreviousWorkspaceStatus
			if originalStatus == "" {
				originalStatus = "active"
			}
			workspace.Spec.Status = originalStatus
			if err := r.Update(ctx, workspace); err != nil {
				logger.Warn("Failed to resume workspace", zap.Error(err))
				// Don't fail the snapshot - it was created successfully
			}
		}
	}

	// Update workspace's lastRestoredSnapshot BEFORE setting snapshot to Ready
	// This ensures the frontend sees the current snapshot when it refreshes
	if err := r.updateWorkspaceLastRestored(ctx, wsRef.Name, wmNamespace, snapshot.Name, logger); err != nil {
		logger.Warn("Failed to update workspace's lastRestoredSnapshot", zap.Error(err))
		// Continue - this is not a fatal error
	}

	// Update status to Ready
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Workspace snapshot created successfully"
		snapshot.Status.SnapshotPath = snapshotPath
		snapshot.Status.SizeBytes = totalSize
		snapshot.Status.SizeHuman = formatSize(totalSize)
		snapshot.Status.CreatedAt = &now
		snapshot.Status.PackageRequestsPath = packageRequestsPath
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Workspace snapshot created successfully",
		zap.String("workspace", wsRef.Name),
		zap.String("path", snapshotPath),
		zap.Int64("sizeBytes", totalSize))

	return reconcile.Result{}, nil
}

// handleWorkspaceSnapshotFailure handles a workspace snapshot failure and resumes the workspace
func (r *SnapshotReconciler) handleWorkspaceSnapshotFailure(ctx context.Context, snapshot *snapshotv1.Snapshot, message string, logger *zap.Logger) (reconcile.Result, error) {
	// Try to resume workspace if we suspended it
	if !snapshot.Status.WorkspaceWasSuspended && snapshot.Spec.WorkspaceRef != nil {
		wsRef := snapshot.Spec.WorkspaceRef
		wmNamespace := snapshot.Status.WorkMachineName

		workspace := &workspacev1.Workspace{}
		if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err == nil {
			originalStatus := snapshot.Status.PreviousWorkspaceStatus
			if originalStatus == "" {
				originalStatus = "active"
			}
			workspace.Spec.Status = originalStatus
			if err := r.Update(ctx, workspace); err != nil {
				logger.Warn("Failed to resume workspace after snapshot failure", zap.Error(err))
			} else {
				logger.Info("Resumed workspace after snapshot failure", zap.String("workspace", wsRef.Name))
			}
		}
	}

	return r.updateStatusFailed(ctx, snapshot, message, logger)
}

// handleWorkspaceRestoring handles workspace snapshot restore
func (r *SnapshotReconciler) handleWorkspaceRestoring(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	wsRef := snapshot.Spec.WorkspaceRef
	if wsRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Workspace reference is required for workspace restore", logger)
	}

	wmNamespace := snapshot.Status.WorkMachineName
	if wmNamespace == "" {
		// Derive namespace from owner if not set in status
		wmNamespace = fmt.Sprintf("wm-%s", snapshot.Spec.OwnedBy)
	}

	workspaceSnapshotPath := filepath.Join(snapshot.Status.SnapshotPath, "home")
	targetPath := filepath.Join(workspaceHomePath, wsRef.Name)

	// Fetch the workspace
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err != nil {
		logger.Error("Failed to get workspace for restore", zap.Error(err))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Workspace not found: %s", wsRef.Name), logger)
	}

	// Store previous status if not already stored
	if snapshot.Status.PreviousWorkspaceStatus == "" && workspace.Spec.Status != "suspended" {
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.PreviousWorkspaceStatus = workspace.Spec.Status
			snapshot.Status.WorkspaceWasSuspended = false
			snapshot.Status.WorkMachineName = wmNamespace
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store previous workspace status", zap.Error(err))
		}
	}

	// Suspend workspace if not already suspended (like scaleEnvironment for environments)
	if workspace.Spec.Status != "suspended" {
		logger.Info("Suspending workspace before restore", zap.String("workspace", wsRef.Name))
		workspace.Spec.Status = "suspended"
		if err := r.Update(ctx, workspace); err != nil {
			logger.Warn("Failed to suspend workspace, will retry", zap.Error(err))
		}
	}

	// Wait for workspace pod to terminate (like waitForPodsTerminated for environments)
	podName := fmt.Sprintf("workspace-%s", wsRef.Name)
	pod := &corev1.Pod{}
	podErr := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: wmNamespace}, pod)

	if podErr == nil {
		// Pod still exists, wait for it to terminate
		logger.Info("Waiting for workspace pod to terminate before restore", zap.String("pod", podName))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Stopping workspace pod before restore..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status message", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if !apierrors.IsNotFound(podErr) {
		logger.Error("Failed to check workspace pod for restore", zap.Error(podErr))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
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
			logger.Info("Workspace restore in progress, waiting...", zap.Bool("inProgress", anyInProgress), zap.Bool("pending", anyPending))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}

		if allCompleted {
			// Resume workspace if we suspended it
			if !snapshot.Status.WorkspaceWasSuspended {
				logger.Info("Resuming workspace after restore", zap.String("workspace", wsRef.Name))

				// Refetch workspace to get latest version
				if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err == nil {
					originalStatus := snapshot.Status.PreviousWorkspaceStatus
					if originalStatus == "" {
						originalStatus = "active"
					}
					workspace.Spec.Status = originalStatus
					if err := r.Update(ctx, workspace); err != nil {
						logger.Warn("Failed to resume workspace after restore", zap.Error(err))
						// Don't fail - restore was successful
					}
				}
			}

			// Restore complete - update status back to Ready
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
				snapshot.Status.State = snapshotv1.SnapshotStateReady
				snapshot.Status.Message = "Workspace snapshot restored successfully"
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update status to Ready", zap.Error(err))
				return reconcile.Result{}, err
			}
			// Delete the completed restore requests to allow future restores
			for _, req := range existingRestoreReqs.Items {
				if err := r.Delete(ctx, &req); err != nil && !apierrors.IsNotFound(err) {
					logger.Warn("Failed to delete completed restore request", zap.Error(err), zap.String("request", req.Name))
				}
			}
			// Track this restore on the workspace for parent lineage
			if err := r.updateWorkspaceLastRestored(ctx, wsRef.Name, wmNamespace, snapshot.Name, logger); err != nil {
				logger.Warn("Failed to update workspace's lastRestoredSnapshot", zap.Error(err))
				// Continue - this is not a fatal error
			}
			logger.Info("Workspace snapshot restored successfully", zap.String("workspace", wsRef.Name))
			return reconcile.Result{}, nil
		}
	}

	// Create restore SnapshotRequest
	restoreReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-restore-home", snapshot.Name),
			Namespace: wmNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/operation": "restore",
				"snapshots.kloudlite.io/workspace": wsRef.Name,
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:     snapshotv1.SnapshotOperationRestore,
			SourcePath:    workspaceSnapshotPath,
			SnapshotPath:  targetPath,
			SnapshotRef:   snapshot.Name,
			WorkspaceName: wsRef.Name,
		},
	}

	if err := r.Create(ctx, restoreReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create restore SnapshotRequest", zap.Error(err))
			return r.updateStatusFailed(ctx, snapshot, "Failed to create restore request for workspace", logger)
		}
	}

	// Requeue to check status
	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

// exportWorkspaceMetadata exports workspace-specific metadata (PackageRequests, settings)
func (r *SnapshotReconciler) exportWorkspaceMetadata(ctx context.Context, workspaceName, wmNamespace, snapshotPath string, logger *zap.Logger) (string, error) {
	metadataPath := filepath.Join(snapshotPath, "metadata")

	// Create metadata directory
	if err := createDir(metadataPath); err != nil {
		return "", fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Export PackageRequests for this workspace
	packageRequests := &packagesv1.PackageRequestList{}
	if err := r.List(ctx, packageRequests, client.InNamespace(wmNamespace)); err == nil {
		// Filter to only include PackageRequests for this workspace
		var filtered []packagesv1.PackageRequest
		for _, pr := range packageRequests.Items {
			if pr.Spec.WorkspaceRef == workspaceName {
				filtered = append(filtered, pr)
			}
		}

		packageRequestsPath := filepath.Join(metadataPath, "package-requests.json")
		if err := exportToJSON(packageRequestsPath, filtered); err != nil {
			logger.Warn("Failed to export PackageRequests", zap.Error(err))
		} else {
			logger.Info("Exported PackageRequests", zap.Int("count", len(filtered)))
		}
	}

	// Export Workspace resource itself (settings, config, etc.)
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspaceName, Namespace: wmNamespace}, workspace); err == nil {
		workspacePath := filepath.Join(metadataPath, "workspace.json")
		if err := exportToJSON(workspacePath, workspace); err != nil {
			logger.Warn("Failed to export Workspace", zap.Error(err))
		} else {
			logger.Info("Exported Workspace settings")
		}
	}

	return metadataPath, nil
}
