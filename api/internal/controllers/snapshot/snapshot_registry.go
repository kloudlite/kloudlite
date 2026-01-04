package snapshot

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handlePushing handles the snapshot push to registry process
func (r *SnapshotReconciler) handlePushing(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling snapshot push to registry")

	if snapshot.Spec.RegistryRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Registry reference is required for push operation", logger)
	}

	// Determine the repository and tag
	repository := snapshot.Spec.RegistryRef.Repository
	if repository == "" {
		// Build repository path based on snapshot type:
		// - Workspace snapshots: snapshots/{user}/ws/{workspace_name}
		// - Environment snapshots: snapshots/{user}/env/{environment_name}
		if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
			repository = fmt.Sprintf("snapshots/%s/ws/%s", snapshot.Spec.OwnedBy, snapshot.Status.WorkspaceName)
		} else if snapshot.Spec.EnvironmentRef != nil {
			repository = fmt.Sprintf("snapshots/%s/env/%s", snapshot.Spec.OwnedBy, snapshot.Spec.EnvironmentRef.Name)
		} else {
			// Fallback to old format if type cannot be determined
			repository = fmt.Sprintf("snapshots/%s", snapshot.Spec.OwnedBy)
		}
	}
	tag := snapshot.Spec.RegistryRef.Tag
	if tag == "" {
		tag = snapshot.Name
	}

	// Remove the tag from any other snapshot that has it (tags are unique per repository)
	targetImageRef := fmt.Sprintf("image-registry:5000/%s:%s", repository, tag)
	if err := r.removeTagFromOtherSnapshots(ctx, snapshot.Name, targetImageRef, logger); err != nil {
		logger.Warn("Failed to remove tag from other snapshots", zap.Error(err))
	}

	// Get parent snapshot's registry layers if parent exists and was pushed
	var parentLayers []string
	var parentSnapshotPath string
	if snapshot.Spec.ParentSnapshotRef != nil {
		parentSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: snapshot.Spec.ParentSnapshotRef.Name}, parentSnapshot); err == nil {
			if parentSnapshot.Status.RegistryStatus != nil && parentSnapshot.Status.RegistryStatus.Pushed {
				parentLayers = parentSnapshot.Status.RegistryStatus.LayerDigests
			}
			parentSnapshotPath = parentSnapshot.Status.SnapshotPath
		}
	}

	// Check for existing push request
	existingPushReqs := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, existingPushReqs, client.MatchingLabels{
		"snapshots.kloudlite.io/snapshot":  snapshot.Name,
		"snapshots.kloudlite.io/operation": "push",
	}); err == nil && len(existingPushReqs.Items) > 0 {
		// Check status of existing push request
		for _, req := range existingPushReqs.Items {
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Push failed: %s", req.Status.Message), logger)
			}
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseCompleted {
				// Push completed - update status
				now := metav1.Now()
				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
					snapshot.Status.State = snapshotv1.SnapshotStateReady
					snapshot.Status.Message = "Snapshot pushed successfully"
					snapshot.Status.RegistryStatus = &snapshotv1.SnapshotRegistryStatus{
						Pushed:         true,
						PushedAt:       &now,
						Tag:            tag,
						ImageRef:       fmt.Sprintf("image-registry:5000/%s:%s", repository, tag),
						Digest:         req.Status.Digest,
						LayerDigests:   req.Status.LayerDigests,
						LayerCount:     int32(len(req.Status.LayerDigests)),
						CompressedSize: req.Status.CompressedSize,
					}
					return nil
				}, logger); err != nil {
					logger.Error("Failed to update status after push", zap.Error(err))
					return reconcile.Result{}, err
				}
				// Delete completed push request
				if err := r.Delete(ctx, &req); err != nil && !apierrors.IsNotFound(err) {
					logger.Warn("Failed to delete completed push request", zap.Error(err))
				}
				logger.Info("Snapshot pushed to registry successfully",
					zap.String("imageRef", snapshot.Status.RegistryStatus.ImageRef))
				return reconcile.Result{}, nil
			}
			// Still in progress
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Create push SnapshotRequest
	pushReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-push", snapshot.Name),
			Namespace: snapshot.Status.WorkMachineName,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/operation": "push",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:          snapshotv1.SnapshotOperationPush,
			SnapshotPath:       snapshot.Status.SnapshotPath,
			ParentSnapshotPath: parentSnapshotPath,
			SnapshotRef:        snapshot.Name,
			RegistryRef: &snapshotv1.SnapshotRequestRegistryRef{
				RegistryURL:  "image-registry.kloudlite.svc.cluster.local:5000",
				Repository:   repository,
				Tag:          tag,
				ParentLayers: parentLayers,
			},
		},
	}

	// Add workspace/environment info
	if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
		pushReq.Spec.WorkspaceName = snapshot.Status.WorkspaceName
	} else if snapshot.Spec.EnvironmentRef != nil {
		pushReq.Spec.EnvironmentName = snapshot.Spec.EnvironmentRef.Name
	}

	if err := r.Create(ctx, pushReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create push SnapshotRequest", zap.Error(err))
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create push request: %v", err), logger)
		}
	}

	// Update status message
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.Message = "Pushing snapshot to registry..."
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update status message", zap.Error(err))
	}

	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// handlePulling handles the snapshot pull from registry process
func (r *SnapshotReconciler) handlePulling(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling snapshot pull from registry")

	if snapshot.Spec.RegistryRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Registry reference is required for pull operation", logger)
	}

	repository := snapshot.Spec.RegistryRef.Repository
	tag := snapshot.Spec.RegistryRef.Tag
	if tag == "" {
		tag = "latest"
	}

	// Check for existing pull request
	existingPullReqs := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, existingPullReqs, client.MatchingLabels{
		"snapshots.kloudlite.io/snapshot":  snapshot.Name,
		"snapshots.kloudlite.io/operation": "pull",
	}); err == nil && len(existingPullReqs.Items) > 0 {
		// Check status of existing pull request
		for _, req := range existingPullReqs.Items {
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Pull failed: %s", req.Status.Message), logger)
			}
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseCompleted {
				// Pull completed - update status
				now := metav1.Now()
				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
					snapshot.Status.State = snapshotv1.SnapshotStateReady
					snapshot.Status.Message = "Snapshot pulled successfully"
					snapshot.Status.CreatedAt = &now
					snapshot.Status.RegistryStatus = &snapshotv1.SnapshotRegistryStatus{
						Pushed:   true,
						PushedAt: &now,
						ImageRef: fmt.Sprintf("image-registry:5000/%s:%s", repository, tag),
						Digest:   req.Status.Digest,
					}
					return nil
				}, logger); err != nil {
					logger.Error("Failed to update status after pull", zap.Error(err))
					return reconcile.Result{}, err
				}
				// Delete completed pull request
				if err := r.Delete(ctx, &req); err != nil && !apierrors.IsNotFound(err) {
					logger.Warn("Failed to delete completed pull request", zap.Error(err))
				}
				logger.Info("Snapshot pulled from registry successfully")
				return reconcile.Result{}, nil
			}
			// Still in progress
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Determine target directory for pulled snapshots
	snapshotPath := snapshot.Status.SnapshotPath
	if snapshotPath == "" {
		snapshotPath = filepath.Join(snapshotsBasePath, snapshot.Name)
		// Store the path
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.SnapshotPath = snapshotPath
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to store snapshot path", zap.Error(err))
		}
	}

	// Create pull SnapshotRequest
	wmNamespace := snapshot.Status.WorkMachineName
	if wmNamespace == "" {
		wmNamespace = fmt.Sprintf("wm-%s", snapshot.Spec.OwnedBy)
	}

	pullReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-pull", snapshot.Name),
			Namespace: wmNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/operation": "pull",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:    snapshotv1.SnapshotOperationPull,
			SnapshotPath: snapshotPath,
			SnapshotRef:  snapshot.Name,
			RegistryRef: &snapshotv1.SnapshotRequestRegistryRef{
				RegistryURL: "image-registry.kloudlite.svc.cluster.local:5000",
				Repository:  repository,
				Tag:         tag,
			},
		},
	}

	if err := r.Create(ctx, pullReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create pull SnapshotRequest", zap.Error(err))
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create pull request: %v", err), logger)
		}
	}

	// Update status message
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.Message = "Pulling snapshot from registry..."
		snapshot.Status.WorkMachineName = wmNamespace
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update status message", zap.Error(err))
	}

	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// removeTagFromOtherSnapshots removes the tag from any other snapshot that has it
// This ensures tags are unique - when pushing with a tag, it's removed from other snapshots
func (r *SnapshotReconciler) removeTagFromOtherSnapshots(ctx context.Context, currentSnapshotName string, imageRef string, logger *zap.Logger) error {
	// List all snapshots
	allSnapshots := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, allSnapshots); err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	for _, s := range allSnapshots.Items {
		// Skip the current snapshot being pushed
		if s.Name == currentSnapshotName {
			continue
		}

		// Check if this snapshot has the same imageRef
		if s.Status.RegistryStatus != nil && s.Status.RegistryStatus.ImageRef == imageRef {
			logger.Info("Removing tag from snapshot that previously had it",
				zap.String("snapshot", s.Name),
				zap.String("imageRef", imageRef))

			// Clear the registry status for this snapshot since the tag now points elsewhere
			snapshotCopy := s.DeepCopy()
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshotCopy, func() error {
				snapshotCopy.Status.RegistryStatus = nil
				return nil
			}, logger); err != nil {
				logger.Warn("Failed to clear registry status from snapshot",
					zap.String("snapshot", s.Name),
					zap.Error(err))
			}
		}
	}

	return nil
}
