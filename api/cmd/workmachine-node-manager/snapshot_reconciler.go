package main

import (
	"context"
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

// SnapshotRequestReconciler watches SnapshotRequest resources and processes them on this node
type SnapshotRequestReconciler struct {
	client.Client
	Logger      *zap2.Logger
	HostCmdExec CommandExecutor // For btrfs commands that must run on host
	NodeName    string
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
	// Generate local snapshot path
	snapshotPath := fmt.Sprintf("%s/%s", snapshotStoragePath, req.Spec.SnapshotName)

	// Check if snapshot already exists (race condition protection)
	checkScript := fmt.Sprintf("test -d %s && echo exists", snapshotPath)
	checkOutput, _ := r.HostCmdExec.Execute(checkScript)
	if strings.TrimSpace(string(checkOutput)) == "exists" {
		logger.Info("Snapshot already exists, transitioning to Uploading", zap2.String("path", snapshotPath))
		req.Status.LocalSnapshotPath = snapshotPath
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

	// Ensure snapshot storage directory exists
	mkdirScript := fmt.Sprintf("mkdir -p %s", snapshotStoragePath)
	if _, err := r.HostCmdExec.Execute(mkdirScript); err != nil {
		logger.Warn("Failed to create snapshot storage directory", zap2.Error(err))
	}

	// Ensure source path exists as a btrfs subvolume
	// This handles the case where environment storage hasn't been created yet (no PVCs)
	checkSourceScript := fmt.Sprintf("test -d %s && echo exists", req.Spec.SourcePath)
	sourceOutput, _ := r.HostCmdExec.Execute(checkSourceScript)
	if strings.TrimSpace(string(sourceOutput)) != "exists" {
		logger.Info("Source path doesn't exist, creating btrfs subvolume", zap2.String("path", req.Spec.SourcePath))
		// Ensure parent directory exists
		parentDir := filepath.Dir(req.Spec.SourcePath)
		mkdirParentScript := fmt.Sprintf("mkdir -p %s", parentDir)
		if _, err := r.HostCmdExec.Execute(mkdirParentScript); err != nil {
			logger.Warn("Failed to create parent directory", zap2.Error(err))
		}
		// Create btrfs subvolume for the source path
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

	// Create btrfs snapshot
	snapshotScript := fmt.Sprintf("btrfs subvolume snapshot -r %s %s", req.Spec.SourcePath, snapshotPath)
	output, err := r.HostCmdExec.Execute(snapshotScript)
	if err != nil {
		logger.Error("Failed to create btrfs snapshot",
			zap2.String("sourcePath", req.Spec.SourcePath),
			zap2.String("snapshotPath", snapshotPath),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to create btrfs snapshot: %v - %s", err, string(output)), logger)
	}

	// Update status
	req.Status.LocalSnapshotPath = snapshotPath
	req.Status.State = snapshotv1.SnapshotRequestStateUploading
	req.Status.Message = "Uploading to registry"

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Created btrfs snapshot", zap2.String("path", snapshotPath))
	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRequestReconciler) handleUploading(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	// First, check if the local snapshot still exists
	// If it was deleted, it means a previous reconcile already completed the upload
	if _, err := os.Stat(req.Status.LocalSnapshotPath); os.IsNotExist(err) {
		// Local snapshot was already deleted - check if Snapshot resource exists
		existingSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.SnapshotName}, existingSnapshot); err == nil {
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

	// Get the SnapshotStore
	store := &snapshotv1.SnapshotStore{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.Store}, store); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("SnapshotStore %q not found", req.Spec.Store), logger)
		}
		logger.Error("Failed to get SnapshotStore", zap2.Error(err))
		return reconcile.Result{}, err
	}

	if !store.Status.Ready {
		logger.Info("SnapshotStore not ready, waiting", zap2.String("store", store.Name))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Build image reference for the registry
	imageRef := fmt.Sprintf("%s/%s/%s:%s",
		store.Spec.Registry.Endpoint,
		store.Spec.Registry.RepositoryPrefix,
		req.Spec.Owner,
		req.Spec.SnapshotName)

	// Push snapshot to registry using embedded oras library
	logger.Info("Pushing snapshot to registry", zap2.String("imageRef", imageRef))
	if err := orasPushSnapshot(ctx, req.Status.LocalSnapshotPath, imageRef, true /* plainHTTP */); err != nil {
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

	// Build lineage from parent
	var lineage []string
	if req.Spec.ParentSnapshot != "" {
		parentSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.ParentSnapshot}, parentSnapshot); err == nil {
			lineage = append(parentSnapshot.Status.Lineage, parentSnapshot.Name)
		}
	}

	// Create the global Snapshot resource
	now := metav1.Now()
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Spec.SnapshotName,
			Labels: req.Labels,
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
			SizeBytes: sizeBytes,
			SizeHuman: formatSnapshotSize(sizeBytes),
			CreatedAt: &now,
			Lineage:   lineage,
			RefCount:  1, // Start with refCount=1 so it's not garbage collected immediately
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
		if getErr = r.Get(ctx, client.ObjectKey{Name: req.Spec.SnapshotName}, existing); getErr == nil {
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
