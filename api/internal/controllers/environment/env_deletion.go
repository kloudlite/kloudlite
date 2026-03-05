package environment

import (
	"context"
	"fmt"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/pagination"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleDeletion handles the deletion of an environment and its child resources
// Note: Snapshots in the environment's namespace are deleted via owner references
func (r *EnvironmentReconciler) handleDeletion(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	// Update status to show deletion in progress
	if environment.Status.State != environmentsv1.EnvironmentStateDeleting {
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateDeleting, "Deleting environment and cleaning up resources", logger); err != nil {
			logger.Error("Failed to update environment status to deleting after retries", zap.Error(err))
			// Continue with deletion even if status update fails, as this is non-critical
		}
	}

	// Clean up workspace environment connections referencing this environment
	// This is critical - if cleanup fails, we should requeue and retry
	if err := r.cleanupWorkspaceConnections(ctx, environment, logger); err != nil {
		logger.Error("Failed to cleanup workspace connections, will retry", zap.Error(err))
		// Requeue with a delay to allow transient issues to resolve
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.DeletionRetryInterval}, fmt.Errorf("failed to cleanup workspace connections: %w", err)
	}

	// Clean up compose resources
	// This is critical - if cleanup fails, we should requeue and retry
	if err := r.cleanupComposeResources(ctx, environment, logger); err != nil {
		logger.Error("Failed to cleanup compose resources, will retry", zap.Error(err))
		// Requeue with a delay to allow transient issues to resolve
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.DeletionRetryInterval}, fmt.Errorf("failed to cleanup compose resources: %w", err)
	}

	// Snapshots are automatically deleted via owner references when the namespace is deleted
	// Storage GC happens via storageRefs - images only deleted when no snapshot references them

	// Delete namespace and wait for completion
	deleted, err := r.deleteNamespace(ctx, environment, logger)
	if err != nil {
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.DeletionRetryInterval}, err
	}

	if !deleted {
		// Namespace deletion in progress
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.PodTerminationRetryInterval}, nil
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
	if err := pagination.ListAll(ctx, r, workspaceList); err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	environmentName := environment.Name
	environmentNamespace := environment.Namespace
	if environmentNamespace == "" {
		environmentNamespace = "default"
	}

	cleanedCount := 0
	var errors []error
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
				// Collect error but continue with other workspaces
				errors = append(errors, fmt.Errorf("workspace %s: %w", workspace.Name, err))
				continue
			}
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		logger.Info("Cleaned up workspace environment connections",
			zap.Int("count", cleanedCount))
	}

	// Return aggregated error if any workspace cleanup failed
	if len(errors) > 0 {
		return fmt.Errorf("failed to cleanup %d workspace connections: %w", len(errors), joinErrors(errors))
	}

	return nil
}
