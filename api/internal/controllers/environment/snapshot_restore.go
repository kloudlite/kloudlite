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
		logger.Info("No PVCs found, skipping data restoration")
		return r.moveToCompleted(ctx, environment, logger)
	}

	// Check/create restore requests for each PVC
	allCompleted := true
	anyFailed := false

	for _, pvc := range pvcList.Items {
		// Skip PVCs that aren't bound yet
		if pvc.Status.Phase != corev1.ClaimBound {
			logger.Info("PVC not bound yet, waiting", zap.String("pvc", pvc.Name))
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
	// The snapshot contains directories like: pvc-{uid}_{original-namespace}_{claim-name}
	// We need to match by claim name since the PVC UID and namespace will be different in the clone

	claimName := pvc.Name
	pvcUID := string(pvc.UID)

	// Target path is where the PVC is actually mounted by local-path-provisioner
	// Format: /var/lib/kloudlite/storage/environments/{namespace}/pvc-{uid}_{namespace}_{claim-name}
	environmentsBasePath := "/var/lib/kloudlite/storage/environments"
	targetPath = filepath.Join(environmentsBasePath, targetNamespace, fmt.Sprintf("pvc-%s_%s_%s", pvcUID, targetNamespace, claimName))

	// Source path - we need to find the matching directory in the snapshot by claim name
	// The snapshot data is at: {pulledSnapshotBase}/pvc-*_env-*_{claim-name}
	// Since we can't list files from the controller, we construct the expected pattern
	// The actual matching will be done by the node manager using glob patterns
	sourcePath = filepath.Join(pulledSnapshotBase, fmt.Sprintf("*_%s", claimName))

	logger.Info("Determined PVC restore paths",
		zap.String("pvc", pvc.Name),
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

	// Clear fromSnapshot to mark restoration as complete
	environment.Spec.FromSnapshot = nil

	// Add completion condition
	environment.Status.Conditions = append(environment.Status.Conditions, environmentsv1.EnvironmentCondition{
		Type:               "RestoredFromSnapshot",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: fn.Ptr(metav1.Now()),
		Reason:             "RestoreCompleted",
		Message:            fmt.Sprintf("Successfully restored from snapshot %s", environment.Status.SnapshotRestoreStatus.SourceSnapshot),
	})

	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear fromSnapshot field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot restore completed, proceeding to normal environment reconciliation")

	// Requeue to start normal environment reconciliation
	return reconcile.Result{Requeue: true}, nil
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
