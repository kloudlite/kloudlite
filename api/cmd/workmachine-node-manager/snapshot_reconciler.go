package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	zap2 "go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// cachePathFromImageRef converts an OCI image reference to a local cache path.
// Uses a SHA256 hash of the imageRef to ensure unique, filesystem-safe cache keys.
// This ensures that snapshots are cached by their actual storage reference, not just by name.
func cachePathFromImageRef(imageRef string) string {
	hash := sha256.Sum256([]byte(imageRef))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash for brevity
	return fmt.Sprintf("%s/%s", snapshotStoragePath, hashStr)
}

// SnapshotRequestReconciler watches SnapshotRequest resources and processes them on this node
type SnapshotRequestReconciler struct {
	client.Client
	Logger           *zap2.Logger
	HostCmdExec      CommandExecutor // For btrfs commands that must run on host
	NodeName         string
	RegistryEndpoint string
	RegistryPrefix   string
	RegistryInsecure bool
}

const snapshotStoragePath = "/var/lib/kloudlite/storage/.snapshots"

func (r *SnapshotRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("snapshotRequest", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	// Fetch SnapshotRequest
	snapshotReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, req.NamespacedName, snapshotReq); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRequest", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Only process requests for this node
	if snapshotReq.Spec.NodeName != r.NodeName {
		return reconcile.Result{}, nil
	}

	// Handle completed or failed requests - no need to reprocess
	if snapshotReq.Status.State == snapshotv1.SnapshotRequestStateCompleted ||
		snapshotReq.Status.State == snapshotv1.SnapshotRequestStateFailed {
		return reconcile.Result{}, nil
	}

	logger.Info("Processing SnapshotRequest",
		zap2.String("state", string(snapshotReq.Status.State)),
		zap2.String("snapshotName", snapshotReq.Spec.SnapshotName))

	// Process based on current state
	switch snapshotReq.Status.State {
	case "", snapshotv1.SnapshotRequestStatePending:
		return r.handlePending(ctx, snapshotReq, logger)
	case snapshotv1.SnapshotRequestStateCreating:
		return r.handleCreating(ctx, snapshotReq, logger)
	case snapshotv1.SnapshotRequestStateUploading:
		return r.handleUploading(ctx, snapshotReq, logger)
	default:
		logger.Warn("Unknown snapshot request state", zap2.String("state", string(snapshotReq.Status.State)))
		return reconcile.Result{}, nil
	}
}

func (r *SnapshotRequestReconciler) handlePending(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Info("Starting snapshot request",
		zap2.String("sourcePath", req.Spec.SourcePath),
		zap2.String("owner", req.Spec.Owner))

	now := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateCreating
	req.Status.Message = "Creating btrfs snapshot"
	req.Status.StartedAt = &now
	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRequestReconciler) handleCreating(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	// Use a temporary path for creating the snapshot
	// Final path will be determined by content hash after creation
	tempSnapshotPath := fmt.Sprintf("%s/creating-%s-%s", snapshotStoragePath, req.Namespace, req.Spec.SnapshotName)

	// If we already have a content hash computed (from previous reconcile), use that
	if req.Status.LocalSnapshotPath != "" && strings.HasPrefix(req.Status.LocalSnapshotPath, snapshotStoragePath) {
		// Already created and moved to final path
		checkScript := fmt.Sprintf("test -d %s && echo exists", req.Status.LocalSnapshotPath)
		checkOutput, _ := r.HostCmdExec.Execute(checkScript)
		if strings.TrimSpace(string(checkOutput)) == "exists" {
			logger.Info("Snapshot already exists, transitioning to Uploading", zap2.String("path", req.Status.LocalSnapshotPath))
			req.Status.State = snapshotv1.SnapshotRequestStateUploading
			req.Status.Message = "Uploading to registry"
			if err := r.Status().Update(ctx, req); err != nil {
				if apierrors.IsConflict(err) {
					return reconcile.Result{Requeue: true}, nil
				}
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// Ensure snapshot storage directory exists
	mkdirScript := fmt.Sprintf("mkdir -p %s", snapshotStoragePath)
	if _, err := r.HostCmdExec.Execute(mkdirScript); err != nil {
		logger.Warn("Failed to create snapshot storage directory", zap2.Error(err))
	}

	// Ensure source path exists as a btrfs subvolume
	checkSourceScript := fmt.Sprintf("test -d %s && echo exists", req.Spec.SourcePath)
	sourceOutput, _ := r.HostCmdExec.Execute(checkSourceScript)
	if strings.TrimSpace(string(sourceOutput)) != "exists" {
		logger.Info("Source path doesn't exist, creating btrfs subvolume", zap2.String("path", req.Spec.SourcePath))
		parentDir := filepath.Dir(req.Spec.SourcePath)
		mkdirParentScript := fmt.Sprintf("mkdir -p %s", parentDir)
		if _, err := r.HostCmdExec.Execute(mkdirParentScript); err != nil {
			logger.Warn("Failed to create parent directory", zap2.Error(err))
		}
		createSubvolScript := fmt.Sprintf("btrfs subvolume create %s", req.Spec.SourcePath)
		if output, err := r.HostCmdExec.Execute(createSubvolScript); err != nil {
			logger.Error("Failed to create source btrfs subvolume",
				zap2.String("path", req.Spec.SourcePath),
				zap2.Error(err),
				zap2.String("output", string(output)))
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create source subvolume: %v - %s", err, string(output)), logger)
		}
		logger.Info("Created source btrfs subvolume", zap2.String("path", req.Spec.SourcePath))
	}

	// Clean up any existing temp snapshot
	cleanupScript := fmt.Sprintf("btrfs subvolume delete %s 2>/dev/null || rm -rf %s", tempSnapshotPath, tempSnapshotPath)
	r.HostCmdExec.Execute(cleanupScript)

	// Create btrfs snapshot at temporary path
	snapshotScript := fmt.Sprintf("btrfs subvolume snapshot -r %s %s", req.Spec.SourcePath, tempSnapshotPath)
	output, err := r.HostCmdExec.Execute(snapshotScript)
	if err != nil {
		logger.Error("Failed to create btrfs snapshot",
			zap2.String("sourcePath", req.Spec.SourcePath),
			zap2.String("tempPath", tempSnapshotPath),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to create btrfs snapshot: %v - %s", err, string(output)), logger)
	}

	// Compute MD5 hash of snapshot content
	// Using tar to create a consistent byte stream for hashing
	hashScript := fmt.Sprintf("tar -C %s -cf - . 2>/dev/null | md5sum | cut -d' ' -f1", tempSnapshotPath)
	hashOutput, err := r.HostCmdExec.Execute(hashScript)
	if err != nil {
		logger.Error("Failed to compute content hash", zap2.Error(err))
		// Clean up temp snapshot
		r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to compute content hash: %v", err), logger)
	}
	contentHash := strings.TrimSpace(string(hashOutput))
	if contentHash == "" {
		r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
		return r.setFailed(ctx, req, "Content hash is empty", logger)
	}

	logger.Info("Computed content hash", zap2.String("contentHash", contentHash))

	// Build imageRef using content hash as tag
	imageRef := fmt.Sprintf("%s/%s:%s", r.RegistryEndpoint, r.RegistryPrefix, contentHash)

	// Final cache path based on content hash
	finalSnapshotPath := cachePathFromImageRef(imageRef)

	// Check if this content already exists in cache (deduplication)
	checkFinalScript := fmt.Sprintf("btrfs subvolume show %s >/dev/null 2>&1 && echo 'exists'", finalSnapshotPath)
	finalOutput, _ := r.HostCmdExec.Execute(checkFinalScript)
	if strings.Contains(string(finalOutput), "exists") {
		logger.Info("Content already cached (deduplication)", zap2.String("contentHash", contentHash))
		// Clean up temp snapshot, use existing cache
		r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
	} else {
		// Move temp snapshot to final path
		// btrfs doesn't support rename, so we create a new snapshot and delete temp
		moveScript := fmt.Sprintf("btrfs subvolume snapshot %s %s && btrfs subvolume delete %s",
			tempSnapshotPath, finalSnapshotPath, tempSnapshotPath)
		if _, err := r.HostCmdExec.Execute(moveScript); err != nil {
			logger.Error("Failed to move snapshot to final path", zap2.Error(err))
			r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to move snapshot: %v", err), logger)
		}
	}

	// Update status with final path and imageRef
	req.Status.LocalSnapshotPath = finalSnapshotPath
	req.Status.State = snapshotv1.SnapshotRequestStateUploading
	req.Status.Message = fmt.Sprintf("Uploading to registry (hash: %s)", contentHash[:8])

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Created btrfs snapshot with content hash",
		zap2.String("path", finalSnapshotPath),
		zap2.String("contentHash", contentHash),
		zap2.String("imageRef", imageRef))
	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRequestReconciler) handleUploading(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	// First, check if the local snapshot still exists
	// If it was deleted, it means a previous reconcile already completed the upload
	if _, err := os.Stat(req.Status.LocalSnapshotPath); os.IsNotExist(err) {
		// Local snapshot was already deleted - check if Snapshot resource exists
		existingSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.SnapshotName, Namespace: req.Namespace}, existingSnapshot); err == nil {
			// Snapshot exists, this means the upload already completed successfully
			// Just update the SnapshotRequest status to Completed
			logger.Info("Local snapshot already deleted and Snapshot resource exists, marking as completed",
				zap2.String("snapshot", req.Spec.SnapshotName))
			completedNow := metav1.Now()
			req.Status.State = snapshotv1.SnapshotRequestStateCompleted
			req.Status.Message = "Snapshot created successfully"
			req.Status.CompletedAt = &completedNow
			req.Status.CreatedSnapshot = req.Spec.SnapshotName
			if err := r.Status().Update(ctx, req); err != nil {
				if apierrors.IsConflict(err) {
					return reconcile.Result{Requeue: true}, nil
				}
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		// No snapshot exists and local snapshot is gone - this is an error
		return r.setFailed(ctx, req, "Local snapshot was deleted but Snapshot resource not found", logger)
	}

	// Compute content hash from the snapshot to build imageRef
	hashScript := fmt.Sprintf("tar -C %s -cf - . 2>/dev/null | md5sum | cut -d' ' -f1", req.Status.LocalSnapshotPath)
	hashOutput, err := r.HostCmdExec.Execute(hashScript)
	if err != nil {
		logger.Error("Failed to compute content hash", zap2.Error(err))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to compute content hash: %v", err), logger)
	}
	contentHash := strings.TrimSpace(string(hashOutput))
	if contentHash == "" {
		return r.setFailed(ctx, req, "Content hash is empty", logger)
	}

	// Build image reference using content hash as tag (content-addressable storage)
	imageRef := fmt.Sprintf("%s/%s:%s", r.RegistryEndpoint, r.RegistryPrefix, contentHash)

	// Push snapshot to registry using embedded oras library
	logger.Info("Pushing snapshot to registry", zap2.String("imageRef", imageRef), zap2.String("contentHash", contentHash))
	if err := orasPushSnapshot(ctx, req.Status.LocalSnapshotPath, imageRef, r.RegistryInsecure); err != nil {
		logger.Error("Failed to push snapshot to registry",
			zap2.String("imageRef", imageRef),
			zap2.Error(err))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to push to registry: %v", err), logger)
	}

	// Get snapshot size
	var sizeBytes int64
	filepath.Walk(req.Status.LocalSnapshotPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			sizeBytes += info.Size()
		}
		return nil
	})

	// Build lineage and storageRefs from parent
	var lineage []string
	var storageRefs []string
	if req.Spec.ParentSnapshot != "" {
		parentSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.ParentSnapshot, Namespace: req.Namespace}, parentSnapshot); err == nil {
			lineage = append(parentSnapshot.Status.Lineage, parentSnapshot.Name)
			// Inherit parent's storage refs and add this snapshot's imageRef
			storageRefs = append(parentSnapshot.Status.StorageRefs, imageRef)
		} else {
			// Parent not found, just use this snapshot's imageRef
			storageRefs = []string{imageRef}
		}
	} else {
		// No parent, this is a root snapshot
		storageRefs = []string{imageRef}
	}

	// Create the namespaced Snapshot resource
	now := metav1.Now()
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.SnapshotName,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
		Spec: snapshotv1.SnapshotSpec{
			Owner:           req.Spec.Owner,
			ParentSnapshot:  req.Spec.ParentSnapshot,
			Description:     req.Spec.Description,
			Artifacts:       req.Spec.Artifacts,
			RetentionPolicy: req.Spec.RetentionPolicy,
		},
		Status: snapshotv1.SnapshotStatus{
			State:       snapshotv1.SnapshotStateReady,
			Message:     "Snapshot ready",
			SizeBytes:   sizeBytes,
			SizeHuman:   formatSnapshotSize(sizeBytes),
			CreatedAt:   &now,
			Lineage:     lineage,
			StorageRefs: storageRefs,
			// ReferencedBy will be populated by environment controller when environments fork from this snapshot
			Registry: &snapshotv1.SnapshotRegistryInfo{
				ImageRef: imageRef,
				PushedAt: &now,
			},
		},
	}

	snapshotStatus := snapshot.Status // Save status before create (Create ignores status subresource)

	if err := r.Create(ctx, snapshot); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create Snapshot: %v", err), logger)
		}
		logger.Info("Snapshot already exists, updating status", zap2.String("name", req.Spec.SnapshotName))
	}

	// Update status separately (Create doesn't set status subresource)
	// Retry Get with backoff since the resource might not be immediately available after Create
	existing := &snapshotv1.Snapshot{}
	var getErr error
	for i := range 5 {
		if getErr = r.Get(ctx, client.ObjectKey{Name: req.Spec.SnapshotName, Namespace: req.Namespace}, existing); getErr == nil {
			break
		}
		if !apierrors.IsNotFound(getErr) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to get Snapshot for status update: %v", getErr), logger)
		}
		// Wait before retry (100ms, 200ms, 400ms, 800ms, 1600ms)
		time.Sleep(time.Duration(100<<i) * time.Millisecond)
	}
	if getErr != nil {
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to get Snapshot for status update after retries: %v", getErr), logger)
	}

	existing.Status = snapshotStatus
	if err := r.Status().Update(ctx, existing); err != nil {
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to update Snapshot status: %v", err), logger)
	}

	// Delete local snapshot to free space (btrfs operation runs on host)
	deleteScript := fmt.Sprintf("btrfs subvolume delete %s", req.Status.LocalSnapshotPath)
	if _, err := r.HostCmdExec.Execute(deleteScript); err != nil {
		logger.Warn("Failed to delete local snapshot", zap2.Error(err))
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
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot created successfully",
		zap2.String("snapshot", req.Spec.SnapshotName),
		zap2.String("imageRef", imageRef))

	return reconcile.Result{}, nil
}

func (r *SnapshotRequestReconciler) setFailed(ctx context.Context, req *snapshotv1.SnapshotRequest, message string, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot request failed", zap2.String("message", message))

	now := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateFailed
	req.Status.Message = message
	req.Status.CompletedAt = &now

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SnapshotRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRequest{}).
		Complete(r)
}
