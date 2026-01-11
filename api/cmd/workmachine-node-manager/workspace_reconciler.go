package main

import (
	"context"
	"fmt"
	"time"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	zap2 "go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// WorkspaceCleanupReconciler watches Workspace resources and manages workspace btrfs subvolumes
// It creates btrfs subvolumes when workspaces are created and deletes them when workspaces are deleted
type WorkspaceCleanupReconciler struct {
	client.Client
	Logger  *zap2.Logger
	FS      FileSystem
	CmdExec CommandExecutor
}

// workspaceStoragePath is the base path for workspace btrfs subvolumes
const workspaceStoragePath = "/var/lib/kloudlite/storage/workspaces"

func (r *WorkspaceCleanupReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("workspace", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling Workspace for btrfs storage management")

	// Fetch Workspace (namespaced)
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, workspace); err != nil {
		// If workspace is not found, it's already fully deleted (including all finalizers)
		// Nothing to do - this is expected when workspace is deleted
		logger.Info("Workspace not found, already deleted")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Check if workspace is being deleted
	if workspace.DeletionTimestamp != nil {
		// Workspace is being deleted
		if containsString(workspace.Finalizers, workspaceCleanupFinalizer) {
			logger.Info("Workspace is being deleted, cleaning up btrfs subvolume")

			// Clean up workspace btrfs subvolume
			workspaceDir := fmt.Sprintf("%s/%s", workspaceStoragePath, workspace.Name)
			logger.Info("Removing workspace btrfs subvolume", zap2.String("path", workspaceDir))

			// Delete btrfs subvolume (will fail if path doesn't exist or is not a subvolume)
			deleteScript := fmt.Sprintf("btrfs subvolume delete %s", workspaceDir)
			output, err := r.CmdExec.Execute(deleteScript)
			if err != nil {
				// Check if path exists - if not, nothing to clean up
				checkExistsScript := fmt.Sprintf("test -e %s", workspaceDir)
				if _, existsErr := r.CmdExec.Execute(checkExistsScript); existsErr != nil {
					logger.Info("Workspace subvolume doesn't exist, skipping cleanup", zap2.String("path", workspaceDir))
				} else {
					logger.Error("Failed to delete btrfs subvolume",
						zap2.String("path", workspaceDir),
						zap2.Error(err),
						zap2.String("output", string(output)))
					return reconcile.Result{}, fmt.Errorf("failed to delete btrfs subvolume: %w", err)
				}
			} else {
				logger.Info("Successfully deleted btrfs subvolume", zap2.String("path", workspaceDir))
			}

			// Remove finalizer
			workspace.Finalizers = removeString(workspace.Finalizers, workspaceCleanupFinalizer)
			if err := r.Update(ctx, workspace); err != nil {
				logger.Error("Failed to remove finalizer", zap2.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("Cleanup complete, finalizer removed")
		}
		return reconcile.Result{}, nil
	}

	// Workspace is NOT being deleted - ensure btrfs subvolume exists
	// This creates the subvolume BEFORE the pod starts
	workspaceDir := fmt.Sprintf("%s/%s", workspaceStoragePath, workspace.Name)

	// Check if btrfs subvolume already exists
	checkScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", workspaceDir)
	if _, err := r.CmdExec.Execute(checkScript); err == nil {
		// Subvolume already exists
		logger.Debug("Workspace btrfs subvolume already exists", zap2.String("path", workspaceDir))
		return reconcile.Result{}, nil
	}

	// Create new btrfs subvolume
	logger.Info("Creating workspace btrfs subvolume", zap2.String("path", workspaceDir))
	createScript := fmt.Sprintf("btrfs subvolume create %s && chown 1001:1001 %s", workspaceDir, workspaceDir)

	if output, err := r.CmdExec.Execute(createScript); err != nil {
		logger.Error("Failed to create workspace btrfs subvolume",
			zap2.String("path", workspaceDir),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	logger.Info("Successfully created workspace btrfs subvolume", zap2.String("path", workspaceDir))
	return reconcile.Result{}, nil
}

func (r *WorkspaceCleanupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacev1.Workspace{}).
		Complete(r)
}
