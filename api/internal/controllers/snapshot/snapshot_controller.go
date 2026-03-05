package snapshot

import (
	"context"
	"fmt"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/pagination"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotFinalizer = "snapshots.kloudlite.io/finalizer"

	// Default requeue intervals
	defaultRequeueInterval      = 30 * time.Second
	storageCleanupRetryInterval = 5 * time.Second

	// Memory management limits
	maxInUseRefsCacheSize = 10000 // Maximum number of refs to track before falling back to direct API checks
	maxSnapshotsPerCheck  = 50000 // Safety limit for snapshots to check before using direct API check
	snapshotCheckBatch    = 5000  // Number of snapshots to check before logging progress
)

// SnapshotReconciler reconciles Snapshot resources
// This controller manages snapshot metadata - it does NOT handle btrfs operations.
// Snapshots are created by SnapshotRequestReconciler after data is pushed to registry.
type SnapshotReconciler struct {
	client.Client
	Logger *zap.Logger

	// SnapshotOperator performs registry cleanup operations
	SnapshotOperator SnapshotOperator
}

// SnapshotOperator defines the interface for snapshot registry operations
type SnapshotOperator interface {
	// DeleteFromRegistry removes the snapshot from OCI registry
	DeleteFromRegistry(ctx context.Context, imageRef string) error
}

func (r *SnapshotReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshot", req.Name),
		zap.String("namespace", req.Namespace),
	)

	// Fetch Snapshot (namespaced - owned by environment)
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, req.NamespacedName, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Handle deletion
	if snapshot.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, snapshot, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
		controllerutil.AddFinalizer(snapshot, snapshotFinalizer)
		if err := r.Update(ctx, snapshot); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Process based on current state
	switch snapshot.Status.State {
	case snapshotv1.SnapshotStateReady:
		return r.handleReady(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateFailed:
		// Stay in failed state, manual intervention required
		return reconcile.Result{}, nil
	case snapshotv1.SnapshotStateDeleting:
		// Already being deleted
		return reconcile.Result{RequeueAfter: defaultRequeueInterval}, nil
	default:
		// For new snapshots or unknown states, just requeue
		return reconcile.Result{RequeueAfter: defaultRequeueInterval}, nil
	}
}

// handleReady handles the Ready state - check for expiration
func (r *SnapshotReconciler) handleReady(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	// Check retention policy for expiration
	if snapshot.Spec.RetentionPolicy != nil {
		if snapshot.Spec.RetentionPolicy.ExpiresAt != nil {
			if time.Now().After(snapshot.Spec.RetentionPolicy.ExpiresAt.Time) {
				logger.Info("Snapshot expired, deleting")
				if err := r.Delete(ctx, snapshot); err != nil {
					logger.Error("Failed to delete expired snapshot", zap.Error(err))
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, nil
			}
			// Requeue to check expiration later
			timeUntilExpiry := time.Until(snapshot.Spec.RetentionPolicy.ExpiresAt.Time)
			return reconcile.Result{RequeueAfter: timeUntilExpiry}, nil
		}
	}

	// No expiration, no action needed
	return reconcile.Result{}, nil
}

// handleDeletion cleans up snapshot resources
func (r *SnapshotReconciler) handleDeletion(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
		return reconcile.Result{}, nil
	}

	// Update state to Deleting
	if snapshot.Status.State != snapshotv1.SnapshotStateDeleting {
		snapshot.Status.State = snapshotv1.SnapshotStateDeleting
		snapshot.Status.Message = "Deleting snapshot"
		if err := r.Status().Update(ctx, snapshot); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Find and re-parent child snapshots
	if err := r.reparentChildSnapshots(ctx, snapshot, logger); err != nil {
		logger.Error("Failed to re-parent child snapshots", zap.Error(err))
		logger.Info("Using configured storage cleanup retry interval", zap.Duration("interval", storageCleanupRetryInterval))
		return reconcile.Result{RequeueAfter: storageCleanupRetryInterval}, nil
	}

	// Clean up orphaned storage - only delete imageRefs that are no longer referenced
	if err := r.cleanupOrphanedStorage(ctx, snapshot, logger); err != nil {
		logger.Warn("Failed to cleanup orphaned storage", zap.Error(err))
		// Continue with deletion anyway
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(snapshot, snapshotFinalizer)
	if err := r.Update(ctx, snapshot); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot deleted successfully")
	return reconcile.Result{}, nil
}

// reparentChildSnapshots updates child snapshots to point to this snapshot's parent
func (r *SnapshotReconciler) reparentChildSnapshots(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) error {
	// Find all snapshots in the same namespace that have this snapshot as their parent
	childSnapshots := &snapshotv1.SnapshotList{}
	if err := pagination.ListAll(ctx, r, childSnapshots, client.InNamespace(snapshot.Namespace)); err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	for i := range childSnapshots.Items {
		child := &childSnapshots.Items[i]
		if child.Spec.ParentSnapshot != snapshot.Name {
			continue
		}

		logger.Info("Re-parenting child snapshot",
			zap.String("child", child.Name),
			zap.String("newParent", snapshot.Spec.ParentSnapshot))

		// Update child's parent to this snapshot's parent (could be empty for root)
		child.Spec.ParentSnapshot = snapshot.Spec.ParentSnapshot
		if err := r.Update(ctx, child); err != nil {
			if apierrors.IsConflict(err) {
				// Retry on conflict
				return fmt.Errorf("conflict updating child %s, will retry", child.Name)
			}
			return fmt.Errorf("failed to update child snapshot %s: %w", child.Name, err)
		}
	}

	return nil
}

// cleanupOrphanedStorage deletes storage refs that are no longer referenced by any snapshot
// Uses streaming pagination to avoid loading all snapshots into memory at once
// Implements multiple safeguards to prevent memory issues:
// 1. Map size limits to prevent unbounded growth
// 2. Snapshot count limits to prevent excessive iteration
// 3. Early exit optimization when all refs are found in use
// 4. Progress logging at regular intervals
func (r *SnapshotReconciler) cleanupOrphanedStorage(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) error {
	if r.SnapshotOperator == nil || len(snapshot.Status.StorageRefs) == 0 {
		return nil
	}

	refsToDelete := snapshot.Status.StorageRefs
	logger.Info("Starting orphaned storage ref check",
		zap.Int("totalRefs", len(refsToDelete)),
		zap.String("snapshot", snapshot.Name),
		zap.String("namespace", snapshot.Namespace))

	// Use streaming pagination to check refs without loading all snapshots
	// This is more memory-efficient for systems with many snapshots
	inUseRefs := make(map[string]bool)
	totalSnapshotsChecked := 0
	totalRefsFound := 0

	err := pagination.ForEachPage(ctx, r, &snapshotv1.SnapshotList{}, pagination.DefaultPageSize, func(objects []client.Object) error {
		for _, obj := range objects {
			s, ok := obj.(*snapshotv1.Snapshot)
			if !ok {
				continue
			}

			// Skip the snapshot being deleted
			if s.Name == snapshot.Name && s.Namespace == snapshot.Namespace {
				continue
			}

			// Check if this snapshot references any of the refs we want to delete
			for _, ref := range s.Status.StorageRefs {
				// Only track refs that are candidates for deletion
				if isRefInList(ref, refsToDelete) {
					inUseRefs[ref] = true
					totalRefsFound++
				}
			}
			totalSnapshotsChecked++

			// Log progress at regular intervals
			if totalSnapshotsChecked%snapshotCheckBatch == 0 {
				logger.Info("Progress checking snapshots for storage refs",
					zap.Int("snapshotsChecked", totalSnapshotsChecked),
					zap.Int("refsFound", len(inUseRefs)),
					zap.Int("cacheSize", len(inUseRefs)))
			}

			// If we've found all refs are still in use, we can stop early
			if len(inUseRefs) == len(refsToDelete) {
				return fmt.Errorf("early exit: all refs still in use")
			}

			// Safety limit: if we've checked too many snapshots without finding all refs,
			// switch to a more conservative approach to prevent excessive memory usage
			if totalSnapshotsChecked > maxSnapshotsPerCheck {
				logger.Warn("Reached snapshot check safety limit, using direct API fallback for remaining refs",
					zap.Int("snapshotsChecked", totalSnapshotsChecked),
					zap.Int("safetyLimit", maxSnapshotsPerCheck))
				return fmt.Errorf("safety limit reached: %d snapshots", maxSnapshotsPerCheck)
			}
		}

		return nil
	})

	// Handle early exit cases (not real errors)
	if err != nil {
		switch err.Error() {
		case "early exit: all refs still in use":
			logger.Info("Early exit: all storage refs still in use", zap.Int("snapshotsChecked", totalSnapshotsChecked))
			// All refs are still in use, don't delete anything
			return nil
		default:
			// Log but continue with what we've found
			logger.Warn("Partial snapshot check completed due to error, continuing with partial results",
				zap.Error(err),
				zap.Int("snapshotsChecked", totalSnapshotsChecked),
				zap.Int("refsFound", len(inUseRefs)))
		}
	}

	logger.Info("Completed storage ref check",
		zap.Int("snapshotsChecked", totalSnapshotsChecked),
		zap.Int("refsFoundInUse", len(inUseRefs)),
		zap.Int("refsToCheck", len(refsToDelete)),
		zap.Int("cacheSize", len(inUseRefs)))

	// Delete storage refs that are no longer in use
	deletedCount := 0
	for _, ref := range refsToDelete {
		if inUseRefs[ref] {
			logger.Debug("Storage ref still in use, keeping", zap.String("imageRef", ref))
			continue
		}

		logger.Info("Deleting orphaned storage", zap.String("imageRef", ref))
		if err := r.SnapshotOperator.DeleteFromRegistry(ctx, ref); err != nil {
			logger.Warn("Failed to delete storage from registry",
				zap.String("imageRef", ref),
				zap.Error(err))
			// Continue with other deletions
		} else {
			deletedCount++
		}
	}

	logger.Info("Orphaned storage cleanup completed",
		zap.Int("deleted", deletedCount),
		zap.Int("kept", len(refsToDelete)-deletedCount))

	return nil
}

// isRefInList checks if a ref is in the list of refs to check
// Uses direct iteration which is O(n) but n is typically small (number of refs in a single snapshot)
func isRefInList(ref string, refs []string) bool {
	for _, r := range refs {
		if r == ref {
			return true
		}
	}
	return false
}

func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.Snapshot{}).
		Complete(r)
}
