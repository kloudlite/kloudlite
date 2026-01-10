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
	)

	// Fetch Snapshot (cluster-scoped)
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, snapshot); err != nil {
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

	// Check if snapshot should be garbage collected (refCount == 0 and not protected)
	if snapshot.Status.RefCount == 0 && snapshot.Status.State == snapshotv1.SnapshotStateReady {
		// Check retention policy
		if shouldGarbageCollect(snapshot) {
			logger.Info("Garbage collecting snapshot with refCount=0")
			if err := r.Delete(ctx, snapshot); err != nil {
				logger.Error("Failed to delete snapshot for garbage collection", zap.Error(err))
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
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

	// Check if refCount > 0 - cannot delete if still referenced
	if snapshot.Status.RefCount > 0 {
		logger.Warn("Cannot delete snapshot with active references",
			zap.Int32("refCount", snapshot.Status.RefCount))
		snapshot.Status.Message = fmt.Sprintf("Cannot delete: %d active references", snapshot.Status.RefCount)
		if err := r.Status().Update(ctx, snapshot); err != nil {
			if !apierrors.IsConflict(err) {
				logger.Error("Failed to update status", zap.Error(err))
			}
		}
		// Requeue to check again later
		return reconcile.Result{RequeueAfter: defaultRequeueAfter}, nil
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

	// Delete from registry if pushed
	if snapshot.Status.Registry != nil && snapshot.Status.Registry.ImageRef != "" {
		logger.Info("Deleting from registry", zap.String("imageRef", snapshot.Status.Registry.ImageRef))
		if err := r.SnapshotOperator.DeleteFromRegistry(ctx, snapshot.Status.Registry.ImageRef); err != nil {
			logger.Warn("Failed to delete from registry", zap.Error(err))
			// Continue with deletion anyway
		}
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

// shouldGarbageCollect determines if a snapshot should be garbage collected
func shouldGarbageCollect(snapshot *snapshotv1.Snapshot) bool {
	// If snapshot has retention policy with expiration, let that handle it
	if snapshot.Spec.RetentionPolicy != nil {
		if snapshot.Spec.RetentionPolicy.ExpiresAt != nil || snapshot.Spec.RetentionPolicy.KeepForDays != nil {
			return false
		}
	}

	// RefCount is 0 and no retention policy - garbage collect
	// This handles orphaned snapshots that are no longer referenced
	return true
}

func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.Snapshot{}).
		Complete(r)
}
