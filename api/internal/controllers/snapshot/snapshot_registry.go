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
// Key behaviors:
// 1. Each snapshot is pushed independently (no waiting for parent to be pushed first)
// 2. Snapshot name is ALWAYS the primary tag (predictable for parent references)
// 3. Uses incremental btrfs send if parent exists locally (smaller data)
// 4. Sets ParentImageRef label for pull to resolve parent chain
// 5. User-provided tag is added as additional tag after primary push
func (r *SnapshotReconciler) handlePushing(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling snapshot push to registry")

	if snapshot.Spec.RegistryRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Registry reference is required for push operation", logger)
	}

	// Determine the repository
	repository := snapshot.Spec.RegistryRef.Repository
	if repository == "" {
		if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
			repository = fmt.Sprintf("snapshots/%s/ws/%s", snapshot.Spec.OwnedBy, snapshot.Status.WorkspaceName)
		} else if snapshot.Spec.EnvironmentRef != nil {
			repository = fmt.Sprintf("snapshots/%s/env/%s", snapshot.Spec.OwnedBy, snapshot.Spec.EnvironmentRef.Name)
		} else {
			repository = fmt.Sprintf("snapshots/%s", snapshot.Spec.OwnedBy)
		}
	}

	// Primary tag is ALWAYS the snapshot name (predictable for parent references)
	primaryTag := snapshot.Name
	userTag := snapshot.Spec.RegistryRef.Tag // Optional user-provided tag

	// Check for existing push request
	existingPushReqs := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, existingPushReqs, client.MatchingLabels{
		"snapshots.kloudlite.io/snapshot":  snapshot.Name,
		"snapshots.kloudlite.io/operation": "push",
	}); err == nil && len(existingPushReqs.Items) > 0 {
		for _, req := range existingPushReqs.Items {
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Push failed: %s", req.Status.Message), logger)
			}
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseCompleted {
				// Push completed - update status
				now := metav1.Now()
				tags := []string{primaryTag}

				// Add user tag if different from primary
				if userTag != "" && userTag != primaryTag {
					if err := r.addAdditionalTag(ctx, snapshot, repository, primaryTag, userTag, logger); err != nil {
						logger.Warn("Failed to add user tag, continuing with primary tag only",
							zap.String("userTag", userTag), zap.Error(err))
					} else {
						tags = append(tags, userTag)
					}
				}

				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
					snapshot.Status.State = snapshotv1.SnapshotStateReady
					snapshot.Status.Message = "Snapshot pushed successfully"
					snapshot.Status.RegistryStatus = &snapshotv1.SnapshotRegistryStatus{
						Pushed:         true,
						PushedAt:       &now,
						Tag:            primaryTag,
						Tags:           tags,
						ImageRef:       fmt.Sprintf("image-registry:5000/%s:%s", repository, primaryTag),
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
					zap.String("imageRef", snapshot.Status.RegistryStatus.ImageRef),
					zap.Strings("tags", tags))
				return reconcile.Result{}, nil
			}
			// Still in progress
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Build parent references for incremental btrfs send and lineage tracking
	// Each snapshot is pushed independently (no waiting for parent), but we still use:
	// - ParentSnapshotPath: for incremental btrfs send (smaller data if parent exists locally)
	// - ParentImageRef: for pull to resolve parent chain (predictable: repo:parentName)
	var parentSnapshotPath string
	var parentImageRef string

	if snapshot.Spec.ParentSnapshotRef != nil && snapshot.Spec.ParentSnapshotRef.Name != "" {
		parentName := snapshot.Spec.ParentSnapshotRef.Name

		// Look up parent snapshot to get its local path (for incremental btrfs send)
		parentSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: parentName}, parentSnapshot); err == nil {
			if parentSnapshot.Status.SnapshotPath != "" {
				parentSnapshotPath = parentSnapshot.Status.SnapshotPath
				logger.Info("Using parent snapshot for incremental send",
					zap.String("parent", parentName),
					zap.String("parentPath", parentSnapshotPath))
			}
		} else {
			logger.Warn("Parent snapshot not found locally, will do full send",
				zap.String("parent", parentName), zap.Error(err))
		}

		// Build predictable parent image ref (same repo, parent name as tag)
		// This allows pull to resolve the parent chain even if parent isn't pushed yet
		parentImageRef = fmt.Sprintf("image-registry.kloudlite.svc.cluster.local:5000/%s:%s", repository, parentName)
		logger.Info("Setting parent image reference for lineage tracking",
			zap.String("parentImageRef", parentImageRef))
	}

	// Create push SnapshotRequest with primary tag (snapshot name)
	// Each snapshot is pushed independently - no waiting for parent push
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
			ParentSnapshotPath: parentSnapshotPath, // For incremental btrfs send
			SnapshotRef:        snapshot.Name,
			RegistryRef: &snapshotv1.SnapshotRequestRegistryRef{
				RegistryURL:    "image-registry.kloudlite.svc.cluster.local:5000",
				Repository:     repository,
				Tag:            primaryTag,      // Always use snapshot name
				ParentImageRef: parentImageRef, // For pull to resolve parent chain
			},
			Metadata: snapshot.Status.CollectedMetadata,
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

// addAdditionalTag creates an additional tag for a pushed snapshot
func (r *SnapshotReconciler) addAdditionalTag(ctx context.Context, snapshot *snapshotv1.Snapshot,
	repository, sourceTag, targetTag string, logger *zap.Logger) error {

	tagReqName := fmt.Sprintf("%s-tag-%s", snapshot.Name, targetTag)
	wmNamespace := snapshot.Status.WorkMachineName

	tagReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tagReqName,
			Namespace: wmNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/operation": "tag",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:    snapshotv1.SnapshotOperationTag,
			SnapshotPath: snapshot.Status.SnapshotPath, // Required field
			RegistryRef: &snapshotv1.SnapshotRequestRegistryRef{
				RegistryURL: "image-registry.kloudlite.svc.cluster.local:5000",
				Repository:  repository,
				Tag:         targetTag,
				SourceTag:   sourceTag,
			},
		},
	}

	if err := r.Create(ctx, tagReq); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil // Already exists, that's fine
		}
		return fmt.Errorf("failed to create tag request: %w", err)
	}

	logger.Info("Created tag request for additional tag",
		zap.String("sourceTag", sourceTag),
		zap.String("targetTag", targetTag))
	return nil
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
				// Create Snapshot CRs for parent chain if they don't exist
				if len(req.Status.PulledSnapshots) > 1 {
					if err := r.createParentSnapshotCRs(ctx, snapshot, &req, logger); err != nil {
						logger.Warn("Failed to create parent snapshot CRs", zap.Error(err))
					}
				}

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

// createParentSnapshotCRs creates Snapshot CRs for the parent chain snapshots
// that were pulled along with the main snapshot
func (r *SnapshotReconciler) createParentSnapshotCRs(
	ctx context.Context,
	mainSnapshot *snapshotv1.Snapshot,
	req *snapshotv1.SnapshotRequest,
	logger *zap.Logger,
) error {
	for _, pulledSnap := range req.Status.PulledSnapshots {
		// Skip the main snapshot itself
		if pulledSnap.Name == mainSnapshot.Name {
			continue
		}

		// Check if snapshot already exists
		existing := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: pulledSnap.Name}, existing); err == nil {
			logger.Debug("Parent snapshot CR already exists", zap.String("name", pulledSnap.Name))
			continue
		}

		// Create Snapshot CR for this parent
		logger.Info("Creating Snapshot CR for parent snapshot", zap.String("name", pulledSnap.Name))

		parentSnapshot := &snapshotv1.Snapshot{
			ObjectMeta: metav1.ObjectMeta{
				Name: pulledSnap.Name,
				Labels: map[string]string{
					"snapshots.kloudlite.io/pulled-from-parent-chain": "true",
				},
			},
			Spec: snapshotv1.SnapshotSpec{
				OwnedBy:         pulledSnap.OwnedBy,
				IncludeMetadata: true,
			},
		}

		// Set parent reference if exists
		if pulledSnap.ParentSnapshotName != "" {
			parentSnapshot.Spec.ParentSnapshotRef = &snapshotv1.ParentSnapshotReference{
				Name: pulledSnap.ParentSnapshotName,
			}
		}

		// Set environment or workspace ref based on type
		if pulledSnap.SnapshotType == string(snapshotv1.SnapshotTypeEnvironment) {
			parentSnapshot.Spec.EnvironmentRef = &snapshotv1.EnvironmentReference{
				Name: pulledSnap.EnvironmentName,
			}
		} else if pulledSnap.SnapshotType == string(snapshotv1.SnapshotTypeWorkspace) {
			parentSnapshot.Spec.WorkspaceRef = &snapshotv1.WorkspaceReference{
				Name:            pulledSnap.WorkspaceName,
				WorkmachineName: pulledSnap.WorkMachineName,
			}
		}

		// Create the snapshot CR
		if err := r.Create(ctx, parentSnapshot); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create parent snapshot CR",
					zap.String("name", pulledSnap.Name),
					zap.Error(err))
				continue
			}
		}

		// Update status to Ready since the snapshot was already pulled
		now := metav1.Now()
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, parentSnapshot, func() error {
			parentSnapshot.Status.State = snapshotv1.SnapshotStateReady
			parentSnapshot.Status.Message = "Pulled as part of parent chain"
			parentSnapshot.Status.SnapshotType = snapshotv1.SnapshotType(pulledSnap.SnapshotType)
			parentSnapshot.Status.TargetName = pulledSnap.TargetName
			parentSnapshot.Status.SnapshotPath = pulledSnap.Path
			parentSnapshot.Status.CreatedAt = &now
			parentSnapshot.Status.WorkMachineName = mainSnapshot.Status.WorkMachineName
			parentSnapshot.Status.WorkspaceName = pulledSnap.WorkspaceName

			// Mark as pushed since we pulled it from registry
			parentSnapshot.Status.RegistryStatus = &snapshotv1.SnapshotRegistryStatus{
				Pushed:   true,
				PushedAt: &now,
			}

			// Set collected metadata for environment snapshots
			if pulledSnap.Resources != nil {
				parentSnapshot.Status.CollectedMetadata = pulledSnap.Resources
			}
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update parent snapshot status",
				zap.String("name", pulledSnap.Name),
				zap.Error(err))
		}

		logger.Info("Created parent snapshot CR", zap.String("name", pulledSnap.Name))
	}

	return nil
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
