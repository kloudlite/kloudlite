package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/pkg/oci"
	zap2 "go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotRequestFinalizer = "snapshots.kloudlite.io/request-finalizer"
)

// SnapshotRequestReconciler handles btrfs snapshot operations on the node
type SnapshotRequestReconciler struct {
	client.Client
	Logger  *zap2.Logger
	CmdExec CommandExecutor
}

func (r *SnapshotRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("snapshotRequest", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling SnapshotRequest")

	// Fetch SnapshotRequest
	snapshotReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, req.NamespacedName, snapshotReq); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("SnapshotRequest not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRequest", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Check if being deleted
	if snapshotReq.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, snapshotReq, logger)
	}

	// Add finalizer if not present
	if !containsString(snapshotReq.Finalizers, snapshotRequestFinalizer) {
		snapshotReq.Finalizers = append(snapshotReq.Finalizers, snapshotRequestFinalizer)
		if err := r.Update(ctx, snapshotReq); err != nil {
			logger.Error("Failed to add finalizer", zap2.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Skip if already completed or failed
	if snapshotReq.Status.Phase == snapshotv1.SnapshotRequestPhaseCompleted ||
		snapshotReq.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
		return reconcile.Result{}, nil
	}

	// Set status to InProgress
	if snapshotReq.Status.Phase != snapshotv1.SnapshotRequestPhaseInProgress {
		now := metav1.Now()
		if err := r.updateStatus(ctx, req.NamespacedName, func(latest *snapshotv1.SnapshotRequest) {
			latest.Status.Phase = snapshotv1.SnapshotRequestPhaseInProgress
			latest.Status.Message = fmt.Sprintf("Starting %s operation", snapshotReq.Spec.Operation)
			latest.Status.StartedAt = &now
		}, logger); err != nil {
			logger.Warn("Failed to update status to InProgress", zap2.Error(err))
		}
	}

	// Execute the operation
	var opErr error
	var pushResult *oci.PushResult
	var pullResult *pullSnapshotResult
	switch snapshotReq.Spec.Operation {
	case snapshotv1.SnapshotOperationCreate:
		opErr = r.createSnapshot(snapshotReq, logger)
	case snapshotv1.SnapshotOperationDelete:
		opErr = r.deleteSnapshot(snapshotReq, logger)
	case snapshotv1.SnapshotOperationRestore:
		opErr = r.restoreSnapshot(snapshotReq, logger)
	case snapshotv1.SnapshotOperationPush:
		pushResult, opErr = r.pushSnapshot(snapshotReq, logger)
	case snapshotv1.SnapshotOperationPull:
		pullResult, opErr = r.pullSnapshot(snapshotReq, logger)
	case snapshotv1.SnapshotOperationTag:
		opErr = r.tagSnapshot(snapshotReq, logger)
	default:
		opErr = fmt.Errorf("unknown operation: %s", snapshotReq.Spec.Operation)
	}

	// Update status based on result
	now := metav1.Now()
	if opErr != nil {
		logger.Error("Snapshot operation failed",
			zap2.String("operation", string(snapshotReq.Spec.Operation)),
			zap2.Error(opErr))

		if err := r.updateStatus(ctx, req.NamespacedName, func(latest *snapshotv1.SnapshotRequest) {
			latest.Status.Phase = snapshotv1.SnapshotRequestPhaseFailed
			latest.Status.Message = opErr.Error()
			latest.Status.FinishedAt = &now
		}, logger); err != nil {
			logger.Error("Failed to update status to Failed", zap2.Error(err))
		}
		return reconcile.Result{}, nil
	}

	// Get snapshot size for create operations
	var sizeBytes int64
	if snapshotReq.Spec.Operation == snapshotv1.SnapshotOperationCreate {
		sizeBytes = r.getSnapshotSize(snapshotReq.Spec.SnapshotPath, logger)
	}

	// Success
	if err := r.updateStatus(ctx, req.NamespacedName, func(latest *snapshotv1.SnapshotRequest) {
		latest.Status.Phase = snapshotv1.SnapshotRequestPhaseCompleted
		latest.Status.Message = fmt.Sprintf("%s operation completed successfully", snapshotReq.Spec.Operation)
		latest.Status.FinishedAt = &now
		latest.Status.SizeBytes = sizeBytes

		// Add push-specific status
		if pushResult != nil {
			latest.Status.Digest = pushResult.Digest
			latest.Status.LayerDigests = pushResult.LayerDigests
			latest.Status.CompressedSize = pushResult.CompressedSize
		}

		// Add pull-specific status (K8s resource metadata and snapshot chain)
		if pullResult != nil {
			latest.Status.PulledMetadata = pullResult.metadata
			latest.Status.PulledSnapshots = pullResult.snapshots
		}
	}, logger); err != nil {
		logger.Error("Failed to update status to Completed", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot operation completed successfully",
		zap2.String("operation", string(snapshotReq.Spec.Operation)),
		zap2.String("path", snapshotReq.Spec.SnapshotPath))

	return reconcile.Result{}, nil
}

// createSnapshot creates a btrfs snapshot
func (r *SnapshotRequestReconciler) createSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) error {
	sourcePath := req.Spec.SourcePath
	snapshotPath := req.Spec.SnapshotPath

	// Ensure parent directory exists
	parentDir := filepath.Dir(snapshotPath)
	mkdirScript := fmt.Sprintf("mkdir -p %s", parentDir)
	if output, err := r.CmdExec.Execute(mkdirScript); err != nil {
		return fmt.Errorf("failed to create parent directory: %s - %w", string(output), err)
	}

	// Note: K8s resource metadata is stored in the OCI layer's metadata.json during push,
	// not as files in the snapshot directory. This allows read-only btrfs snapshots.

	// Check if source is a btrfs subvolume - this is required for proper snapshots
	checkScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", sourcePath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		// Source might not exist or not be a subvolume
		// Try to check if it's a regular directory
		checkDirScript := fmt.Sprintf("test -d %s", sourcePath)
		if _, dirErr := r.CmdExec.Execute(checkDirScript); dirErr != nil {
			// Source doesn't exist - this is an error, not a fallback
			return fmt.Errorf("source path does not exist: %s", sourcePath)
		}

		// Source exists but is not a subvolume - this is an error
		// Snapshots require btrfs subvolumes to support push/pull operations
		logger.Error("Source is not a btrfs subvolume, cannot create snapshot",
			zap2.String("source", sourcePath))
		return fmt.Errorf("source path is not a btrfs subvolume: %s (snapshots require btrfs subvolumes for push/pull support)", sourcePath)
	}

	// Sync filesystem before snapshot to ensure all writes are flushed
	// This is critical for database consistency - checkpoint writes may be in kernel buffers
	logger.Info("Syncing filesystem before snapshot")
	if output, err := r.CmdExec.Execute("sync"); err != nil {
		logger.Warn("Failed to sync filesystem", zap2.Error(err), zap2.String("output", string(output)))
	}

	// Create btrfs snapshot
	var snapshotScript string
	if req.Spec.ReadOnly {
		snapshotScript = fmt.Sprintf("btrfs subvolume snapshot -r %s %s", sourcePath, snapshotPath)
	} else {
		snapshotScript = fmt.Sprintf("btrfs subvolume snapshot %s %s", sourcePath, snapshotPath)
	}

	logger.Info("Creating btrfs snapshot",
		zap2.String("source", sourcePath),
		zap2.String("destination", snapshotPath),
		zap2.Bool("readOnly", req.Spec.ReadOnly))

	output, err := r.CmdExec.Execute(snapshotScript)
	if err != nil {
		return fmt.Errorf("btrfs snapshot failed: %s - %w", string(output), err)
	}

	logger.Info("Btrfs snapshot created successfully")
	return nil
}

// deleteSnapshot deletes a btrfs snapshot
func (r *SnapshotRequestReconciler) deleteSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) error {
	snapshotPath := req.Spec.SnapshotPath

	// Check if path exists
	checkScript := fmt.Sprintf("test -e %s", snapshotPath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		logger.Info("Snapshot path does not exist, nothing to delete",
			zap2.String("path", snapshotPath))
		return nil
	}

	// Check if it's a btrfs subvolume
	checkSubvolScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", snapshotPath)
	if _, err := r.CmdExec.Execute(checkSubvolScript); err != nil {
		// Not a subvolume, use rm -rf
		logger.Info("Path is not a btrfs subvolume, using rm -rf",
			zap2.String("path", snapshotPath))
		rmScript := fmt.Sprintf("rm -rf %s", snapshotPath)
		if output, rmErr := r.CmdExec.Execute(rmScript); rmErr != nil {
			return fmt.Errorf("rm -rf failed: %s - %w", string(output), rmErr)
		}
		return nil
	}

	// Delete btrfs subvolume
	deleteScript := fmt.Sprintf("btrfs subvolume delete %s", snapshotPath)
	logger.Info("Deleting btrfs subvolume", zap2.String("path", snapshotPath))

	output, err := r.CmdExec.Execute(deleteScript)
	if err != nil {
		return fmt.Errorf("btrfs subvolume delete failed: %s - %w", string(output), err)
	}

	logger.Info("Btrfs subvolume deleted successfully")
	return nil
}

// restoreSnapshot restores from a btrfs snapshot
func (r *SnapshotRequestReconciler) restoreSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) error {
	// For restore: SourcePath is the snapshot source, SnapshotPath is the restore target
	sourcePath := req.Spec.SourcePath   // The snapshot to restore FROM
	targetPath := req.Spec.SnapshotPath // Where to restore TO

	// Check if source snapshot exists
	checkScript := fmt.Sprintf("test -e %s", sourcePath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		return fmt.Errorf("source snapshot does not exist: %s", sourcePath)
	}

	// Check if target is a btrfs subvolume - if so, delete it first
	checkSubvolScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", targetPath)
	if _, err := r.CmdExec.Execute(checkSubvolScript); err == nil {
		// Target is a subvolume, delete it
		logger.Info("Deleting existing target subvolume", zap2.String("path", targetPath))
		deleteScript := fmt.Sprintf("btrfs subvolume delete %s", targetPath)
		if output, delErr := r.CmdExec.Execute(deleteScript); delErr != nil {
			return fmt.Errorf("failed to delete existing subvolume: %s - %w", string(output), delErr)
		}
	} else {
		// Not a subvolume, try rm -rf
		rmScript := fmt.Sprintf("rm -rf %s", targetPath)
		if output, rmErr := r.CmdExec.Execute(rmScript); rmErr != nil {
			logger.Warn("Failed to remove existing target path", zap2.String("output", string(output)))
		}
	}

	// Check if source snapshot is a btrfs subvolume
	checkSnapshotSubvolScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", sourcePath)
	if _, err := r.CmdExec.Execute(checkSnapshotSubvolScript); err != nil {
		// Snapshot is not a subvolume, use rsync to restore
		logger.Info("Snapshot is not a btrfs subvolume, using rsync",
			zap2.String("source", sourcePath))

		// Create target directory
		mkdirScript := fmt.Sprintf("mkdir -p %s", targetPath)
		if output, mkdirErr := r.CmdExec.Execute(mkdirScript); mkdirErr != nil {
			return fmt.Errorf("failed to create target directory: %s - %w", string(output), mkdirErr)
		}

		// Rsync restore
		rsyncScript := fmt.Sprintf("rsync -a --delete %s/ %s/", sourcePath, targetPath)
		if output, rsyncErr := r.CmdExec.Execute(rsyncScript); rsyncErr != nil {
			return fmt.Errorf("rsync restore failed: %s - %w", string(output), rsyncErr)
		}

		return nil
	}

	// Create a writable snapshot from the read-only snapshot
	restoreScript := fmt.Sprintf("btrfs subvolume snapshot %s %s", sourcePath, targetPath)
	logger.Info("Restoring btrfs snapshot",
		zap2.String("source", sourcePath),
		zap2.String("destination", targetPath))

	output, err := r.CmdExec.Execute(restoreScript)
	if err != nil {
		return fmt.Errorf("btrfs snapshot restore failed: %s - %w", string(output), err)
	}

	logger.Info("Btrfs snapshot restored successfully")
	return nil
}

// pushSnapshot pushes a snapshot to the OCI registry
func (r *SnapshotRequestReconciler) pushSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (*oci.PushResult, error) {
	if req.Spec.RegistryRef == nil {
		return nil, fmt.Errorf("registryRef is required for push operation")
	}

	snapshotPath := req.Spec.SnapshotPath
	parentSnapshotPath := req.Spec.ParentSnapshotPath

	// Check if snapshot exists
	checkScript := fmt.Sprintf("test -e %s", snapshotPath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		return nil, fmt.Errorf("snapshot path does not exist: %s", snapshotPath)
	}

	// Build metadata from the SnapshotRequest
	metadata := &oci.SnapshotMetadata{
		Name: req.Spec.SnapshotRef,
		Spec: oci.SnapshotMetadataSpec{
			OwnedBy: "", // Will be populated by the controller
		},
		Status: oci.SnapshotMetadataStatus{
			WorkMachineName: "", // Will be populated by the controller
		},
	}

	// If we have environment or workspace info, add it
	if req.Spec.EnvironmentName != "" {
		metadata.Spec.EnvironmentRef = &oci.EnvironmentReference{
			Name: req.Spec.EnvironmentName,
		}
		metadata.Status.SnapshotType = "Environment"
	}
	if req.Spec.WorkspaceName != "" {
		metadata.Spec.WorkspaceRef = &oci.WorkspaceReference{
			Name: req.Spec.WorkspaceName,
		}
		metadata.Status.SnapshotType = "Workspace"
	}

	// Include K8s resource metadata in the OCI metadata (stored in metadata.json)
	// This avoids having to write files to the read-only btrfs snapshot
	if req.Spec.Metadata != nil {
		metadata.Resources = &oci.ResourceMetadata{
			ConfigMaps:   req.Spec.Metadata.ConfigMaps,
			Secrets:      req.Spec.Metadata.Secrets,
			Deployments:  req.Spec.Metadata.Deployments,
			Services:     req.Spec.Metadata.Services,
			StatefulSets: req.Spec.Metadata.StatefulSets,
			Compositions: req.Spec.Metadata.Compositions,
		}
		logger.Info("Including K8s resource metadata in OCI layer")
	}

	logger.Info("Pushing snapshot to registry",
		zap2.String("snapshotPath", snapshotPath),
		zap2.String("parentPath", parentSnapshotPath),
		zap2.String("registry", req.Spec.RegistryRef.RegistryURL),
		zap2.String("repository", req.Spec.RegistryRef.Repository),
		zap2.String("tag", req.Spec.RegistryRef.Tag),
	)

	// Create OCI client and push
	client := oci.NewClient(true) // Use insecure for internal registry
	result, err := client.Push(oci.PushOptions{
		RegistryURL:        req.Spec.RegistryRef.RegistryURL,
		Repository:         req.Spec.RegistryRef.Repository,
		Tag:                req.Spec.RegistryRef.Tag,
		SnapshotPath:       snapshotPath,
		ParentSnapshotPath: parentSnapshotPath,
		Metadata:           metadata,
		ParentImageRef:     req.Spec.RegistryRef.ParentImageRef,
		Insecure:           true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to push snapshot: %w", err)
	}

	logger.Info("Snapshot pushed successfully",
		zap2.String("imageRef", result.ImageRef),
		zap2.String("digest", result.Digest),
		zap2.Int("layerCount", len(result.LayerDigests)),
		zap2.Int64("compressedSize", result.CompressedSize),
	)

	return result, nil
}

// pullSnapshot pulls a snapshot chain from the OCI registry
// Returns the K8s resource metadata from the pulled OCI layer
// pullSnapshotResult contains the result of a pull operation
type pullSnapshotResult struct {
	metadata  *snapshotv1.SnapshotMetadata
	snapshots []snapshotv1.PulledSnapshotInfo
}

func (r *SnapshotRequestReconciler) pullSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (*pullSnapshotResult, error) {
	if req.Spec.RegistryRef == nil {
		return nil, fmt.Errorf("registryRef is required for pull operation")
	}

	targetDir := req.Spec.SnapshotPath

	// Ensure target directory exists
	mkdirScript := fmt.Sprintf("mkdir -p %s", targetDir)
	if output, err := r.CmdExec.Execute(mkdirScript); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %s - %w", string(output), err)
	}

	logger.Info("Pulling snapshot from registry",
		zap2.String("registry", req.Spec.RegistryRef.RegistryURL),
		zap2.String("repository", req.Spec.RegistryRef.Repository),
		zap2.String("tag", req.Spec.RegistryRef.Tag),
		zap2.String("targetDir", targetDir),
	)

	// Create OCI client and pull
	client := oci.NewClient(true) // Use insecure for internal registry
	result, err := client.Pull(oci.PullOptions{
		RegistryURL: req.Spec.RegistryRef.RegistryURL,
		Repository:  req.Spec.RegistryRef.Repository,
		Tag:         req.Spec.RegistryRef.Tag,
		TargetDir:   targetDir,
		Insecure:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pull snapshot: %w", err)
	}

	logger.Info("Snapshot chain pulled successfully",
		zap2.Int("snapshotCount", len(result.Snapshots)),
	)

	// Build result with K8s metadata and snapshot chain info
	pullResult := &pullSnapshotResult{
		snapshots: make([]snapshotv1.PulledSnapshotInfo, 0, len(result.Snapshots)),
	}

	for _, snap := range result.Snapshots {
		path := result.SnapshotPaths[snap.Name]
		logger.Info("Pulled snapshot",
			zap2.String("name", snap.Name),
			zap2.String("path", path),
			zap2.String("type", snap.Status.SnapshotType),
		)

		// Build snapshot info for parent chain tracking
		snapInfo := snapshotv1.PulledSnapshotInfo{
			Name:            snap.Name,
			Path:            path,
			SnapshotType:    snap.Status.SnapshotType,
			TargetName:      snap.Status.TargetName,
			WorkspaceName:   snap.Status.WorkspaceName,
			WorkMachineName: snap.Status.WorkMachineName,
			OwnedBy:         snap.Spec.OwnedBy,
		}

		// Add parent reference if exists
		if snap.Spec.ParentSnapshotRef != nil {
			snapInfo.ParentSnapshotName = snap.Spec.ParentSnapshotRef.Name
		}

		// Add environment reference if exists
		if snap.Spec.EnvironmentRef != nil {
			snapInfo.EnvironmentName = snap.Spec.EnvironmentRef.Name
		}

		// Extract K8s resource metadata from the OCI metadata
		if snap.Resources != nil {
			snapInfo.Resources = &snapshotv1.SnapshotMetadata{
				ConfigMaps:   snap.Resources.ConfigMaps,
				Secrets:      snap.Resources.Secrets,
				Deployments:  snap.Resources.Deployments,
				Services:     snap.Resources.Services,
				StatefulSets: snap.Resources.StatefulSets,
				Compositions: snap.Resources.Compositions,
			}
			// Use the first snapshot with resources as the main metadata
			if pullResult.metadata == nil {
				pullResult.metadata = snapInfo.Resources
			}
			logger.Info("Found K8s resource metadata in OCI layer", zap2.String("snapshot", snap.Name))
		}

		pullResult.snapshots = append(pullResult.snapshots, snapInfo)
	}

	return pullResult, nil
}

// tagSnapshot creates an additional tag for an existing image in the registry
func (r *SnapshotRequestReconciler) tagSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) error {
	if req.Spec.RegistryRef == nil {
		return fmt.Errorf("registryRef is required for tag operation")
	}

	if req.Spec.RegistryRef.SourceTag == "" {
		return fmt.Errorf("sourceTag is required for tag operation")
	}

	sourceRef := fmt.Sprintf("%s/%s:%s",
		req.Spec.RegistryRef.RegistryURL,
		req.Spec.RegistryRef.Repository,
		req.Spec.RegistryRef.SourceTag)

	targetRef := fmt.Sprintf("%s/%s:%s",
		req.Spec.RegistryRef.RegistryURL,
		req.Spec.RegistryRef.Repository,
		req.Spec.RegistryRef.Tag)

	logger.Info("Creating additional tag for image",
		zap2.String("sourceRef", sourceRef),
		zap2.String("targetRef", targetRef),
	)

	// Create OCI client and tag
	client := oci.NewClient(true) // Use insecure for internal registry
	if err := client.Tag(sourceRef, targetRef); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	logger.Info("Image tagged successfully",
		zap2.String("newTag", req.Spec.RegistryRef.Tag),
	)

	return nil
}

// getSnapshotSize returns the size of the snapshot in bytes
func (r *SnapshotRequestReconciler) getSnapshotSize(path string, logger *zap2.Logger) int64 {
	// Use du to get directory size
	sizeScript := fmt.Sprintf("du -sb %s 2>/dev/null | cut -f1", path)
	output, err := r.CmdExec.Execute(sizeScript)
	if err != nil {
		logger.Warn("Failed to get snapshot size", zap2.Error(err))
		return 0
	}

	sizeStr := strings.TrimSpace(string(output))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		logger.Warn("Failed to parse snapshot size", zap2.String("output", sizeStr))
		return 0
	}

	return size
}

// handleDeletion handles SnapshotRequest deletion
func (r *SnapshotRequestReconciler) handleDeletion(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	if !containsString(req.Finalizers, snapshotRequestFinalizer) {
		return reconcile.Result{}, nil
	}

	// Remove finalizer - the actual snapshot data is managed by the Snapshot controller
	req.Finalizers = removeString(req.Finalizers, snapshotRequestFinalizer)
	if err := r.Update(ctx, req); err != nil {
		logger.Error("Failed to remove finalizer", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("SnapshotRequest finalizer removed")
	return reconcile.Result{}, nil
}

// updateStatus updates the SnapshotRequest status with retry
func (r *SnapshotRequestReconciler) updateStatus(
	ctx context.Context,
	namespacedName client.ObjectKey,
	updateFn func(*snapshotv1.SnapshotRequest),
	logger *zap2.Logger,
) error {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		// Fetch the latest version
		latest := &snapshotv1.SnapshotRequest{}
		if err := r.Get(ctx, namespacedName, latest); err != nil {
			return fmt.Errorf("failed to fetch latest SnapshotRequest: %w", err)
		}

		// Apply the update function
		updateFn(latest)

		// Try to update status
		if err := r.Status().Update(ctx, latest); err != nil {
			if apierrors.IsConflict(err) && i < maxRetries-1 {
				logger.Info("Status update conflict, retrying",
					zap2.Int("attempt", i+1))
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("failed to update status: %w", err)
		}

		return nil
	}

	return fmt.Errorf("failed to update status after %d retries", maxRetries)
}

func (r *SnapshotRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRequest{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldReq, okOld := e.ObjectOld.(*snapshotv1.SnapshotRequest)
				newReq, okNew := e.ObjectNew.(*snapshotv1.SnapshotRequest)
				if !okOld || !okNew {
					return false
				}
				// Reconcile on spec changes or deletion
				return oldReq.Generation != newReq.Generation || newReq.DeletionTimestamp != nil
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
		}).
		Complete(r)
}
