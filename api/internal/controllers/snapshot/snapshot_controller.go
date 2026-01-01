package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	snapshotFinalizer = "snapshots.kloudlite.io/finalizer"
	snapshotsBasePath = "/var/lib/kloudlite/storage/.snapshots"
	metadataBasePath  = "/var/lib/kloudlite/storage/.snapshots-metadata"
	workspaceHomePath = "/var/lib/kloudlite/home/workspaces"
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
	case snapshotv1.SnapshotStateRestoring:
		return r.handleRestoring(ctx, snapshot, logger)
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

	// Determine snapshot type
	if snapshot.Spec.WorkspaceRef != nil {
		return r.handleWorkspacePending(ctx, snapshot, logger)
	}

	// Environment snapshot
	if snapshot.Spec.EnvironmentRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Either environmentRef or workspaceRef must be set", logger)
	}

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
		snapshot.Status.Message = "Preparing to create environment snapshot"
		snapshot.Status.WorkMachineName = env.Spec.WorkMachineName
		snapshot.Status.SnapshotType = snapshotv1.SnapshotTypeEnvironment
		snapshot.Status.TargetName = envName
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Creating", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleWorkspacePending starts the workspace snapshot creation process
func (r *SnapshotReconciler) handleWorkspacePending(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	wsRef := snapshot.Spec.WorkspaceRef
	wmNamespace := fmt.Sprintf("wm-%s", snapshot.Spec.OwnedBy)

	logger.Info("Starting workspace snapshot",
		zap.String("workspace", wsRef.Name),
		zap.String("workmachine", wsRef.WorkmachineName))

	// Fetch the workspace
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err != nil {
		logger.Error("Failed to get workspace", zap.Error(err), zap.String("workspace", wsRef.Name))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Workspace not found: %s", wsRef.Name), logger)
	}

	// Validate ownership
	if workspace.Spec.OwnedBy != snapshot.Spec.OwnedBy {
		return r.updateStatusFailed(ctx, snapshot, "Workspace is not owned by the snapshot creator", logger)
	}

	// Store previous workspace status and suspend the workspace
	previousStatus := workspace.Spec.Status
	wasSuspended := previousStatus == "suspended"

	if !wasSuspended {
		// Suspend the workspace
		logger.Info("Suspending workspace for snapshot", zap.String("workspace", wsRef.Name))
		workspace.Spec.Status = "suspended"
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to suspend workspace", zap.Error(err))
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to suspend workspace: %v", err), logger)
		}
	}

	// Set state to Creating
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateCreating
		snapshot.Status.Message = "Suspending workspace and preparing to create snapshot"
		snapshot.Status.WorkMachineName = wmNamespace
		snapshot.Status.SnapshotType = snapshotv1.SnapshotTypeWorkspace
		snapshot.Status.TargetName = wsRef.Name
		snapshot.Status.WorkspaceName = wsRef.Name
		snapshot.Status.WorkspaceWasSuspended = wasSuspended
		snapshot.Status.PreviousWorkspaceStatus = previousStatus
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Creating", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// scaleEnvironment scales all deployments and statefulsets in a namespace to the specified replica count
func (r *SnapshotReconciler) scaleEnvironment(ctx context.Context, namespace string, replicas int32, logger *zap.Logger) error {
	// Scale deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}
	for _, deploy := range deployments.Items {
		if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas != replicas {
			deploy.Spec.Replicas = &replicas
			if err := r.Update(ctx, &deploy); err != nil {
				logger.Warn("Failed to scale deployment", zap.String("deployment", deploy.Name), zap.Error(err))
			} else {
				logger.Info("Scaled deployment", zap.String("deployment", deploy.Name), zap.Int32("replicas", replicas))
			}
		}
	}

	// Scale statefulsets
	statefulsets := &appsv1.StatefulSetList{}
	if err := r.List(ctx, statefulsets, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list statefulsets: %w", err)
	}
	for _, sts := range statefulsets.Items {
		if sts.Spec.Replicas != nil && *sts.Spec.Replicas != replicas {
			sts.Spec.Replicas = &replicas
			if err := r.Update(ctx, &sts); err != nil {
				logger.Warn("Failed to scale statefulset", zap.String("statefulset", sts.Name), zap.Error(err))
			} else {
				logger.Info("Scaled statefulset", zap.String("statefulset", sts.Name), zap.Int32("replicas", replicas))
			}
		}
	}

	return nil
}

// waitForPodsTerminated waits for all pods in a namespace to terminate
func (r *SnapshotReconciler) waitForPodsTerminated(ctx context.Context, namespace string, logger *zap.Logger) bool {
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(namespace)); err != nil {
		logger.Warn("Failed to list pods", zap.Error(err))
		return false
	}
	// Check if any non-terminated pods exist (excluding jobs/completed pods)
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			return false
		}
	}
	return true
}

// handleCreating manages the snapshot creation process
func (r *SnapshotReconciler) handleCreating(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	// Dispatch based on snapshot type
	if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
		return r.handleWorkspaceCreating(ctx, snapshot, logger)
	}

	// Environment snapshot
	if snapshot.Spec.EnvironmentRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Environment reference is required for environment snapshots", logger)
	}

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

	// Scale down environment to stop all pods before snapshot
	if err := r.scaleEnvironment(ctx, namespace, 0, logger); err != nil {
		logger.Warn("Failed to scale down environment", zap.Error(err))
	}

	// Wait for pods to terminate
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Stopping environment pods..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Generate snapshot path based on timestamp
	snapshotTimestamp := time.Now().UTC().Format("20060102-150405")
	snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("%s-%s", envName, snapshotTimestamp))

	// List PVCs in the environment namespace
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, client.InNamespace(namespace)); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		// Scale environment back up on failure
		if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
			logger.Warn("Failed to scale up environment after PVC list failure", zap.Error(scaleErr))
		}
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
	}

	// Create SnapshotRequest for each PVC
	var pvcSnapshots []snapshotv1.PVCSnapshotInfo
	for _, pvc := range pvcList.Items {
		// Get the actual PV path from the PersistentVolume
		pvName := pvc.Spec.VolumeName
		if pvName == "" {
			logger.Warn("PVC has no bound PV, skipping", zap.String("pvc", pvc.Name))
			continue
		}

		pv := &corev1.PersistentVolume{}
		if err := r.Get(ctx, client.ObjectKey{Name: pvName}, pv); err != nil {
			logger.Error("Failed to get PV", zap.Error(err), zap.String("pv", pvName))
			continue
		}

		// Get the actual host path from the PV (local-path-provisioner uses spec.local.path)
		var sourcePath string
		if pv.Spec.Local != nil && pv.Spec.Local.Path != "" {
			sourcePath = pv.Spec.Local.Path
		} else if pv.Spec.HostPath != nil && pv.Spec.HostPath.Path != "" {
			sourcePath = pv.Spec.HostPath.Path
		} else {
			logger.Warn("PV has no local or hostPath, skipping", zap.String("pv", pvName))
			continue
		}

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
				// Scale environment back up on failure
				if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
					logger.Warn("Failed to scale up environment after SnapshotRequest creation failure", zap.Error(scaleErr))
				}
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create SnapshotRequest for PVC %s", pvc.Name), logger)
			}
		}

		pvcSnapshots = append(pvcSnapshots, snapshotv1.PVCSnapshotInfo{
			PVCName:      pvc.Name,
			SnapshotPath: pvcSnapshotPath,
			SourcePath:   sourcePath,
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

	// Scale environment back up
	if err := r.scaleEnvironment(ctx, namespace, 1, logger); err != nil {
		logger.Warn("Failed to scale up environment after snapshot", zap.Error(err))
	}

	logger.Info("Snapshot created successfully",
		zap.String("path", snapshotPath),
		zap.Int64("sizeBytes", totalSize))

	return reconcile.Result{}, nil
}

// handleWorkspaceCreating manages the workspace snapshot creation process
func (r *SnapshotReconciler) handleWorkspaceCreating(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	wsRef := snapshot.Spec.WorkspaceRef
	wmNamespace := snapshot.Status.WorkMachineName

	logger.Info("Creating workspace snapshot",
		zap.String("workspace", wsRef.Name),
		zap.String("namespace", wmNamespace))

	// Fetch the workspace to check pod status
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		return r.handleWorkspaceSnapshotFailure(ctx, snapshot, fmt.Sprintf("Workspace not found: %s", wsRef.Name), logger)
	}

	// Wait for workspace pod to terminate
	podName := fmt.Sprintf("workspace-%s", wsRef.Name)
	pod := &corev1.Pod{}
	podErr := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: wmNamespace}, pod)

	if podErr == nil {
		// Pod still exists, wait for it to terminate
		logger.Info("Waiting for workspace pod to terminate", zap.String("pod", podName))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Waiting for workspace pod to terminate..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status message", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if !apierrors.IsNotFound(podErr) {
		logger.Error("Failed to check workspace pod", zap.Error(podErr))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Pod is terminated, proceed with snapshot

	// Generate snapshot path based on timestamp
	snapshotTimestamp := time.Now().UTC().Format("20060102-150405")
	snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("ws-%s-%s", wsRef.Name, snapshotTimestamp))

	// Workspace home directory path
	sourcePath := filepath.Join(workspaceHomePath, wsRef.Name)
	workspaceSnapshotPath := filepath.Join(snapshotPath, "home")

	// Create SnapshotRequest for workspace home directory
	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-home", snapshot.Name),
			Namespace: wmNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/workspace": wsRef.Name,
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:     snapshotv1.SnapshotOperationCreate,
			SourcePath:    sourcePath,
			SnapshotPath:  workspaceSnapshotPath,
			SnapshotRef:   snapshot.Name,
			WorkspaceName: wsRef.Name,
			ReadOnly:      true,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(snapshot, snapshotReq, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference", zap.Error(err))
	}

	// Create SnapshotRequest
	if err := r.Create(ctx, snapshotReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create SnapshotRequest", zap.Error(err))
			return r.handleWorkspaceSnapshotFailure(ctx, snapshot, "Failed to create SnapshotRequest for workspace home", logger)
		}
	}

	// Export workspace metadata if requested
	var packageRequestsPath string
	if snapshot.Spec.IncludeMetadata {
		var err error
		packageRequestsPath, err = r.exportWorkspaceMetadata(ctx, wsRef.Name, wmNamespace, snapshotPath, logger)
		if err != nil {
			logger.Warn("Failed to export workspace metadata, continuing anyway", zap.Error(err))
		}
	}

	// Check if SnapshotRequest is complete
	allComplete, err := r.checkSnapshotRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check SnapshotRequest status", zap.Error(err))
		return r.handleWorkspaceSnapshotFailure(ctx, snapshot, fmt.Sprintf("SnapshotRequest failed: %v", err), logger)
	}

	if !allComplete {
		// Update status with progress
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Creating workspace snapshot..."
			snapshot.Status.SnapshotPath = snapshotPath
			snapshot.Status.PackageRequestsPath = packageRequestsPath
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status with progress", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Snapshot complete - get size from SnapshotRequest
	var totalSize int64
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{"snapshots.kloudlite.io/snapshot": snapshot.Name}); err == nil {
		for _, req := range snapshotReqList.Items {
			totalSize += req.Status.SizeBytes
		}
	}

	// Resume workspace if it wasn't already suspended
	if !snapshot.Status.WorkspaceWasSuspended {
		logger.Info("Resuming workspace after snapshot", zap.String("workspace", wsRef.Name))

		// Refetch workspace to get latest version
		if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err == nil {
			originalStatus := snapshot.Status.PreviousWorkspaceStatus
			if originalStatus == "" {
				originalStatus = "active"
			}
			workspace.Spec.Status = originalStatus
			if err := r.Update(ctx, workspace); err != nil {
				logger.Warn("Failed to resume workspace", zap.Error(err))
				// Don't fail the snapshot - it was created successfully
			}
		}
	}

	// Update status to Ready
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Workspace snapshot created successfully"
		snapshot.Status.SnapshotPath = snapshotPath
		snapshot.Status.SizeBytes = totalSize
		snapshot.Status.SizeHuman = formatSize(totalSize)
		snapshot.Status.CreatedAt = &now
		snapshot.Status.PackageRequestsPath = packageRequestsPath
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Workspace snapshot created successfully",
		zap.String("workspace", wsRef.Name),
		zap.String("path", snapshotPath),
		zap.Int64("sizeBytes", totalSize))

	return reconcile.Result{}, nil
}

// handleWorkspaceSnapshotFailure handles a workspace snapshot failure and resumes the workspace
func (r *SnapshotReconciler) handleWorkspaceSnapshotFailure(ctx context.Context, snapshot *snapshotv1.Snapshot, message string, logger *zap.Logger) (reconcile.Result, error) {
	// Try to resume workspace if we suspended it
	if !snapshot.Status.WorkspaceWasSuspended && snapshot.Spec.WorkspaceRef != nil {
		wsRef := snapshot.Spec.WorkspaceRef
		wmNamespace := snapshot.Status.WorkMachineName

		workspace := &workspacev1.Workspace{}
		if err := r.Get(ctx, client.ObjectKey{Name: wsRef.Name, Namespace: wmNamespace}, workspace); err == nil {
			originalStatus := snapshot.Status.PreviousWorkspaceStatus
			if originalStatus == "" {
				originalStatus = "active"
			}
			workspace.Spec.Status = originalStatus
			if err := r.Update(ctx, workspace); err != nil {
				logger.Warn("Failed to resume workspace after snapshot failure", zap.Error(err))
			} else {
				logger.Info("Resumed workspace after snapshot failure", zap.String("workspace", wsRef.Name))
			}
		}
	}

	return r.updateStatusFailed(ctx, snapshot, message, logger)
}

// exportWorkspaceMetadata exports workspace-specific metadata (PackageRequests, settings)
func (r *SnapshotReconciler) exportWorkspaceMetadata(ctx context.Context, workspaceName, wmNamespace, snapshotPath string, logger *zap.Logger) (string, error) {
	metadataPath := filepath.Join(snapshotPath, "metadata")

	// Create metadata directory
	if err := os.MkdirAll(metadataPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Export PackageRequests for this workspace
	packageRequests := &packagesv1.PackageRequestList{}
	if err := r.List(ctx, packageRequests, client.InNamespace(wmNamespace)); err == nil {
		// Filter to only include PackageRequests for this workspace
		var filtered []packagesv1.PackageRequest
		for _, pr := range packageRequests.Items {
			if pr.Spec.WorkspaceRef == workspaceName {
				filtered = append(filtered, pr)
			}
		}

		packageRequestsPath := filepath.Join(metadataPath, "package-requests.json")
		if err := exportToJSON(packageRequestsPath, filtered); err != nil {
			logger.Warn("Failed to export PackageRequests", zap.Error(err))
		} else {
			logger.Info("Exported PackageRequests", zap.Int("count", len(filtered)))
		}
	}

	// Export Workspace resource itself (settings, config, etc.)
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspaceName, Namespace: wmNamespace}, workspace); err == nil {
		workspacePath := filepath.Join(metadataPath, "workspace.json")
		if err := exportToJSON(workspacePath, workspace); err != nil {
			logger.Warn("Failed to export Workspace", zap.Error(err))
		} else {
			logger.Info("Exported Workspace settings")
		}
	}

	return metadataPath, nil
}

// handleRestoring handles the snapshot restore process
func (r *SnapshotReconciler) handleRestoring(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling snapshot restore",
		zap.String("snapshotType", string(snapshot.Status.SnapshotType)),
		zap.String("targetName", snapshot.Status.TargetName))

	// Dispatch based on snapshot type
	if snapshot.Status.SnapshotType == snapshotv1.SnapshotTypeWorkspace {
		return r.handleWorkspaceRestoring(ctx, snapshot, logger)
	}

	// Environment restore
	if snapshot.Spec.EnvironmentRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Environment reference is required for environment restore", logger)
	}

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

	// Scale down environment to stop all pods before restore
	if err := r.scaleEnvironment(ctx, namespace, 0, logger); err != nil {
		logger.Warn("Failed to scale down environment", zap.Error(err))
	}

	// Wait for pods to terminate
	if !r.waitForPodsTerminated(ctx, namespace, logger) {
		logger.Info("Waiting for environment pods to terminate before restore")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
			snapshot.Status.Message = "Stopping environment pods before restore..."
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Create restore SnapshotRequests for each PVC snapshot
	for _, pvcInfo := range snapshot.Status.PVCSnapshots {
		// Target path is the original PV location (stored in SourcePath)
		targetPath := pvcInfo.SourcePath
		if targetPath == "" {
			// Fallback: look up PV path if not stored in snapshot
			pvcList := &corev1.PersistentVolumeClaimList{}
			if err := r.List(ctx, pvcList, client.InNamespace(namespace)); err == nil {
				for _, pvc := range pvcList.Items {
					if pvc.Name == pvcInfo.PVCName && pvc.Spec.VolumeName != "" {
						pv := &corev1.PersistentVolume{}
						if err := r.Get(ctx, client.ObjectKey{Name: pvc.Spec.VolumeName}, pv); err == nil {
							if pv.Spec.Local != nil && pv.Spec.Local.Path != "" {
								targetPath = pv.Spec.Local.Path
							} else if pv.Spec.HostPath != nil && pv.Spec.HostPath.Path != "" {
								targetPath = pv.Spec.HostPath.Path
							}
						}
						break
					}
				}
			}
		}
		if targetPath == "" {
			logger.Error("Could not determine target path for PVC", zap.String("pvc", pvcInfo.PVCName))
			// Scale environment back up on failure
			if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
				logger.Warn("Failed to scale up environment after target path lookup failure", zap.Error(scaleErr))
			}
			return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Could not determine target path for PVC %s", pvcInfo.PVCName), logger)
		}

		restoreReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-restore-%s", snapshot.Name, pvcInfo.PVCName),
				Namespace: snapshot.Status.WorkMachineName,
				Labels: map[string]string{
					"snapshots.kloudlite.io/snapshot":  snapshot.Name,
					"snapshots.kloudlite.io/operation": "restore",
					"snapshots.kloudlite.io/pvc":       pvcInfo.PVCName,
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:       snapshotv1.SnapshotOperationRestore,
				SourcePath:      pvcInfo.SnapshotPath,
				SnapshotPath:    targetPath,
				SnapshotRef:     snapshot.Name,
				EnvironmentName: envName,
			},
		}

		if err := r.Create(ctx, restoreReq); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create restore SnapshotRequest", zap.Error(err), zap.String("pvc", pvcInfo.PVCName))
				// Scale environment back up on failure
				if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
					logger.Warn("Failed to scale up environment after restore request creation failure", zap.Error(scaleErr))
				}
				return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Failed to create restore request for PVC %s", pvcInfo.PVCName), logger)
			}
		}
	}

	// Check if all restore requests are complete
	allComplete, err := r.checkRestoreRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check restore request status", zap.Error(err))
		// Scale environment back up even on failure
		if scaleErr := r.scaleEnvironment(ctx, namespace, 1, logger); scaleErr != nil {
			logger.Warn("Failed to scale up environment after restore failure", zap.Error(scaleErr))
		}
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Restore failed: %v", err), logger)
	}

	if !allComplete {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Restore complete - update status back to Ready
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Snapshot restored successfully"
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Scale environment back up after restore
	if err := r.scaleEnvironment(ctx, namespace, 1, logger); err != nil {
		logger.Warn("Failed to scale up environment after restore", zap.Error(err))
	}

	logger.Info("Snapshot restored successfully", zap.String("environment", envName))
	return reconcile.Result{}, nil
}

// handleWorkspaceRestoring handles workspace snapshot restore
func (r *SnapshotReconciler) handleWorkspaceRestoring(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (reconcile.Result, error) {
	wsRef := snapshot.Spec.WorkspaceRef
	if wsRef == nil {
		return r.updateStatusFailed(ctx, snapshot, "Workspace reference is required for workspace restore", logger)
	}

	wmNamespace := snapshot.Status.WorkMachineName
	workspaceSnapshotPath := filepath.Join(snapshot.Status.SnapshotPath, "home")
	targetPath := filepath.Join(workspaceHomePath, wsRef.Name)

	// Create restore SnapshotRequest
	restoreReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-restore-home", snapshot.Name),
			Namespace: wmNamespace,
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot":  snapshot.Name,
				"snapshots.kloudlite.io/operation": "restore",
				"snapshots.kloudlite.io/workspace": wsRef.Name,
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:     snapshotv1.SnapshotOperationRestore,
			SourcePath:    workspaceSnapshotPath,
			SnapshotPath:  targetPath,
			SnapshotRef:   snapshot.Name,
			WorkspaceName: wsRef.Name,
		},
	}

	if err := r.Create(ctx, restoreReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create restore SnapshotRequest", zap.Error(err))
			return r.updateStatusFailed(ctx, snapshot, "Failed to create restore request for workspace", logger)
		}
	}

	// Check if restore request is complete
	allComplete, err := r.checkRestoreRequestsComplete(ctx, snapshot, logger)
	if err != nil {
		logger.Error("Failed to check restore request status", zap.Error(err))
		return r.updateStatusFailed(ctx, snapshot, fmt.Sprintf("Restore failed: %v", err), logger)
	}

	if !allComplete {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// Restore complete - update status back to Ready
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, snapshot, func() error {
		snapshot.Status.State = snapshotv1.SnapshotStateReady
		snapshot.Status.Message = "Workspace snapshot restored successfully"
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Ready", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Workspace snapshot restored successfully", zap.String("workspace", wsRef.Name))
	return reconcile.Result{}, nil
}

// checkRestoreRequestsComplete checks if all restore SnapshotRequests are complete
func (r *SnapshotReconciler) checkRestoreRequestsComplete(ctx context.Context, snapshot *snapshotv1.Snapshot, logger *zap.Logger) (bool, error) {
	snapshotReqList := &snapshotv1.SnapshotRequestList{}
	if err := r.List(ctx, snapshotReqList, client.MatchingLabels{
		"snapshots.kloudlite.io/snapshot":  snapshot.Name,
		"snapshots.kloudlite.io/operation": "restore",
	}); err != nil {
		return false, err
	}

	if len(snapshotReqList.Items) == 0 {
		// No restore requests created yet, wait
		return false, nil
	}

	for _, req := range snapshotReqList.Items {
		if req.Status.Phase != snapshotv1.SnapshotRequestPhaseCompleted {
			if req.Status.Phase == snapshotv1.SnapshotRequestPhaseFailed {
				logger.Error("Restore SnapshotRequest failed", zap.String("request", req.Name), zap.String("message", req.Status.Message))
				return false, fmt.Errorf("restore request %s failed: %s", req.Name, req.Status.Message)
			}
			return false, nil
		}
	}

	return true, nil
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
		// Create delete SnapshotRequests for each PVC snapshot (environment)
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
		For(&snapshotv1.Snapshot{}).
		Owns(&snapshotv1.SnapshotRequest{}).
		Complete(r)
}
