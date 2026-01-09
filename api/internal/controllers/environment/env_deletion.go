package environment

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// environmentsStoragePath is where environment btrfs subvolumes are stored
	environmentsStoragePath = "/var/lib/kloudlite/storage/environments"
)

// handleDeletion handles the deletion of an environment and its child resources
func (r *EnvironmentReconciler) handleDeletion(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	// Update status to show deletion in progress
	if environment.Status.State != environmentsv1.EnvironmentStateDeleting {
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateDeleting, "Deleting environment and cleaning up resources", logger); err != nil {
			logger.Error("Failed to update environment status to deleting after retries", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

	// Clean up workspace environment connections referencing this environment
	if err := r.cleanupWorkspaceConnections(ctx, environment, logger); err != nil {
		logger.Error("Failed to cleanup workspace connections", zap.Error(err))
		// Continue with deletion even if cleanup fails
	}

	// Clean up snapshots for this environment
	if err := r.cleanupEnvironmentSnapshots(ctx, environment, logger); err != nil {
		logger.Error("Failed to cleanup environment snapshots", zap.Error(err))
		// Continue with deletion even if cleanup fails
	}

	// Delete namespace and wait for completion
	deleted, err := r.deleteNamespace(ctx, environment, logger)
	if err != nil {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	if !deleted {
		// Namespace deletion in progress
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Namespace is deleted, now clean up the btrfs subvolumes
	logger.Info("Namespace deleted, cleaning up environment btrfs subvolumes")

	// Clean up environment btrfs subvolume (live data)
	subvolumeDeleted, err := r.cleanupEnvironmentSubvolume(ctx, environment, logger)
	if err != nil {
		logger.Error("Failed to cleanup environment subvolume", zap.Error(err))
		// Continue with finalizer removal even if subvolume cleanup fails
	} else if !subvolumeDeleted {
		// Subvolume deletion in progress
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Clean up pulled snapshots directory (.snapshots/envs/{envName}/)
	snapshotsDeleted, err := r.cleanupEnvironmentSnapshotDirectory(ctx, environment, logger)
	if err != nil {
		logger.Error("Failed to cleanup environment snapshot directory", zap.Error(err))
		// Continue with finalizer removal even if cleanup fails
	} else if !snapshotsDeleted {
		// Snapshot directory deletion in progress
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// All cleanup complete, remove finalizer
	logger.Info("All cleanup complete, removing finalizer from environment")

	if controllerutil.ContainsFinalizer(environment, environmentFinalizer) {
		controllerutil.RemoveFinalizer(environment, environmentFinalizer)
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to remove finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("Environment cleanup completed successfully")
	return reconcile.Result{}, nil
}

// cleanupWorkspaceConnections removes environment connections from all workspaces referencing this environment
func (r *EnvironmentReconciler) cleanupWorkspaceConnections(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Get Workspace type to list workspaces
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	environmentName := environment.Name
	environmentNamespace := environment.Namespace
	if environmentNamespace == "" {
		environmentNamespace = "default"
	}

	cleanedCount := 0
	for i := range workspaceList.Items {
		workspace := &workspaceList.Items[i]

		// Check if this workspace references the environment being deleted
		if workspace.Spec.EnvironmentConnection == nil {
			continue
		}

		envRef := workspace.Spec.EnvironmentConnection.EnvironmentRef
		if envRef.Name == environmentName && envRef.Namespace == environmentNamespace {
			logger.Info("Removing environment connection from workspace",
				zap.String("workspace", workspace.Name),
				zap.String("environment", environmentName))

			// Remove the environment connection
			workspace.Spec.EnvironmentConnection = nil

			if err := r.Update(ctx, workspace); err != nil {
				logger.Error("Failed to remove environment connection from workspace",
					zap.String("workspace", workspace.Name),
					zap.Error(err))
				// Continue with other workspaces even if one fails
				continue
			}
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		logger.Info("Cleaned up workspace environment connections",
			zap.Int("count", cleanedCount))
	}

	return nil
}

// cleanupEnvironmentSnapshots deletes all snapshots associated with this environment
func (r *EnvironmentReconciler) cleanupEnvironmentSnapshots(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// List all snapshots for this environment using the label
	snapshotList := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, snapshotList, client.MatchingLabels{
		"snapshots.kloudlite.io/environment": environment.Name,
	}); err != nil {
		return fmt.Errorf("failed to list snapshots for environment: %w", err)
	}

	if len(snapshotList.Items) == 0 {
		return nil
	}

	logger.Info("Deleting snapshots for environment",
		zap.String("environment", environment.Name),
		zap.Int("count", len(snapshotList.Items)))

	deletedCount := 0
	for i := range snapshotList.Items {
		snapshot := &snapshotList.Items[i]
		logger.Info("Deleting snapshot",
			zap.String("snapshot", snapshot.Name),
			zap.String("environment", environment.Name))

		if err := r.Delete(ctx, snapshot); err != nil {
			logger.Error("Failed to delete snapshot",
				zap.String("snapshot", snapshot.Name),
				zap.Error(err))
			// Continue with other snapshots even if one fails
			continue
		}
		deletedCount++
	}

	logger.Info("Deleted environment snapshots",
		zap.String("environment", environment.Name),
		zap.Int("deleted", deletedCount),
		zap.Int("total", len(snapshotList.Items)))

	return nil
}

// cleanupEnvironmentSubvolume creates a SnapshotRequest to delete the environment's btrfs subvolume
// Returns (true, nil) if cleanup is complete, (false, nil) if in progress, (false, err) on error
func (r *EnvironmentReconciler) cleanupEnvironmentSubvolume(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (bool, error) {
	if environment.Spec.TargetNamespace == "" {
		// No target namespace means no subvolume to delete
		return true, nil
	}

	// The environment subvolume path
	subvolumePath := filepath.Join(environmentsStoragePath, environment.Spec.TargetNamespace)

	// SnapshotRequest name and namespace
	deleteReqName := fmt.Sprintf("%s-delete-subvolume", environment.Name)
	deleteReqNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)

	// Check if delete request already exists
	existingReq := &snapshotv1.SnapshotRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: deleteReqName, Namespace: deleteReqNamespace}, existingReq)
	if err == nil {
		// Request exists, check its status
		switch existingReq.Status.Phase {
		case snapshotv1.SnapshotRequestPhaseCompleted:
			logger.Info("Environment subvolume deleted successfully")
			// Clean up the completed request
			if delErr := r.Delete(ctx, existingReq); delErr != nil && !apierrors.IsNotFound(delErr) {
				logger.Warn("Failed to delete completed SnapshotRequest", zap.Error(delErr))
			}
			return true, nil
		case snapshotv1.SnapshotRequestPhaseFailed:
			logger.Warn("Environment subvolume deletion failed",
				zap.String("message", existingReq.Status.Message))
			// Clean up the failed request and continue (subvolume might not exist)
			if delErr := r.Delete(ctx, existingReq); delErr != nil && !apierrors.IsNotFound(delErr) {
				logger.Warn("Failed to delete failed SnapshotRequest", zap.Error(delErr))
			}
			return true, nil
		default:
			// Still in progress
			logger.Info("Waiting for environment subvolume deletion",
				zap.String("phase", string(existingReq.Status.Phase)))
			return false, nil
		}
	} else if !apierrors.IsNotFound(err) {
		return false, fmt.Errorf("failed to get SnapshotRequest: %w", err)
	}

	// Create the delete request
	deleteReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deleteReqName,
			Namespace: deleteReqNamespace,
			Labels: map[string]string{
				"environments.kloudlite.io/environment": environment.Name,
				"snapshots.kloudlite.io/operation":      "delete",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:       snapshotv1.SnapshotOperationDelete,
			SnapshotPath:    subvolumePath,
			EnvironmentName: environment.Name,
		},
	}

	logger.Info("Creating SnapshotRequest to delete environment subvolume",
		zap.String("path", subvolumePath),
		zap.String("request", deleteReqName))

	if err := r.Create(ctx, deleteReq); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Race condition, request was just created
			return false, nil
		}
		return false, fmt.Errorf("failed to create delete SnapshotRequest: %w", err)
	}

	// Request created, wait for completion
	return false, nil
}

// cleanupEnvironmentSnapshotDirectory creates a SnapshotRequest to delete the environment's pulled snapshots directory
// This cleans up .snapshots/envs/{envName}/ which contains pulled snapshots from forks/restores
// Returns (true, nil) if cleanup is complete, (false, nil) if in progress, (false, err) on error
func (r *EnvironmentReconciler) cleanupEnvironmentSnapshotDirectory(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (bool, error) {
	// The pulled snapshots directory path
	snapshotsDirPath := filepath.Join(envSnapshotsBasePath, environment.Name)

	// SnapshotRequest name and namespace
	deleteReqName := fmt.Sprintf("%s-delete-snapshots-dir", environment.Name)
	deleteReqNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)

	// Check if delete request already exists
	existingReq := &snapshotv1.SnapshotRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: deleteReqName, Namespace: deleteReqNamespace}, existingReq)
	if err == nil {
		// Request exists, check its status
		switch existingReq.Status.Phase {
		case snapshotv1.SnapshotRequestPhaseCompleted:
			logger.Info("Environment snapshot directory deleted successfully")
			// Clean up the completed request
			if delErr := r.Delete(ctx, existingReq); delErr != nil && !apierrors.IsNotFound(delErr) {
				logger.Warn("Failed to delete completed SnapshotRequest", zap.Error(delErr))
			}
			return true, nil
		case snapshotv1.SnapshotRequestPhaseFailed:
			logger.Warn("Environment snapshot directory deletion failed (directory may not exist)",
				zap.String("message", existingReq.Status.Message))
			// Clean up the failed request and continue (directory might not exist)
			if delErr := r.Delete(ctx, existingReq); delErr != nil && !apierrors.IsNotFound(delErr) {
				logger.Warn("Failed to delete failed SnapshotRequest", zap.Error(delErr))
			}
			return true, nil
		default:
			// Still in progress
			logger.Info("Waiting for environment snapshot directory deletion",
				zap.String("phase", string(existingReq.Status.Phase)))
			return false, nil
		}
	} else if !apierrors.IsNotFound(err) {
		return false, fmt.Errorf("failed to get SnapshotRequest: %w", err)
	}

	// Create the delete request
	deleteReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deleteReqName,
			Namespace: deleteReqNamespace,
			Labels: map[string]string{
				"environments.kloudlite.io/environment": environment.Name,
				"snapshots.kloudlite.io/operation":      "delete-snapshots-dir",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:       snapshotv1.SnapshotOperationDelete,
			SnapshotPath:    snapshotsDirPath,
			EnvironmentName: environment.Name,
		},
	}

	logger.Info("Creating SnapshotRequest to delete environment snapshot directory",
		zap.String("path", snapshotsDirPath),
		zap.String("request", deleteReqName))

	if err := r.Create(ctx, deleteReq); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Race condition, request was just created
			return false, nil
		}
		return false, fmt.Errorf("failed to create delete SnapshotRequest: %w", err)
	}

	// Request created, wait for completion
	return false, nil
}
