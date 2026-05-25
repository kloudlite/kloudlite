package checkpoint

import (
	"context"
	"time"

	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const checkpointFinalizer = "checkpoints.kloudlite.io/finalizer"

// CheckpointOperator defines the interface for checkpoint registry operations
type CheckpointOperator interface {
	// DeleteFromRegistry removes the checkpoint layer from OCI registry
	DeleteFromRegistry(ctx context.Context, imageRef string) error
}

// CheckpointReconciler reconciles Checkpoint resources on the api-server.
// It handles expiration checks and registry cleanup on deletion.
// The worker node (workmachine-node-manager) handles the btrfs + upload lifecycle.
type CheckpointReconciler struct {
	client.Client
	Logger             *zap.Logger
	CheckpointOperator CheckpointOperator
}

func (r *CheckpointReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("checkpoint", req.Name),
		zap.String("namespace", req.Namespace),
	)

	cp := &checkpointv1.Checkpoint{}
	if err := r.Get(ctx, req.NamespacedName, cp); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Checkpoint", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Handle deletion
	if cp.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, cp, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(cp, checkpointFinalizer) {
		controllerutil.AddFinalizer(cp, checkpointFinalizer)
		if err := r.Update(ctx, cp); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Only check expiration for Ready checkpoints
	if cp.Status.Phase == checkpointv1.CheckpointPhaseReady {
		return r.handleReady(ctx, cp, logger)
	}

	return reconcile.Result{}, nil
}

// handleReady checks retention policy for expiration
func (r *CheckpointReconciler) handleReady(ctx context.Context, cp *checkpointv1.Checkpoint, logger *zap.Logger) (reconcile.Result, error) {
	if cp.Spec.RetentionPolicy == nil {
		return reconcile.Result{}, nil
	}

	if cp.Spec.RetentionPolicy.ExpiresAt != nil {
		if time.Now().After(cp.Spec.RetentionPolicy.ExpiresAt.Time) {
			logger.Info("Checkpoint expired, deleting")
			if err := r.Delete(ctx, cp); err != nil {
				logger.Error("Failed to delete expired checkpoint", zap.Error(err))
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		timeUntilExpiry := time.Until(cp.Spec.RetentionPolicy.ExpiresAt.Time)
		return reconcile.Result{RequeueAfter: timeUntilExpiry}, nil
	}

	return reconcile.Result{}, nil
}

// handleDeletion cleans up registry storage and removes finalizer
func (r *CheckpointReconciler) handleDeletion(ctx context.Context, cp *checkpointv1.Checkpoint, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(cp, checkpointFinalizer) {
		return reconcile.Result{}, nil
	}

	// Update phase to Deleting
	if cp.Status.Phase != checkpointv1.CheckpointPhaseDeleting {
		cp.Status.Phase = checkpointv1.CheckpointPhaseDeleting
		cp.Status.Message = "Deleting checkpoint"
		if err := r.Status().Update(ctx, cp); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Delete from registry if layer info exists
	if r.CheckpointOperator != nil && cp.Status.Layer != nil && cp.Status.Layer.ImageRef != "" {
		logger.Info("Deleting checkpoint from registry", zap.String("imageRef", cp.Status.Layer.ImageRef))
		if err := r.CheckpointOperator.DeleteFromRegistry(ctx, cp.Status.Layer.ImageRef); err != nil {
			logger.Warn("Failed to delete from registry", zap.String("imageRef", cp.Status.Layer.ImageRef), zap.Error(err))
			// Continue with deletion anyway — registry cleanup is best-effort
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(cp, checkpointFinalizer)
	if err := r.Update(ctx, cp); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Checkpoint deleted successfully")
	return reconcile.Result{}, nil
}

func (r *CheckpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&checkpointv1.Checkpoint{}).
		Complete(r)
}
