package workmachinenodemanager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	zap2 "go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CheckpointRestoreReconciler handles checkpoint restore operations on this node
type CheckpointRestoreReconciler struct {
	client.Client
	Logger           *zap2.Logger
	HostCmdExec      CommandExecutor
	NodeName         string
	RegistryInsecure bool
}

func (r *CheckpointRestoreReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("checkpointRestore", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	restore := &checkpointv1.CheckpointRestore{}
	if err := r.Get(ctx, req.NamespacedName, restore); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get CheckpointRestore", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Only process requests for this node
	if restore.Spec.NodeName != r.NodeName {
		return reconcile.Result{}, nil
	}

	// Skip terminal states
	if restore.Status.Phase == checkpointv1.CheckpointRestorePhaseCompleted ||
		restore.Status.Phase == checkpointv1.CheckpointRestorePhaseFailed {
		return reconcile.Result{}, nil
	}

	logger.Info("Processing CheckpointRestore",
		zap2.String("phase", string(restore.Status.Phase)),
		zap2.String("checkpoint", restore.Spec.CheckpointName))

	switch restore.Status.Phase {
	case "", checkpointv1.CheckpointRestorePhasePending:
		return r.handlePending(ctx, restore, logger)
	case checkpointv1.CheckpointRestorePhaseDownloading:
		return r.handleDownloading(ctx, restore, logger)
	case checkpointv1.CheckpointRestorePhaseRestoring:
		return r.handleRestoring(ctx, restore, logger)
	default:
		logger.Warn("Unknown restore phase", zap2.String("phase", string(restore.Status.Phase)))
		return reconcile.Result{}, nil
	}
}

func (r *CheckpointRestoreReconciler) handlePending(ctx context.Context, restore *checkpointv1.CheckpointRestore, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Info("Starting checkpoint restore",
		zap2.String("targetPath", restore.Spec.TargetPath),
		zap2.String("checkpoint", restore.Spec.CheckpointName))

	now := metav1.Now()
	restore.Status.Phase = checkpointv1.CheckpointRestorePhaseDownloading
	restore.Status.Message = "Downloading checkpoint from registry"
	restore.Status.StartedAt = &now
	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *CheckpointRestoreReconciler) handleDownloading(ctx context.Context, restore *checkpointv1.CheckpointRestore, logger *zap2.Logger) (reconcile.Result, error) {
	// Get the Checkpoint to find the layer info
	checkpoint := &checkpointv1.Checkpoint{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.CheckpointName, Namespace: restore.Namespace}, checkpoint); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, restore, fmt.Sprintf("Checkpoint %q not found in namespace %s", restore.Spec.CheckpointName, restore.Namespace), logger)
		}
		logger.Error("Failed to get Checkpoint", zap2.Error(err))
		return reconcile.Result{}, err
	}

	if checkpoint.Status.Phase != checkpointv1.CheckpointPhaseReady {
		logger.Info("Checkpoint not ready, waiting", zap2.String("phase", string(checkpoint.Status.Phase)))
		restore.Status.Message = fmt.Sprintf("Waiting for checkpoint to be ready (phase: %s)", checkpoint.Status.Phase)
		if err := r.Status().Update(ctx, restore); err != nil {
			if !apierrors.IsConflict(err) {
				logger.Error("Failed to update status", zap2.Error(err))
			}
		}
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if checkpoint.Status.Layer == nil || checkpoint.Status.Layer.ImageRef == "" {
		return r.setFailed(ctx, restore, "Checkpoint has no layer info", logger)
	}

	imageRef := checkpoint.Status.Layer.ImageRef
	cachePath := cachePathFromImageRef(imageRef)

	// Check if checkpoint is already cached locally as a btrfs subvolume
	checkCacheScript := fmt.Sprintf("btrfs subvolume show %s >/dev/null 2>&1 && echo 'subvol'", cachePath)
	cacheOutput, _ := r.HostCmdExec.Execute(checkCacheScript)
	if strings.Contains(string(cacheOutput), "subvol") {
		logger.Info("Checkpoint already cached, skipping download",
			zap2.String("imageRef", imageRef),
			zap2.String("cachePath", cachePath))

		restore.Status.Phase = checkpointv1.CheckpointRestorePhaseRestoring
		restore.Status.Message = "Using cached checkpoint"
		if err := r.Status().Update(ctx, restore); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// If cache exists but is not a subvolume (old format), remove it
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		logger.Info("Cache exists but is not a btrfs subvolume, removing", zap2.String("cachePath", cachePath))
		r.HostCmdExec.Execute(fmt.Sprintf("rm -rf %s", cachePath))
	}

	// Create temp directory for extraction
	tempExtractPath := fmt.Sprintf("%s-extracting", cachePath)
	if err := os.MkdirAll(tempExtractPath, 0755); err != nil {
		logger.Warn("Failed to create temp extract directory", zap2.Error(err))
	}

	// Pull from registry
	logger.Info("Pulling checkpoint from registry", zap2.String("imageRef", imageRef))
	if err := orasPullSnapshot(ctx, imageRef, tempExtractPath, r.RegistryInsecure); err != nil {
		os.RemoveAll(tempExtractPath)
		return r.setFailed(ctx, restore, fmt.Sprintf("Failed to pull from registry: %v", err), logger)
	}

	// Convert extracted data to a btrfs subvolume for efficient snapshots
	convertScript := fmt.Sprintf(`
		set -e
		btrfs subvolume create %s
		cp -a %s/. %s/
		rm -rf %s
	`, cachePath, tempExtractPath, cachePath, tempExtractPath)

	convertOutput, err := r.HostCmdExec.Execute(convertScript)
	if err != nil {
		r.HostCmdExec.Execute(fmt.Sprintf("rm -rf %s %s", tempExtractPath, cachePath))
		return r.setFailed(ctx, restore, fmt.Sprintf("Failed to create cache subvolume: %v - %s", err, string(convertOutput)), logger)
	}

	logger.Info("Created btrfs subvolume cache", zap2.String("cachePath", cachePath))

	restore.Status.Phase = checkpointv1.CheckpointRestorePhaseRestoring
	restore.Status.Message = "Restoring checkpoint data"
	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *CheckpointRestoreReconciler) handleRestoring(ctx context.Context, restore *checkpointv1.CheckpointRestore, logger *zap2.Logger) (reconcile.Result, error) {
	targetPath := restore.Spec.TargetPath

	// Get the Checkpoint to find the layer imageRef for cache lookup
	checkpoint := &checkpointv1.Checkpoint{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.CheckpointName, Namespace: restore.Namespace}, checkpoint); err != nil {
		return r.setFailed(ctx, restore, fmt.Sprintf("Checkpoint %q not found", restore.Spec.CheckpointName), logger)
	}

	if checkpoint.Status.Layer == nil || checkpoint.Status.Layer.ImageRef == "" {
		return r.setFailed(ctx, restore, "Checkpoint has no layer info", logger)
	}

	cachePath := cachePathFromImageRef(checkpoint.Status.Layer.ImageRef)

	// Ensure parent directory of target exists
	parentDir := filepath.Dir(targetPath)
	if _, err := r.HostCmdExec.Execute(fmt.Sprintf("mkdir -p %s", parentDir)); err != nil {
		logger.Warn("Failed to create parent directory", zap2.Error(err))
	}

	// Check if target path exists and is a btrfs subvolume
	checkScript := fmt.Sprintf("btrfs subvolume show %s 2>/dev/null && echo 'is_subvol'", targetPath)
	checkOutput, _ := r.HostCmdExec.Execute(checkScript)
	isSubvolume := strings.Contains(string(checkOutput), "is_subvol")

	if isSubvolume {
		logger.Info("Deleting existing subvolume", zap2.String("path", targetPath))
		deleteScript := fmt.Sprintf("btrfs subvolume delete %s", targetPath)
		if output, err := r.HostCmdExec.Execute(deleteScript); err != nil {
			logger.Warn("Failed to delete existing subvolume",
				zap2.Error(err),
				zap2.String("output", string(output)))
		}
	} else {
		if _, err := r.HostCmdExec.Execute(fmt.Sprintf("rm -rf %s", targetPath)); err != nil {
			logger.Warn("Failed to remove existing directory", zap2.Error(err))
		}
	}

	// Create btrfs snapshot from cache (instant, copy-on-write)
	restoreScript := fmt.Sprintf("btrfs subvolume snapshot %s %s", cachePath, targetPath)
	output, err := r.HostCmdExec.Execute(restoreScript)
	if err != nil {
		return r.setFailed(ctx, restore, fmt.Sprintf("Failed to restore: %v - %s", err, string(output)), logger)
	}

	// Mark as completed
	now := metav1.Now()
	restore.Status.Phase = checkpointv1.CheckpointRestorePhaseCompleted
	restore.Status.Message = "Restore completed successfully"
	restore.Status.CompletedAt = &now
	restore.Status.RestoredPath = targetPath

	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	logger.Info("Checkpoint restored",
		zap2.String("cachePath", cachePath),
		zap2.String("targetPath", targetPath))
	return reconcile.Result{}, nil
}

func (r *CheckpointRestoreReconciler) setFailed(ctx context.Context, restore *checkpointv1.CheckpointRestore, message string, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Error("Checkpoint restore failed", zap2.String("message", message))

	now := metav1.Now()
	restore.Status.Phase = checkpointv1.CheckpointRestorePhaseFailed
	restore.Status.Message = message
	restore.Status.CompletedAt = &now

	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *CheckpointRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&checkpointv1.CheckpointRestore{}).
		Complete(r)
}
