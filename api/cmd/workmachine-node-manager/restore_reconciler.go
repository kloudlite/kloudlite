package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	zap2 "go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SnapshotRestoreReconciler handles snapshot restore operations on this node
type SnapshotRestoreReconciler struct {
	client.Client
	Logger           *zap2.Logger
	HostCmdExec      CommandExecutor // For btrfs commands that must run on host
	NodeName         string
	RegistryInsecure bool
}

func (r *SnapshotRestoreReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("snapshotRestore", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	// Fetch SnapshotRestore
	restore := &snapshotv1.SnapshotRestore{}
	if err := r.Get(ctx, req.NamespacedName, restore); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRestore", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Only process requests for this node
	if restore.Spec.NodeName != r.NodeName {
		return reconcile.Result{}, nil
	}

	// Handle completed or failed requests - no need to reprocess
	if restore.Status.State == snapshotv1.SnapshotRestoreStateCompleted ||
		restore.Status.State == snapshotv1.SnapshotRestoreStateFailed {
		return reconcile.Result{}, nil
	}

	logger.Info("Processing SnapshotRestore",
		zap2.String("state", string(restore.Status.State)),
		zap2.String("snapshotName", restore.Spec.SnapshotName))

	// Process based on current state
	switch restore.Status.State {
	case "", snapshotv1.SnapshotRestoreStatePending:
		return r.handleRestorePending(ctx, restore, logger)
	case snapshotv1.SnapshotRestoreStateDownloading:
		return r.handleRestoreDownloading(ctx, restore, logger)
	case snapshotv1.SnapshotRestoreStateRestoring:
		return r.handleRestoreRestoring(ctx, restore, logger)
	default:
		logger.Warn("Unknown restore state", zap2.String("state", string(restore.Status.State)))
		return reconcile.Result{}, nil
	}
}

func (r *SnapshotRestoreReconciler) handleRestorePending(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Info("Starting snapshot restore",
		zap2.String("targetPath", restore.Spec.TargetPath),
		zap2.String("snapshotName", restore.Spec.SnapshotName))

	now := metav1.Now()
	restore.Status.State = snapshotv1.SnapshotRestoreStateDownloading
	restore.Status.Message = "Downloading snapshot from registry"
	restore.Status.StartedAt = &now
	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRestoreReconciler) handleRestoreDownloading(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap2.Logger) (reconcile.Result, error) {
	snapshotName := restore.Spec.SnapshotName

	// Get the Snapshot first to find the registry info (storage reference)
	// The imageRef is the only constant identifier - snapshot name/owner may differ across envs
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName, Namespace: restore.Namespace}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setRestoreFailed(ctx, restore, fmt.Sprintf("Snapshot %q not found in namespace %s", snapshotName, restore.Namespace), logger)
		}
		logger.Error("Failed to get Snapshot", zap2.Error(err))
		return reconcile.Result{}, err
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		logger.Info("Snapshot not ready, waiting", zap2.String("state", string(snapshot.Status.State)))
		restore.Status.Message = fmt.Sprintf("Waiting for snapshot to be ready (state: %s)", snapshot.Status.State)
		if err := r.Status().Update(ctx, restore); err != nil {
			if !apierrors.IsConflict(err) {
				logger.Error("Failed to update status", zap2.Error(err))
			}
		}
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if snapshot.Status.Registry == nil || snapshot.Status.Registry.ImageRef == "" {
		return r.setRestoreFailed(ctx, restore, "Snapshot has no registry info", logger)
	}

	// Use the storage reference (imageRef) for cache lookup
	// This ensures we find the cached snapshot regardless of name/owner changes
	imageRef := snapshot.Status.Registry.ImageRef
	cachePath := cachePathFromImageRef(imageRef)

	// Check if snapshot is already cached locally as a btrfs subvolume
	// We verify it's a subvolume (not just a directory) to ensure btrfs snapshot will work
	checkCacheScript := fmt.Sprintf("btrfs subvolume show %s >/dev/null 2>&1 && echo 'subvol'", cachePath)
	cacheOutput, _ := r.HostCmdExec.Execute(checkCacheScript)
	if strings.Contains(string(cacheOutput), "subvol") {
		logger.Info("Snapshot already cached as btrfs subvolume, skipping download",
			zap2.String("imageRef", imageRef),
			zap2.String("cachePath", cachePath))

		// Skip download, go directly to Restoring state
		restore.Status.State = snapshotv1.SnapshotRestoreStateRestoring
		restore.Status.Message = "Using cached snapshot"
		if err := r.Status().Update(ctx, restore); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Error("Failed to update status", zap2.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// If cache exists but is not a subvolume (old format), remove it
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		logger.Info("Cache exists but is not a btrfs subvolume, removing old cache",
			zap2.String("cachePath", cachePath))
		cleanupScript := fmt.Sprintf("rm -rf %s", cachePath)
		r.HostCmdExec.Execute(cleanupScript)
	}

	// Create a temp directory for extraction
	tempExtractPath := fmt.Sprintf("%s-extracting", cachePath)
	if err := os.MkdirAll(tempExtractPath, 0755); err != nil {
		logger.Warn("Failed to create temp extract directory", zap2.Error(err))
	}

	// Pull snapshot from registry using embedded oras library
	logger.Info("Pulling snapshot from registry", zap2.String("imageRef", imageRef))
	if err := orasPullSnapshot(ctx, imageRef, tempExtractPath, r.RegistryInsecure); err != nil {
		logger.Error("Failed to pull snapshot from registry",
			zap2.String("imageRef", imageRef),
			zap2.Error(err))
		// Clean up failed temp directory
		os.RemoveAll(tempExtractPath)
		return r.setRestoreFailed(ctx, restore, fmt.Sprintf("Failed to pull from registry: %v", err), logger)
	}

	// Convert extracted data to a btrfs subvolume for efficient snapshots
	// 1. Create btrfs subvolume at cache path
	// 2. Copy data from temp to subvolume
	// 3. Remove temp directory
	convertScript := fmt.Sprintf(`
		set -e
		btrfs subvolume create %s
		cp -a %s/. %s/
		rm -rf %s
	`, cachePath, tempExtractPath, cachePath, tempExtractPath)

	convertOutput, err := r.HostCmdExec.Execute(convertScript)
	if err != nil {
		logger.Error("Failed to convert cache to btrfs subvolume",
			zap2.String("cachePath", cachePath),
			zap2.Error(err),
			zap2.String("output", string(convertOutput)))
		// Clean up
		cleanupScript := fmt.Sprintf("rm -rf %s %s", tempExtractPath, cachePath)
		r.HostCmdExec.Execute(cleanupScript)
		return r.setRestoreFailed(ctx, restore, fmt.Sprintf("Failed to create cache subvolume: %v", err), logger)
	}

	logger.Info("Created btrfs subvolume cache from extracted data", zap2.String("cachePath", cachePath))

	// Update status to Restoring
	restore.Status.State = snapshotv1.SnapshotRestoreStateRestoring
	restore.Status.Message = "Restoring snapshot data"
	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Downloaded snapshot to cache", zap2.String("imageRef", imageRef), zap2.String("cachePath", cachePath))
	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRestoreReconciler) handleRestoreRestoring(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap2.Logger) (reconcile.Result, error) {
	snapshotName := restore.Spec.SnapshotName
	targetPath := restore.Spec.TargetPath

	// Get the Snapshot to find the storage reference (imageRef) for cache lookup
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName, Namespace: restore.Namespace}, snapshot); err != nil {
		return r.setRestoreFailed(ctx, restore, fmt.Sprintf("Snapshot %q not found", snapshotName), logger)
	}

	if snapshot.Status.Registry == nil || snapshot.Status.Registry.ImageRef == "" {
		return r.setRestoreFailed(ctx, restore, "Snapshot has no registry info", logger)
	}

	// Use the storage reference for cache path
	cachePath := cachePathFromImageRef(snapshot.Status.Registry.ImageRef)

	// Ensure parent directory of target exists
	parentDir := filepath.Dir(targetPath)
	mkdirScript := fmt.Sprintf("mkdir -p %s", parentDir)
	if _, err := r.HostCmdExec.Execute(mkdirScript); err != nil {
		logger.Warn("Failed to create parent directory", zap2.Error(err))
	}

	// Check if target path exists and is a btrfs subvolume
	checkScript := fmt.Sprintf("btrfs subvolume show %s 2>/dev/null && echo 'is_subvol'", targetPath)
	checkOutput, _ := r.HostCmdExec.Execute(checkScript)
	isSubvolume := strings.Contains(string(checkOutput), "is_subvol")

	if isSubvolume {
		// Delete existing subvolume
		logger.Info("Deleting existing subvolume", zap2.String("path", targetPath))
		deleteScript := fmt.Sprintf("btrfs subvolume delete %s", targetPath)
		if output, err := r.HostCmdExec.Execute(deleteScript); err != nil {
			logger.Warn("Failed to delete existing subvolume",
				zap2.Error(err),
				zap2.String("output", string(output)))
		}
	} else {
		// Remove existing directory if it exists
		rmScript := fmt.Sprintf("rm -rf %s", targetPath)
		if _, err := r.HostCmdExec.Execute(rmScript); err != nil {
			logger.Warn("Failed to remove existing directory", zap2.Error(err))
		}
	}

	// Create environment as a btrfs snapshot of the cached snapshot data
	// This is fast (instant) and space-efficient (copy-on-write)
	restoreScript := fmt.Sprintf("btrfs subvolume snapshot %s %s", cachePath, targetPath)

	output, err := r.HostCmdExec.Execute(restoreScript)
	if err != nil {
		logger.Error("Failed to create snapshot from cache",
			zap2.String("cachePath", cachePath),
			zap2.String("targetPath", targetPath),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return r.setRestoreFailed(ctx, restore, fmt.Sprintf("Failed to restore: %v", err), logger)
	}

	// Cache is kept for future restores (not cleaned up)

	// Mark as completed
	now := metav1.Now()
	restore.Status.State = snapshotv1.SnapshotRestoreStateCompleted
	restore.Status.Message = "Restore completed successfully"
	restore.Status.CompletedAt = &now
	restore.Status.RestoredPath = targetPath

	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot restored from cache",
		zap2.String("cachePath", cachePath),
		zap2.String("targetPath", targetPath))
	return reconcile.Result{}, nil
}

func (r *SnapshotRestoreReconciler) setRestoreFailed(ctx context.Context, restore *snapshotv1.SnapshotRestore, message string, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot restore failed", zap2.String("message", message))

	now := metav1.Now()
	restore.Status.State = snapshotv1.SnapshotRestoreStateFailed
	restore.Status.Message = message
	restore.Status.CompletedAt = &now

	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SnapshotRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRestore{}).
		Complete(r)
}
