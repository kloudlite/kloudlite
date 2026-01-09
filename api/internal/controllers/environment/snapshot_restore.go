package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// environmentsBasePath is where live environment data is stored (btrfs subvolumes)
	environmentsBasePath = "/var/lib/kloudlite/storage/environments"
	// envSnapshotsBasePath is where pulled snapshots are stored before creating live volumes
	envSnapshotsBasePath = "/var/lib/kloudlite/storage/.snapshots/envs"
)

// handleSnapshotRestore handles creating an environment from a pushed snapshot
// This function orchestrates the entire snapshot restoration workflow through various phases
func (r *EnvironmentReconciler) handleSnapshotRestore(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Handling environment snapshot restore",
		zap.String("environment", environment.Name),
		zap.String("snapshotName", environment.Spec.FromSnapshot.SnapshotName))

	// Initialize snapshot restore status if not set
	if environment.Status.SnapshotRestoreStatus == nil {
		logger.Info("Initializing snapshot restore status")
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
				Phase:          environmentsv1.SnapshotRestorePhasePending,
				Message:        "Snapshot restore initialized",
				SourceSnapshot: environment.Spec.FromSnapshot.SnapshotName,
				StartTime:      fn.Ptr(metav1.Now()),
			}
			return nil
		}, logger); err != nil {
			logger.Error("Failed to initialize snapshot restore status", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	status := environment.Status.SnapshotRestoreStatus

	// Handle based on current phase
	switch status.Phase {
	case environmentsv1.SnapshotRestorePhasePending:
		return r.handleRestorePending(ctx, environment, logger)

	case environmentsv1.SnapshotRestorePhasePulling:
		return r.handleRestorePulling(ctx, environment, logger)

	case environmentsv1.SnapshotRestorePhaseRestoring:
		return r.handleRestoreRestoring(ctx, environment, logger)

	case environmentsv1.SnapshotRestorePhaseDataRestoring:
		return r.handleRestoreDataRestoring(ctx, environment, logger)

	case environmentsv1.SnapshotRestorePhaseCompleted:
		return r.handleRestoreCompleted(ctx, environment, logger)

	case environmentsv1.SnapshotRestorePhaseFailed:
		// Restoration failed, don't retry automatically
		logger.Info("Snapshot restore failed, not retrying",
			zap.String("error", status.ErrorMessage))
		return reconcile.Result{}, nil

	default:
		logger.Error("Unknown snapshot restore phase", zap.String("phase", string(status.Phase)))
		return reconcile.Result{}, nil
	}
}

// handleRestorePending validates the snapshot exists and is pushed, then moves to Pulling phase
func (r *EnvironmentReconciler) handleRestorePending(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Pending - Validating snapshot")

	snapshotName := environment.Spec.FromSnapshot.SnapshotName

	// Fetch the snapshot to validate it exists and is pushed
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return r.failSnapshotRestore(ctx, environment, fmt.Sprintf("Snapshot %s not found", snapshotName), logger)
		}
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Validate snapshot is pushed to registry
	if snapshot.Status.RegistryStatus == nil || !snapshot.Status.RegistryStatus.Pushed {
		return r.failSnapshotRestore(ctx, environment,
			fmt.Sprintf("Snapshot %s is not pushed to registry", snapshotName), logger)
	}

	// Validate snapshot type is environment
	if snapshot.Status.SnapshotType != snapshotv1.SnapshotTypeEnvironment {
		return r.failSnapshotRestore(ctx, environment,
			fmt.Sprintf("Snapshot %s is not an environment snapshot", snapshotName), logger)
	}

	// Create namespace first if it doesn't exist
	if _, err := r.ensureNamespaceExists(ctx, environment, logger); err != nil {
		logger.Error("Failed to ensure namespace exists", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Store snapshot details and move to Pulling phase
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhasePulling
		environment.Status.SnapshotRestoreStatus.Message = "Pulling snapshot from registry"
		environment.Status.SnapshotRestoreStatus.ImageRef = snapshot.Status.RegistryStatus.ImageRef
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update snapshot restore status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot validated, moving to Pulling phase",
		zap.String("imageRef", snapshot.Status.RegistryStatus.ImageRef))
	return reconcile.Result{Requeue: true}, nil
}

// handleRestorePulling creates a SnapshotRequest to pull the snapshot from registry
func (r *EnvironmentReconciler) handleRestorePulling(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Pulling - Pulling snapshot from registry")

	status := environment.Status.SnapshotRestoreStatus
	snapshotName := environment.Spec.FromSnapshot.SnapshotName

	// Check for existing pull request
	pullReqName := fmt.Sprintf("%s-restore-pull", environment.Name)
	if status.SnapshotRequestName == "" {
		// Get the workmachine namespace
		wmNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)

		// Pull to .snapshots/envs/{envName}/ directory
		// Each environment has its own snapshot folder to avoid conflicts when forking
		// btrfs receive creates the subvolume INSIDE targetDir with the original snapshot name
		// So if we pull snap3 for env main-fork, it will be created at .snapshots/envs/main-fork/snap3/
		snapshotPath := filepath.Join(envSnapshotsBasePath, environment.Name)

		// Parse imageRef to get repository and tag
		repository, tag := parseImageRef(status.ImageRef)

		pullReq := &snapshotv1.SnapshotRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pullReqName,
				Namespace: wmNamespace,
				Labels: map[string]string{
					"environments.kloudlite.io/environment": environment.Name,
					"environments.kloudlite.io/operation":   "restore-pull",
				},
			},
			Spec: snapshotv1.SnapshotRequestSpec{
				Operation:       snapshotv1.SnapshotOperationPull,
				SnapshotPath:    snapshotPath,
				SnapshotRef:     snapshotName,
				EnvironmentName: environment.Name,
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
				return r.failSnapshotRestore(ctx, environment,
					fmt.Sprintf("Failed to create pull request: %v", err), logger)
			}
		}

		// Update status with SnapshotRequest name
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.SnapshotRestoreStatus.SnapshotRequestName = pullReqName
			environment.Status.SnapshotRestoreStatus.Message = "Pull request created, waiting for completion"
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status with pull request name", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Created pull SnapshotRequest", zap.String("name", pullReqName))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Check status of existing pull request
	wmNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)
	pullReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: status.SnapshotRequestName, Namespace: wmNamespace}, pullReq); err != nil {
		if apierrors.IsNotFound(err) {
			// Request was deleted, reset and retry
			logger.Warn("Pull request not found, resetting")
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
				environment.Status.SnapshotRestoreStatus.SnapshotRequestName = ""
				return nil
			}, logger); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to get pull request", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	switch pullReq.Status.Phase {
	case snapshotv1.SnapshotRequestPhaseFailed:
		return r.failSnapshotRestore(ctx, environment,
			fmt.Sprintf("Pull failed: %s", pullReq.Status.Message), logger)

	case snapshotv1.SnapshotRequestPhaseCompleted:
		// Pull completed - now create the live btrfs subvolume from the pulled snapshot
		// Source: .snapshots/envs/{envName}/{snapshotName}/ (the pulled snapshot)
		// Target: environments/{targetNamespace}/ (the live environment)
		snapshotSourcePath := filepath.Join(envSnapshotsBasePath, environment.Name, snapshotName)
		liveEnvPath := filepath.Join(environmentsBasePath, environment.Spec.TargetNamespace)

		// Create a SnapshotRequest to create the live subvolume
		createLiveReqName := fmt.Sprintf("%s-create-live", environment.Name)
		createLiveReq := &snapshotv1.SnapshotRequest{}
		if err := r.Get(ctx, client.ObjectKey{Name: createLiveReqName, Namespace: wmNamespace}, createLiveReq); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to get create-live request", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			}

			// Create the request to make a btrfs subvolume snapshot
			newReq := &snapshotv1.SnapshotRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      createLiveReqName,
					Namespace: wmNamespace,
					Labels: map[string]string{
						"environments.kloudlite.io/environment": environment.Name,
						"environments.kloudlite.io/operation":   "create-live",
					},
				},
				Spec: snapshotv1.SnapshotRequestSpec{
					Operation:       snapshotv1.SnapshotOperationRestore,
					SourcePath:      snapshotSourcePath,
					SnapshotPath:    liveEnvPath,
					EnvironmentName: environment.Name,
				},
			}

			if err := r.Create(ctx, newReq); err != nil && !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create live volume request", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			}

			logger.Info("Created request to create live environment volume",
				zap.String("source", snapshotSourcePath),
				zap.String("target", liveEnvPath))
			return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
		}

		// Check if the live volume creation is complete
		switch createLiveReq.Status.Phase {
		case snapshotv1.SnapshotRequestPhaseFailed:
			return r.failSnapshotRestore(ctx, environment,
				fmt.Sprintf("Failed to create live volume: %s", createLiveReq.Status.Message), logger)
		case snapshotv1.SnapshotRequestPhaseCompleted:
			// Clean up the create-live request
			if err := r.Delete(ctx, createLiveReq); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete create-live request", zap.Error(err))
			}

			// Move to Restoring phase to apply K8s resources
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
				environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseRestoring
				environment.Status.SnapshotRestoreStatus.Message = "Live volume created, restoring K8s resources"
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update status after live volume creation", zap.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("Live volume created, moving to Restoring phase")
			return reconcile.Result{Requeue: true}, nil
		default:
			logger.Info("Waiting for live volume creation", zap.String("phase", string(createLiveReq.Status.Phase)))
			return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
		}

	default:
		// Still in progress
		logger.Info("Pull in progress", zap.String("phase", string(pullReq.Status.Phase)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
}

// handleRestoreRestoring applies resources from snapshot metadata
func (r *EnvironmentReconciler) handleRestoreRestoring(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Restoring - Applying K8s resources from snapshot")

	targetNamespace := environment.Spec.TargetNamespace
	status := environment.Status.SnapshotRestoreStatus

	// Fetch the pull request to get the pulled metadata
	pullReqName := status.SnapshotRequestName
	wmNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)

	pullReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: pullReqName, Namespace: wmNamespace}, pullReq); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Warn("Pull request not found, cannot restore metadata")
		} else {
			logger.Error("Failed to get pull request", zap.Error(err))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Restore resources from the pulled metadata (stored in OCI layer's metadata.json)
	if pullReq.Status.PulledMetadata != nil {
		if err := r.restoreResourcesFromMetadata(ctx, pullReq.Status.PulledMetadata, targetNamespace, environment, logger); err != nil {
			logger.Warn("Failed to restore some resources", zap.Error(err))
			// Don't fail the entire restore if some resource restoration fails
		}
	} else {
		logger.Info("No K8s resource metadata found in pulled snapshot")
	}

	// Clean up the pull request - no longer needed
	if err := r.Delete(ctx, pullReq); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete pull request", zap.Error(err))
	}

	// Data is already in place (live subvolume created from snapshot)
	// Skip DataRestoring and go directly to Completed
	return r.moveToCompleted(ctx, environment, logger)
}

// handleRestoreDataRestoring moves PVC data from snapshot subdirectory to the correct location
// After pull, data structure is:
//
//	{namespace}/{snapshotName}/{claimName}/  (data from snapshot)
//
// PVC expects data at:
//
//	{namespace}/{claimName}/
//
// We need to move {snapshotName}/{claimName}/ contents to {claimName}/ and delete {snapshotName}/
func (r *EnvironmentReconciler) handleRestoreDataRestoring(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: DataRestoring - Moving PVC data to correct paths")

	targetNamespace := environment.Spec.TargetNamespace
	status := environment.Status.SnapshotRestoreStatus
	wmNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)
	snapshotName := status.SourceSnapshot

	// Get the pull request for cleanup later
	pullReq := &snapshotv1.SnapshotRequest{}
	pullReqExists := true
	if err := r.Get(ctx, client.ObjectKey{Name: status.SnapshotRequestName, Namespace: wmNamespace}, pullReq); err != nil {
		if apierrors.IsNotFound(err) {
			pullReqExists = false
		} else {
			logger.Error("Failed to get pull request", zap.Error(err))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// List PVCs in the target namespace
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, client.InNamespace(targetNamespace)); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// List StatefulSets to check if PVCs should exist
	stsList := &appsv1.StatefulSetList{}
	if err := r.List(ctx, stsList, client.InNamespace(targetNamespace)); err != nil {
		logger.Error("Failed to list StatefulSets", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// List Compositions to check if they might create PVCs
	compList := &environmentsv1.CompositionList{}
	if err := r.List(ctx, compList, client.InNamespace(targetNamespace)); err != nil {
		logger.Error("Failed to list Compositions", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Check if any compositions might need PVCs (have volumes defined or are processing)
	compositionsWithVolumes := 0
	for _, comp := range compList.Items {
		// Check if composition defines volumes in its Docker Compose content
		// Compositions use Deployments (not StatefulSets) and create PVCs independently
		if comp.Spec.ComposeContent != "" {
			compositionsWithVolumes++
		}
	}

	// Also check if the snapshot metadata had PVCs that should be restored
	snapshotHadPVCs := false
	if pullReqExists && pullReq.Status.PulledMetadata != nil && pullReq.Status.PulledMetadata.PVCs != "" {
		snapshotHadPVCs = true
		logger.Info("Snapshot metadata contains PVCs to restore")
	}

	// If no PVCs, no StatefulSets, no compositions with volumes, and snapshot had no PVCs
	// then we can safely skip data restoration
	if len(pvcList.Items) == 0 && len(stsList.Items) == 0 && compositionsWithVolumes == 0 && !snapshotHadPVCs {
		logger.Info("No PVCs, StatefulSets, Compositions with volumes, or snapshot PVCs found - skipping data restore")
		r.cleanupSnapshotSubdirectory(ctx, environment, snapshotName, wmNamespace, logger)
		if pullReqExists {
			if err := r.Delete(ctx, pullReq); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete pull request", zap.Error(err))
			}
		}
		return r.moveToCompleted(ctx, environment, logger)
	}

	// If no PVCs exist yet but we expect them (from compositions or snapshot), wait
	if len(pvcList.Items) == 0 {
		logger.Info("Waiting for PVCs to be created",
			zap.Int("compositions", compositionsWithVolumes),
			zap.Bool("snapshotHadPVCs", snapshotHadPVCs))
		return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
	}

	anyPending := false
	allDataMoved := true

	for _, pvc := range pvcList.Items {
		// Handle PVCs that aren't bound yet (WaitForFirstConsumer)
		if pvc.Status.Phase != corev1.ClaimBound {
			logger.Info("PVC not bound yet", zap.String("pvc", pvc.Name), zap.String("phase", string(pvc.Status.Phase)))
			anyPending = true
			allDataMoved = false
			continue
		}

		// PVC is bound - check if we need to move data
		claimName := pvc.Name

		// Create a SnapshotRequest to move data from snapshot subdirectory
		moveReqName := fmt.Sprintf("%s-data-move-%s", environment.Name, pvc.Name)

		moveReq := &snapshotv1.SnapshotRequest{}
		if err := r.Get(ctx, client.ObjectKey{Name: moveReqName, Namespace: wmNamespace}, moveReq); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to get move request", zap.Error(err), zap.String("pvc", pvc.Name))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			}

			// Create move request
			// Source: {namespace}/{snapshotName}/{claimName}/
			// Target: {namespace}/{claimName}/
			namespacePath := filepath.Join(environmentsBasePath, targetNamespace)
			sourcePath := filepath.Join(namespacePath, snapshotName, claimName)
			targetPath := filepath.Join(namespacePath, claimName)

			newMoveReq := &snapshotv1.SnapshotRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      moveReqName,
					Namespace: wmNamespace,
					Labels: map[string]string{
						"environments.kloudlite.io/environment": environment.Name,
						"environments.kloudlite.io/operation":   "data-move",
						"environments.kloudlite.io/pvc":         pvc.Name,
					},
				},
				Spec: snapshotv1.SnapshotRequestSpec{
					Operation:       snapshotv1.SnapshotOperationRestore,
					SourcePath:      sourcePath,
					SnapshotPath:    targetPath,
					EnvironmentName: environment.Name,
				},
			}

			if err := r.Create(ctx, newMoveReq); err != nil {
				if !apierrors.IsAlreadyExists(err) {
					logger.Error("Failed to create move request", zap.Error(err), zap.String("pvc", pvc.Name))
					return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
				}
			}

			logger.Info("Created data move request",
				zap.String("pvc", pvc.Name),
				zap.String("source", sourcePath),
				zap.String("target", targetPath))
			allDataMoved = false
			continue
		}

		// Check move request status
		switch moveReq.Status.Phase {
		case snapshotv1.SnapshotRequestPhaseCompleted:
			logger.Info("PVC data move completed", zap.String("pvc", pvc.Name))
			// Clean up completed move request
			if err := r.Delete(ctx, moveReq); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete completed move request", zap.Error(err))
			}
		case snapshotv1.SnapshotRequestPhaseFailed:
			logger.Warn("PVC data move failed", zap.String("pvc", pvc.Name), zap.String("error", moveReq.Status.Message))
			// Data might already be in the right place, proceed anyway
			if err := r.Delete(ctx, moveReq); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete failed move request", zap.Error(err))
			}
		default:
			logger.Info("PVC data move in progress", zap.String("pvc", pvc.Name), zap.String("phase", string(moveReq.Status.Phase)))
			allDataMoved = false
		}
	}

	// Handle pending PVCs - create helper pod to trigger WaitForFirstConsumer binding
	if anyPending {
		if err := r.ensurePVCBindingHelperPod(ctx, environment, pvcList.Items, logger); err != nil {
			logger.Error("Failed to ensure PVC binding helper pod", zap.Error(err))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// All PVCs are bound, clean up helper pod if it exists
	r.cleanupPVCBindingHelperPod(ctx, environment, logger)

	if !allDataMoved {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// All data moved, clean up snapshot subdirectory and pull request
	r.cleanupSnapshotSubdirectory(ctx, environment, snapshotName, wmNamespace, logger)
	if pullReqExists {
		if err := r.Delete(ctx, pullReq); err != nil && !apierrors.IsNotFound(err) {
			logger.Warn("Failed to delete pull request", zap.Error(err))
		}
	}

	return r.moveToCompleted(ctx, environment, logger)
}

// cleanupSnapshotSubdirectory removes the snapshot subdirectory after data has been moved
func (r *EnvironmentReconciler) cleanupSnapshotSubdirectory(
	ctx context.Context,
	environment *environmentsv1.Environment,
	snapshotName, wmNamespace string,
	logger *zap.Logger,
) {
	targetNamespace := environment.Spec.TargetNamespace
	snapshotSubdir := filepath.Join(environmentsBasePath, targetNamespace, snapshotName)

	// Create a delete SnapshotRequest to clean up the snapshot subdirectory
	cleanupReqName := fmt.Sprintf("%s-cleanup", environment.Name)

	cleanupReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cleanupReqName,
			Namespace: wmNamespace,
			Labels: map[string]string{
				"environments.kloudlite.io/environment": environment.Name,
				"environments.kloudlite.io/operation":   "cleanup",
			},
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			Operation:       snapshotv1.SnapshotOperationDelete,
			SnapshotPath:    snapshotSubdir,
			EnvironmentName: environment.Name,
		},
	}

	if err := r.Create(ctx, cleanupReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Warn("Failed to create cleanup request", zap.Error(err))
		}
	} else {
		logger.Info("Created cleanup request for snapshot subdirectory", zap.String("path", snapshotSubdir))
	}
}

// getPVCRestorePaths determines the source (snapshot) and target (actual PVC) paths
func (r *EnvironmentReconciler) getPVCRestorePaths(
	pulledSnapshotBase, targetNamespace string,
	pvc *corev1.PersistentVolumeClaim,
	logger *zap.Logger,
) (sourcePath, targetPath string) {
	claimName := pvc.Name
	pvName := pvc.Spec.VolumeName // Get PV name from bound PVC

	environmentsBasePath := "/var/lib/kloudlite/storage/environments"

	// New pathPattern format: {namespace}/{claim-name}/{pv-name}
	// This matches the StorageClass pathPattern: "{{ .PVC.Namespace }}/{{ .PVC.Name }}/{{ .PVName }}"
	targetPath = filepath.Join(environmentsBasePath, targetNamespace, claimName, pvName)

	// Source path - find the actual data directory inside the snapshot
	// Snapshot structure (new pathPattern format):
	//   {snapshotBase}/{snapshotName}/{claimName}/{old-pv-name}/(data files)
	// We need to copy CONTENTS of {claimName}/{old-pv-name}/* to target
	// Use glob pattern to find the claim directory, then the PV subdirectory inside it
	// The snapshotrequest controller will resolve this and copy the innermost data
	sourcePath = filepath.Join(pulledSnapshotBase, claimName, "*")

	logger.Info("Determined PVC restore paths",
		zap.String("pvc", pvc.Name),
		zap.String("pvName", pvName),
		zap.String("sourcePattern", sourcePath),
		zap.String("target", targetPath))

	return sourcePath, targetPath
}

// moveToCompleted updates status to Completed phase
func (r *EnvironmentReconciler) moveToCompleted(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseCompleted
		environment.Status.SnapshotRestoreStatus.Message = "Snapshot restore completed"
		environment.Status.SnapshotRestoreStatus.CompletionTime = &now

		environment.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
			Name:       environment.Spec.FromSnapshot.SnapshotName,
			RestoredAt: now,
		}
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Moving to Completed phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleRestoreCompleted clears the fromSnapshot field and proceeds to normal reconciliation
func (r *EnvironmentReconciler) handleRestoreCompleted(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: Completed - Clearing fromSnapshot and proceeding to normal reconciliation")

	sourceSnapshotName := environment.Status.SnapshotRestoreStatus.SourceSnapshot

	// Auto-create snapshots on the forked environment to maintain lineage (like git fork)
	leafSnapshotName, err := r.forkSnapshotLineage(ctx, environment, sourceSnapshotName, logger)
	if err != nil {
		logger.Warn("Failed to fork snapshot lineage to new environment", zap.Error(err))
		// Don't fail the restore - this is best-effort
	}

	// Set LastRestoredSnapshot to the leaf snapshot (HEAD of the forked lineage)
	if leafSnapshotName != "" {
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
				Name:       leafSnapshotName,
				RestoredAt: metav1.Now(),
			}
			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update LastRestoredSnapshot status", zap.Error(err))
		}
	}

	// Clear fromSnapshot to mark restoration as complete
	environment.Spec.FromSnapshot = nil

	// Add completion condition
	environment.Status.Conditions = append(environment.Status.Conditions, environmentsv1.EnvironmentCondition{
		Type:               "RestoredFromSnapshot",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: fn.Ptr(metav1.Now()),
		Reason:             "RestoreCompleted",
		Message:            fmt.Sprintf("Successfully restored from snapshot %s", sourceSnapshotName),
	})

	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear fromSnapshot field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot restore completed, proceeding to normal environment reconciliation")

	// Requeue to start normal environment reconciliation
	return reconcile.Result{Requeue: true}, nil
}

// forkSnapshotLineage copies the entire snapshot lineage to the forked environment (like git fork)
// Only the direct lineage of the source snapshot is copied, not other branches
// Returns the name of the leaf snapshot (the last one in the lineage) for setting as HEAD
func (r *EnvironmentReconciler) forkSnapshotLineage(
	ctx context.Context,
	environment *environmentsv1.Environment,
	sourceSnapshotName string,
	logger *zap.Logger,
) (string, error) {
	// Build the lineage from root to source (walk up, then reverse)
	lineage, err := r.buildSnapshotLineage(ctx, sourceSnapshotName, logger)
	if err != nil {
		return "", fmt.Errorf("failed to build snapshot lineage: %w", err)
	}

	if len(lineage) == 0 {
		return "", fmt.Errorf("no snapshots found in lineage for %s", sourceSnapshotName)
	}

	logger.Info("Forking snapshot lineage to new environment",
		zap.String("environment", environment.Name),
		zap.Int("snapshotCount", len(lineage)))

	// Create copies of each snapshot, linking them together
	var previousCopyName string
	for i, originalSnapshot := range lineage {
		// Generate name for the forked snapshot
		forkName := fmt.Sprintf("%s-%s", environment.Name, originalSnapshot.Name)

		// Check if already exists (idempotent)
		existing := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: forkName}, existing); err == nil {
			logger.Info("Forked snapshot already exists, skipping",
				zap.String("fork", forkName))
			previousCopyName = forkName
			continue
		}

		// Build parent reference - point to previous fork in chain
		var parentRef *snapshotv1.ParentSnapshotReference
		if previousCopyName != "" {
			parentRef = &snapshotv1.ParentSnapshotReference{
				Name: previousCopyName,
			}
		}

		// Create the forked snapshot
		// Copy registry status from original - no need to re-push, data is already in registry
		var registryStatus *snapshotv1.SnapshotRegistryStatus
		if originalSnapshot.Status.RegistryStatus != nil {
			registryStatus = originalSnapshot.Status.RegistryStatus.DeepCopy()
		}

		var registryRef *snapshotv1.SnapshotRegistryRef
		if originalSnapshot.Spec.RegistryRef != nil {
			registryRef = originalSnapshot.Spec.RegistryRef.DeepCopy()
		}

		forkedSnapshot := &snapshotv1.Snapshot{
			ObjectMeta: metav1.ObjectMeta{
				Name: forkName,
				Labels: map[string]string{
					"snapshots.kloudlite.io/environment": environment.Name,
					"kloudlite.io/owned-by":              environment.Spec.OwnedBy,
					"snapshots.kloudlite.io/forked-from": originalSnapshot.Name,
					"snapshots.kloudlite.io/fork-of-env": originalSnapshot.Labels["snapshots.kloudlite.io/environment"],
				},
			},
			Spec: snapshotv1.SnapshotSpec{
				EnvironmentRef: &snapshotv1.EnvironmentReference{
					Name: environment.Name,
				},
				ParentSnapshotRef: parentRef,
				Description:       originalSnapshot.Spec.Description,
				OwnedBy:           environment.Spec.OwnedBy,
				IncludeMetadata:   originalSnapshot.Spec.IncludeMetadata,
				// Copy registry ref - forked snapshot uses same image
				RegistryRef: registryRef,
			},
			Status: snapshotv1.SnapshotStatus{
				// Copy status from original - these are pre-existing snapshots
				State:             snapshotv1.SnapshotStateReady,
				SnapshotType:      originalSnapshot.Status.SnapshotType,
				TargetName:        environment.Name,
				Message:           fmt.Sprintf("Forked from %s", originalSnapshot.Name),
				SizeBytes:         originalSnapshot.Status.SizeBytes,
				SizeHuman:         originalSnapshot.Status.SizeHuman,
				SnapshotPath:      originalSnapshot.Status.SnapshotPath, // Points to same btrfs snapshot
				CreatedAt:         originalSnapshot.Status.CreatedAt,
				WorkMachineName:   originalSnapshot.Status.WorkMachineName,
				ResourceMetadata:  originalSnapshot.Status.ResourceMetadata,
				CollectedMetadata: originalSnapshot.Status.CollectedMetadata,
				RegistryStatus:    registryStatus, // Copy registry status - already pushed
			},
		}

		// Set parent label for lineage tracking
		if parentRef != nil {
			forkedSnapshot.Labels["snapshots.kloudlite.io/parent"] = parentRef.Name
		}

		if err := r.Create(ctx, forkedSnapshot); err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Info("Forked snapshot already exists", zap.String("fork", forkName))
				previousCopyName = forkName
				continue
			}
			return "", fmt.Errorf("failed to create forked snapshot %s: %w", forkName, err)
		}

		// Status subresource requires separate update - status is ignored on Create
		// Must set status to Ready to prevent normal reconciliation from auto-detecting wrong parent
		if err := r.Status().Update(ctx, forkedSnapshot); err != nil {
			logger.Warn("Failed to update forked snapshot status", zap.String("fork", forkName), zap.Error(err))
			// Continue anyway - the snapshot was created
		}

		// Add a new tag to the existing image for the forked environment
		// This avoids re-pushing the same data - just copies the manifest to the new repo
		if originalSnapshot.Status.RegistryStatus != nil && originalSnapshot.Status.RegistryStatus.Pushed {
			newRepository := fmt.Sprintf("snapshots/%s/env/%s", environment.Spec.OwnedBy, environment.Name)
			newTag := forkName

			// Get source repository from original snapshot
			sourceRepository := ""
			if originalSnapshot.Spec.RegistryRef != nil {
				sourceRepository = originalSnapshot.Spec.RegistryRef.Repository
			}

			tagReq := &snapshotv1.SnapshotRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("tag-%s", forkName),
					Namespace: fmt.Sprintf("wm-%s", environment.Spec.OwnedBy),
					Labels: map[string]string{
						"snapshots.kloudlite.io/snapshot":  forkName,
						"snapshots.kloudlite.io/operation": "tag",
					},
				},
				Spec: snapshotv1.SnapshotRequestSpec{
					Operation:    snapshotv1.SnapshotOperationTag,
					SnapshotPath: originalSnapshot.Status.SnapshotPath,
					SnapshotRef:  forkName,
					RegistryRef: &snapshotv1.SnapshotRequestRegistryRef{
						RegistryURL:      "image-registry:5000",
						Repository:       newRepository,
						Tag:              newTag,
						SourceTag:        originalSnapshot.Status.RegistryStatus.Tag,
						SourceRepository: sourceRepository,
					},
				},
			}

			if err := r.Create(ctx, tagReq); err != nil && !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create tag request for forked snapshot",
					zap.String("fork", forkName), zap.Error(err))
			} else {
				logger.Info("Created tag request for forked snapshot",
					zap.String("fork", forkName),
					zap.String("sourceRepository", sourceRepository),
					zap.String("newRepository", newRepository),
					zap.String("newTag", newTag))
			}
		}

		logger.Info("Created forked snapshot",
			zap.String("fork", forkName),
			zap.String("original", originalSnapshot.Name),
			zap.Int("index", i+1),
			zap.Int("total", len(lineage)))

		previousCopyName = forkName
	}

	logger.Info("Successfully forked snapshot lineage",
		zap.String("environment", environment.Name),
		zap.Int("snapshotsForked", len(lineage)),
		zap.String("leafSnapshot", previousCopyName))

	return previousCopyName, nil
}

// buildSnapshotLineage walks up from a snapshot to root and returns lineage from root to leaf
func (r *EnvironmentReconciler) buildSnapshotLineage(
	ctx context.Context,
	snapshotName string,
	logger *zap.Logger,
) ([]*snapshotv1.Snapshot, error) {
	var lineage []*snapshotv1.Snapshot

	currentName := snapshotName
	for currentName != "" {
		snapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: currentName}, snapshot); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Warn("Snapshot not found while building lineage", zap.String("snapshot", currentName))
				break
			}
			return nil, fmt.Errorf("failed to get snapshot %s: %w", currentName, err)
		}

		// Prepend to build root-to-leaf order
		lineage = append([]*snapshotv1.Snapshot{snapshot}, lineage...)

		// Move to parent
		if snapshot.Spec.ParentSnapshotRef == nil {
			break // Reached root
		}
		currentName = snapshot.Spec.ParentSnapshotRef.Name
	}

	return lineage, nil
}

// restoreResourcesFromMetadata restores K8s resources from the metadata struct (from OCI layer)
func (r *EnvironmentReconciler) restoreResourcesFromMetadata(
	ctx context.Context,
	metadata *snapshotv1.SnapshotMetadata,
	targetNamespace string,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) error {
	logger.Info("Restoring resources from OCI metadata")
	envName := environment.Name

	// Restore PVCs first with selected-node annotation for immediate binding
	// This must happen before compositions so data can be restored to the PVCs
	if metadata.PVCs != "" {
		if err := r.restorePVCsFromJSON(ctx, metadata.PVCs, targetNamespace, environment, logger); err != nil {
			logger.Warn("Failed to restore PVCs", zap.Error(err))
		}
	}

	// Restore ConfigMaps
	if metadata.ConfigMaps != "" {
		if err := r.restoreConfigMapsFromJSON(ctx, metadata.ConfigMaps, targetNamespace, envName, logger); err != nil {
			logger.Warn("Failed to restore ConfigMaps", zap.Error(err))
		}
	}

	// Restore Secrets
	if metadata.Secrets != "" {
		if err := r.restoreSecretsFromJSON(ctx, metadata.Secrets, targetNamespace, envName, logger); err != nil {
			logger.Warn("Failed to restore Secrets", zap.Error(err))
		}
	}

	// Restore Deployments
	if metadata.Deployments != "" {
		if err := r.restoreDeploymentsFromJSON(ctx, metadata.Deployments, targetNamespace, logger); err != nil {
			logger.Warn("Failed to restore Deployments", zap.Error(err))
		}
	}

	// Restore StatefulSets
	if metadata.StatefulSets != "" {
		if err := r.restoreStatefulSetsFromJSON(ctx, metadata.StatefulSets, targetNamespace, logger); err != nil {
			logger.Warn("Failed to restore StatefulSets", zap.Error(err))
		}
	}

	// Restore Services
	if metadata.Services != "" {
		if err := r.restoreServicesFromJSON(ctx, metadata.Services, targetNamespace, logger); err != nil {
			logger.Warn("Failed to restore Services", zap.Error(err))
		}
	}

	// Restore Compositions
	if metadata.Compositions != "" {
		if err := r.restoreCompositionsFromJSON(ctx, metadata.Compositions, targetNamespace, logger); err != nil {
			logger.Warn("Failed to restore Compositions", zap.Error(err))
		}
	}

	return nil
}

// restoreConfigMapsFromJSON restores ConfigMaps from JSON string
func (r *EnvironmentReconciler) restoreConfigMapsFromJSON(
	ctx context.Context,
	jsonData, targetNamespace, envName string,
	logger *zap.Logger,
) error {
	var configMapList corev1.ConfigMapList
	if err := json.Unmarshal([]byte(jsonData), &configMapList); err != nil {
		return fmt.Errorf("failed to parse configmaps JSON: %w", err)
	}

	restored := 0
	for _, cm := range configMapList.Items {
		// Skip system ConfigMaps
		if cm.Name == "kube-root-ca.crt" {
			continue
		}

		newCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        cm.Name,
				Namespace:   targetNamespace,
				Labels:      cm.Labels,
				Annotations: cm.Annotations,
			},
			Data:       cm.Data,
			BinaryData: cm.BinaryData,
		}

		// Update environment label
		if newCM.Labels == nil {
			newCM.Labels = make(map[string]string)
		}
		newCM.Labels["kloudlite.io/environment"] = envName

		if err := r.Create(ctx, newCM); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create ConfigMap", zap.String("name", cm.Name), zap.Error(err))
				continue
			}
		}
		restored++
	}

	logger.Info("Restored ConfigMaps", zap.Int("count", restored))
	return nil
}

// restoreSecretsFromJSON restores Secrets from JSON string
func (r *EnvironmentReconciler) restoreSecretsFromJSON(
	ctx context.Context,
	jsonData, targetNamespace, envName string,
	logger *zap.Logger,
) error {
	var secrets []corev1.Secret
	if err := json.Unmarshal([]byte(jsonData), &secrets); err != nil {
		return fmt.Errorf("failed to parse secrets JSON: %w", err)
	}

	restored := 0
	for _, secret := range secrets {
		// Skip service account tokens and system secrets
		if secret.Type == corev1.SecretTypeServiceAccountToken {
			continue
		}

		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.Name,
				Namespace:   targetNamespace,
				Labels:      secret.Labels,
				Annotations: secret.Annotations,
			},
			Type: secret.Type,
			Data: secret.Data,
		}

		// Update environment label
		if newSecret.Labels == nil {
			newSecret.Labels = make(map[string]string)
		}
		newSecret.Labels["kloudlite.io/environment"] = envName

		if err := r.Create(ctx, newSecret); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create Secret", zap.String("name", secret.Name), zap.Error(err))
				continue
			}
		}
		restored++
	}

	logger.Info("Restored Secrets", zap.Int("count", restored))
	return nil
}

// restoreDeploymentsFromJSON restores Deployments from JSON string
func (r *EnvironmentReconciler) restoreDeploymentsFromJSON(
	ctx context.Context,
	jsonData, targetNamespace string,
	logger *zap.Logger,
) error {
	var deploymentList appsv1.DeploymentList
	if err := json.Unmarshal([]byte(jsonData), &deploymentList); err != nil {
		return fmt.Errorf("failed to parse deployments JSON: %w", err)
	}

	restored := 0
	for _, deploy := range deploymentList.Items {
		newDeploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        deploy.Name,
				Namespace:   targetNamespace,
				Labels:      deploy.Labels,
				Annotations: deploy.Annotations,
			},
			Spec: deploy.Spec,
		}

		// Clear resource version and UID
		newDeploy.Spec.Template.ObjectMeta.ResourceVersion = ""
		newDeploy.Spec.Template.ObjectMeta.UID = ""

		if err := r.Create(ctx, newDeploy); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create Deployment", zap.String("name", deploy.Name), zap.Error(err))
				continue
			}
		}
		restored++
	}

	logger.Info("Restored Deployments", zap.Int("count", restored))
	return nil
}

// restoreStatefulSetsFromJSON restores StatefulSets from JSON string
func (r *EnvironmentReconciler) restoreStatefulSetsFromJSON(
	ctx context.Context,
	jsonData, targetNamespace string,
	logger *zap.Logger,
) error {
	var stsList appsv1.StatefulSetList
	if err := json.Unmarshal([]byte(jsonData), &stsList); err != nil {
		return fmt.Errorf("failed to parse statefulsets JSON: %w", err)
	}

	restored := 0
	for _, sts := range stsList.Items {
		newSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        sts.Name,
				Namespace:   targetNamespace,
				Labels:      sts.Labels,
				Annotations: sts.Annotations,
			},
			Spec: sts.Spec,
		}

		// Clear resource version and UID
		newSts.Spec.Template.ObjectMeta.ResourceVersion = ""
		newSts.Spec.Template.ObjectMeta.UID = ""

		if err := r.Create(ctx, newSts); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create StatefulSet", zap.String("name", sts.Name), zap.Error(err))
				continue
			}
		}
		restored++
	}

	logger.Info("Restored StatefulSets", zap.Int("count", restored))
	return nil
}

// restoreServicesFromJSON restores Services from JSON string
func (r *EnvironmentReconciler) restoreServicesFromJSON(
	ctx context.Context,
	jsonData, targetNamespace string,
	logger *zap.Logger,
) error {
	var svcList corev1.ServiceList
	if err := json.Unmarshal([]byte(jsonData), &svcList); err != nil {
		return fmt.Errorf("failed to parse services JSON: %w", err)
	}

	restored := 0
	for _, svc := range svcList.Items {
		// Skip Kubernetes service
		if svc.Name == "kubernetes" {
			continue
		}

		newSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        svc.Name,
				Namespace:   targetNamespace,
				Labels:      svc.Labels,
				Annotations: svc.Annotations,
			},
			Spec: corev1.ServiceSpec{
				Ports:    svc.Spec.Ports,
				Selector: svc.Spec.Selector,
				Type:     svc.Spec.Type,
			},
		}

		// Clear ClusterIP to let Kubernetes assign a new one
		newSvc.Spec.ClusterIP = ""
		newSvc.Spec.ClusterIPs = nil

		if err := r.Create(ctx, newSvc); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create Service", zap.String("name", svc.Name), zap.Error(err))
				continue
			}
		}
		restored++
	}

	logger.Info("Restored Services", zap.Int("count", restored))
	return nil
}

// restoreCompositionsFromJSON restores Compositions from JSON string
func (r *EnvironmentReconciler) restoreCompositionsFromJSON(
	ctx context.Context,
	jsonData, targetNamespace string,
	logger *zap.Logger,
) error {
	var compositionList environmentsv1.CompositionList
	if err := json.Unmarshal([]byte(jsonData), &compositionList); err != nil {
		return fmt.Errorf("failed to parse compositions JSON: %w", err)
	}

	restored := 0
	for _, comp := range compositionList.Items {
		newComp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:        comp.Name,
				Namespace:   targetNamespace,
				Labels:      comp.Labels,
				Annotations: comp.Annotations,
			},
			Spec: comp.Spec,
		}

		if err := r.Create(ctx, newComp); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create Composition", zap.String("name", comp.Name), zap.Error(err))
				continue
			}
		}
		restored++
	}

	logger.Info("Restored Compositions", zap.Int("count", restored))
	return nil
}

// restorePVCsFromJSON restores PVCs from JSON string with selected-node annotation
func (r *EnvironmentReconciler) restorePVCsFromJSON(
	ctx context.Context,
	jsonData, targetNamespace string,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) error {
	var pvcList corev1.PersistentVolumeClaimList
	if err := json.Unmarshal([]byte(jsonData), &pvcList); err != nil {
		return fmt.Errorf("failed to parse PVCs JSON: %w", err)
	}

	nodeName := environment.Spec.WorkMachineName
	restored := 0

	for _, pvc := range pvcList.Items {
		// Create new PVC with selected-node annotation for immediate binding
		annotations := make(map[string]string)
		if nodeName != "" {
			annotations["volume.kubernetes.io/selected-node"] = nodeName
		}
		// Copy existing annotations (except binding-related ones)
		for k, v := range pvc.Annotations {
			if k != "pv.kubernetes.io/bind-completed" &&
				k != "pv.kubernetes.io/bound-by-controller" &&
				k != "volume.beta.kubernetes.io/storage-provisioner" &&
				k != "volume.kubernetes.io/storage-provisioner" {
				annotations[k] = v
			}
		}

		newPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        pvc.Name,
				Namespace:   targetNamespace,
				Labels:      pvc.Labels,
				Annotations: annotations,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes:      pvc.Spec.AccessModes,
				Resources:        pvc.Spec.Resources,
				StorageClassName: pvc.Spec.StorageClassName,
				VolumeMode:       pvc.Spec.VolumeMode,
				// Don't copy VolumeName - let the provisioner create a new PV
			},
		}

		if err := r.Create(ctx, newPVC); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				logger.Warn("Failed to create PVC", zap.String("name", pvc.Name), zap.Error(err))
				continue
			}
		}
		restored++
		logger.Info("Restored PVC with selected-node annotation",
			zap.String("pvc", pvc.Name),
			zap.String("node", nodeName))
	}

	logger.Info("Restored PVCs", zap.Int("count", restored))
	return nil
}

// failSnapshotRestore sets the restore phase to failed with an error message
func (r *EnvironmentReconciler) failSnapshotRestore(
	ctx context.Context,
	environment *environmentsv1.Environment,
	errorMessage string,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Snapshot restore failed", zap.String("error", errorMessage))

	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseFailed
		environment.Status.SnapshotRestoreStatus.ErrorMessage = errorMessage
		environment.Status.SnapshotRestoreStatus.CompletionTime = &now
		return nil
	}, logger); err != nil {
		return reconcile.Result{}, err
	}

	// Add failure condition
	environment.Status.Conditions = append(environment.Status.Conditions, environmentsv1.EnvironmentCondition{
		Type:               "RestoredFromSnapshot",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: fn.Ptr(metav1.Now()),
		Reason:             "RestoreFailed",
		Message:            errorMessage,
	})

	if err := r.Status().Update(ctx, environment); err != nil {
		logger.Error("Failed to update status with failure condition", zap.Error(err))
	}

	// Update environment state to error
	if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError, errorMessage, logger); err != nil {
		logger.Error("Failed to update environment status to error", zap.Error(err))
	}

	return reconcile.Result{}, nil
}

// ensurePVCBindingHelperPod creates a temporary pod that mounts all pending PVCs
// to trigger WaitForFirstConsumer volume binding
func (r *EnvironmentReconciler) ensurePVCBindingHelperPod(
	ctx context.Context,
	environment *environmentsv1.Environment,
	pvcs []corev1.PersistentVolumeClaim,
	logger *zap.Logger,
) error {
	podName := fmt.Sprintf("%s-pvc-binder", environment.Name)
	targetNamespace := environment.Spec.TargetNamespace

	// Check if pod already exists
	existingPod := &corev1.Pod{}
	if err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, existingPod); err == nil {
		// Pod exists, check if it's running or completed
		if existingPod.Status.Phase == corev1.PodSucceeded || existingPod.Status.Phase == corev1.PodFailed {
			// Delete and recreate
			if err := r.Delete(ctx, existingPod); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete completed helper pod", zap.Error(err))
			}
		} else {
			logger.Info("PVC binding helper pod already exists", zap.String("phase", string(existingPod.Status.Phase)))
			return nil
		}
	} else if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing helper pod: %w", err)
	}

	// Build volume mounts for all pending PVCs
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	for i, pvc := range pvcs {
		if pvc.Status.Phase == corev1.ClaimBound {
			continue // Skip already bound PVCs
		}
		volName := fmt.Sprintf("vol-%d", i)
		volumes = append(volumes, corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volName,
			MountPath: fmt.Sprintf("/mnt/%s", pvc.Name),
		})
	}

	if len(volumes) == 0 {
		logger.Info("No pending PVCs to bind")
		return nil
	}

	// Create helper pod that just sleeps to allow PVC binding
	// Must be scheduled on the workmachine node where the storage is located
	helperPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: targetNamespace,
			Labels: map[string]string{
				"kloudlite.io/pvc-binder":               "true",
				"environments.kloudlite.io/environment": environment.Name,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			NodeSelector: map[string]string{
				"kubernetes.io/hostname": environment.Spec.WorkMachineName,
			},
			Tolerations: []corev1.Toleration{
				{
					Key:      "kloudlite.io/workmachine",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
			Containers: []corev1.Container{
				{
					Name:         "binder",
					Image:        "busybox:1.36",
					Command:      []string{"sh", "-c", "echo 'PVCs bound successfully'; sleep 10"},
					VolumeMounts: volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}

	if err := r.Create(ctx, helperPod); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("failed to create PVC binding helper pod: %w", err)
	}

	logger.Info("Created PVC binding helper pod",
		zap.String("pod", podName),
		zap.Int("pendingPVCs", len(volumes)))
	return nil
}

// cleanupPVCBindingHelperPod removes the temporary helper pod if it exists
func (r *EnvironmentReconciler) cleanupPVCBindingHelperPod(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) {
	podName := fmt.Sprintf("%s-pvc-binder", environment.Name)
	targetNamespace := environment.Spec.TargetNamespace

	helperPod := &corev1.Pod{}
	if err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, helperPod); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Warn("Failed to get helper pod for cleanup", zap.Error(err))
		}
		return
	}

	if err := r.Delete(ctx, helperPod); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete PVC binding helper pod", zap.Error(err))
	} else {
		logger.Info("Cleaned up PVC binding helper pod", zap.String("pod", podName))
	}
}

// parseImageRef parses an image reference like "image-registry:5000/repo:tag" into repository and tag
func parseImageRef(imageRef string) (repository, tag string) {
	// Remove registry prefix if present (e.g., "image-registry:5000/")
	parts := splitN(imageRef, "/", 2)
	if len(parts) == 2 {
		imageRef = parts[1]
	}

	// Split by colon to get repo:tag
	parts = splitN(imageRef, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return imageRef, "latest"
}

// splitN is a simple string split helper
func splitN(s, sep string, n int) []string {
	result := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	result = append(result, s)
	return result
}

func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
