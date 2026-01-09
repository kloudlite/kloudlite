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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotRequestFinalizer = "snapshots.kloudlite.io/request-finalizer"
	retryRequeueAfter        = 5 * time.Second
)

// SnapshotRequestReconciler reconciles SnapshotRequest resources
// This controller runs on each node and handles node-local btrfs operations.
// After pushing data to the registry, it creates the global Snapshot resource.
type SnapshotRequestReconciler struct {
	client.Client
	Logger *zap.Logger

	// NodeName is the name of the node this reconciler runs on
	// Only requests targeting this node will be processed
	NodeName string

	// SnapshotRequestOperator performs actual snapshot operations (btrfs + OCI)
	SnapshotRequestOperator SnapshotRequestOperator
}

// SnapshotRequestOperator defines the interface for node-local snapshot operations
type SnapshotRequestOperator interface {
	// CreateSnapshot creates a btrfs snapshot at the given path
	CreateSnapshot(ctx context.Context, sourcePath, snapshotPath string) error

	// PushToRegistry pushes the snapshot to OCI registry
	// Returns the image reference, digest, and size
	PushToRegistry(ctx context.Context, request *snapshotv1.SnapshotRequest, store *snapshotv1.SnapshotStore, parentSnapshot *snapshotv1.Snapshot) (*RegistryPushResult, error)

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
	SizeBytes      int64
}

func (r *SnapshotRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshotRequest", req.NamespacedName.String()),
	)

	// Fetch SnapshotRequest
	snapshotReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, req.NamespacedName, snapshotReq); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRequest", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Only process requests for this node
	if snapshotReq.Spec.NodeName != r.NodeName {
		// Not our node, skip
		return reconcile.Result{}, nil
	}

	// Handle completed or failed requests - no need to reprocess
	if snapshotReq.Status.State == snapshotv1.SnapshotRequestStateCompleted ||
		snapshotReq.Status.State == snapshotv1.SnapshotRequestStateFailed {
		return reconcile.Result{}, nil
	}

	// Process based on current state
	switch snapshotReq.Status.State {
	case "", snapshotv1.SnapshotRequestStatePending:
		return r.handlePending(ctx, snapshotReq, logger)
	case snapshotv1.SnapshotRequestStateCreating:
		return r.handleCreating(ctx, snapshotReq, logger)
	case snapshotv1.SnapshotRequestStateUploading:
		return r.handleUploading(ctx, snapshotReq, logger)
	default:
		logger.Warn("Unknown snapshot request state", zap.String("state", string(snapshotReq.Status.State)))
		return reconcile.Result{}, nil
	}
}

// handlePending transitions to Creating state
func (r *SnapshotRequestReconciler) handlePending(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Starting snapshot request processing",
		zap.String("sourcePath", req.Spec.SourcePath),
		zap.String("snapshotName", req.Spec.SnapshotName),
		zap.String("owner", req.Spec.Owner))

	// Transition to Creating state
	now := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateCreating
	req.Status.Message = "Creating btrfs snapshot"
	req.Status.StartedAt = &now
	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleCreating creates the btrfs snapshot
func (r *SnapshotRequestReconciler) handleCreating(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap.Logger) (reconcile.Result, error) {
	// Generate local snapshot path
	snapshotPath := fmt.Sprintf("/data/.snapshots/%s", req.Spec.SnapshotName)

	// Create btrfs snapshot
	if err := r.SnapshotRequestOperator.CreateSnapshot(ctx, req.Spec.SourcePath, snapshotPath); err != nil {
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to create btrfs snapshot: %v", err), logger)
	}

	// Update status with local path and transition to Uploading
	req.Status.LocalSnapshotPath = snapshotPath
	req.Status.State = snapshotv1.SnapshotRequestStateUploading
	req.Status.Message = "Uploading to OCI registry"

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Created btrfs snapshot", zap.String("path", snapshotPath))
	return reconcile.Result{Requeue: true}, nil
}

// handleUploading pushes snapshot to OCI registry and creates global Snapshot
func (r *SnapshotRequestReconciler) handleUploading(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap.Logger) (reconcile.Result, error) {
	// Get the SnapshotStore
	store := &snapshotv1.SnapshotStore{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.Store}, store); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("SnapshotStore %q not found", req.Spec.Store), logger)
		}
		logger.Error("Failed to get SnapshotStore", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if store is ready
	if !store.Status.Ready {
		logger.Info("SnapshotStore not ready, waiting", zap.String("store", store.Name))
		return reconcile.Result{RequeueAfter: retryRequeueAfter}, nil
	}

	// Get parent snapshot if specified (for incremental send)
	var parentSnapshot *snapshotv1.Snapshot
	if req.Spec.ParentSnapshot != "" {
		parentSnapshot = &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.ParentSnapshot}, parentSnapshot); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Warn("Parent snapshot not found, will create full snapshot",
					zap.String("parent", req.Spec.ParentSnapshot))
				parentSnapshot = nil
			} else {
				logger.Error("Failed to get parent snapshot", zap.Error(err))
				return reconcile.Result{}, err
			}
		}
	}

	// Get snapshot size
	size, err := r.SnapshotRequestOperator.GetSnapshotSize(ctx, req.Status.LocalSnapshotPath)
	if err != nil {
		logger.Warn("Failed to get snapshot size", zap.Error(err))
		// Continue anyway, size is optional
	}

	// Push to registry
	result, err := r.SnapshotRequestOperator.PushToRegistry(ctx, req, store, parentSnapshot)
	if err != nil {
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to push to registry: %v", err), logger)
	}

	// Build lineage
	var lineage []string
	if parentSnapshot != nil {
		lineage = append(parentSnapshot.Status.Lineage, parentSnapshot.Name)
	}

	// Create the global Snapshot resource
	now := metav1.Now()
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Spec.SnapshotName,
			Labels: req.Labels, // Inherit labels from request
		},
		Spec: snapshotv1.SnapshotSpec{
			Owner:           req.Spec.Owner,
			ParentSnapshot:  req.Spec.ParentSnapshot,
			Description:     req.Spec.Description,
			Artifacts:       req.Spec.Artifacts,
			RetentionPolicy: req.Spec.RetentionPolicy,
		},
		Status: snapshotv1.SnapshotStatus{
			State:     snapshotv1.SnapshotStateReady,
			Message:   "Snapshot ready",
			SizeBytes: size,
			SizeHuman: formatSize(size),
			CreatedAt: &now,
			Lineage:   lineage,
			Registry: &snapshotv1.SnapshotRegistryInfo{
				ImageRef:       result.ImageRef,
				Digest:         result.Digest,
				PushedAt:       &now,
				CompressedSize: result.CompressedSize,
			},
		},
	}

	// Set parent image ref if we have a parent
	if parentSnapshot != nil && parentSnapshot.Status.Registry != nil {
		snapshot.Status.Registry.ParentImageRef = parentSnapshot.Status.Registry.ImageRef
	}

	// Create artifact status
	for _, artifact := range req.Spec.Artifacts {
		snapshot.Status.Artifacts = append(snapshot.Status.Artifacts, snapshotv1.ArtifactStatus{
			Name:   artifact.Name,
			Type:   artifact.Type,
			Stored: true,
		})
	}

	if err := r.Create(ctx, snapshot); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create Snapshot resource: %v", err), logger)
		}
		// Already exists, update it
		existing := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.SnapshotName}, existing); err != nil {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to get existing Snapshot: %v", err), logger)
		}
		existing.Status = snapshot.Status
		if err := r.Status().Update(ctx, existing); err != nil {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to update existing Snapshot: %v", err), logger)
		}
	}

	// Clean up local snapshot
	if req.Status.LocalSnapshotPath != "" {
		if err := r.SnapshotRequestOperator.DeleteLocalSnapshot(ctx, req.Status.LocalSnapshotPath); err != nil {
			logger.Warn("Failed to delete local snapshot", zap.Error(err))
			// Continue anyway
		}
	}

	// Mark request as completed
	completedNow := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateCompleted
	req.Status.Message = "Snapshot created successfully"
	req.Status.CompletedAt = &completedNow
	req.Status.CreatedSnapshot = req.Spec.SnapshotName

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot created successfully",
		zap.String("snapshot", req.Spec.SnapshotName),
		zap.String("imageRef", result.ImageRef),
		zap.String("digest", result.Digest))

	return reconcile.Result{}, nil
}

// setFailed updates the request to Failed state
func (r *SnapshotRequestReconciler) setFailed(ctx context.Context, req *snapshotv1.SnapshotRequest, message string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot request failed", zap.String("message", message))

	now := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateFailed
	req.Status.Message = message
	req.Status.CompletedAt = &now

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
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

func (r *SnapshotRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRequest{}).
		Complete(r)
}
