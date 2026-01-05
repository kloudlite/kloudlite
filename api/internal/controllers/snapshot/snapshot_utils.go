package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleDeleting handles the snapshot deletion
func (r *SnapshotReconciler) handleDeleting(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	// List and wait for all SnapshotRequests to be deleted
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{"snapshots.kloudlite.io/snapshot": snapshot.Name}); err != nil {
		logger.Error("Failed to list SnapshotRequests", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if len(snapshotReqList.Items) > 0 {
		logger.Info("Waiting for SnapshotRequests to be deleted", zap.Int("count", len(snapshotReqList.Items)))
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

// handleDeletion cleans up snapshot resources
func (r *SnapshotReconciler) handleDeletion(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
		return reconcile.Result{}, nil
	}

	// Set state to Deleting
	if snapshot.Status.State != snapshotv1.SnapshotStateDeleting {
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.State = snapshotv1.SnapshotStateDeleting
			snapshot.Status.Message = "Deleting snapshot data"
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status to Deleting", zap.Error(err))
		}
	}

	// Handle deletion based on snapshot type
	if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
		// Delete workspace snapshot
		if snapshot.Status.SnapshotPath != "" {
			workspaceSnapshotPath := filepath.Join(snapshot.Status.SnapshotPath, "home")
			deleteReq := &snapshotv1.SnapshotRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-delete-home", snapshot.Name),
					Namespace: snapshot.Status.WorkMachineName,
					Labels: map[string]string{
						"snapshots.kloudlite.io/snapshot":  snapshot.Name,
						"snapshots.kloudlite.io/operation": "delete",
					},
				},
				Spec: snapshotv1.SnapshotRequestSpec{
					Operation:     snapshotv1.SnapshotOperationDelete,
					SnapshotPath:  workspaceSnapshotPath,
					SnapshotRef:   snapshot.Name,
					WorkspaceName: snapshot.Status.WorkspaceName,
				},
			}

			if err := r.Create(ctx, deleteReq); err != nil {
				if !apierrors.IsAlreadyExists(err) {
					logger.Warn("Failed to create delete SnapshotRequest for workspace", zap.Error(err))
				}
			}
		}
	} else {
		// Delete environment snapshot - single delete request for entire environment directory
		if snapshot.Status.SnapshotPath != "" {
			deleteReq := &snapshotv1.SnapshotRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-delete-env", snapshot.Name),
					Namespace: snapshot.Status.WorkMachineName,
					Labels: map[string]string{
						"snapshots.kloudlite.io/snapshot":  snapshot.Name,
						"snapshots.kloudlite.io/operation": "delete",
					},
				},
				Spec: snapshotv1.SnapshotRequestSpec{
					Operation:    snapshotv1.SnapshotOperationDelete,
					SnapshotPath: snapshot.Status.SnapshotPath,
					SnapshotRef:  snapshot.Name,
				},
			}

			if err := r.Create(ctx, deleteReq); err != nil {
				if !apierrors.IsAlreadyExists(err) {
					logger.Warn("Failed to create delete SnapshotRequest", zap.Error(err))
				}
			}
		}
	}

	// Check if all delete requests are complete
	allComplete, err := r.checkDeleteRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check delete request status", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !allComplete {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Re-link children to this snapshot's parent before deletion
	if err := r.relinkChildSnapshots(ctx, snapshot, logger); err != nil {
		logger.Error("Failed to re-link child snapshots", zap.Error(err))
		// Continue with deletion even if re-linking fails
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(snapshot, snapshotFinalizer)
	if err := r.Update(ctx, snapshot); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot cleanup complete")
	return reconcile.Result{}, nil
}

// relinkChildSnapshots updates all child snapshots to point to this snapshot's parent
func (r *SnapshotReconciler) relinkChildSnapshots(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) error {
	// Find all snapshots that have this snapshot as their parent
	childSnapshots := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, childSnapshots, client.MatchingLabels{
		"snapshots.kloudlite.io/parent": snapshot.Name,
	}); err != nil {
		return fmt.Errorf("failed to list child snapshots: %w", err)
	}

	if len(childSnapshots.Items) == 0 {
		return nil
	}

	logger.Info("Re-linking child snapshots", zap.Int("count", len(childSnapshots.Items)))

	// Get the parent of the snapshot being deleted
	var newParent *snapshotv1.ParentSnapshotReference
	newParentLabel := ""
	if snapshot.Spec.ParentSnapshotRef != nil {
		newParent = snapshot.Spec.ParentSnapshotRef.DeepCopy()
		newParentLabel = snapshot.Spec.ParentSnapshotRef.Name
	}

	// Update each child to point to the new parent
	for i := range childSnapshots.Items {
		child := &childSnapshots.Items[i]
		logger.Info("Re-linking child snapshot",
			zap.String("child", child.Name),
			zap.String("oldParent", snapshot.Name),
			zap.String("newParent", newParentLabel),
		)

		// Update spec
		child.Spec.ParentSnapshotRef = newParent

		// Update label
		if child.Labels == nil {
			child.Labels = make(map[string]string)
		}
		if newParentLabel != "" {
			child.Labels["snapshots.kloudlite.io/parent"] = newParentLabel
		} else {
			delete(child.Labels, "snapshots.kloudlite.io/parent")
		}

		if err := r.Update(ctx, child); err != nil {
			logger.Warn("Failed to re-link child snapshot",
				zap.String("child", child.Name),
				zap.Error(err),
			)
			// Continue with other children
		}
	}

	return nil
}

// checkSnapshotRequestsComplete checks if all SnapshotRequests for a Snapshot are complete
// expectedCount is the minimum number of SnapshotRequests expected (to handle cache sync issues)
func (r *SnapshotReconciler) checkSnapshotRequestsComplete(ctx context.Context, snapshot *snapshotv1.Snapshot, expectedCount int, logger *zap.Logger) (bool, error) {
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{"snapshots.kloudlite.io/snapshot": snapshot.Name}); err != nil {
		return false, err
	}

	// If we haven't seen all expected SnapshotRequests yet, wait for cache to sync
	if len(snapshotReqList.Items) < expectedCount {
		logger.Info("Waiting for SnapshotRequests to appear in cache",
			zap.Int("found", len(snapshotReqList.Items)),
			zap.Int("expected", expectedCount))
		return false, nil
	}

	for _, req := range snapshotReqList.Items {
		if req.Status.Phase != snapshotv1.SnapshotRequestPhaseCompleted {
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
				// If any request failed, fail the whole snapshot
				logger.Error("SnapshotRequest failed", zap.String("request", req.Name), zap.String("message", req.Status.Message))
				return false, fmt.Errorf("SnapshotRequest %s failed: %s", req.Name, req.Status.Message)
			}
			return false, nil
		}
	}

	return true, nil
}

// checkDeleteRequestsComplete checks if all delete SnapshotRequests are complete
func (r *SnapshotReconciler) checkDeleteRequestsComplete(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (bool, error) {
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{
		"snapshots.kloudlite.io/snapshot":  snapshot.Name,
		"snapshots.kloudlite.io/operation": "delete",
	}); err != nil {
		return false, err
	}

	for _, req := range snapshotReqList.Items {
		if req.Status.Phase != snapshotv1.SnapshotRequestPhaseCompleted {
			return false, nil
		}
	}

	return true, nil
}

// ExportedMetadata contains both the resource info counts and the JSON data
type ExportedMetadata struct {
	Info     *snapshotv1.ResourceMetadataInfo
	Metadata *snapshotv1.SnapshotMetadata
}

// exportMetadata collects K8s resources as JSON strings to be passed to SnapshotRequest
func (r *SnapshotReconciler) exportMetadata(ctx context.Context, namespace string, logger *zap.Logger) (*ExportedMetadata, error) {
	info := &snapshotv1.ResourceMetadataInfo{}
	metadata := &snapshotv1.SnapshotMetadata{}

	// Export ConfigMaps
	configMaps := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMaps, client.InNamespace(namespace)); err == nil {
		info.ConfigMaps = int32(len(configMaps.Items))
		if data, err := json.Marshal(configMaps); err == nil {
			metadata.ConfigMaps = string(data)
		} else {
			logger.Warn("Failed to marshal ConfigMaps", zap.Error(err))
		}
	}

	// Export Secrets (excluding service account tokens)
	secrets := &corev1.SecretList{}
	if err := r.List(ctx, secrets, client.InNamespace(namespace)); err == nil {
		// Filter out service account tokens
		filtered := []corev1.Secret{}
		for _, s := range secrets.Items {
			if s.Type != corev1.SecretTypeServiceAccountToken {
				filtered = append(filtered, s)
			}
		}
		info.Secrets = int32(len(filtered))
		if data, err := json.Marshal(filtered); err == nil {
			metadata.Secrets = string(data)
		} else {
			logger.Warn("Failed to marshal Secrets", zap.Error(err))
		}
	}

	// Export Deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err == nil {
		info.Deployments = int32(len(deployments.Items))
		if data, err := json.Marshal(deployments); err == nil {
			metadata.Deployments = string(data)
		} else {
			logger.Warn("Failed to marshal Deployments", zap.Error(err))
		}
	}

	// Export Services
	services := &corev1.ServiceList{}
	if err := r.List(ctx, services, client.InNamespace(namespace)); err == nil {
		info.Services = int32(len(services.Items))
		if data, err := json.Marshal(services); err == nil {
			metadata.Services = string(data)
		} else {
			logger.Warn("Failed to marshal Services", zap.Error(err))
		}
	}

	// Export StatefulSets
	statefulSets := &appsv1.StatefulSetList{}
	if err := r.List(ctx, statefulSets, client.InNamespace(namespace)); err == nil {
		info.StatefulSets = int32(len(statefulSets.Items))
		if data, err := json.Marshal(statefulSets); err == nil {
			metadata.StatefulSets = string(data)
		} else {
			logger.Warn("Failed to marshal StatefulSets", zap.Error(err))
		}
	}

	// Export Compositions
	compositions := &environmentsv1.CompositionList{}
	if err := r.List(ctx, compositions, client.InNamespace(namespace)); err == nil {
		info.Compositions = int32(len(compositions.Items))
		if data, err := json.Marshal(compositions); err == nil {
			metadata.Compositions = string(data)
		} else {
			logger.Warn("Failed to marshal Compositions", zap.Error(err))
		}
	}

	logger.Info("Collected metadata",
		zap.Int32("configMaps", info.ConfigMaps),
		zap.Int32("secrets", info.Secrets),
		zap.Int32("deployments", info.Deployments),
		zap.Int32("services", info.Services),
		zap.Int32("statefulSets", info.StatefulSets),
		zap.Int32("compositions", info.Compositions))

	return &ExportedMetadata{Info: info, Metadata: metadata}, nil
}

// createDir creates a directory with proper permissions
func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// exportToJSON writes an object to a JSON file
func exportToJSON(path string, obj interface{}) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// updateStatusFailed sets the snapshot status to Failed
func (r *SnapshotReconciler) updateStatusFailed(ctx context.Context, snapshot *snapshotv1.Snapshot, message string, logger *zap.Logger) (reconcile.Result, error) {
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateFailed
		snapshot.Status.Message = message
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Failed", zap.Error(err))
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
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// updateEnvironmentLastRestored updates the environment's LastRestoredSnapshot status
// This is used to track parent-child lineage when creating new snapshots
func (r *SnapshotReconciler) updateEnvironmentLastRestored(ctx context.Context, envName, snapshotName string, logger *zap.Logger) error {
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	now := metav1.Now()
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, env, func() error {
		env.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
			Name:       snapshotName,
			RestoredAt: now,
		}
		return nil
	}, logger)
}

// updateWorkspaceLastRestored updates the workspace's LastRestoredSnapshot status
// This is used to track parent-child lineage when creating new snapshots
func (r *SnapshotReconciler) updateWorkspaceLastRestored(ctx context.Context, workspaceName, wmNamespace, snapshotName string, logger *zap.Logger) error {
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspaceName, Namespace: wmNamespace}, workspace); err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	now := metav1.Now()
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
		workspace.Status.LastRestoredSnapshot = &workspacev1.WorkspaceLastRestoredSnapshotInfo{
			Name:       snapshotName,
			RestoredAt: now,
		}
		return nil
	}, logger)
}
