package environment

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
	snapshotsBasePath = "/kl-data/snapshots"
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
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseRestoring
			environment.Status.SnapshotRestoreStatus.Message = "Snapshot pulled, restoring resources"
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status after pull", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Delete the completed pull request
		if err := r.Delete(ctx, pullReq); err != nil && !apierrors.IsNotFound(err) {
			logger.Warn("Failed to delete completed pull request", zap.Error(err))
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
	logger.Info("Phase: Restoring - Applying resources from snapshot")

	targetNamespace := environment.Spec.TargetNamespace

	// Restore resources from snapshot metadata
	if err := r.restoreResourcesFromSnapshot(ctx, environment, targetNamespace, logger); err != nil {
		logger.Warn("Failed to restore some resources", zap.Error(err))
		// Don't fail the entire restore if some resource restoration fails
	}

	// Track the restored snapshot for lineage
	now := metav1.Now()
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus.Phase = environmentsv1.SnapshotRestorePhaseCompleted
		environment.Status.SnapshotRestoreStatus.Message = "Snapshot restore completed"
		environment.Status.SnapshotRestoreStatus.CompletionTime = &now

		// Track the last restored snapshot for lineage
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

// restoreResourcesFromSnapshot reads metadata from the pulled snapshot and creates resources
func (r *EnvironmentReconciler) restoreResourcesFromSnapshot(
	ctx context.Context,
	environment *environmentsv1.Environment,
	targetNamespace string,
	logger *zap.Logger,
) error {
	// Determine the snapshot path
	snapshotPath := filepath.Join(snapshotsBasePath, fmt.Sprintf("env-%s-restore", environment.Name))
	metadataPath := filepath.Join(snapshotPath, "metadata")

	// Check if metadata directory exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		logger.Info("No metadata directory found in snapshot", zap.String("path", metadataPath))
		return nil
	}

	// Restore ConfigMaps
	if err := r.restoreConfigMaps(ctx, metadataPath, targetNamespace, environment.Name, logger); err != nil {
		logger.Warn("Failed to restore ConfigMaps", zap.Error(err))
	}

	// Restore Secrets
	if err := r.restoreSecrets(ctx, metadataPath, targetNamespace, environment.Name, logger); err != nil {
		logger.Warn("Failed to restore Secrets", zap.Error(err))
	}

	// Restore Deployments
	if err := r.restoreDeployments(ctx, metadataPath, targetNamespace, logger); err != nil {
		logger.Warn("Failed to restore Deployments", zap.Error(err))
	}

	// Restore StatefulSets
	if err := r.restoreStatefulSets(ctx, metadataPath, targetNamespace, logger); err != nil {
		logger.Warn("Failed to restore StatefulSets", zap.Error(err))
	}

	// Restore Services
	if err := r.restoreServices(ctx, metadataPath, targetNamespace, logger); err != nil {
		logger.Warn("Failed to restore Services", zap.Error(err))
	}

	// Restore Compositions
	if err := r.restoreCompositions(ctx, metadataPath, targetNamespace, logger); err != nil {
		logger.Warn("Failed to restore Compositions", zap.Error(err))
	}

	return nil
}

// restoreConfigMaps restores ConfigMaps from snapshot metadata
func (r *EnvironmentReconciler) restoreConfigMaps(
	ctx context.Context,
	metadataPath, targetNamespace, envName string,
	logger *zap.Logger,
) error {
	filePath := filepath.Join(metadataPath, "configmaps.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read configmaps.json: %w", err)
	}

	var configMapList corev1.ConfigMapList
	if err := json.Unmarshal(data, &configMapList); err != nil {
		return fmt.Errorf("failed to parse configmaps.json: %w", err)
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

// restoreSecrets restores Secrets from snapshot metadata
func (r *EnvironmentReconciler) restoreSecrets(
	ctx context.Context,
	metadataPath, targetNamespace, envName string,
	logger *zap.Logger,
) error {
	filePath := filepath.Join(metadataPath, "secrets.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read secrets.json: %w", err)
	}

	var secrets []corev1.Secret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return fmt.Errorf("failed to parse secrets.json: %w", err)
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

// restoreDeployments restores Deployments from snapshot metadata
func (r *EnvironmentReconciler) restoreDeployments(
	ctx context.Context,
	metadataPath, targetNamespace string,
	logger *zap.Logger,
) error {
	filePath := filepath.Join(metadataPath, "deployments.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read deployments.json: %w", err)
	}

	var deploymentList appsv1.DeploymentList
	if err := json.Unmarshal(data, &deploymentList); err != nil {
		return fmt.Errorf("failed to parse deployments.json: %w", err)
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

// restoreStatefulSets restores StatefulSets from snapshot metadata
func (r *EnvironmentReconciler) restoreStatefulSets(
	ctx context.Context,
	metadataPath, targetNamespace string,
	logger *zap.Logger,
) error {
	filePath := filepath.Join(metadataPath, "statefulsets.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read statefulsets.json: %w", err)
	}

	var stsList appsv1.StatefulSetList
	if err := json.Unmarshal(data, &stsList); err != nil {
		return fmt.Errorf("failed to parse statefulsets.json: %w", err)
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

// restoreServices restores Services from snapshot metadata
func (r *EnvironmentReconciler) restoreServices(
	ctx context.Context,
	metadataPath, targetNamespace string,
	logger *zap.Logger,
) error {
	filePath := filepath.Join(metadataPath, "services.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read services.json: %w", err)
	}

	var svcList corev1.ServiceList
	if err := json.Unmarshal(data, &svcList); err != nil {
		return fmt.Errorf("failed to parse services.json: %w", err)
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

// restoreCompositions restores Compositions from snapshot metadata
func (r *EnvironmentReconciler) restoreCompositions(
	ctx context.Context,
	metadataPath, targetNamespace string,
	logger *zap.Logger,
) error {
	filePath := filepath.Join(metadataPath, "compositions.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // No compositions to restore
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compositions.json: %w", err)
	}

	var compositionList environmentsv1.CompositionList
	if err := json.Unmarshal(data, &compositionList); err != nil {
		return fmt.Errorf("failed to parse compositions.json: %w", err)
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
