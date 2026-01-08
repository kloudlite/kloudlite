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
	"github.com/kloudlite/kloudlite/api/pkg/oci"
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

	// Delete image from registry if pushed
	if snapshot.Status.RegistryStatus != nil && snapshot.Status.RegistryStatus.Pushed {
		r.deleteRegistryImage(snapshot, logger)
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

	// Export PVCs - needed for data restoration to know original claim names
	pvcs := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcs, client.InNamespace(namespace)); err == nil {
		info.PVCs = int32(len(pvcs.Items))
		if data, err := json.Marshal(pvcs); err == nil {
			metadata.PVCs = string(data)
		} else {
			logger.Warn("Failed to marshal PVCs", zap.Error(err))
		}
	}

	logger.Info("Collected metadata",
		zap.Int32("configMaps", info.ConfigMaps),
		zap.Int32("secrets", info.Secrets),
		zap.Int32("deployments", info.Deployments),
		zap.Int32("services", info.Services),
		zap.Int32("statefulSets", info.StatefulSets),
		zap.Int32("compositions", info.Compositions),
		zap.Int32("pvcs", info.PVCs))

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

// importMetadata restores K8s resources from the snapshot's collected metadata
// This replaces the current resources in the namespace with those from the snapshot
func (r *SnapshotReconciler) importMetadata(ctx context.Context, namespace string, metadata *snapshotv1.SnapshotMetadata, logger *zap.Logger) error {
	if metadata == nil {
		logger.Info("No metadata to restore")
		return nil
	}

	logger.Info("Restoring K8s metadata from snapshot", zap.String("namespace", namespace))

	// Restore ConfigMaps
	if metadata.ConfigMaps != "" {
		if err := r.restoreConfigMaps(ctx, namespace, metadata.ConfigMaps, logger); err != nil {
			logger.Warn("Failed to restore ConfigMaps", zap.Error(err))
		}
	}

	// Restore Secrets
	if metadata.Secrets != "" {
		if err := r.restoreSecrets(ctx, namespace, metadata.Secrets, logger); err != nil {
			logger.Warn("Failed to restore Secrets", zap.Error(err))
		}
	}

	// Restore Compositions (before Deployments/StatefulSets as they may depend on them)
	if metadata.Compositions != "" {
		if err := r.restoreCompositions(ctx, namespace, metadata.Compositions, logger); err != nil {
			logger.Warn("Failed to restore Compositions", zap.Error(err))
		}
	}

	// Restore Services
	if metadata.Services != "" {
		if err := r.restoreServices(ctx, namespace, metadata.Services, logger); err != nil {
			logger.Warn("Failed to restore Services", zap.Error(err))
		}
	}

	// Restore Deployments
	if metadata.Deployments != "" {
		if err := r.restoreDeployments(ctx, namespace, metadata.Deployments, logger); err != nil {
			logger.Warn("Failed to restore Deployments", zap.Error(err))
		}
	}

	// Restore StatefulSets
	if metadata.StatefulSets != "" {
		if err := r.restoreStatefulSets(ctx, namespace, metadata.StatefulSets, logger); err != nil {
			logger.Warn("Failed to restore StatefulSets", zap.Error(err))
		}
	}

	logger.Info("Metadata restoration complete")
	return nil
}

// restoreConfigMaps restores ConfigMaps from JSON
func (r *SnapshotReconciler) restoreConfigMaps(ctx context.Context, namespace, jsonData string, logger *zap.Logger) error {
	var cmList corev1.ConfigMapList
	if err := json.Unmarshal([]byte(jsonData), &cmList); err != nil {
		return fmt.Errorf("failed to unmarshal ConfigMaps: %w", err)
	}

	// Get current ConfigMaps
	currentList := &corev1.ConfigMapList{}
	if err := r.List(ctx, currentList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list current ConfigMaps: %w", err)
	}

	// Build map of snapshot ConfigMaps
	snapshotCMs := make(map[string]corev1.ConfigMap)
	for _, cm := range cmList.Items {
		snapshotCMs[cm.Name] = cm
	}

	// Delete ConfigMaps that don't exist in snapshot
	for _, cm := range currentList.Items {
		// Skip kube-root-ca.crt as it's managed by Kubernetes
		if cm.Name == "kube-root-ca.crt" {
			continue
		}
		if _, exists := snapshotCMs[cm.Name]; !exists {
			logger.Info("Deleting ConfigMap not in snapshot", zap.String("name", cm.Name))
			if err := r.Delete(ctx, &cm); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete ConfigMap", zap.String("name", cm.Name), zap.Error(err))
			}
		}
	}

	// Create or update ConfigMaps from snapshot
	for _, cm := range cmList.Items {
		// Skip kube-root-ca.crt
		if cm.Name == "kube-root-ca.crt" {
			continue
		}
		newCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        cm.Name,
				Namespace:   namespace,
				Labels:      cm.Labels,
				Annotations: cm.Annotations,
			},
			Data:       cm.Data,
			BinaryData: cm.BinaryData,
		}

		existing := &corev1.ConfigMap{}
		err := r.Get(ctx, client.ObjectKey{Name: cm.Name, Namespace: namespace}, existing)
		if apierrors.IsNotFound(err) {
			logger.Info("Creating ConfigMap from snapshot", zap.String("name", cm.Name))
			if err := r.Create(ctx, newCM); err != nil {
				logger.Warn("Failed to create ConfigMap", zap.String("name", cm.Name), zap.Error(err))
			}
		} else if err == nil {
			existing.Data = cm.Data
			existing.BinaryData = cm.BinaryData
			existing.Labels = cm.Labels
			existing.Annotations = cm.Annotations
			logger.Info("Updating ConfigMap from snapshot", zap.String("name", cm.Name))
			if err := r.Update(ctx, existing); err != nil {
				logger.Warn("Failed to update ConfigMap", zap.String("name", cm.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// restoreSecrets restores Secrets from JSON
func (r *SnapshotReconciler) restoreSecrets(ctx context.Context, namespace, jsonData string, logger *zap.Logger) error {
	var secretList []corev1.Secret
	if err := json.Unmarshal([]byte(jsonData), &secretList); err != nil {
		return fmt.Errorf("failed to unmarshal Secrets: %w", err)
	}

	// Get current Secrets
	currentList := &corev1.SecretList{}
	if err := r.List(ctx, currentList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list current Secrets: %w", err)
	}

	// Build map of snapshot Secrets
	snapshotSecrets := make(map[string]corev1.Secret)
	for _, s := range secretList {
		snapshotSecrets[s.Name] = s
	}

	// Delete Secrets that don't exist in snapshot (except service account tokens)
	for _, s := range currentList.Items {
		if s.Type == corev1.SecretTypeServiceAccountToken {
			continue
		}
		if _, exists := snapshotSecrets[s.Name]; !exists {
			logger.Info("Deleting Secret not in snapshot", zap.String("name", s.Name))
			if err := r.Delete(ctx, &s); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete Secret", zap.String("name", s.Name), zap.Error(err))
			}
		}
	}

	// Create or update Secrets from snapshot
	for _, s := range secretList {
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        s.Name,
				Namespace:   namespace,
				Labels:      s.Labels,
				Annotations: s.Annotations,
			},
			Type:       s.Type,
			Data:       s.Data,
			StringData: s.StringData,
		}

		existing := &corev1.Secret{}
		err := r.Get(ctx, client.ObjectKey{Name: s.Name, Namespace: namespace}, existing)
		if apierrors.IsNotFound(err) {
			logger.Info("Creating Secret from snapshot", zap.String("name", s.Name))
			if err := r.Create(ctx, newSecret); err != nil {
				logger.Warn("Failed to create Secret", zap.String("name", s.Name), zap.Error(err))
			}
		} else if err == nil {
			existing.Data = s.Data
			existing.StringData = s.StringData
			existing.Labels = s.Labels
			existing.Annotations = s.Annotations
			logger.Info("Updating Secret from snapshot", zap.String("name", s.Name))
			if err := r.Update(ctx, existing); err != nil {
				logger.Warn("Failed to update Secret", zap.String("name", s.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// restoreServices restores Services from JSON
func (r *SnapshotReconciler) restoreServices(ctx context.Context, namespace, jsonData string, logger *zap.Logger) error {
	var svcList corev1.ServiceList
	if err := json.Unmarshal([]byte(jsonData), &svcList); err != nil {
		return fmt.Errorf("failed to unmarshal Services: %w", err)
	}

	// Get current Services
	currentList := &corev1.ServiceList{}
	if err := r.List(ctx, currentList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list current Services: %w", err)
	}

	// Build map of snapshot Services
	snapshotSvcs := make(map[string]corev1.Service)
	for _, svc := range svcList.Items {
		snapshotSvcs[svc.Name] = svc
	}

	// Delete Services that don't exist in snapshot
	for _, svc := range currentList.Items {
		if _, exists := snapshotSvcs[svc.Name]; !exists {
			logger.Info("Deleting Service not in snapshot", zap.String("name", svc.Name))
			if err := r.Delete(ctx, &svc); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete Service", zap.String("name", svc.Name), zap.Error(err))
			}
		}
	}

	// Create or update Services from snapshot
	for _, svc := range svcList.Items {
		newSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        svc.Name,
				Namespace:   namespace,
				Labels:      svc.Labels,
				Annotations: svc.Annotations,
			},
			Spec: corev1.ServiceSpec{
				Selector:  svc.Spec.Selector,
				Ports:     svc.Spec.Ports,
				Type:      svc.Spec.Type,
				ClusterIP: svc.Spec.ClusterIP,
			},
		}

		existing := &corev1.Service{}
		err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: namespace}, existing)
		if apierrors.IsNotFound(err) {
			// Clear ClusterIP for new services to let K8s assign
			newSvc.Spec.ClusterIP = ""
			logger.Info("Creating Service from snapshot", zap.String("name", svc.Name))
			if err := r.Create(ctx, newSvc); err != nil {
				logger.Warn("Failed to create Service", zap.String("name", svc.Name), zap.Error(err))
			}
		} else if err == nil {
			// Update mutable fields only
			existing.Spec.Selector = svc.Spec.Selector
			existing.Spec.Ports = svc.Spec.Ports
			existing.Labels = svc.Labels
			existing.Annotations = svc.Annotations
			logger.Info("Updating Service from snapshot", zap.String("name", svc.Name))
			if err := r.Update(ctx, existing); err != nil {
				logger.Warn("Failed to update Service", zap.String("name", svc.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// restoreDeployments restores Deployments from JSON
func (r *SnapshotReconciler) restoreDeployments(ctx context.Context, namespace, jsonData string, logger *zap.Logger) error {
	var deployList appsv1.DeploymentList
	if err := json.Unmarshal([]byte(jsonData), &deployList); err != nil {
		return fmt.Errorf("failed to unmarshal Deployments: %w", err)
	}

	// Get current Deployments
	currentList := &appsv1.DeploymentList{}
	if err := r.List(ctx, currentList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list current Deployments: %w", err)
	}

	// Build map of snapshot Deployments
	snapshotDeploys := make(map[string]appsv1.Deployment)
	for _, d := range deployList.Items {
		snapshotDeploys[d.Name] = d
	}

	// Delete Deployments that don't exist in snapshot
	for _, d := range currentList.Items {
		if _, exists := snapshotDeploys[d.Name]; !exists {
			logger.Info("Deleting Deployment not in snapshot", zap.String("name", d.Name))
			if err := r.Delete(ctx, &d); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete Deployment", zap.String("name", d.Name), zap.Error(err))
			}
		}
	}

	// Create or update Deployments from snapshot
	for _, d := range deployList.Items {
		// Start with 0 replicas, scaleUpWorkloads will restore them
		zero := int32(0)
		newDeploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        d.Name,
				Namespace:   namespace,
				Labels:      d.Labels,
				Annotations: d.Annotations,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &zero,
				Selector: d.Spec.Selector,
				Template: d.Spec.Template,
				Strategy: d.Spec.Strategy,
			},
		}

		existing := &appsv1.Deployment{}
		err := r.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: namespace}, existing)
		if apierrors.IsNotFound(err) {
			logger.Info("Creating Deployment from snapshot", zap.String("name", d.Name))
			if err := r.Create(ctx, newDeploy); err != nil {
				logger.Warn("Failed to create Deployment", zap.String("name", d.Name), zap.Error(err))
			}
		} else if err == nil {
			existing.Spec.Template = d.Spec.Template
			existing.Spec.Strategy = d.Spec.Strategy
			existing.Labels = d.Labels
			existing.Annotations = d.Annotations
			// Keep replicas at 0 during restore
			existing.Spec.Replicas = &zero
			logger.Info("Updating Deployment from snapshot", zap.String("name", d.Name))
			if err := r.Update(ctx, existing); err != nil {
				logger.Warn("Failed to update Deployment", zap.String("name", d.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// restoreStatefulSets restores StatefulSets from JSON
func (r *SnapshotReconciler) restoreStatefulSets(ctx context.Context, namespace, jsonData string, logger *zap.Logger) error {
	var stsList appsv1.StatefulSetList
	if err := json.Unmarshal([]byte(jsonData), &stsList); err != nil {
		return fmt.Errorf("failed to unmarshal StatefulSets: %w", err)
	}

	// Get current StatefulSets
	currentList := &appsv1.StatefulSetList{}
	if err := r.List(ctx, currentList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list current StatefulSets: %w", err)
	}

	// Build map of snapshot StatefulSets
	snapshotSts := make(map[string]appsv1.StatefulSet)
	for _, s := range stsList.Items {
		snapshotSts[s.Name] = s
	}

	// Delete StatefulSets that don't exist in snapshot
	for _, s := range currentList.Items {
		if _, exists := snapshotSts[s.Name]; !exists {
			logger.Info("Deleting StatefulSet not in snapshot", zap.String("name", s.Name))
			if err := r.Delete(ctx, &s); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete StatefulSet", zap.String("name", s.Name), zap.Error(err))
			}
		}
	}

	// Create or update StatefulSets from snapshot
	for _, s := range stsList.Items {
		// Start with 0 replicas, scaleUpWorkloads will restore them
		zero := int32(0)
		newSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        s.Name,
				Namespace:   namespace,
				Labels:      s.Labels,
				Annotations: s.Annotations,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas:             &zero,
				Selector:             s.Spec.Selector,
				Template:             s.Spec.Template,
				ServiceName:          s.Spec.ServiceName,
				VolumeClaimTemplates: s.Spec.VolumeClaimTemplates,
			},
		}

		existing := &appsv1.StatefulSet{}
		err := r.Get(ctx, client.ObjectKey{Name: s.Name, Namespace: namespace}, existing)
		if apierrors.IsNotFound(err) {
			logger.Info("Creating StatefulSet from snapshot", zap.String("name", s.Name))
			if err := r.Create(ctx, newSts); err != nil {
				logger.Warn("Failed to create StatefulSet", zap.String("name", s.Name), zap.Error(err))
			}
		} else if err == nil {
			existing.Spec.Template = s.Spec.Template
			existing.Labels = s.Labels
			existing.Annotations = s.Annotations
			// Keep replicas at 0 during restore
			existing.Spec.Replicas = &zero
			logger.Info("Updating StatefulSet from snapshot", zap.String("name", s.Name))
			if err := r.Update(ctx, existing); err != nil {
				logger.Warn("Failed to update StatefulSet", zap.String("name", s.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// restoreCompositions restores Compositions from JSON
func (r *SnapshotReconciler) restoreCompositions(ctx context.Context, namespace, jsonData string, logger *zap.Logger) error {
	var compList environmentsv1.CompositionList
	if err := json.Unmarshal([]byte(jsonData), &compList); err != nil {
		return fmt.Errorf("failed to unmarshal Compositions: %w", err)
	}

	// Get current Compositions
	currentList := &environmentsv1.CompositionList{}
	if err := r.List(ctx, currentList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list current Compositions: %w", err)
	}

	// Build map of snapshot Compositions
	snapshotComps := make(map[string]environmentsv1.Composition)
	for _, c := range compList.Items {
		snapshotComps[c.Name] = c
	}

	// Delete Compositions that don't exist in snapshot
	for _, c := range currentList.Items {
		if _, exists := snapshotComps[c.Name]; !exists {
			logger.Info("Deleting Composition not in snapshot", zap.String("name", c.Name))
			if err := r.Delete(ctx, &c); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete Composition", zap.String("name", c.Name), zap.Error(err))
			}
		}
	}

	// Create or update Compositions from snapshot
	for _, c := range compList.Items {
		newComp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:        c.Name,
				Namespace:   namespace,
				Labels:      c.Labels,
				Annotations: c.Annotations,
			},
			Spec: c.Spec,
		}

		existing := &environmentsv1.Composition{}
		err := r.Get(ctx, client.ObjectKey{Name: c.Name, Namespace: namespace}, existing)
		if apierrors.IsNotFound(err) {
			logger.Info("Creating Composition from snapshot", zap.String("name", c.Name))
			if err := r.Create(ctx, newComp); err != nil {
				logger.Warn("Failed to create Composition", zap.String("name", c.Name), zap.Error(err))
			}
		} else if err == nil {
			existing.Spec = c.Spec
			existing.Labels = c.Labels
			existing.Annotations = c.Annotations
			logger.Info("Updating Composition from snapshot", zap.String("name", c.Name))
			if err := r.Update(ctx, existing); err != nil {
				logger.Warn("Failed to update Composition", zap.String("name", c.Name), zap.Error(err))
			}
		}
	}

	return nil
}

// deleteRegistryImage deletes the snapshot image from the OCI registry
func (r *SnapshotReconciler) deleteRegistryImage(snapshot *snapshotv1.Snapshot, logger *zap.Logger) {
	if snapshot.Status.RegistryStatus == nil || snapshot.Status.RegistryStatus.ImageRef == "" {
		return
	}

	imageRef := snapshot.Status.RegistryStatus.ImageRef
	logger.Info("Deleting snapshot image from registry", zap.String("imageRef", imageRef))

	// Create OCI client and delete the image
	client := oci.NewClient(true) // Use insecure for internal registry
	if err := client.Delete(imageRef); err != nil {
		logger.Warn("Failed to delete image from registry",
			zap.String("imageRef", imageRef),
			zap.Error(err))
		// Continue with snapshot deletion even if registry delete fails
		// The image will become orphaned but that's better than blocking deletion
		return
	}

	// Also delete any additional tags
	if len(snapshot.Status.RegistryStatus.Tags) > 1 {
		// Extract registry and repository from imageRef
		// imageRef format: registry/repo:tag
		for _, tag := range snapshot.Status.RegistryStatus.Tags {
			if tag == snapshot.Status.RegistryStatus.Tag {
				continue // Already deleted with primary imageRef
			}
			// Build the full ref for this tag
			// Get base from imageRef by removing the tag
			baseRef := imageRef[:len(imageRef)-len(snapshot.Status.RegistryStatus.Tag)]
			tagRef := baseRef + tag
			if err := client.Delete(tagRef); err != nil {
				logger.Warn("Failed to delete additional tag from registry",
					zap.String("tagRef", tagRef),
					zap.Error(err))
			}
		}
	}

	logger.Info("Successfully deleted snapshot image from registry", zap.String("imageRef", imageRef))
}
