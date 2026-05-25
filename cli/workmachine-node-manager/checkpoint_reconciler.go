package workmachinenodemanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	zap2 "go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const checkpointStoragePath = "/var/lib/kloudlite/storage/.snapshots"

// cachePathFromImageRef converts an OCI image reference to a local cache path.
// Uses a SHA256 hash of the imageRef to ensure unique, filesystem-safe cache keys.
func cachePathFromImageRef(imageRef string) string {
	hash := sha256.Sum256([]byte(imageRef))
	hashStr := hex.EncodeToString(hash[:])[:16]
	return fmt.Sprintf("%s/%s", checkpointStoragePath, hashStr)
}

// CheckpointReconciler watches Checkpoint resources and processes them on this node.
// It handles the full lifecycle: btrfs snapshot creation, ORAS push, and status updates.
type CheckpointReconciler struct {
	client.Client
	Logger           *zap2.Logger
	HostCmdExec      CommandExecutor
	NodeName         string
	RegistryEndpoint string
	RegistryPrefix   string
	RegistryInsecure bool
}

func (r *CheckpointReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("checkpoint", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	checkpoint := &checkpointv1.Checkpoint{}
	if err := r.Get(ctx, req.NamespacedName, checkpoint); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Checkpoint", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Only process checkpoints for this node
	if checkpoint.Spec.NodeName != r.NodeName {
		return reconcile.Result{}, nil
	}

	// Skip terminal states
	if checkpoint.Status.Phase == checkpointv1.CheckpointPhaseReady ||
		checkpoint.Status.Phase == checkpointv1.CheckpointPhaseFailed ||
		checkpoint.Status.Phase == checkpointv1.CheckpointPhaseDeleting {
		return reconcile.Result{}, nil
	}

	logger.Info("Processing Checkpoint", zap2.String("phase", string(checkpoint.Status.Phase)))

	switch checkpoint.Status.Phase {
	case "", checkpointv1.CheckpointPhasePending:
		return r.handlePending(ctx, checkpoint, logger)
	case checkpointv1.CheckpointPhaseCreating:
		return r.handleCreating(ctx, checkpoint, logger)
	case checkpointv1.CheckpointPhaseUploading:
		return r.handleUploading(ctx, checkpoint, logger)
	default:
		logger.Warn("Unknown checkpoint phase", zap2.String("phase", string(checkpoint.Status.Phase)))
		return reconcile.Result{}, nil
	}
}

func (r *CheckpointReconciler) handlePending(ctx context.Context, cp *checkpointv1.Checkpoint, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Info("Starting checkpoint", zap2.String("sourcePath", cp.Spec.SourcePath))

	now := metav1.Now()
	cp.Status.Phase = checkpointv1.CheckpointPhaseCreating
	cp.Status.Message = "Creating btrfs snapshot"
	cp.Status.StartedAt = &now
	if err := r.Status().Update(ctx, cp); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *CheckpointReconciler) handleCreating(ctx context.Context, cp *checkpointv1.Checkpoint, logger *zap2.Logger) (reconcile.Result, error) {
	tempSnapshotPath := fmt.Sprintf("%s/creating-%s-%s", checkpointStoragePath, cp.Namespace, cp.Name)

	// If we already have a localPath set and it exists, skip to uploading
	if cp.Status.LocalPath != "" && strings.HasPrefix(cp.Status.LocalPath, checkpointStoragePath) {
		checkScript := fmt.Sprintf("test -d %s && echo exists", cp.Status.LocalPath)
		checkOutput, _ := r.HostCmdExec.Execute(checkScript)
		if strings.TrimSpace(string(checkOutput)) == "exists" {
			logger.Info("Snapshot already exists, transitioning to Uploading", zap2.String("path", cp.Status.LocalPath))
			cp.Status.Phase = checkpointv1.CheckpointPhaseUploading
			cp.Status.Message = "Uploading to registry"
			if err := r.Status().Update(ctx, cp); err != nil {
				if apierrors.IsConflict(err) {
					return reconcile.Result{Requeue: true}, nil
				}
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// Ensure snapshot storage directory exists
	mkdirScript := fmt.Sprintf("mkdir -p %s", checkpointStoragePath)
	if _, err := r.HostCmdExec.Execute(mkdirScript); err != nil {
		logger.Warn("Failed to create checkpoint storage directory", zap2.Error(err))
	}

	// Ensure source path exists as a btrfs subvolume
	checkSourceScript := fmt.Sprintf("test -d %s && echo exists", cp.Spec.SourcePath)
	sourceOutput, _ := r.HostCmdExec.Execute(checkSourceScript)
	if strings.TrimSpace(string(sourceOutput)) != "exists" {
		logger.Info("Source path doesn't exist, creating btrfs subvolume", zap2.String("path", cp.Spec.SourcePath))
		parentDir := filepath.Dir(cp.Spec.SourcePath)
		mkdirParentScript := fmt.Sprintf("mkdir -p %s", parentDir)
		if _, err := r.HostCmdExec.Execute(mkdirParentScript); err != nil {
			logger.Warn("Failed to create parent directory", zap2.Error(err))
		}
		createSubvolScript := fmt.Sprintf("btrfs subvolume create %s", cp.Spec.SourcePath)
		if output, err := r.HostCmdExec.Execute(createSubvolScript); err != nil {
			return r.setFailed(ctx, cp, fmt.Sprintf("Failed to create source subvolume: %v - %s", err, string(output)), logger)
		}
	}

	// Clean up any existing temp snapshot
	cleanupScript := fmt.Sprintf("btrfs subvolume delete %s 2>/dev/null || rm -rf %s", tempSnapshotPath, tempSnapshotPath)
	r.HostCmdExec.Execute(cleanupScript)

	// Create btrfs snapshot
	snapshotScript := fmt.Sprintf("btrfs subvolume snapshot -r %s %s", cp.Spec.SourcePath, tempSnapshotPath)
	output, err := r.HostCmdExec.Execute(snapshotScript)
	if err != nil {
		return r.setFailed(ctx, cp, fmt.Sprintf("Failed to create btrfs snapshot: %v - %s", err, string(output)), logger)
	}

	// Compute MD5 content hash
	hashScript := fmt.Sprintf("tar -C %s -cf - . 2>/dev/null | md5sum | cut -d' ' -f1", tempSnapshotPath)
	hashOutput, err := r.HostCmdExec.Execute(hashScript)
	if err != nil {
		r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
		return r.setFailed(ctx, cp, fmt.Sprintf("Failed to compute content hash: %v", err), logger)
	}
	contentHash := strings.TrimSpace(string(hashOutput))
	if contentHash == "" {
		r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
		return r.setFailed(ctx, cp, "Content hash is empty", logger)
	}

	logger.Info("Computed content hash", zap2.String("contentHash", contentHash))

	imageRef := fmt.Sprintf("%s/%s:%s", r.RegistryEndpoint, r.RegistryPrefix, contentHash)
	finalSnapshotPath := cachePathFromImageRef(imageRef)

	// Check if content already exists in cache (deduplication)
	checkFinalScript := fmt.Sprintf("btrfs subvolume show %s >/dev/null 2>&1 && echo 'exists'", finalSnapshotPath)
	finalOutput, _ := r.HostCmdExec.Execute(checkFinalScript)
	if strings.Contains(string(finalOutput), "exists") {
		logger.Info("Content already cached (deduplication)", zap2.String("contentHash", contentHash))
		r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
	} else {
		// Move temp snapshot to final path
		moveScript := fmt.Sprintf("btrfs subvolume snapshot %s %s && btrfs subvolume delete %s",
			tempSnapshotPath, finalSnapshotPath, tempSnapshotPath)
		if _, err := r.HostCmdExec.Execute(moveScript); err != nil {
			r.HostCmdExec.Execute(fmt.Sprintf("btrfs subvolume delete %s", tempSnapshotPath))
			return r.setFailed(ctx, cp, fmt.Sprintf("Failed to move snapshot: %v", err), logger)
		}
	}

	// Store content hash and local path in status, transition to Uploading
	cp.Status.LocalPath = finalSnapshotPath
	cp.Status.ContentHash = contentHash
	cp.Status.Phase = checkpointv1.CheckpointPhaseUploading
	cp.Status.Message = fmt.Sprintf("Uploading to registry (hash: %s)", contentHash[:8])

	if err := r.Status().Update(ctx, cp); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	logger.Info("Created btrfs snapshot",
		zap2.String("path", finalSnapshotPath),
		zap2.String("contentHash", contentHash))
	return reconcile.Result{Requeue: true}, nil
}

func (r *CheckpointReconciler) handleUploading(ctx context.Context, cp *checkpointv1.Checkpoint, logger *zap2.Logger) (reconcile.Result, error) {
	// Verify local snapshot still exists
	if _, err := os.Stat(cp.Status.LocalPath); os.IsNotExist(err) {
		// If layer info is already set, this was already uploaded
		if cp.Status.Layer != nil && cp.Status.Layer.ImageRef != "" {
			logger.Info("Already uploaded, marking as Ready")
			now := metav1.Now()
			cp.Status.Phase = checkpointv1.CheckpointPhaseReady
			cp.Status.Message = "Checkpoint ready"
			cp.Status.CompletedAt = &now
			if err := r.Status().Update(ctx, cp); err != nil {
				if apierrors.IsConflict(err) {
					return reconcile.Result{Requeue: true}, nil
				}
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		return r.setFailed(ctx, cp, "Local snapshot was deleted before upload completed", logger)
	}

	// Reuse content hash from Creating phase (stored in status)
	contentHash := cp.Status.ContentHash
	if contentHash == "" {
		// Fallback: recompute if missing (shouldn't happen normally)
		hashScript := fmt.Sprintf("tar -C %s -cf - . 2>/dev/null | md5sum | cut -d' ' -f1", cp.Status.LocalPath)
		hashOutput, err := r.HostCmdExec.Execute(hashScript)
		if err != nil {
			return r.setFailed(ctx, cp, fmt.Sprintf("Failed to compute content hash: %v", err), logger)
		}
		contentHash = strings.TrimSpace(string(hashOutput))
		if contentHash == "" {
			return r.setFailed(ctx, cp, "Content hash is empty", logger)
		}
	}

	imageRef := fmt.Sprintf("%s/%s:%s", r.RegistryEndpoint, r.RegistryPrefix, contentHash)

	// Push to registry
	logger.Info("Pushing checkpoint to registry", zap2.String("imageRef", imageRef))
	if err := orasPushSnapshot(ctx, cp.Status.LocalPath, imageRef, r.RegistryInsecure); err != nil {
		return r.setFailed(ctx, cp, fmt.Sprintf("Failed to push to registry: %v", err), logger)
	}

	// Get snapshot size
	var sizeBytes int64
	filepath.Walk(cp.Status.LocalPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			sizeBytes += info.Size()
		}
		return nil
	})

	// Update status to Ready with layer info
	now := metav1.Now()
	cp.Status.Phase = checkpointv1.CheckpointPhaseReady
	cp.Status.Message = "Checkpoint ready"
	cp.Status.CompletedAt = &now
	cp.Status.SizeBytes = sizeBytes
	cp.Status.SizeHuman = formatSnapshotSize(sizeBytes)
	cp.Status.Layer = &checkpointv1.LayerInfo{
		ImageRef: imageRef,
		PushedAt: &now,
	}

	if err := r.Status().Update(ctx, cp); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	// Delete local snapshot to free space
	deleteScript := fmt.Sprintf("btrfs subvolume delete %s", cp.Status.LocalPath)
	if _, err := r.HostCmdExec.Execute(deleteScript); err != nil {
		logger.Warn("Failed to delete local snapshot", zap2.Error(err))
	}

	logger.Info("Checkpoint created successfully",
		zap2.String("imageRef", imageRef),
		zap2.Int64("sizeBytes", sizeBytes))
	return reconcile.Result{}, nil
}

func (r *CheckpointReconciler) setFailed(ctx context.Context, cp *checkpointv1.Checkpoint, message string, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Error("Checkpoint failed", zap2.String("message", message))

	now := metav1.Now()
	cp.Status.Phase = checkpointv1.CheckpointPhaseFailed
	cp.Status.Message = message
	cp.Status.CompletedAt = &now

	if err := r.Status().Update(ctx, cp); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *CheckpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&checkpointv1.Checkpoint{}).
		Complete(r)
}
