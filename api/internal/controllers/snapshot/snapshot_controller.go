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
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotFinalizer = "snapshots.kloudlite.io/finalizer"
	snapshotsBasePath = "/var/lib/kloudlite/storage/.snapshots"
	metadataBasePath  = "/var/lib/kloudlite/storage/.snapshots-metadata"
)

// SnapshotReconciler reconciles Snapshot objects
type SnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles Snapshot events
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshot", req.Name),
	)

	logger.Info("Reconciling Snapshot")

	// Fetch the Snapshot instance (cluster-scoped)
	snapshot := &snapshotv1.Snapshot{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name}, snapshot)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Snapshot not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if snapshot is being deleted
	if snapshot.DeletionTimestamp != nil {
		logger.Info("Snapshot is being deleted, starting cleanup")
		return r.handleDeletion(ctx, snapshot, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
		logger.Info("Adding finalizer to snapshot")
		controllerutil.AddFinalizer(snapshot, snapshotFinalizer)
		if err := r.Update(ctx, snapshot); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Handle based on current state
	switch snapshot.Status.State {
	case "", snapshotv1.SnapshotStatePending:
		return r.handlePending(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateCreating:
		return r.handleCreating(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateReady:
		return reconcile.Result{}, nil
	case snapshotv1.SnapshotStateDeleting:
		return r.handleDeleting(ctx, snapshot, logger)
	case snapshotv1.SnapshotStateFailed:
		return reconcile.Result{}, nil
	default:
		logger.Warn("Unknown snapshot state", zap.String("state", string(snapshot.Status.State)))
		return reconcile.Result{}, nil
	}
}

// handlePending starts the snapshot creation process
func (r *SnapshotReconciler) handlePending(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Snapshot is pending, starting creation")

	envName := snapshot.Spec.EnvironmentRef.Name

	// Fetch the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		logger.Error("Failed to get environment", zap.Error(err), zap.String("environment", envName))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Environment not found: %s", envName), logger)
	}

	// Set state to Creating
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateCreating
		snapshot.Status.Message = "Preparing to create snapshot"
		snapshot.Status.WorkMachineName = env.Spec.WorkMachineName
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Creating", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleCreating manages the snapshot creation process
func (r *SnapshotReconciler) handleCreating(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	envName := snapshot.Spec.EnvironmentRef.Name

	// Fetch the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		logger.Error("Failed to get environment", zap.Error(err))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Environment not found: %s", envName), logger)
	}

	namespace := env.Spec.TargetNamespace
	if namespace == "" {
		return r.updateStatusFailed(ctx, snapshot, "Environment has no target namespace", logger)
	}

	// Generate snapshot path based on timestamp
	snapshotTimestamp := time.Now().UTC().Format("20060102-150405")
	snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("%s-%s", envName, snapshotTimestamp))

	// List PVCs in the environment namespace
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, client.InNamespace(namespace)); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
	}

	// Create SnapshotRequest for each PVC
	var pvcSnapshots []snapshotv1.PVCSnapshotInfo
	for _, pvc := range pvcList.Items {
		// Determine source path from PVC
		// Local-path-provisioner uses /var/lib/kloudlite/storage/environments/<namespace>/<pvc>
		sourcePath := filepath.Join("/var/lib/kloudlite/storage/environments", namespace, pvc.Name)
		pvcSnapshotPath := filepath.Join(snapshotPath, "pvcs", pvc.Name)

		// Create SnapshotRequest
		snapshotReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", snapshot.Name, pvc.Name),
				Namespace: fmt.Sprintf("wm-%s", env.Spec.OwnedBy),
				Labels: map[string]string{
					"snapshots.kloudlite.io/snapshot": snapshot.Name,
					"snapshots.kloudlite.io/pvc":      pvc.Name,
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:       snapshotv1.SnapshotOperationCreate,
				SourcePath:      sourcePath,
				SnapshotPath:    pvcSnapshotPath,
				SnapshotRef:     snapshot.Name,
				EnvironmentName: envName,
				ReadOnly:        true,
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(snapshot, snapshotReq, r.Scheme); err != nil {
			logger.Error("Failed to set owner reference", zap.Error(err))
		}

		// Create or update SnapshotRequest
		if err := r.Create(ctx, snapshotReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create SnapshotRequest", zap.Error(err), zap.String("pvc", pvc.Name))
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create SnapshotRequest for PVC %s", pvc.Name), logger)
			}
		}

		pvcSnapshots = append(pvcSnapshots, snapshotv1.PVCSnapshotInfo{
			PVCName:      pvc.Name,
			SnapshotPath: pvcSnapshotPath,
		})
	}

	// Export K8s metadata if requested
	var resourceMetadata *snapshotv1.ResourceMetadataInfo
	if snapshot.Spec.IncludeMetadata {
		var err error
		resourceMetadata, err = r.exportMetadata(ctx, namespace, snapshotPath, logger)
		if err != nil {
			logger.Warn("Failed to export metadata, continuing anyway", zap.Error(err))
		}
	}

	// Check if all SnapshotRequests are complete
	allComplete, err := r.checkSnapshotRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check SnapshotRequest status", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !allComplete {
		// Update status with progress
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Creating btrfs snapshots..."
			snapshot.Status.SnapshotPath = snapshotPath
			snapshot.Status.PVCSnapshots = pvcSnapshots
			snapshot.Status.ResourceMetadata = resourceMetadata
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status with progress", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// All snapshots complete - calculate total size
	var totalSize int64
	for _, pvcInfo := range pvcSnapshots {
		totalSize += pvcInfo.SizeBytes
	}

	// Update status to Ready
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Snapshot created successfully"
		snapshot.Status.SnapshotPath = snapshotPath
		snapshot.Status.SizeBytes = totalSize
		snapshot.Status.SizeHuman = formatSize(totalSize)
		snapshot.Status.CreatedAt = &now
		snapshot.Status.PVCSnapshots = pvcSnapshots
		snapshot.Status.ResourceMetadata = resourceMetadata
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot created successfully",
		zap.String("path", snapshotPath),
		zap.Int64("sizeBytes", totalSize))

	return reconcile.Result{}, nil
}

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

	// Create delete SnapshotRequests for each PVC snapshot
	for _, pvcInfo := range snapshot.Status.PVCSnapshots {
		deleteReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-delete-%s", snapshot.Name, pvcInfo.PVCName),
				Namespace: snapshot.Status.WorkMachineName,
				Labels: map[string]string{
					"snapshots.kloudlite.io/snapshot":  snapshot.Name,
					"snapshots.kloudlite.io/operation": "delete",
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:    snapshotv1.SnapshotOperationDelete,
				SnapshotPath: pvcInfo.SnapshotPath,
				SnapshotRef:  snapshot.Name,
			},
		}

		if err := r.Create(ctx, deleteReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create delete SnapshotRequest", zap.Error(err))
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

	// Remove finalizer
	controllerutil.RemoveFinalizer(snapshot, snapshotFinalizer)
	if err := r.Update(ctx, snapshot); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot cleanup complete")
	return reconcile.Result{}, nil
}

// checkSnapshotRequestsComplete checks if all SnapshotRequests for a Snapshot are complete
func (r *SnapshotReconciler) checkSnapshotRequestsComplete(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (bool, error) {
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{"snapshots.kloudlite.io/snapshot": snapshot.Name}); err != nil {
		return false, err
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

// exportMetadata exports K8s resources to JSON files
func (r *SnapshotReconciler) exportMetadata(ctx context.Context, namespace, snapshotPath string, logger *zap.Logger) (*snapshotv1.ResourceMetadataInfo, error) {
	metadataPath := filepath.Join(snapshotPath, "metadata")

	// Create metadata directory
	if err := os.MkdirAll(metadataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	info := &snapshotv1.ResourceMetadataInfo{}

	// Export ConfigMaps
	configMaps := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMaps, client.InNamespace(namespace)); err == nil {
		info.ConfigMaps = int32(len(configMaps.Items))
		if err := exportToJSON(filepath.Join(metadataPath, "configmaps.json"), configMaps); err != nil {
			logger.Warn("Failed to export ConfigMaps", zap.Error(err))
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
		if err := exportToJSON(filepath.Join(metadataPath, "secrets.json"), filtered); err != nil {
			logger.Warn("Failed to export Secrets", zap.Error(err))
		}
	}

	// Export Deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err == nil {
		info.Deployments = int32(len(deployments.Items))
		if err := exportToJSON(filepath.Join(metadataPath, "deployments.json"), deployments); err != nil {
			logger.Warn("Failed to export Deployments", zap.Error(err))
		}
	}

	// Export Services
	services := &corev1.ServiceList{}
	if err := r.List(ctx, services, client.InNamespace(namespace)); err == nil {
		info.Services = int32(len(services.Items))
		if err := exportToJSON(filepath.Join(metadataPath, "services.json"), services); err != nil {
			logger.Warn("Failed to export Services", zap.Error(err))
		}
	}

	// Export StatefulSets
	statefulSets := &appsv1.StatefulSetList{}
	if err := r.List(ctx, statefulSets, client.InNamespace(namespace)); err == nil {
		info.StatefulSets = int32(len(statefulSets.Items))
		if err := exportToJSON(filepath.Join(metadataPath, "statefulsets.json"), statefulSets); err != nil {
			logger.Warn("Failed to export StatefulSets", zap.Error(err))
		}
	}

	logger.Info("Exported metadata",
		zap.Int32("configMaps", info.ConfigMaps),
		zap.Int32("secrets", info.Secrets),
		zap.Int32("deployments", info.Deployments),
		zap.Int32("services", info.Services),
		zap.Int32("statefulSets", info.StatefulSets))

	return info, nil
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

// SetupWithManager sets up the controller with the Manager
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.Snapshot{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&snapshotv1.SnapshotRequest{}).
		Complete(r)
}
