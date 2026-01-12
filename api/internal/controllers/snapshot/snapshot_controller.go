package snapshot

import (
	"context"
	"fmt"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotFinalizer = "snapshots.kloudlite.io/finalizer"

	// Requeue intervals
	defaultRequeueAfter = 30 * time.Second
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
		return reconcile.Result{RequeueAfter: defaultRequeueAfter}, nil
	default:
		// For new snapshots or unknown states, just requeue
		return reconcile.Result{RequeueAfter: defaultRequeueAfter}, nil
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
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
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
	if err := r.List(ctx, childSnapshots, client.InNamespace(snapshot.Namespace)); err != nil {
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
func (r *SnapshotReconciler) cleanupOrphanedStorage(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) error {
	if r.SnapshotOperator == nil || len(snapshot.Status.StorageRefs) == 0 {
		return nil
	}

	// Get all snapshots across ALL namespaces to check which storage refs are still in use
	// Storage is shared across environments (via deep clone), so we need to check globally
	allSnapshots := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, allSnapshots); err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Build a set of all imageRefs still in use by other snapshots
	inUseRefs := make(map[string]bool)
	for _, s := range allSnapshots.Items {
		// Skip the snapshot being deleted (check both name and namespace)
		if s.Name == snapshot.Name && s.Namespace == snapshot.Namespace {
			continue
		}
		for _, ref := range s.Status.StorageRefs {
			inUseRefs[ref] = true
		}
	}

	// Delete storage refs that are no longer in use
	for _, ref := range snapshot.Status.StorageRefs {
		if inUseRefs[ref] {
			logger.Info("Storage ref still in use, keeping", zap.String("imageRef", ref))
			continue
		}

		logger.Info("Deleting orphaned storage", zap.String("imageRef", ref))
		if err := r.SnapshotOperator.DeleteFromRegistry(ctx, ref); err != nil {
			logger.Warn("Failed to delete storage from registry",
				zap.String("imageRef", ref),
				zap.Error(err))
			// Continue with other deletions
		}
	}

	return nil
}

func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.Snapshot{}).
		Complete(r)
}
