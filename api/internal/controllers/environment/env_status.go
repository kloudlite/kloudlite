package environment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	sigyaml "sigs.k8s.io/yaml"
)

// generateHash generates an 8-character hash from the input string
func generateHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])[:8]
}

// updateHashAndSubdomain computes and sets the hash and subdomain in environment status
func (r *EnvironmentReconciler) updateHashAndSubdomain(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Compute hash from envName-owner
	hash := generateHash(fmt.Sprintf("%s-%s", environment.Spec.Name, environment.Spec.OwnedBy))

	// Get subdomain from HOSTED_SUBDOMAIN env var (shared across all environments)
	subdomain := os.Getenv("HOSTED_SUBDOMAIN")
	if subdomain == "" {
		logger.Debug("HOSTED_SUBDOMAIN env var not set, subdomain will be empty")
	}

	// Only update if values changed
	if environment.Status.Hash == hash && environment.Status.Subdomain == subdomain {
		return nil
	}

	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.Hash = hash
		environment.Status.Subdomain = subdomain
		return nil
	}, logger)
}

// updateEnvironmentStatus safely updates environment status with retry logic
func (r *EnvironmentReconciler) updateEnvironmentStatus(ctx context.Context, environment *environmentsv1.Environment, state environmentsv1.EnvironmentState, message string, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = state
		environment.Status.Message = message

		now := metav1.Now()
		if state == environmentsv1.EnvironmentStateActive {
			environment.Status.LastActivatedTime = &now
		} else if state == environmentsv1.EnvironmentStateInactive {
			environment.Status.LastDeactivatedTime = &now
		}

		return nil
	}, logger)
}

// addOrUpdateCondition adds or updates a condition in the environment status
func (r *EnvironmentReconciler) addOrUpdateCondition(environment *environmentsv1.Environment, conditionType environmentsv1.EnvironmentConditionType, status metav1.ConditionStatus, reason, message string) {
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	now := metav1.Now()
	newCondition := environmentsv1.EnvironmentCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: &now,
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or add new one
	found := false
	for i, condition := range environment.Status.Conditions {
		if condition.Type == conditionType {
			environment.Status.Conditions[i] = newCondition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, newCondition)
	}
}

// handleSnapshotRestore handles environment creation from a snapshot
// This creates a SnapshotRestore resource and waits for it to complete
func (r *EnvironmentReconciler) handleSnapshotRestore(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	snapshotName := environment.Spec.FromSnapshot.SnapshotName

	// Get the snapshot to verify it exists and is ready
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("Snapshot not found", zap.String("snapshot", snapshotName))
			return r.failSnapshotRestore(ctx, environment, fmt.Sprintf("Snapshot %s not found", snapshotName), logger)
		}
		logger.Error("Failed to get snapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		logger.Info("Snapshot not ready, waiting", zap.String("state", string(snapshot.Status.State)))
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePending,
			fmt.Sprintf("Waiting for snapshot to be ready (state: %s)", snapshot.Status.State), logger)
	}

	// Get the node name from the workmachine
	if environment.Spec.WorkMachineName == "" {
		return r.failSnapshotRestore(ctx, environment, "Environment has no workmachine assigned", logger)
	}

	nodeName, err := r.getNodeForWorkMachine(ctx, environment.Spec.WorkMachineName)
	if err != nil {
		logger.Warn("WorkMachine not ready, waiting", zap.Error(err))
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePending,
			"Waiting for workmachine to be ready", logger)
	}

	// Ensure namespace exists for the SnapshotRestore resource
	namespaceExists, err := r.ensureNamespaceExists(ctx, environment, logger)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !namespaceExists {
		// Namespace was just created, requeue to continue
		return reconcile.Result{Requeue: true}, nil
	}

	// Define the restore name
	restoreName := fmt.Sprintf("env-restore-%s", environment.Name)

	// Check if SnapshotRestore already exists
	restore := &snapshotv1.SnapshotRestore{}
	err = r.Get(ctx, client.ObjectKey{Name: restoreName, Namespace: environment.Spec.TargetNamespace}, restore)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error("Failed to get SnapshotRestore", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Create the SnapshotRestore resource
		logger.Info("Creating SnapshotRestore", zap.String("restore", restoreName), zap.String("snapshot", snapshotName))

		targetPath := fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", environment.Spec.TargetNamespace)

		restore = &snapshotv1.SnapshotRestore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      restoreName,
				Namespace: environment.Spec.TargetNamespace,
				Labels: map[string]string{
					"kloudlite.io/owned-by":              environment.Spec.OwnedBy,
					"snapshots.kloudlite.io/environment": environment.Name,
					"snapshots.kloudlite.io/source":      snapshotName,
				},
			},
			Spec: snapshotv1.SnapshotRestoreSpec{
				SnapshotName: snapshotName,
				TargetPath:   targetPath,
				NodeName:     nodeName,
			},
		}

		if err := r.Create(ctx, restore); err != nil {
			logger.Error("Failed to create SnapshotRestore", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Update status to show we're starting the restore
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePulling,
			"Downloading snapshot from registry", logger)
	}

	// SnapshotRestore exists, check its status
	switch restore.Status.State {
	case snapshotv1.SnapshotRestoreStatePending:
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePending,
			"Waiting to start restore", logger)

	case snapshotv1.SnapshotRestoreStateDownloading:
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhasePulling,
			"Downloading snapshot from registry", logger)

	case snapshotv1.SnapshotRestoreStateRestoring:
		return r.updateSnapshotRestoreStatus(ctx, environment, environmentsv1.SnapshotRestorePhaseDataRestoring,
			"Restoring snapshot data", logger)

	case snapshotv1.SnapshotRestoreStateCompleted:
		// Restore completed! Now apply artifacts from SnapshotArtifacts CR
		logger.Info("Snapshot restore completed successfully", zap.String("snapshot", snapshotName))

		// Apply artifacts (Compositions, ConfigMaps, Secrets) from SnapshotArtifacts CR
		if err := r.applySnapshotArtifacts(ctx, snapshotName, environment, logger); err != nil {
			logger.Warn("Failed to apply snapshot artifacts", zap.Error(err))
			// Don't fail the restore, just log the warning
		}

		// Build full lineage: snapshot's ancestors + the snapshot itself
		lineage := append(snapshot.Status.Lineage, snapshotName)
		logger.Info("Built snapshot lineage for forked environment",
			zap.Strings("lineage", lineage),
			zap.Int("count", len(lineage)))

		// Increment refCount for all snapshots in the lineage
		// This protects parent snapshots from garbage collection while this environment exists
		if err := r.incrementSnapshotLineageRefCounts(ctx, lineage, logger); err != nil {
			logger.Error("Failed to increment snapshot lineage refCounts", zap.Error(err))
			// Continue anyway - the environment is already restored
		}

		now := metav1.Now()

		// Update status with completed restore, LastRestoredSnapshot, and full lineage
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
				Phase:          environmentsv1.SnapshotRestorePhaseCompleted,
				Message:        "Snapshot restored successfully",
				SourceSnapshot: snapshotName,
				CompletionTime: &now,
			}
			environment.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
				Name:       snapshotName,
				RestoredAt: now,
				Lineage:    lineage,
			}
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Clear FromSnapshot to proceed with normal reconciliation
		environment.Spec.FromSnapshot = nil
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to clear fromSnapshot", zap.Error(err))
			return reconcile.Result{}, err
		}

		logger.Info("Cleared fromSnapshot, proceeding with normal environment reconciliation")
		return reconcile.Result{Requeue: true}, nil

	case snapshotv1.SnapshotRestoreStateFailed:
		return r.failSnapshotRestore(ctx, environment,
			fmt.Sprintf("Snapshot restore failed: %s", restore.Status.Message), logger)

	default:
		logger.Warn("Unknown restore state", zap.String("state", string(restore.Status.State)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
}

// getNodeForWorkMachine finds the k8s node for a workmachine by label
func (r *EnvironmentReconciler) getNodeForWorkMachine(ctx context.Context, workmachineName string) (string, error) {
	var nodes corev1.NodeList
	if err := r.List(ctx, &nodes, client.MatchingLabels{
		"kloudlite.io/workmachine": workmachineName,
	}); err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no node found for workmachine %s", workmachineName)
	}
	return nodes.Items[0].Name, nil
}

// updateSnapshotRestoreStatus updates the environment's snapshot restore status
func (r *EnvironmentReconciler) updateSnapshotRestoreStatus(ctx context.Context, environment *environmentsv1.Environment, phase environmentsv1.SnapshotRestorePhase, message string, logger *zap.Logger) (reconcile.Result, error) {
	now := metav1.Now()

	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		if environment.Status.SnapshotRestoreStatus == nil {
			environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
				StartTime:      &now,
				SourceSnapshot: environment.Spec.FromSnapshot.SnapshotName,
			}
		}
		environment.Status.SnapshotRestoreStatus.Phase = phase
		environment.Status.SnapshotRestoreStatus.Message = message
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update snapshot restore status", zap.Error(err))
	}

	// Requeue to check progress
	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// failSnapshotRestore marks the snapshot restore as failed and clears FromSnapshot
func (r *EnvironmentReconciler) failSnapshotRestore(ctx context.Context, environment *environmentsv1.Environment, errorMessage string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot restore failed", zap.String("error", errorMessage))

	now := metav1.Now()

	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.SnapshotRestoreStatus = &environmentsv1.SnapshotRestoreStatus{
			Phase:          environmentsv1.SnapshotRestorePhaseFailed,
			Message:        errorMessage,
			ErrorMessage:   errorMessage,
			SourceSnapshot: environment.Spec.FromSnapshot.SnapshotName,
			CompletionTime: &now,
		}
		environment.Status.State = environmentsv1.EnvironmentStateError
		environment.Status.Message = errorMessage
		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update status", zap.Error(err))
	}

	// Clear FromSnapshot even on failure to avoid infinite loops
	environment.Spec.FromSnapshot = nil
	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear fromSnapshot", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// applySnapshotArtifacts reads SnapshotArtifacts CR and applies resources to the target environment
func (r *EnvironmentReconciler) applySnapshotArtifacts(ctx context.Context, snapshotName string, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Get the SnapshotArtifacts CR - it's cluster-scoped, named after the snapshot
	artifacts := &snapshotv1.SnapshotArtifacts{}
	if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, artifacts); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("No SnapshotArtifacts found for snapshot", zap.String("snapshot", snapshotName))
			return nil
		}
		return fmt.Errorf("failed to get SnapshotArtifacts: %w", err)
	}

	targetNamespace := environment.Spec.TargetNamespace
	logger.Info("Applying snapshot artifacts",
		zap.String("snapshot", snapshotName),
		zap.String("targetNamespace", targetNamespace))

	// Apply Compositions
	if artifacts.Spec.Compositions != "" {
		count, err := r.applyCompositionsFromYAML(ctx, artifacts.Spec.Compositions, targetNamespace, environment.Name, logger)
		if err != nil {
			logger.Warn("Failed to apply compositions", zap.Error(err))
		} else {
			logger.Info("Applied compositions from snapshot", zap.Int("count", count))
		}
	}

	// Apply ConfigMaps
	if artifacts.Spec.ConfigMaps != "" {
		count, err := r.applyConfigMapsFromYAML(ctx, artifacts.Spec.ConfigMaps, targetNamespace, logger)
		if err != nil {
			logger.Warn("Failed to apply configmaps", zap.Error(err))
		} else {
			logger.Info("Applied configmaps from snapshot", zap.Int("count", count))
		}
	}

	// Apply Secrets
	if artifacts.Spec.Secrets != "" {
		count, err := r.applySecretsFromYAML(ctx, artifacts.Spec.Secrets, targetNamespace, logger)
		if err != nil {
			logger.Warn("Failed to apply secrets", zap.Error(err))
		} else {
			logger.Info("Applied secrets from snapshot", zap.Int("count", count))
		}
	}

	return nil
}

// applyCompositionsFromYAML decodes base64 YAML and creates Composition resources
func (r *EnvironmentReconciler) applyCompositionsFromYAML(ctx context.Context, encodedYAML, targetNamespace, envName string, logger *zap.Logger) (int, error) {
	yamlData, err := base64.StdEncoding.DecodeString(encodedYAML)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64: %w", err)
	}

	// First try to decode as a YAML array ([]Composition)
	var compositions []environmentsv1.Composition
	if err := sigyaml.Unmarshal(yamlData, &compositions); err == nil && len(compositions) > 0 {
		// Successfully decoded as YAML array
	} else {
		// Try to decode as a Kubernetes List type
		compList := &environmentsv1.CompositionList{}
		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
		if err := decoder.Decode(compList); err == nil && len(compList.Items) > 0 {
			compositions = compList.Items
		} else {
			// Fall back to decoding individual compositions (multi-doc YAML)
			decoder = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
			for {
				comp := &environmentsv1.Composition{}
				if err := decoder.Decode(comp); err != nil {
					if err == io.EOF {
						break
					}
					return 0, fmt.Errorf("failed to decode composition: %w", err)
				}
				if comp.Name != "" {
					compositions = append(compositions, *comp)
				}
			}
		}
	}

	count := 0
	for _, comp := range compositions {
		// Create a copy in the target namespace
		newComp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:        comp.Name,
				Namespace:   targetNamespace,
				Labels:      comp.Labels,
				Annotations: comp.Annotations,
			},
			Spec: comp.Spec,
		}

		// Update labels to reference the new environment
		if newComp.Labels == nil {
			newComp.Labels = make(map[string]string)
		}
		newComp.Labels["kloudlite.io/environment"] = envName

		if err := r.Create(ctx, newComp); err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Debug("Composition already exists, skipping", zap.String("name", comp.Name))
				continue
			}
			logger.Warn("Failed to create composition", zap.String("name", comp.Name), zap.Error(err))
			continue
		}
		count++
		logger.Debug("Created composition from snapshot", zap.String("name", comp.Name))
	}

	return count, nil
}

// applyConfigMapsFromYAML decodes base64 YAML and creates ConfigMap resources
func (r *EnvironmentReconciler) applyConfigMapsFromYAML(ctx context.Context, encodedYAML, targetNamespace string, logger *zap.Logger) (int, error) {
	yamlData, err := base64.StdEncoding.DecodeString(encodedYAML)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64: %w", err)
	}

	// First try to decode as a YAML array ([]ConfigMap)
	var configMaps []corev1.ConfigMap
	if err := sigyaml.Unmarshal(yamlData, &configMaps); err == nil && len(configMaps) > 0 {
		// Successfully decoded as YAML array
	} else {
		// Try to decode as a Kubernetes List type
		cmList := &corev1.ConfigMapList{}
		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
		if err := decoder.Decode(cmList); err == nil && len(cmList.Items) > 0 {
			configMaps = cmList.Items
		} else {
			// Fall back to decoding individual configmaps (multi-doc YAML)
			decoder = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
			for {
				cm := &corev1.ConfigMap{}
				if err := decoder.Decode(cm); err != nil {
					if err == io.EOF {
						break
					}
					return 0, fmt.Errorf("failed to decode configmap: %w", err)
				}
				if cm.Name != "" {
					configMaps = append(configMaps, *cm)
				}
			}
		}
	}

	count := 0
	for _, cm := range configMaps {
		// Create a copy in the target namespace
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

		if err := r.Create(ctx, newCM); err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Debug("ConfigMap already exists, skipping", zap.String("name", cm.Name))
				continue
			}
			logger.Warn("Failed to create configmap", zap.String("name", cm.Name), zap.Error(err))
			continue
		}
		count++
		logger.Debug("Created configmap from snapshot", zap.String("name", cm.Name))
	}

	return count, nil
}

// applySecretsFromYAML decodes base64 YAML and creates Secret resources
func (r *EnvironmentReconciler) applySecretsFromYAML(ctx context.Context, encodedYAML, targetNamespace string, logger *zap.Logger) (int, error) {
	yamlData, err := base64.StdEncoding.DecodeString(encodedYAML)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64: %w", err)
	}

	// First try to decode as a YAML array ([]Secret)
	var secrets []corev1.Secret
	if err := sigyaml.Unmarshal(yamlData, &secrets); err == nil && len(secrets) > 0 {
		// Successfully decoded as YAML array
	} else {
		// Try to decode as a Kubernetes List type
		secretList := &corev1.SecretList{}
		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
		if err := decoder.Decode(secretList); err == nil && len(secretList.Items) > 0 {
			secrets = secretList.Items
		} else {
			// Fall back to decoding individual secrets (multi-doc YAML)
			decoder = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlData), 4096)
			for {
				secret := &corev1.Secret{}
				if err := decoder.Decode(secret); err != nil {
					if err == io.EOF {
						break
					}
					return 0, fmt.Errorf("failed to decode secret: %w", err)
				}
				if secret.Name != "" {
					secrets = append(secrets, *secret)
				}
			}
		}
	}

	count := 0
	for _, secret := range secrets {
		// Create a copy in the target namespace
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.Name,
				Namespace:   targetNamespace,
				Labels:      secret.Labels,
				Annotations: secret.Annotations,
			},
			Type:       secret.Type,
			Data:       secret.Data,
			StringData: secret.StringData,
		}

		if err := r.Create(ctx, newSecret); err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Debug("Secret already exists, skipping", zap.String("name", secret.Name))
				continue
			}
			logger.Warn("Failed to create secret", zap.String("name", secret.Name), zap.Error(err))
			continue
		}
		count++
		logger.Debug("Created secret from snapshot", zap.String("name", secret.Name))
	}

	return count, nil
}

// incrementSnapshotLineageRefCounts increments refCount for all snapshots in the lineage
// This ensures parent snapshots are protected from garbage collection while the forked environment exists
func (r *EnvironmentReconciler) incrementSnapshotLineageRefCounts(ctx context.Context, lineage []string, logger *zap.Logger) error {
	for _, snapshotName := range lineage {
		snapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: snapshotName}, snapshot); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Warn("Snapshot in lineage not found, skipping refCount increment",
					zap.String("snapshot", snapshotName))
				continue
			}
			return fmt.Errorf("failed to get snapshot %s: %w", snapshotName, err)
		}

		snapshot.Status.RefCount++
		if err := r.Status().Update(ctx, snapshot); err != nil {
			return fmt.Errorf("failed to update snapshot %s refCount: %w", snapshotName, err)
		}

		logger.Info("Incremented snapshot refCount",
			zap.String("snapshot", snapshotName),
			zap.Int32("newRefCount", snapshot.Status.RefCount))
	}
	return nil
}
