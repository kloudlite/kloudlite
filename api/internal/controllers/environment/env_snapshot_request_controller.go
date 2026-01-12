package environment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

const (
	envSnapshotRequestFinalizer = "environments.kloudlite.io/snapshot-request-finalizer"
)

// EnvironmentSnapshotRequestReconciler reconciles EnvironmentSnapshotRequest objects
type EnvironmentSnapshotRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles EnvironmentSnapshotRequest events
func (r *EnvironmentSnapshotRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(zap.String("envSnapshotRequest", req.Name))
	logger.Info("Reconciling EnvironmentSnapshotRequest")

	// Fetch the EnvironmentSnapshotRequest
	envSnapshotReq := &environmentsv1.EnvironmentSnapshotRequest{}
	if err := r.Get(ctx, req.NamespacedName, envSnapshotReq); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Handle deletion
	if envSnapshotReq.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, envSnapshotReq, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(envSnapshotReq, envSnapshotRequestFinalizer) {
		controllerutil.AddFinalizer(envSnapshotReq, envSnapshotRequestFinalizer)
		if err := r.Update(ctx, envSnapshotReq); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Skip if already completed or failed
	// Environment controller will handle transitioning out of snapping state
	if envSnapshotReq.Status.Phase == environmentsv1.EnvironmentSnapshotRequestPhaseCompleted ||
		envSnapshotReq.Status.Phase == environmentsv1.EnvironmentSnapshotRequestPhaseFailed {
		return reconcile.Result{}, nil
	}

	// Get the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envSnapshotReq.Spec.EnvironmentName}, env); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, envSnapshotReq, "Environment not found", logger)
		}
		return reconcile.Result{}, err
	}

	// Process based on current phase
	switch envSnapshotReq.Status.Phase {
	case "", environmentsv1.EnvironmentSnapshotRequestPhasePending:
		return r.handlePending(ctx, envSnapshotReq, env, logger)

	case environmentsv1.EnvironmentSnapshotRequestPhaseStoppingWorkloads:
		return r.handleStoppingWorkloads(ctx, envSnapshotReq, env, logger)

	case environmentsv1.EnvironmentSnapshotRequestPhaseWaitingForPods:
		return r.handleWaitingForPods(ctx, envSnapshotReq, env, logger)

	case environmentsv1.EnvironmentSnapshotRequestPhaseCreatingSnapshot,
		environmentsv1.EnvironmentSnapshotRequestPhaseUploadingSnapshot:
		return r.handleSnapshotInProgress(ctx, envSnapshotReq, env, logger)

	case environmentsv1.EnvironmentSnapshotRequestPhaseRestoringEnvironment:
		return r.handleRestoringEnvironment(ctx, envSnapshotReq, env, logger)
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentSnapshotRequestReconciler) handlePending(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Starting snapshot request, saving environment state and setting to snapping")

	// Save the current environment state
	req.Status.PreviousEnvironmentState = env.Status.State
	req.Status.StartTime = &metav1.Time{Time: time.Now()}
	req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseStoppingWorkloads
	req.Status.Message = "Stopping environment workloads..."

	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	// Set environment state to snapping
	env.Status.State = environmentsv1.EnvironmentStateSnapping
	env.Status.Message = "Creating snapshot..."
	if err := r.Status().Update(ctx, env); err != nil {
		logger.Error("Failed to set environment to snapping state", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *EnvironmentSnapshotRequestReconciler) handleStoppingWorkloads(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Scaling down environment workloads")

	// Use the existing suspendEnvironment logic
	envReconciler := &EnvironmentReconciler{Client: r.Client, Scheme: r.Scheme, Logger: r.Logger}
	if err := envReconciler.suspendEnvironment(ctx, env, logger); err != nil {
		logger.Error("Failed to suspend environment", zap.Error(err))
		// Continue anyway - some resources might already be scaled down
	}

	// Move to waiting for pods
	req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseWaitingForPods
	req.Status.Message = "Waiting for pods to terminate..."
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

func (r *EnvironmentSnapshotRequestReconciler) handleWaitingForPods(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	// Check if all pods are terminated
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		return reconcile.Result{}, err
	}

	// Check for running pods (ignore completed jobs)
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			logger.Debug("Pod still running", zap.String("pod", pod.Name), zap.String("phase", string(pod.Status.Phase)))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
	}

	logger.Info("All pods terminated, creating snapshot resources")

	// Find the node for this workmachine by label
	nodeName, err := r.getNodeForWorkMachine(ctx, env.Spec.WorkMachineName)
	if err != nil {
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to find node for WorkMachine: %v", err), logger)
	}

	// Get parent snapshot for incremental
	parentSnapshot := ""
	if env.Status.LastRestoredSnapshot != nil {
		parentSnapshot = env.Status.LastRestoredSnapshot.Name
	}

	labels := map[string]string{
		"kloudlite.io/owned-by":              env.Spec.OwnedBy,
		"snapshots.kloudlite.io/environment": env.Name,
		"snapshots.kloudlite.io/type":        "environment",
	}

	// Create the Snapshot object first (so UI can see it immediately)
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Spec.SnapshotName,
			Labels: labels,
		},
		Spec: snapshotv1.SnapshotSpec{
			Owner:          env.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Spec.Description,
		},
	}

	if err := r.Create(ctx, snapshot); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create Snapshot: %v", err), logger)
		}
		logger.Debug("Snapshot already exists", zap.String("name", req.Spec.SnapshotName))
	} else {
		logger.Info("Created Snapshot", zap.String("name", req.Spec.SnapshotName))
	}

	// Create SnapshotArtifacts to capture K8s resources (including Environment's compose spec)
	if err := r.createSnapshotArtifacts(ctx, req.Spec.SnapshotName, env.Spec.TargetNamespace, env, logger); err != nil {
		logger.Warn("Failed to create snapshot artifacts", zap.Error(err))
		// Continue without artifacts - not a fatal error
	}

	// Create SnapshotRef to track ownership (prevents immediate GC when snapshot is ready)
	snapshotRef := &snapshotv1.SnapshotRef{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s--%s", env.Name, req.Spec.SnapshotName),
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/environment":        env.Name,
				"snapshots.kloudlite.io/snapshot": req.Spec.SnapshotName,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "environments.kloudlite.io/v1",
				Kind:       "Environment",
				Name:       env.Name,
				UID:        env.UID,
				Controller: boolPtr(true),
			}},
		},
		Spec: snapshotv1.SnapshotRefSpec{
			SnapshotName: req.Spec.SnapshotName,
			Purpose:      "owned",
		},
	}

	if err := r.Create(ctx, snapshotRef); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Warn("Failed to create SnapshotRef", zap.Error(err))
		}
	} else {
		logger.Info("Created SnapshotRef", zap.String("name", snapshotRef.Name))
	}

	// Create the SnapshotRequest (node-specific operation)
	snapshotRequestName := fmt.Sprintf("req-%s", req.Spec.SnapshotName)
	sourcePath := fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", env.Spec.TargetNamespace)

	snapshotReq := &snapshotv1.SnapshotRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotRequestName,
			Namespace: env.Spec.TargetNamespace,
			Labels:    labels,
		},
		Spec: snapshotv1.SnapshotRequestSpec{
			SnapshotName:   req.Spec.SnapshotName,
			SourcePath:     sourcePath,
			NodeName:       nodeName,
			Store:          "default",
			Owner:          env.Spec.OwnedBy,
			ParentSnapshot: parentSnapshot,
			Description:    req.Spec.Description,
		},
	}

	if err := r.Create(ctx, snapshotReq); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create SnapshotRequest: %v", err), logger)
		}
	}

	// Update status
	req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseCreatingSnapshot
	req.Status.Message = "Creating btrfs snapshot..."
	req.Status.SnapshotRequestName = snapshotRequestName
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

func (r *EnvironmentSnapshotRequestReconciler) handleSnapshotInProgress(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	// Get the SnapshotRequest status
	snapshotReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      req.Status.SnapshotRequestName,
		Namespace: env.Spec.TargetNamespace,
	}, snapshotReq); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, req, "SnapshotRequest not found", logger)
		}
		return reconcile.Result{}, err
	}

	// Check SnapshotRequest state
	switch snapshotReq.Status.State {
	case snapshotv1.SnapshotRequestStateCompleted:
		logger.Info("Snapshot created successfully")
		req.Status.CreatedSnapshotName = snapshotReq.Spec.SnapshotName
		req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseRestoringEnvironment
		req.Status.Message = "Restoring environment state..."
		if err := r.Status().Update(ctx, req); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil

	case snapshotv1.SnapshotRequestStateFailed:
		return r.setFailed(ctx, req, fmt.Sprintf("SnapshotRequest failed: %s", snapshotReq.Status.Message), logger)

	case snapshotv1.SnapshotRequestStateUploading:
		if req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseUploadingSnapshot {
			req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseUploadingSnapshot
			req.Status.Message = "Uploading snapshot to registry..."
			if err := r.Status().Update(ctx, req); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

func (r *EnvironmentSnapshotRequestReconciler) handleRestoringEnvironment(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Restoring environment to previous state",
		zap.String("previousState", string(req.Status.PreviousEnvironmentState)))

	// Restore environment state
	targetState := req.Status.PreviousEnvironmentState
	if targetState == "" || targetState == environmentsv1.EnvironmentStateSnapping {
		// Default to active if previous state was snapping or empty
		targetState = environmentsv1.EnvironmentStateActive
	}

	env.Status.State = targetState
	env.Status.Message = "Snapshot created successfully"

	// Update LastRestoredSnapshot to track lineage
	now := metav1.Now()
	lineage := []string{}
	if env.Status.LastRestoredSnapshot != nil && len(env.Status.LastRestoredSnapshot.Lineage) > 0 {
		lineage = env.Status.LastRestoredSnapshot.Lineage
	}
	lineage = append(lineage, req.Status.CreatedSnapshotName)

	env.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
		Name:       req.Status.CreatedSnapshotName,
		RestoredAt: now,
		Lineage:    lineage,
	}

	if err := r.Status().Update(ctx, env); err != nil {
		logger.Error("Failed to restore environment state", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Mark request as completed
	req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseCompleted
	req.Status.Message = fmt.Sprintf("Snapshot '%s' created successfully", req.Status.CreatedSnapshotName)
	req.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot request completed successfully")
	return reconcile.Result{}, nil
}

func (r *EnvironmentSnapshotRequestReconciler) handleDeletion(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Handling deletion of EnvironmentSnapshotRequest")

	// If snapshot was in progress, try to restore environment state
	if req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseCompleted &&
		req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseFailed &&
		req.Status.Phase != "" {
		env := &environmentsv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.EnvironmentName}, env); err == nil {
			if env.Status.State == environmentsv1.EnvironmentStateSnapping {
				targetState := req.Status.PreviousEnvironmentState
				if targetState == "" || targetState == environmentsv1.EnvironmentStateSnapping {
					targetState = environmentsv1.EnvironmentStateActive
				}
				env.Status.State = targetState
				env.Status.Message = "Snapshot request cancelled"
				if err := r.Status().Update(ctx, env); err != nil {
					logger.Warn("Failed to restore environment state on deletion", zap.Error(err))
				}
			}
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(req, envSnapshotRequestFinalizer)
	if err := r.Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentSnapshotRequestReconciler) setFailed(
	ctx context.Context,
	req *environmentsv1.EnvironmentSnapshotRequest,
	message string,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Snapshot request failed", zap.String("message", message))

	req.Status.Phase = environmentsv1.EnvironmentSnapshotRequestPhaseFailed
	req.Status.Message = message
	req.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	// Try to restore environment state
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.EnvironmentName}, env); err == nil {
		if env.Status.State == environmentsv1.EnvironmentStateSnapping {
			targetState := req.Status.PreviousEnvironmentState
			if targetState == "" || targetState == environmentsv1.EnvironmentStateSnapping {
				targetState = environmentsv1.EnvironmentStateActive
			}
			env.Status.State = targetState
			env.Status.Message = fmt.Sprintf("Snapshot failed: %s", message)
			if err := r.Status().Update(ctx, env); err != nil {
				logger.Warn("Failed to restore environment state after failure", zap.Error(err))
			}
		}
	}

	return reconcile.Result{}, nil
}

// getNodeForWorkMachine finds the k8s node for a workmachine by label
func (r *EnvironmentSnapshotRequestReconciler) getNodeForWorkMachine(ctx context.Context, workmachineName string) (string, error) {
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

// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentSnapshotRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.EnvironmentSnapshotRequest{}).
		Complete(r)
}

// createSnapshotArtifacts captures K8s resources and creates a SnapshotArtifacts CR
func (r *EnvironmentSnapshotRequestReconciler) createSnapshotArtifacts(ctx context.Context, snapshotName, namespace string, env *environmentsv1.Environment, logger *zap.Logger) error {
	artifacts := &snapshotv1.SnapshotArtifacts{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName, // Same name as the snapshot
			Labels: map[string]string{
				"snapshots.kloudlite.io/snapshot": snapshotName,
			},
		},
		Spec: snapshotv1.SnapshotArtifactsSpec{
			SnapshotName: snapshotName,
		},
	}

	var compositionCount, configMapCount, secretCount int32

	// Capture Environment's inline compose spec if present
	if env != nil && env.Spec.Compose != nil {
		composeData, err := json.Marshal(env.Spec.Compose)
		if err != nil {
			logger.Warn("Failed to marshal compose spec", zap.Error(err))
		} else {
			artifacts.Spec.ComposeSpec = base64.StdEncoding.EncodeToString(composeData)
			logger.Info("Captured environment compose spec")
		}
	}

	// Capture Compositions
	compositions := &environmentsv1.CompositionList{}
	if err := r.List(ctx, compositions, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list compositions: %w", err)
	}

	if len(compositions.Items) > 0 {
		cleanCompositions := make([]environmentsv1.Composition, len(compositions.Items))
		for i, comp := range compositions.Items {
			cleanCompositions[i] = environmentsv1.Composition{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "environments.kloudlite.io/v1",
					Kind:       "Composition",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        comp.Name,
					Labels:      comp.Labels,
					Annotations: comp.Annotations,
				},
				Spec: comp.Spec,
			}
		}

		data, err := yaml.Marshal(cleanCompositions)
		if err != nil {
			return fmt.Errorf("failed to marshal compositions: %w", err)
		}
		artifacts.Spec.Compositions = base64.StdEncoding.EncodeToString(data)
		compositionCount = int32(len(compositions.Items))
		logger.Info("Captured compositions", zap.Int("count", len(compositions.Items)))
	}

	// Capture ConfigMaps (excluding system ones)
	configMaps := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMaps, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list configmaps: %w", err)
	}

	var userConfigMaps []corev1.ConfigMap
	for _, cm := range configMaps.Items {
		if cm.Name == "kube-root-ca.crt" {
			continue
		}
		userConfigMaps = append(userConfigMaps, corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        cm.Name,
				Labels:      cm.Labels,
				Annotations: cm.Annotations,
			},
			Data:       cm.Data,
			BinaryData: cm.BinaryData,
		})
	}

	if len(userConfigMaps) > 0 {
		data, err := yaml.Marshal(userConfigMaps)
		if err != nil {
			return fmt.Errorf("failed to marshal configmaps: %w", err)
		}
		artifacts.Spec.ConfigMaps = base64.StdEncoding.EncodeToString(data)
		configMapCount = int32(len(userConfigMaps))
		logger.Info("Captured configmaps", zap.Int("count", len(userConfigMaps)))
	}

	// Capture Secrets (excluding service account tokens and system secrets)
	secrets := &corev1.SecretList{}
	if err := r.List(ctx, secrets, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	var userSecrets []corev1.Secret
	for _, secret := range secrets.Items {
		if secret.Type == corev1.SecretTypeServiceAccountToken {
			continue
		}
		if secret.Name == "default-token" || secret.Name == "builder-dockercfg" {
			continue
		}
		userSecrets = append(userSecrets, corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.Name,
				Labels:      secret.Labels,
				Annotations: secret.Annotations,
			},
			Data: secret.Data,
			Type: secret.Type,
		})
	}

	if len(userSecrets) > 0 {
		data, err := yaml.Marshal(userSecrets)
		if err != nil {
			return fmt.Errorf("failed to marshal secrets: %w", err)
		}
		artifacts.Spec.Secrets = base64.StdEncoding.EncodeToString(data)
		secretCount = int32(len(userSecrets))
		logger.Info("Captured secrets", zap.Int("count", len(userSecrets)))
	}

	// Set status counts
	artifacts.Status = snapshotv1.SnapshotArtifactsStatus{
		CompositionCount: compositionCount,
		ConfigMapCount:   configMapCount,
		SecretCount:      secretCount,
	}

	// Create the SnapshotArtifacts CR
	if err := r.Create(ctx, artifacts); err != nil {
		if apierrors.IsAlreadyExists(err) {
			logger.Debug("SnapshotArtifacts already exists", zap.String("name", snapshotName))
			return nil
		}
		return fmt.Errorf("failed to create SnapshotArtifacts: %w", err)
	}

	logger.Info("Created SnapshotArtifacts",
		zap.String("name", snapshotName),
		zap.Int32("compositions", compositionCount),
		zap.Int32("configMaps", configMapCount),
		zap.Int32("secrets", secretCount))

	return nil
}
