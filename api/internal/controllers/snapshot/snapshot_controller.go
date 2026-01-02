package snapshot

import (
	"context"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotFinalizer = "snapshots.kloudlite.io/finalizer"
	snapshotsBasePath = "/var/lib/kloudlite/storage/.snapshots"
	metadataBasePath  = "/var/lib/kloudlite/storage/.snapshots-metadata"
	workspaceHomePath = "/var/lib/kloudlite/home/workspaces"
)

// SnapshotReconciler reconciles Snapshot objects
type SnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles Snapshot events
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshot", req.Name),
	)

	logger.Info("Reconciling Snapshot")

	// Fetch the Snapshot instance (cluster-scoped)
	snapshot := &snapshotv1.Snapshot{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name}, snapshot)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Snapshot not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if snapshot is being deleted
	if snapshot.DeletionTimestamp != nil {
		logger.Info("Snapshot is being deleted, starting cleanup")
		return r.handleDeletion(ctx, snapshot, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
		logger.Info("Adding finalizer to snapshot")
		controllerutil.AddFinalizer(snapshot, snapshotFinalizer)
		if err := r.Update(ctx, snapshot); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Handle based on current state
	switch snapshot.Status.State {
	case "", snapshotv1.SnapshotStatePending:
		return r.handlePending(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateCreating:
		return r.handleCreating(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateReady:
		return reconcile.Result{}, nil
	case snapshotv1.SnapshotStateRestoring:
		return r.handleRestoring(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateDeleting:
		return r.handleDeleting(ctx, snapshot, logger)
	case snapshotv1.SnapshotStatePushing:
		return r.handlePushing(ctx, snapshot, logger)
	case snapshotv1.SnapshotStatePulling:
		return r.handlePulling(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateFailed:
		return reconcile.Result{}, nil
	default:
		logger.Warn("Unknown snapshot state", zap.String("state", string(snapshot.Status.State)))
		return reconcile.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.Snapshot{}).
		Owns(&snapshotv1.SnapshotRequest{}).
		Complete(r)
}
