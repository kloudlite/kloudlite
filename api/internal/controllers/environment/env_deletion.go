package environment

import (
	"context"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	// Decrement snapshot refCount if this environment was created from a snapshot
	if err := r.decrementSnapshotRefCount(ctx, environment, logger); err != nil {
		logger.Error("Failed to decrement snapshot refCount", zap.Error(err))
		// Continue with deletion even if decrement fails
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

// decrementSnapshotRefCount decrements the refCount of all snapshots in the lineage this environment was created from
func (r *EnvironmentReconciler) decrementSnapshotRefCount(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Get lineage from LastRestoredSnapshot status (preferred) or fallback to spec.fromSnapshot
	var lineage []string

	if environment.Status.LastRestoredSnapshot != nil && len(environment.Status.LastRestoredSnapshot.Lineage) > 0 {
		// Use stored lineage from status
		lineage = environment.Status.LastRestoredSnapshot.Lineage
		logger.Info("Using stored snapshot lineage for refCount decrement",
			zap.Strings("lineage", lineage),
			zap.String("environment", environment.Name))
	} else if environment.Status.LastRestoredSnapshot != nil && environment.Status.LastRestoredSnapshot.Name != "" {
		// Fallback for environments created before lineage tracking was added
		lineage = []string{environment.Status.LastRestoredSnapshot.Name}
		logger.Info("Using fallback snapshot name for refCount decrement (no lineage stored)",
			zap.String("snapshot", environment.Status.LastRestoredSnapshot.Name),
			zap.String("environment", environment.Name))
	} else if environment.Spec.FromSnapshot != nil && environment.Spec.FromSnapshot.SnapshotName != "" {
		// Legacy fallback: use spec.fromSnapshot (should rarely happen as it's cleared after restore)
		lineage = []string{environment.Spec.FromSnapshot.SnapshotName}
		logger.Info("Using legacy spec.fromSnapshot for refCount decrement",
			zap.String("snapshot", environment.Spec.FromSnapshot.SnapshotName),
			zap.String("environment", environment.Name))
	} else {
		// Environment was not created from a snapshot
		return nil
	}

	// Decrement refCount for all snapshots in lineage
	for _, snapshotName := range lineage {
		snapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("failed to get snapshot %s: %w", snapshotName, err)
			}
			// Snapshot doesn't exist, nothing to decrement
			logger.Info("Snapshot not found, skipping refCount decrement", zap.String("snapshot", snapshotName))
			continue
		}

		// Decrement refCount (minimum 0)
		newRefCount := snapshot.Status.RefCount - 1
		if newRefCount < 0 {
			newRefCount = 0
		}

		snapshot.Status.RefCount = newRefCount
		if err := r.Status().Update(ctx, snapshot); err != nil {
			return fmt.Errorf("failed to update snapshot %s refCount: %w", snapshotName, err)
		}

		logger.Info("Decremented snapshot refCount",
			zap.String("snapshot", snapshotName),
			zap.Int32("newRefCount", newRefCount))
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
