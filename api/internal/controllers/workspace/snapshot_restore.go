package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotsBasePath = "/kl-data/snapshots"
)

// handleSnapshotRestore handles creating a workspace from a pushed snapshot
// This function orchestrates the entire snapshot restoration workflow through various phases
func (r *WorkspaceReconciler) handleSnapshotRestore(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Handling workspace snapshot restore",
		zap.String("workspace", workspace.Name),
		zap.String("snapshotName", workspace.Spec.FromSnapshot.SnapshotName))

	// Initialize snapshot restore status if not set
	if workspace.Status.SnapshotRestoreStatus == nil {
		logger.Info("Initializing snapshot restore status")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
			workspace.Status.SnapshotRestoreStatus = &workspacev1.SnapshotRestoreStatus{
				Phase:          workspacev1.SnapshotRestorePhasePending,
				Message:        "Snapshot restore initialized",
				SourceSnapshot: workspace.Spec.FromSnapshot.SnapshotName,
				StartTime:      fn.Ptr(metav1.Now()),
			}
			return nil
		}, logger); err != nil {
			logger.Error("Failed to initialize snapshot restore status", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	status := workspace.Status.SnapshotRestoreStatus

	// Handle based on current phase
	switch status.Phase {
	case workspacev1.SnapshotRestorePhasePending:
		return r.handleRestorePending(ctx, workspace, logger)

	case workspacev1.SnapshotRestorePhasePulling:
		return r.handleRestorePulling(ctx, workspace, logger)

	case workspacev1.SnapshotRestorePhaseRestoring:
		return r.handleRestoreRestoring(ctx, workspace, logger)

	case workspacev1.SnapshotRestorePhaseCompleted:
		return r.handleRestoreCompleted(ctx, workspace, logger)

	case workspacev1.SnapshotRestorePhaseFailed:
		// Restoration failed, don't retry automatically
		logger.Info("Snapshot restore failed, not retrying",
			zap.String("error", status.ErrorMessage))
		return reconcile.Result{}, nil

	default:
		logger.Error("Unknown snapshot restore phase", zap.String("phase", string(status.Phase)))
		return reconcile.Result{}, nil
	}
}

// handleRestorePending validates the snapshot exists and is pushed, then moves to Pulling phase
func (r *WorkspaceReconciler) handleRestorePending(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Pending - Validating snapshot")

	snapshotName := workspace.Spec.FromSnapshot.SnapshotName

	// Fetch the snapshot to validate it exists and is pushed
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return r.failSnapshotRestore(ctx, workspace, fmt.Sprintf("Snapshot %s not found", snapshotName), logger)
		}
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Validate snapshot is pushed to registry
	if snapshot.Status.RegistryStatus == nil || !snapshot.Status.RegistryStatus.Pushed {
		return r.failSnapshotRestore(ctx, workspace,
			fmt.Sprintf("Snapshot %s is not pushed to registry", snapshotName), logger)
	}

	// Validate snapshot type is workspace
	if snapshot.Status.SnapshotType != snapshotv1.SnapshotTypeWorkspace {
		return r.failSnapshotRestore(ctx, workspace,
			fmt.Sprintf("Snapshot %s is not a workspace snapshot", snapshotName), logger)
	}

	// Store snapshot details and move to Pulling phase
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
		workspace.Status.SnapshotRestoreStatus.Phase = workspacev1.SnapshotRestorePhasePulling
		workspace.Status.SnapshotRestoreStatus.Message = "Pulling snapshot from registry"
		workspace.Status.SnapshotRestoreStatus.ImageRef = snapshot.Status.RegistryStatus.ImageRef
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update snapshot restore status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot validated, moving to Pulling phase",
		zap.String("imageRef", snapshot.Status.RegistryStatus.ImageRef))
	return reconcile.Result{Requeue: true}, nil
}

// handleRestorePulling creates a SnapshotRequest to pull the snapshot from registry
func (r *WorkspaceReconciler) handleRestorePulling(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Pulling - Pulling snapshot from registry")

	status := workspace.Status.SnapshotRestoreStatus
	snapshotName := workspace.Spec.FromSnapshot.SnapshotName

	// Check for existing pull request
	pullReqName := fmt.Sprintf("%s-restore-pull", workspace.Name)
	if status.SnapshotRequestName == "" {
		// Create new SnapshotRequest for pulling
		// Get the workspace's target namespace from WorkMachine
		targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
		if err != nil {
			logger.Error("Failed to get workspace target namespace", zap.Error(err))
			return r.failSnapshotRestore(ctx, workspace,
				fmt.Sprintf("Failed to get target namespace: %v", err), logger)
		}

		// Determine snapshot path for the pulled data
		// Use workspace name to make it unique
		snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("ws-%s-restore", workspace.Name))

		// Parse imageRef to get repository and tag
		repository, tag := parseImageRef(status.ImageRef)

		pullReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pullReqName,
				Namespace: targetNamespace,
				Labels: map[string]string{
					"workspaces.kloudlite.io/workspace": workspace.Name,
					"workspaces.kloudlite.io/operation": "restore-pull",
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:    snapshotv1.SnapshotOperationPull,
				SnapshotPath: snapshotPath,
				SnapshotRef:  snapshotName,
				RegistryRef: &snapshotv1.SnapshotRequestRegistryRef{
					RegistryURL: "image-registry.kloudlite.svc.cluster.local:5000",
					Repository:  repository,
					Tag:         tag,
				},
			},
		}

		if err := r.Create(ctx, pullReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create pull SnapshotRequest", zap.Error(err))
				return r.failSnapshotRestore(ctx, workspace,
					fmt.Sprintf("Failed to create pull request: %v", err), logger)
			}
		}

		// Update status with SnapshotRequest name
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
			workspace.Status.SnapshotRestoreStatus.SnapshotRequestName = pullReqName
			workspace.Status.SnapshotRestoreStatus.Message = "Pull request created, waiting for completion"
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status with pull request name", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Created pull SnapshotRequest", zap.String("name", pullReqName))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Check status of existing pull request
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		logger.Error("Failed to get workspace target namespace", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	pullReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: status.SnapshotRequestName, Namespace: targetNamespace}, pullReq); err != nil {
		if apierrors.IsNotFound(err) {
			// Request was deleted, reset and retry
			logger.Warn("Pull request not found, resetting")
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
				workspace.Status.SnapshotRestoreStatus.SnapshotRequestName = ""
				return nil
			}, logger); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to get pull request", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	switch pullReq.Status.Phase {
	case snapshotv1.SnapshotRequestPhaseFailed:
		return r.failSnapshotRestore(ctx, workspace,
			fmt.Sprintf("Pull failed: %s", pullReq.Status.Message), logger)

	case snapshotv1.SnapshotRequestPhaseCompleted:
		// Pull completed, move to Restoring phase
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
			workspace.Status.SnapshotRestoreStatus.Phase = workspacev1.SnapshotRestorePhaseRestoring
			workspace.Status.SnapshotRestoreStatus.Message = "Snapshot pulled, restoring packages"
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status after pull", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Delete the completed pull request
		if err := r.Delete(ctx, pullReq); err != nil && !apierrors.IsNotFound(err) {
			logger.Warn("Failed to delete completed pull request", zap.Error(err))
		}

		logger.Info("Pull completed, moving to Restoring phase")
		return reconcile.Result{Requeue: true}, nil

	default:
		// Still in progress
		logger.Info("Pull in progress", zap.String("phase", string(pullReq.Status.Phase)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
}

// handleRestoreRestoring restores workspace data and packages from snapshot
func (r *WorkspaceReconciler) handleRestoreRestoring(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Restoring - Restoring workspace data from snapshot")

	status := workspace.Status.SnapshotRestoreStatus
	snapshotName := workspace.Spec.FromSnapshot.SnapshotName

	// Check if we need to create a restore request for workspace data
	restoreReqName := fmt.Sprintf("%s-restore-data", workspace.Name)
	if status.DataRestoreRequestName == "" {
		// Create SnapshotRequest to restore workspace data
		targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
		if err != nil {
			logger.Error("Failed to get workspace target namespace", zap.Error(err))
			return r.failSnapshotRestore(ctx, workspace,
				fmt.Sprintf("Failed to get target namespace: %v", err), logger)
		}

		// Source: pulled snapshot home directory
		// e.g., /kl-data/snapshots/ws-main-clone-btrfs-restore/ws-main-btrfs-test/home
		sourcePath := filepath.Join(snapshotsBasePath, fmt.Sprintf("ws-%s-restore", workspace.Name), snapshotName, "home")

		// Target: workspace btrfs storage path
		// e.g., /var/lib/kloudlite/storage/workspaces/main-clone-btrfs
		targetPath := fmt.Sprintf("/var/lib/kloudlite/storage/workspaces/%s", workspace.Name)

		restoreReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      restoreReqName,
				Namespace: targetNamespace,
				Labels: map[string]string{
					"workspaces.kloudlite.io/workspace": workspace.Name,
					"workspaces.kloudlite.io/operation": "restore-data",
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:     snapshotv1.SnapshotOperationRestore,
				SourcePath:    sourcePath, // Restore FROM pulled snapshot
				SnapshotPath:  targetPath, // Restore TO workspace directory
				SnapshotRef:   snapshotName,
				WorkspaceName: workspace.Name,
			},
		}

		if err := r.Create(ctx, restoreReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create restore SnapshotRequest", zap.Error(err))
				return r.failSnapshotRestore(ctx, workspace,
					fmt.Sprintf("Failed to create restore request: %v", err), logger)
			}
		}

		// Update status with restore request name
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
			workspace.Status.SnapshotRestoreStatus.DataRestoreRequestName = restoreReqName
			workspace.Status.SnapshotRestoreStatus.Message = "Restoring workspace data"
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status with restore request name", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Created data restore SnapshotRequest",
			zap.String("name", restoreReqName),
			zap.String("sourcePath", sourcePath),
			zap.String("targetPath", targetPath))
		return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// Check status of existing restore request
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		logger.Error("Failed to get workspace target namespace", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	restoreReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: status.DataRestoreRequestName, Namespace: targetNamespace}, restoreReq); err != nil {
		if apierrors.IsNotFound(err) {
			// Request was deleted, reset and retry
			logger.Warn("Restore request not found, resetting")
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
				workspace.Status.SnapshotRestoreStatus.DataRestoreRequestName = ""
				return nil
			}, logger); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to get restore request", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	switch restoreReq.Status.Phase {
	case snapshotv1.SnapshotRequestPhaseFailed:
		return r.failSnapshotRestore(ctx, workspace,
			fmt.Sprintf("Data restore failed: %s", restoreReq.Status.Message), logger)

	case snapshotv1.SnapshotRequestPhaseCompleted:
		// Data restore completed, now restore packages
		logger.Info("Data restore completed, restoring packages")

		// Delete the completed restore request
		if err := r.Delete(ctx, restoreReq); err != nil && !apierrors.IsNotFound(err) {
			logger.Warn("Failed to delete completed restore request", zap.Error(err))
		}

		// Restore packages from snapshot metadata
		if err := r.restorePackagesFromSnapshot(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to restore packages", zap.Error(err))
			// Don't fail the entire restore if package restoration fails
		}

		// Track the restored snapshot for lineage
		now := metav1.Now()
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
			workspace.Status.SnapshotRestoreStatus.Phase = workspacev1.SnapshotRestorePhaseCompleted
			workspace.Status.SnapshotRestoreStatus.Message = "Snapshot restore completed"
			workspace.Status.SnapshotRestoreStatus.CompletionTime = &now

			// Track the last restored snapshot for lineage
			workspace.Status.LastRestoredSnapshot = &workspacev1.WorkspaceLastRestoredSnapshotInfo{
				Name:       workspace.Spec.FromSnapshot.SnapshotName,
				RestoredAt: now,
			}
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Moving to Completed phase")
		return reconcile.Result{Requeue: true}, nil

	default:
		// Still in progress
		logger.Info("Data restore in progress", zap.String("phase", string(restoreReq.Status.Phase)))
		return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
	}
}

// handleRestoreCompleted clears the fromSnapshot field and proceeds to normal reconciliation
func (r *WorkspaceReconciler) handleRestoreCompleted(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Completed - Clearing fromSnapshot and proceeding to normal reconciliation")

	// Clear fromSnapshot to mark restoration as complete
	workspace.Spec.FromSnapshot = nil

	// Add completion condition
	workspace.Status.Conditions = append(workspace.Status.Conditions, metav1.Condition{
		Type:               "RestoredFromSnapshot",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: workspace.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "RestoreCompleted",
		Message:            fmt.Sprintf("Successfully restored from snapshot %s", workspace.Status.SnapshotRestoreStatus.SourceSnapshot),
	})

	if err := r.Update(ctx, workspace); err != nil {
		logger.Error("Failed to clear fromSnapshot field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot restore completed, proceeding to normal workspace reconciliation")

	// Requeue to start normal workspace reconciliation
	return reconcile.Result{Requeue: true}, nil
}

// restorePackagesFromSnapshot reads package metadata from the pulled snapshot and creates PackageRequest
func (r *WorkspaceReconciler) restorePackagesFromSnapshot(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	logger *zap.Logger,
) error {
	// Determine the snapshot path
	snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("ws-%s-restore", workspace.Name))
	metadataPath := filepath.Join(snapshotPath, "metadata", "package-requests.json")

	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		logger.Info("No package metadata found in snapshot", zap.String("path", metadataPath))
		return nil
	}

	// Read package requests from metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read package metadata: %w", err)
	}

	var sourcePackageRequests []packagesv1.PackageRequest
	if err := json.Unmarshal(data, &sourcePackageRequests); err != nil {
		return fmt.Errorf("failed to parse package metadata: %w", err)
	}

	if len(sourcePackageRequests) == 0 {
		logger.Info("No packages to restore from snapshot")
		return nil
	}

	// Get the first (and typically only) package request
	sourcePackageRequest := sourcePackageRequests[0]
	if len(sourcePackageRequest.Spec.Packages) == 0 {
		logger.Info("Source package request has no packages")
		return nil
	}

	// Get target namespace
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	// Create PackageRequest for target workspace
	targetPackageRequestName := fmt.Sprintf("%s-packages", workspace.Name)
	targetPackageRequest := &packagesv1.PackageRequest{}

	// Check if target PackageRequest already exists
	err = r.Get(ctx, client.ObjectKey{
		Name:      targetPackageRequestName,
		Namespace: targetNamespace,
	}, targetPackageRequest)
	if err == nil {
		// PackageRequest exists, update it with source packages
		logger.Info("Updating existing PackageRequest with restored packages",
			zap.Int("packageCount", len(sourcePackageRequest.Spec.Packages)))
		targetPackageRequest.Spec.Packages = sourcePackageRequest.Spec.Packages
		if err := r.Update(ctx, targetPackageRequest); err != nil {
			return fmt.Errorf("failed to update target PackageRequest: %w", err)
		}
		logger.Info("PackageRequest updated with restored packages")
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check target PackageRequest: %w", err)
	}

	// Create new PackageRequest
	logger.Info("Creating PackageRequest with restored packages",
		zap.Int("packageCount", len(sourcePackageRequest.Spec.Packages)))

	blockOwnerDeletion := true
	targetPackageRequest = &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetPackageRequestName,
			Namespace: targetNamespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "workspaces.kloudlite.io/v1",
					Kind:               "Workspace",
					Name:               workspace.Name,
					UID:                workspace.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				},
			},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: workspace.Name,
			ProfileName:  targetPackageRequestName,
			Packages:     sourcePackageRequest.Spec.Packages,
		},
	}

	if err := r.Create(ctx, targetPackageRequest); err != nil {
		return fmt.Errorf("failed to create target PackageRequest: %w", err)
	}

	logger.Info("PackageRequest created with restored packages",
		zap.String("name", targetPackageRequestName),
		zap.Int("packageCount", len(sourcePackageRequest.Spec.Packages)))

	return nil
}

// failSnapshotRestore sets the restore phase to failed with an error message
func (r *WorkspaceReconciler) failSnapshotRestore(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	errorMessage string,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Snapshot restore failed", zap.String("error", errorMessage))

	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
		workspace.Status.SnapshotRestoreStatus.Phase = workspacev1.SnapshotRestorePhaseFailed
		workspace.Status.SnapshotRestoreStatus.ErrorMessage = errorMessage
		workspace.Status.SnapshotRestoreStatus.CompletionTime = &now
		return nil
	}, logger); err != nil {
		return reconcile.Result{}, err
	}

	// Add failure condition
	workspace.Status.Conditions = append(workspace.Status.Conditions, metav1.Condition{
		Type:               "RestoredFromSnapshot",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: workspace.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "RestoreFailed",
		Message:            errorMessage,
	})

	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error("Failed to update status with failure condition", zap.Error(err))
	}

	return reconcile.Result{}, nil
}

// parseImageRef parses an image reference like "image-registry:5000/repo:tag" into repository and tag
func parseImageRef(imageRef string) (repository, tag string) {
	// Remove registry prefix if present (e.g., "image-registry:5000/")
	parts := splitN(imageRef, "/", 2)
	if len(parts) == 2 {
		imageRef = parts[1]
	}

	// Split by colon to get repo:tag
	parts = splitN(imageRef, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return imageRef, "latest"
}

// splitN is a simple string split helper
func splitN(s, sep string, n int) []string {
	result := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	result = append(result, s)
	return result
}

func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
