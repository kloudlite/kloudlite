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
	snapshotFinalizer = "snapshots.kloudlite.io/finalizer"

	// Requeue intervals
	defaultRequeueAfter = 30 * time.Second
	retryRequeueAfter   = 5 * time.Second
)

// SnapshotReconciler reconciles Snapshot resources
// This is a GENERIC controller - it does not know about environments or workspaces
type SnapshotReconciler struct {
	client.Client
	Logger *zap.Logger

	// SnapshotOperator performs actual snapshot operations (btrfs + OCI)
	// This is injected to allow different implementations (e.g., for testing)
	SnapshotOperator SnapshotOperator
}

// SnapshotOperator defines the interface for snapshot operations
type SnapshotOperator interface {
	// CreateSnapshot creates a btrfs snapshot at the given path
	CreateSnapshot(ctx context.Context, sourcePath, snapshotPath string) error

	// PushToRegistry pushes the snapshot to OCI registry
	// Returns the image reference, digest, and size
	PushToRegistry(ctx context.Context, snapshot *snapshotv1.Snapshot, store *snapshotv1.SnapshotStore) (*RegistryPushResult, error)

	// DeleteFromRegistry removes the snapshot from OCI registry
	DeleteFromRegistry(ctx context.Context, imageRef string) error

	// DeleteLocalSnapshot removes the local btrfs snapshot
	DeleteLocalSnapshot(ctx context.Context, snapshotPath string) error

	// GetSnapshotSize returns the size of a local snapshot in bytes
	GetSnapshotSize(ctx context.Context, snapshotPath string) (int64, error)
}

// RegistryPushResult contains information about a pushed snapshot
type RegistryPushResult struct {
	ImageRef       string
	Digest         string
	CompressedSize int64
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
	case "", snapshotv1.SnapshotStatePending:
		return r.handlePending(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateCreating:
		return r.handleCreating(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateUploading:
		return r.handleUploading(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateReady:
		return r.handleReady(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateFailed:
		// Stay in failed state, manual intervention required
		return reconcile.Result{}, nil
	case snapshotv1.SnapshotStateDeleting:
		// Already being deleted
		return reconcile.Result{RequeueAfter: retryRequeueAfter}, nil
	default:
		logger.Warn("Unknown snapshot state", zap.String("state", string(snapshot.Status.State)))
		return reconcile.Result{}, nil
	}
}

// handlePending transitions to Creating state
func (r *SnapshotReconciler) handlePending(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Starting snapshot creation",
		zap.String("sourcePath", snapshot.Spec.SourcePath),
		zap.String("owner", snapshot.Spec.Owner))

	// Build lineage from parent snapshot
	if err := r.buildLineage(ctx, snapshot, logger); err != nil {
		return r.setFailed(ctx, snapshot, fmt.Sprintf("Failed to build lineage: %v", err), logger)
	}

	// Transition to Creating state
	snapshot.Status.State = snapshotv1.SnapshotStateCreating
	snapshot.Status.Message = "Creating btrfs snapshot"
	if err := r.Status().Update(ctx, snapshot); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleCreating creates the btrfs snapshot
func (r *SnapshotReconciler) handleCreating(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	// Generate local snapshot path
	snapshotPath := fmt.Sprintf("/data/.snapshots/%s", snapshot.Name)

	// Create btrfs snapshot
	if err := r.SnapshotOperator.CreateSnapshot(ctx, snapshot.Spec.SourcePath, snapshotPath); err != nil {
		return r.setFailed(ctx, snapshot, fmt.Sprintf("Failed to create btrfs snapshot: %v", err), logger)
	}

	// Get snapshot size
	size, err := r.SnapshotOperator.GetSnapshotSize(ctx, snapshotPath)
	if err != nil {
		logger.Warn("Failed to get snapshot size", zap.Error(err))
		// Continue anyway, size is optional
	}

	// Update status with local path and transition to Uploading
	snapshot.Status.LocalPath = snapshotPath
	snapshot.Status.SizeBytes = size
	snapshot.Status.SizeHuman = formatSize(size)
	snapshot.Status.State = snapshotv1.SnapshotStateUploading
	snapshot.Status.Message = "Uploading to OCI registry"

	if err := r.Status().Update(ctx, snapshot); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Created btrfs snapshot", zap.String("path", snapshotPath), zap.Int64("sizeBytes", size))
	return reconcile.Result{Requeue: true}, nil
}

// handleUploading pushes snapshot to OCI registry
func (r *SnapshotReconciler) handleUploading(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	// Get the SnapshotStore
	store := &snapshotv1.SnapshotStore{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshot.Spec.Store}, store); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, snapshot, fmt.Sprintf("SnapshotStore %q not found", snapshot.Spec.Store), logger)
		}
		logger.Error("Failed to get SnapshotStore", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if store is ready
	if !store.Status.Ready {
		logger.Info("SnapshotStore not ready, waiting", zap.String("store", store.Name))
		return reconcile.Result{RequeueAfter: retryRequeueAfter}, nil
	}

	// Push to registry
	result, err := r.SnapshotOperator.PushToRegistry(ctx, snapshot, store)
	if err != nil {
		return r.setFailed(ctx, snapshot, fmt.Sprintf("Failed to push to registry: %v", err), logger)
	}

	// Update status with registry info
	now := metav1.Now()
	snapshot.Status.Registry = &snapshotv1.SnapshotRegistryInfo{
		ImageRef:       result.ImageRef,
		Digest:         result.Digest,
		PushedAt:       &now,
		CompressedSize: result.CompressedSize,
	}

	// Set parent image ref if we have a parent
	if snapshot.Spec.ParentSnapshot != "" {
		parentSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: snapshot.Spec.ParentSnapshot}, parentSnapshot); err == nil {
			if parentSnapshot.Status.Registry != nil {
				snapshot.Status.Registry.ParentImageRef = parentSnapshot.Status.Registry.ImageRef
			}
		}
	}

	snapshot.Status.State = snapshotv1.SnapshotStateReady
	snapshot.Status.Message = "Snapshot ready"
	snapshot.Status.CreatedAt = &now

	// Update artifact status
	for _, artifact := range snapshot.Spec.Artifacts {
		snapshot.Status.Artifacts = append(snapshot.Status.Artifacts, snapshotv1.ArtifactStatus{
			Name:   artifact.Name,
			Type:   artifact.Type,
			Stored: true,
		})
	}

	if err := r.Status().Update(ctx, snapshot); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot uploaded successfully",
		zap.String("imageRef", result.ImageRef),
		zap.String("digest", result.Digest))

	return reconcile.Result{}, nil
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

	// Delete local snapshot if exists
	if snapshot.Status.LocalPath != "" {
		logger.Info("Deleting local snapshot", zap.String("path", snapshot.Status.LocalPath))
		if err := r.SnapshotOperator.DeleteLocalSnapshot(ctx, snapshot.Status.LocalPath); err != nil {
			logger.Warn("Failed to delete local snapshot", zap.Error(err))
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

// buildLineage constructs the lineage from parent snapshots
func (r *SnapshotReconciler) buildLineage(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) error {
	if snapshot.Spec.ParentSnapshot == "" {
		snapshot.Status.Lineage = []string{}
		return nil
	}

	// Get parent snapshot
	parent := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshot.Spec.ParentSnapshot}, parent); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("parent snapshot %q not found", snapshot.Spec.ParentSnapshot)
		}
		return err
	}

	// Build lineage: parent's lineage + parent name
	snapshot.Status.Lineage = append(parent.Status.Lineage, parent.Name)
	logger.Info("Built lineage", zap.Strings("lineage", snapshot.Status.Lineage))

	return nil
}

// setFailed updates the snapshot to Failed state
func (r *SnapshotReconciler) setFailed(ctx context.Context, snapshot *snapshotv1.Snapshot, message string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot failed", zap.String("message", message))

	snapshot.Status.State = snapshotv1.SnapshotStateFailed
	snapshot.Status.Message = message

	if err := r.Status().Update(ctx, snapshot); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

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

	// No retention policy and no references - can be garbage collected
	// However, we might want to keep snapshots for a grace period
	// For now, don't auto-delete - require explicit deletion or expiration
	return false
}

// formatSize converts bytes to human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.Snapshot{}).
		Complete(r)
}
