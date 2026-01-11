package snapshot

import (
	"context"
	"fmt"
	"slices"

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
	snapshotRefFinalizer = "snapshots.kloudlite.io/ref-finalizer"
)

// SnapshotRefReconciler manages SnapshotRef resources and updates refCount on Snapshots
type SnapshotRefReconciler struct {
	client.Client
	Logger *zap.Logger
}

func (r *SnapshotRefReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshotRef", req.Name),
		zap.String("namespace", req.Namespace),
	)

	// Fetch SnapshotRef
	ref := &snapshotv1.SnapshotRef{}
	if err := r.Get(ctx, req.NamespacedName, ref); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRef", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Handle deletion
	if ref.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, ref, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(ref, snapshotRefFinalizer) {
		controllerutil.AddFinalizer(ref, snapshotRefFinalizer)
		if err := r.Update(ctx, ref); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// If not bound, bind to snapshot (add to ReferencedBy)
	if !ref.Status.Bound {
		return r.bindToSnapshot(ctx, ref, logger)
	}

	// Update snapshot state in status
	return r.updateSnapshotState(ctx, ref, logger)
}

// bindToSnapshot adds the ref to ReferencedBy on the referenced Snapshot
func (r *SnapshotRefReconciler) bindToSnapshot(ctx context.Context, ref *snapshotv1.SnapshotRef, logger *zap.Logger) (reconcile.Result, error) {
	snapshotName := ref.Spec.SnapshotName

	// Get the snapshot
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Warn("Referenced snapshot not found", zap.String("snapshot", snapshotName))
			// Update status to reflect snapshot not found
			ref.Status.Bound = false
			ref.Status.SnapshotState = ""
			if err := r.Status().Update(ctx, ref); err != nil {
				logger.Error("Failed to update status", zap.Error(err))
			}
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Add ref to ReferencedBy (use ref: prefix to distinguish from environment names)
	refId := fmt.Sprintf("ref:%s/%s", ref.Namespace, ref.Name)
	if !slices.Contains(snapshot.Status.ReferencedBy, refId) {
		snapshot.Status.ReferencedBy = append(snapshot.Status.ReferencedBy, refId)
		if err := r.Status().Update(ctx, snapshot); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Error("Failed to add ref to ReferencedBy", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("Bound to snapshot, added to ReferencedBy",
		zap.String("snapshot", snapshotName),
		zap.Strings("referencedBy", snapshot.Status.ReferencedBy))

	// Update ref status
	now := metav1.Now()
	ref.Status.Bound = true
	ref.Status.BoundAt = &now
	ref.Status.SnapshotState = snapshot.Status.State
	if err := r.Status().Update(ctx, ref); err != nil {
		logger.Error("Failed to update SnapshotRef status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// updateSnapshotState syncs the snapshot state to the ref status
func (r *SnapshotRefReconciler) updateSnapshotState(ctx context.Context, ref *snapshotv1.SnapshotRef, logger *zap.Logger) (reconcile.Result, error) {
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: ref.Spec.SnapshotName}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			// Snapshot was deleted - this shouldn't happen if references are managed correctly
			logger.Warn("Snapshot was deleted while ref exists")
			ref.Status.SnapshotState = ""
			if err := r.Status().Update(ctx, ref); err != nil {
				logger.Error("Failed to update status", zap.Error(err))
			}
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Update state if changed
	if ref.Status.SnapshotState != snapshot.Status.State {
		ref.Status.SnapshotState = snapshot.Status.State
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// handleDeletion removes the ref from ReferencedBy on the referenced Snapshot
func (r *SnapshotRefReconciler) handleDeletion(ctx context.Context, ref *snapshotv1.SnapshotRef, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(ref, snapshotRefFinalizer) {
		return reconcile.Result{}, nil
	}

	// Only remove reference if we were bound
	if ref.Status.Bound {
		snapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: ref.Spec.SnapshotName}, snapshot); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to get snapshot for reference removal", zap.Error(err))
				return reconcile.Result{}, err
			}
			// Snapshot already deleted, just remove finalizer
			logger.Info("Snapshot already deleted, skipping reference removal")
		} else {
			// Remove ref from ReferencedBy
			refId := fmt.Sprintf("ref:%s/%s", ref.Namespace, ref.Name)
			originalLen := len(snapshot.Status.ReferencedBy)
			snapshot.Status.ReferencedBy = slices.DeleteFunc(snapshot.Status.ReferencedBy, func(s string) bool {
				return s == refId
			})

			// Only update if we actually removed something
			if len(snapshot.Status.ReferencedBy) < originalLen {
				if err := r.Status().Update(ctx, snapshot); err != nil {
					if apierrors.IsConflict(err) {
						return reconcile.Result{Requeue: true}, nil
					}
					logger.Error("Failed to remove ref from ReferencedBy", zap.Error(err))
					return reconcile.Result{}, err
				}
				logger.Info("Removed ref from snapshot ReferencedBy",
					zap.String("snapshot", ref.Spec.SnapshotName),
					zap.Strings("referencedBy", snapshot.Status.ReferencedBy))
			}
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(ref, snapshotRefFinalizer)
	if err := r.Update(ctx, ref); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("SnapshotRef deleted successfully")
	return reconcile.Result{}, nil
}

func (r *SnapshotRefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRef{}).
		Complete(r)
}

// Helper to create a SnapshotRef for any resource
func NewSnapshotRef(name, namespace, snapshotName, purpose string, owner metav1.Object, ownerGVK metav1.GroupVersionKind) *snapshotv1.SnapshotRef {
	ref := &snapshotv1.SnapshotRef{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: snapshotv1.SnapshotRefSpec{
			SnapshotName: snapshotName,
			Purpose:      purpose,
		},
	}

	// Set owner reference for automatic cleanup
	if owner != nil {
		ref.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         fmt.Sprintf("%s/%s", ownerGVK.Group, ownerGVK.Version),
				Kind:               ownerGVK.Kind,
				Name:               owner.GetName(),
				UID:                owner.GetUID(),
				Controller:         boolPtr(true),
				BlockOwnerDeletion: boolPtr(true),
			},
		}
	}

	return ref
}

func boolPtr(b bool) *bool {
	return &b
}
