package snapshot

import (
	"context"
	"fmt"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotRestoreFinalizer       = "snapshots.kloudlite.io/restore-finalizer"
	snapshotRestoreWaitInterval = 5 * time.Second
)

// RestoreOperator defines the interface for restore operations
type RestoreOperator interface {
	// PullFromRegistry pulls a snapshot and its lineage from the registry
	// Uses the imageRef from the snapshot's registry info
	PullFromRegistry(ctx context.Context, snapshot *snapshotv1.Snapshot, targetPath string) error

	// RestoreSnapshot restores a btrfs snapshot to the target path
	RestoreSnapshot(ctx context.Context, snapshotPath, targetPath string) error

	// GetArtifacts retrieves artifacts from the snapshot
	GetArtifacts(ctx context.Context, snapshot *snapshotv1.Snapshot, artifactNames []string) (map[string]string, error)

	// CleanupTempData cleans up temporary restore data
	CleanupTempData(ctx context.Context, tempPath string) error
}

// SnapshotRestoreReconciler reconciles SnapshotRestore resources
// This controller runs on each node and handles restore operations for that node.
type SnapshotRestoreReconciler struct {
	client.Client
	Logger *zap.Logger

	// NodeName is the name of the node this reconciler runs on
	// Only restore requests targeting this node will be processed
	NodeName string

	// RestoreOperator performs actual restore operations
	RestoreOperator RestoreOperator
}

func (r *SnapshotRestoreReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshotRestore", req.Name),
		zap.String("namespace", req.Namespace),
	)

	// Fetch SnapshotRestore
	restore := &snapshotv1.SnapshotRestore{}
	if err := r.Get(ctx, req.NamespacedName, restore); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRestore", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Only process requests for this node
	if restore.Spec.NodeName != r.NodeName {
		// Not our node, skip
		return reconcile.Result{}, nil
	}

	// Handle deletion
	if restore.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, restore, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(restore, snapshotRestoreFinalizer) {
		controllerutil.AddFinalizer(restore, snapshotRestoreFinalizer)
		if err := r.Update(ctx, restore); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Process based on current state
	switch restore.Status.State {
	case "", snapshotv1.SnapshotRestoreStatePending:
		return r.handlePending(ctx, restore, logger)
	case snapshotv1.SnapshotRestoreStateDownloading:
		return r.handleDownloading(ctx, restore, logger)
	case snapshotv1.SnapshotRestoreStateRestoring:
		return r.handleRestoring(ctx, restore, logger)
	case snapshotv1.SnapshotRestoreStateCompleted:
		// Nothing more to do
		return reconcile.Result{}, nil
	case snapshotv1.SnapshotRestoreStateFailed:
		// Stay in failed state
		return reconcile.Result{}, nil
	default:
		logger.Warn("Unknown restore state", zap.String("state", string(restore.Status.State)))
		return reconcile.Result{}, nil
	}
}

// handlePending validates the restore request and starts downloading
func (r *SnapshotRestoreReconciler) handlePending(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Starting snapshot restore",
		zap.String("snapshot", restore.Spec.SnapshotName),
		zap.String("targetPath", restore.Spec.TargetPath))

	// Get the snapshot
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.SnapshotName, Namespace: restore.Namespace}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, restore, fmt.Sprintf("Snapshot %q not found", restore.Spec.SnapshotName), logger)
		}
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if snapshot is ready
	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		logger.Info("Snapshot not ready, waiting", zap.String("state", string(snapshot.Status.State)))
		restore.Status.Message = fmt.Sprintf("Waiting for snapshot to be ready (current: %s)", snapshot.Status.State)
		if err := r.Status().Update(ctx, restore); err != nil {
			if !apierrors.IsConflict(err) {
				logger.Error("Failed to update status", zap.Error(err))
			}
		}
		logger.Info("Using configured snapshot restore wait interval", zap.Duration("interval", snapshotRestoreWaitInterval))
		return reconcile.Result{RequeueAfter: snapshotRestoreWaitInterval}, nil
	}

	// Ensure snapshot has registry info
	if snapshot.Status.Registry == nil || snapshot.Status.Registry.ImageRef == "" {
		return r.setFailed(ctx, restore, "Snapshot has no registry info", logger)
	}

	// Start downloading from registry
	now := metav1.Now()
	restore.Status.State = snapshotv1.SnapshotRestoreStateDownloading
	restore.Status.Message = "Downloading snapshot from registry"
	restore.Status.StartedAt = &now
	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleDownloading pulls the snapshot from registry
func (r *SnapshotRestoreReconciler) handleDownloading(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap.Logger) (reconcile.Result, error) {
	// Get the snapshot
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.SnapshotName, Namespace: restore.Namespace}, snapshot); err != nil {
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Pull from registry (including lineage)
	// The target path is a temporary location for the snapshot data
	tempPath := fmt.Sprintf("/data/.snapshots-restore/%s", restore.Name)
	if err := r.RestoreOperator.PullFromRegistry(ctx, snapshot, tempPath); err != nil {
		return r.setFailed(ctx, restore, fmt.Sprintf("Failed to pull from registry: %v", err), logger)
	}

	// Transition to Restoring
	restore.Status.State = snapshotv1.SnapshotRestoreStateRestoring
	restore.Status.Message = "Restoring snapshot data"
	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot downloaded successfully", zap.String("tempPath", tempPath))
	return reconcile.Result{Requeue: true}, nil
}

// handleRestoring restores the snapshot to the target path
func (r *SnapshotRestoreReconciler) handleRestoring(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap.Logger) (reconcile.Result, error) {
	// Get the snapshot
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.SnapshotName, Namespace: restore.Namespace}, snapshot); err != nil {
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// The temp path where we downloaded the snapshot
	tempPath := fmt.Sprintf("/data/.snapshots-restore/%s", restore.Name)

	// Restore the snapshot to target path
	if err := r.RestoreOperator.RestoreSnapshot(ctx, tempPath, restore.Spec.TargetPath); err != nil {
		return r.setFailed(ctx, restore, fmt.Sprintf("Failed to restore snapshot: %v", err), logger)
	}

	// Get artifacts if requested
	var artifacts map[string]string
	if len(restore.Spec.IncludeArtifacts) > 0 || len(snapshot.Spec.Artifacts) > 0 {
		var artifactNames []string
		if len(restore.Spec.IncludeArtifacts) > 0 {
			artifactNames = restore.Spec.IncludeArtifacts
		} else {
			// Include all artifacts
			for _, a := range snapshot.Spec.Artifacts {
				artifactNames = append(artifactNames, a.Name)
			}
		}

		var err error
		artifacts, err = r.RestoreOperator.GetArtifacts(ctx, snapshot, artifactNames)
		if err != nil {
			logger.Warn("Failed to get artifacts", zap.Error(err))
			// Continue - artifacts are optional
		}
	}

	// Clean up temp data
	if err := r.RestoreOperator.CleanupTempData(ctx, tempPath); err != nil {
		logger.Warn("Failed to cleanup temp data", zap.Error(err))
		// Continue - cleanup failure is not critical
	}

	// Mark as completed
	now := metav1.Now()
	restore.Status.State = snapshotv1.SnapshotRestoreStateCompleted
	restore.Status.Message = "Restore completed successfully"
	restore.Status.CompletedAt = &now
	restore.Status.RestoredPath = restore.Spec.TargetPath
	restore.Status.Artifacts = artifacts

	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot restored successfully",
		zap.String("targetPath", restore.Spec.TargetPath),
		zap.Int("artifactCount", len(artifacts)))

	return reconcile.Result{}, nil
}

// handleDeletion cleans up any temporary restore data
func (r *SnapshotRestoreReconciler) handleDeletion(ctx context.Context, restore *snapshotv1.SnapshotRestore, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(restore, snapshotRestoreFinalizer) {
		return reconcile.Result{}, nil
	}

	// Clean up temp data if it still exists
	tempPath := fmt.Sprintf("/data/.snapshots-restore/%s", restore.Name)
	if err := r.RestoreOperator.CleanupTempData(ctx, tempPath); err != nil {
		logger.Warn("Failed to cleanup temp data on deletion", zap.Error(err))
		// Continue - cleanup failure is not critical
	}

	// Note: We don't clean up the restored data at targetPath
	// That's up to the owner of the restore request to manage

	// Remove finalizer
	controllerutil.RemoveFinalizer(restore, snapshotRestoreFinalizer)
	if err := r.Update(ctx, restore); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("SnapshotRestore deleted successfully")
	return reconcile.Result{}, nil
}

// setFailed updates the restore to Failed state
func (r *SnapshotRestoreReconciler) setFailed(ctx context.Context, restore *snapshotv1.SnapshotRestore, message string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Error("Restore failed", zap.String("message", message))

	restore.Status.State = snapshotv1.SnapshotRestoreStateFailed
	restore.Status.Message = message

	if err := r.Status().Update(ctx, restore); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SnapshotRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRestore{}).
		Complete(r)
}
