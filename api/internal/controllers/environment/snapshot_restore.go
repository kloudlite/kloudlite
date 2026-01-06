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
	// snapshotsBasePath is where snapshots are stored on the host filesystem
	// This is used for pull operations - the extracted data goes here
	snapshotsBasePath = "/var/lib/kloudlite/storage/.snapshots"
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

		// Determine snapshot path for the pulled data
		snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("env-%s-restore", environment.Name))

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
		// Pull completed, move to Restoring phase
		// Note: We keep the pull request until after restoration to access the pulled metadata
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseRestoring
			environment.Status.SnapshotRestoreStatus.Message = "Snapshot pulled, restoring resources"
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status after pull", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Pull completed, moving to Restoring phase")
		return reconcile.Result{Requeue: true}, nil

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
		if err := r.restoreResourcesFromMetadata(ctx, pullReq.Status.PulledMetadata, targetNamespace, environment.Name, logger); err != nil {
			logger.Warn("Failed to restore some resources", zap.Error(err))
			// Don't fail the entire restore if some resource restoration fails
		}
	} else {
		logger.Info("No K8s resource metadata found in pulled snapshot")
	}

	// Move to DataRestoring phase to restore PVC data
	// Keep the pull request for now - we need the snapshot path for data restoration
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseDataRestoring
		environment.Status.SnapshotRestoreStatus.Message = "K8s resources restored, restoring PVC data"
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Moving to DataRestoring phase")
	return reconcile.Result{Requeue: true}, nil
}

// handleRestoreDataRestoring restores PVC data from the pulled snapshot
func (r *EnvironmentReconciler) handleRestoreDataRestoring(
	ctx context.Context,
	environment *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Phase: DataRestoring - Restoring PVC data from snapshot")

	targetNamespace := environment.Spec.TargetNamespace
	status := environment.Status.SnapshotRestoreStatus
	wmNamespace := fmt.Sprintf("wm-%s", environment.Spec.OwnedBy)
	snapshotName := environment.Spec.FromSnapshot.SnapshotName

	// Get the pull request to find the snapshot path
	pullReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: status.SnapshotRequestName, Namespace: wmNamespace}, pullReq); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Warn("Pull request not found, skipping data restoration")
			return r.moveToCompleted(ctx, environment, logger)
		}
		logger.Error("Failed to get pull request", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// The pulled snapshot data is at: {snapshotPath}/{snapshotName}/
	pulledSnapshotBase := filepath.Join(pullReq.Spec.SnapshotPath, snapshotName)

	// List PVCs in the target namespace
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, client.InNamespace(targetNamespace)); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if len(pvcList.Items) == 0 {
		// Check if there are compositions that might create PVCs
		compositionList := &environmentsv1.CompositionList{}
		if err := r.List(ctx, compositionList, client.InNamespace(targetNamespace)); err != nil {
			logger.Error("Failed to list compositions", zap.Error(err))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		if len(compositionList.Items) == 0 {
			// No compositions and no PVCs - nothing to restore
			logger.Info("No PVCs and no compositions found, skipping data restoration")
			return r.moveToCompleted(ctx, environment, logger)
		}

		// Compositions exist - wait for PVCs to be created
		logger.Info("No PVCs found yet, waiting for composition to create them",
			zap.Int("compositions", len(compositionList.Items)))
		return reconcile.Result{RequeueAfter: 3 * time.Second}, nil
	}

	// Check/create restore requests for each PVC
	allCompleted := true
	anyFailed := false
	anyPending := false

	for _, pvc := range pvcList.Items {
		// Handle PVCs that aren't bound yet (WaitForFirstConsumer)
		if pvc.Status.Phase != corev1.ClaimBound {
			logger.Info("PVC not bound yet", zap.String("pvc", pvc.Name), zap.String("phase", string(pvc.Status.Phase)))
			anyPending = true
			allCompleted = false
			continue
		}

		restoreReqName := fmt.Sprintf("%s-data-restore-%s", environment.Name, pvc.Name)

		// Check if restore request already exists
		restoreReq := &snapshotv1.SnapshotRequest{}
		if err := r.Get(ctx, client.ObjectKey{Name: restoreReqName, Namespace: wmNamespace}, restoreReq); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to get restore request", zap.Error(err), zap.String("pvc", pvc.Name))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			}

			// Create restore request
			// Source: snapshot data path matching by claim name
			// Target: actual PVC mount path
			sourcePath, targetPath := r.getPVCRestorePaths(pulledSnapshotBase, targetNamespace, &pvc, logger)
			if sourcePath == "" {
				logger.Warn("Could not determine source path for PVC", zap.String("pvc", pvc.Name))
				continue
			}

			newRestoreReq := &snapshotv1.SnapshotRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      restoreReqName,
					Namespace: wmNamespace,
					Labels: map[string]string{
						"environments.kloudlite.io/environment": environment.Name,
						"environments.kloudlite.io/operation":   "data-restore",
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

			if err := r.Create(ctx, newRestoreReq); err != nil {
				if !apierrors.IsAlreadyExists(err) {
					logger.Error("Failed to create restore request", zap.Error(err), zap.String("pvc", pvc.Name))
					return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
				}
			}

			logger.Info("Created data restore request",
				zap.String("pvc", pvc.Name),
				zap.String("source", sourcePath),
				zap.String("target", targetPath))
			allCompleted = false
			continue
		}

		// Check restore request status
		switch restoreReq.Status.Phase {
		case snapshotv1.SnapshotRequestPhaseCompleted:
			logger.Info("PVC data restore completed", zap.String("pvc", pvc.Name))
			// Clean up completed restore request
			if err := r.Delete(ctx, restoreReq); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete completed restore request", zap.Error(err))
			}
		case snapshotv1.SnapshotRequestPhaseFailed:
			logger.Error("PVC data restore failed", zap.String("pvc", pvc.Name), zap.String("error", restoreReq.Status.Message))
			anyFailed = true
		default:
			logger.Info("PVC data restore in progress", zap.String("pvc", pvc.Name), zap.String("phase", string(restoreReq.Status.Phase)))
			allCompleted = false
		}
	}

	if anyFailed {
		logger.Warn("Some PVC data restores failed, proceeding anyway")
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

	if !allCompleted {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// All restores completed, clean up pull request and move to Completed
	if err := r.Delete(ctx, pullReq); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete pull request", zap.Error(err))
	}

	return r.moveToCompleted(ctx, environment, logger)
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

	// Auto-create a snapshot on the cloned environment to maintain lineage
	if err := r.createClonedSnapshot(ctx, environment, sourceSnapshotName, logger); err != nil {
		logger.Warn("Failed to auto-create snapshot on cloned environment", zap.Error(err))
		// Don't fail the restore - this is best-effort
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

// createClonedSnapshot creates a snapshot on the cloned environment with proper lineage
func (r *EnvironmentReconciler) createClonedSnapshot(
	ctx context.Context,
	environment *environmentsv1.Environment,
	sourceSnapshotName string,
	logger *zap.Logger,
) error {
	// Get source snapshot to inherit description
	sourceSnapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: sourceSnapshotName}, sourceSnapshot); err != nil {
		return fmt.Errorf("failed to get source snapshot: %w", err)
	}

	// Generate snapshot name for the cloned environment
	snapshotName := fmt.Sprintf("%s-clone-%d", environment.Name, time.Now().Unix())

	// Create snapshot with inherited description and parent reference
	restoredAt := metav1.Now()
	newSnapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment":  environment.Name,
				"kloudlite.io/owned-by":               environment.Spec.OwnedBy,
				"snapshots.kloudlite.io/auto-created": "true",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: environment.Name,
			},
			ParentSnapshotRef: &snapshotv1.ParentSnapshotReference{
				Name:       sourceSnapshotName,
				RestoredAt: &restoredAt,
			},
			Description:     sourceSnapshot.Spec.Description,
			OwnedBy:         environment.Spec.OwnedBy,
			IncludeMetadata: true,
			RegistryRef: &snapshotv1.SnapshotRegistryRef{
				Repository: fmt.Sprintf("snapshots/%s", environment.Spec.OwnedBy),
				AutoPush:   true,
			},
		},
	}

	if err := r.Create(ctx, newSnapshot); err != nil {
		if apierrors.IsAlreadyExists(err) {
			logger.Info("Snapshot already exists for cloned environment", zap.String("snapshot", snapshotName))
			return nil
		}
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	logger.Info("Auto-created snapshot on cloned environment",
		zap.String("snapshot", snapshotName),
		zap.String("parent", sourceSnapshotName),
		zap.String("description", sourceSnapshot.Spec.Description))

	return nil
}

// restoreResourcesFromMetadata restores K8s resources from the metadata struct (from OCI layer)
func (r *EnvironmentReconciler) restoreResourcesFromMetadata(
	ctx context.Context,
	metadata *snapshotv1.SnapshotMetadata,
	targetNamespace, envName string,
	logger *zap.Logger,
) error {
	logger.Info("Restoring resources from OCI metadata")

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
