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
		opErr = r.pullSnapshot(snapshotReq, logger)
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

	// Check if source is a btrfs subvolume
	checkScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", sourcePath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		// Source might not exist or not be a subvolume
		// Try to check if it's a regular directory
		checkDirScript := fmt.Sprintf("test -d %s", sourcePath)
		if _, dirErr := r.CmdExec.Execute(checkDirScript); dirErr != nil {
			logger.Warn("Source path does not exist, creating empty snapshot directory",
				zap2.String("source", sourcePath))
			// Create empty directory as placeholder
			mkdirScript := fmt.Sprintf("mkdir -p %s", snapshotPath)
			if output, err := r.CmdExec.Execute(mkdirScript); err != nil {
				return fmt.Errorf("failed to create placeholder directory: %s - %w", string(output), err)
			}
			return nil
		}

		// Source exists but is not a subvolume - create a regular copy
		logger.Info("Source is not a btrfs subvolume, using rsync copy",
			zap2.String("source", sourcePath))
		copyScript := fmt.Sprintf("rsync -a --delete %s/ %s/", sourcePath, snapshotPath)
		if output, err := r.CmdExec.Execute(copyScript); err != nil {
			return fmt.Errorf("rsync failed: %s - %w", string(output), err)
		}
		return nil
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

	// Write metadata files if provided
	if req.Spec.Metadata != nil {
		if err := r.writeMetadataFiles(snapshotPath, req.Spec.Metadata, logger); err != nil {
			logger.Warn("Failed to write metadata files", zap2.Error(err))
			// Continue with push even if metadata writing fails
		}
	}

	// Build metadata from the SnapshotRequest
	// In a real implementation, we'd fetch the actual Snapshot resource
	// For now, we use basic metadata from the request
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
		ParentLayers:       req.Spec.RegistryRef.ParentLayers,
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
func (r *SnapshotRequestReconciler) pullSnapshot(req *snapshotv1.SnapshotRequest, logger *zap2.Logger) error {
	if req.Spec.RegistryRef == nil {
		return fmt.Errorf("registryRef is required for pull operation")
	}

	targetDir := req.Spec.SnapshotPath

	// Ensure target directory exists
	mkdirScript := fmt.Sprintf("mkdir -p %s", targetDir)
	if output, err := r.CmdExec.Execute(mkdirScript); err != nil {
		return fmt.Errorf("failed to create target directory: %s - %w", string(output), err)
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
		return fmt.Errorf("failed to pull snapshot: %w", err)
	}

	logger.Info("Snapshot chain pulled successfully",
		zap2.Int("snapshotCount", len(result.Snapshots)),
	)

	// Log each pulled snapshot
	for _, snap := range result.Snapshots {
		path := result.SnapshotPaths[snap.Name]
		logger.Info("Pulled snapshot",
			zap2.String("name", snap.Name),
			zap2.String("path", path),
			zap2.String("type", snap.Status.SnapshotType),
		)
	}

	return nil
}

// writeMetadataFiles writes K8s resource metadata JSON files to the snapshot directory
func (r *SnapshotRequestReconciler) writeMetadataFiles(snapshotPath string, metadata *snapshotv1.SnapshotMetadata, logger *zap2.Logger) error {
	metadataDir := filepath.Join(snapshotPath, "metadata")

	// Create metadata directory
	mkdirScript := fmt.Sprintf("mkdir -p %s", metadataDir)
	if output, err := r.CmdExec.Execute(mkdirScript); err != nil {
		return fmt.Errorf("failed to create metadata directory: %s - %w", string(output), err)
	}

	logger.Info("Writing metadata files", zap2.String("dir", metadataDir))

	// Write each metadata file using cat with heredoc
	writeFile := func(filename, content string) error {
		if content == "" {
			return nil
		}
		filePath := filepath.Join(metadataDir, filename)
		// Use printf to handle special characters properly
		// First write to a temp file, then move it
		script := fmt.Sprintf("cat > %s << 'METADATA_EOF'\n%s\nMETADATA_EOF", filePath, content)
		if output, err := r.CmdExec.Execute(script); err != nil {
			return fmt.Errorf("failed to write %s: %s - %w", filename, string(output), err)
		}
		logger.Info("Wrote metadata file", zap2.String("file", filename))
		return nil
	}

	// Write all metadata files
	if err := writeFile("configmaps.json", metadata.ConfigMaps); err != nil {
		return err
	}
	if err := writeFile("secrets.json", metadata.Secrets); err != nil {
		return err
	}
	if err := writeFile("deployments.json", metadata.Deployments); err != nil {
		return err
	}
	if err := writeFile("services.json", metadata.Services); err != nil {
		return err
	}
	if err := writeFile("statefulsets.json", metadata.StatefulSets); err != nil {
		return err
	}
	if err := writeFile("compositions.json", metadata.Compositions); err != nil {
		return err
	}

	logger.Info("Metadata files written successfully")
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
