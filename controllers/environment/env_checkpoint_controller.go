package environment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/controllers/controllerconfig"
	"github.com/kloudlite/kloudlite/pkg/pagination"
	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
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

const envCheckpointFinalizer = "environments.kloudlite.io/checkpoint-finalizer"

// EnvironmentCheckpointReconciler reconciles EnvironmentCheckpoint objects
type EnvironmentCheckpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
	Cfg    *controllerconfig.ControllerConfig
}

func (r *EnvironmentCheckpointReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(zap.String("envCheckpoint", req.Name))

	envCheckpoint := &environmentsv1.EnvironmentCheckpoint{}
	if err := r.Get(ctx, req.NamespacedName, envCheckpoint); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Handle deletion
	if envCheckpoint.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, envCheckpoint, logger)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(envCheckpoint, envCheckpointFinalizer) {
		controllerutil.AddFinalizer(envCheckpoint, envCheckpointFinalizer)
		if err := r.Update(ctx, envCheckpoint); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Skip terminal states
	if envCheckpoint.Status.Phase == environmentsv1.EnvironmentCheckpointPhaseCompleted ||
		envCheckpoint.Status.Phase == environmentsv1.EnvironmentCheckpointPhaseFailed {
		return reconcile.Result{}, nil
	}

	// Get the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: envCheckpoint.Spec.EnvironmentNamespace, Name: envCheckpoint.Spec.EnvironmentName}, env); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, envCheckpoint, "Environment not found", logger)
		}
		return reconcile.Result{}, err
	}

	switch envCheckpoint.Status.Phase {
	case "", environmentsv1.EnvironmentCheckpointPhasePending:
		return r.handlePending(ctx, envCheckpoint, env, logger)
	case environmentsv1.EnvironmentCheckpointPhaseStoppingWorkloads:
		return r.handleStoppingWorkloads(ctx, envCheckpoint, env, logger)
	case environmentsv1.EnvironmentCheckpointPhaseWaitingForPods:
		return r.handleWaitingForPods(ctx, envCheckpoint, env, logger)
	case environmentsv1.EnvironmentCheckpointPhaseCreatingCheckpoint,
		environmentsv1.EnvironmentCheckpointPhaseUploadingCheckpoint:
		return r.handleCheckpointInProgress(ctx, envCheckpoint, env, logger)
	case environmentsv1.EnvironmentCheckpointPhaseRestoringEnvironment:
		return r.handleRestoringEnvironment(ctx, envCheckpoint, env, logger)
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentCheckpointReconciler) handlePending(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Starting checkpoint, saving environment state")

	req.Status.PreviousEnvironmentState = env.Status.State
	req.Status.StartTime = &metav1.Time{Time: time.Now()}
	req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseStoppingWorkloads
	req.Status.Message = "Stopping environment workloads..."

	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	// Set environment state to snapping
	env.Status.State = environmentsv1.EnvironmentStateSnapping
	env.Status.Message = "Creating checkpoint..."
	if err := r.Status().Update(ctx, env); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *EnvironmentCheckpointReconciler) handleStoppingWorkloads(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Scaling down environment workloads")

	envReconciler := &EnvironmentReconciler{Client: r.Client, Scheme: r.Scheme, Logger: r.Logger}
	if err := envReconciler.suspendEnvironment(ctx, env, logger); err != nil {
		logger.Error("Failed to suspend environment", zap.Error(err))
	}

	req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseWaitingForPods
	req.Status.Message = "Waiting for pods to terminate..."
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRequestRetryInterval}, nil
}

func (r *EnvironmentCheckpointReconciler) handleWaitingForPods(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	pods := &corev1.PodList{}
	if err := pagination.ListAll(ctx, r, pods, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		return reconcile.Result{}, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRequestRetryInterval}, nil
		}
	}

	logger.Info("All pods terminated, capturing artifacts and creating checkpoint")

	// Find node for this workmachine
	nodeName, err := getNodeForWorkMachine(ctx, r.Client, env.Spec.WorkMachineName)
	if err != nil {
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to find node: %v", err), logger)
	}

	// Capture K8s artifacts into status
	if err := r.captureArtifacts(ctx, req, env, logger); err != nil {
		logger.Warn("Failed to capture artifacts", zap.Error(err))
	}

	// Create the Checkpoint CR directly (worker will process it)
	sourcePath := fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", env.Spec.TargetNamespace)
	checkpoint := &checkpointv1.Checkpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.CheckpointName,
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":                env.Spec.OwnedBy,
				"checkpoints.kloudlite.io/environment": env.Name,
				"checkpoints.kloudlite.io/type":        "environment",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "environments.kloudlite.io/v1",
				Kind:       "Environment",
				Name:       env.Name,
				UID:        env.UID,
				Controller: boolPtr(true),
			}},
		},
		Spec: checkpointv1.CheckpointSpec{
			Owner:       env.Spec.OwnedBy,
			Description: req.Spec.Description,
			SourcePath:  sourcePath,
			NodeName:    nodeName,
		},
	}

	if req.Spec.RetentionDays > 0 {
		expiresAt := metav1.NewTime(time.Now().AddDate(0, 0, int(req.Spec.RetentionDays)))
		checkpoint.Spec.RetentionPolicy = &checkpointv1.RetentionPolicy{
			ExpiresAt:   &expiresAt,
			KeepForDays: &req.Spec.RetentionDays,
		}
	}

	if err := r.Create(ctx, checkpoint); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create Checkpoint: %v", err), logger)
		}
	}

	req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseCreatingCheckpoint
	req.Status.Message = "Creating btrfs snapshot..."
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRequestRetryInterval}, nil
}

func (r *EnvironmentCheckpointReconciler) handleCheckpointInProgress(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	checkpoint := &checkpointv1.Checkpoint{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      req.Spec.CheckpointName,
		Namespace: env.Spec.TargetNamespace,
	}, checkpoint); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, req, "Checkpoint not found", logger)
		}
		return reconcile.Result{}, err
	}

	switch checkpoint.Status.Phase {
	case checkpointv1.CheckpointPhaseReady:
		logger.Info("Checkpoint created successfully")
		req.Status.CreatedCheckpointName = checkpoint.Name
		req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseRestoringEnvironment
		req.Status.Message = "Restoring environment state..."
		if err := r.Status().Update(ctx, req); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil

	case checkpointv1.CheckpointPhaseFailed:
		return r.setFailed(ctx, req, fmt.Sprintf("Checkpoint failed: %s", checkpoint.Status.Message), logger)

	case checkpointv1.CheckpointPhaseUploading:
		if req.Status.Phase != environmentsv1.EnvironmentCheckpointPhaseUploadingCheckpoint {
			req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseUploadingCheckpoint
			req.Status.Message = "Uploading checkpoint to registry..."
			if err := r.Status().Update(ctx, req); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRequestRetryInterval}, nil
}

func (r *EnvironmentCheckpointReconciler) handleRestoringEnvironment(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Restoring environment to previous state")

	targetState := req.Status.PreviousEnvironmentState
	if targetState == "" || targetState == environmentsv1.EnvironmentStateSnapping {
		targetState = environmentsv1.EnvironmentStateActive
	}

	env.Status.State = targetState
	env.Status.Message = "Checkpoint created successfully"

	now := metav1.Now()
	env.Status.LastRestoredCheckpoint = &environmentsv1.LastRestoredCheckpointInfo{
		Name:       req.Status.CreatedCheckpointName,
		RestoredAt: now,
	}

	if err := r.Status().Update(ctx, env); err != nil {
		return reconcile.Result{}, err
	}

	req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseCompleted
	req.Status.Message = fmt.Sprintf("Checkpoint '%s' created successfully", req.Status.CreatedCheckpointName)
	req.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	logger.Info("Checkpoint request completed successfully")
	return reconcile.Result{}, nil
}

func (r *EnvironmentCheckpointReconciler) handleDeletion(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	logger *zap.Logger,
) (reconcile.Result, error) {
	// If checkpoint was in progress, restore environment state
	if req.Status.Phase != environmentsv1.EnvironmentCheckpointPhaseCompleted &&
		req.Status.Phase != environmentsv1.EnvironmentCheckpointPhaseFailed &&
		req.Status.Phase != "" {
		env := &environmentsv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Namespace: req.Spec.EnvironmentNamespace, Name: req.Spec.EnvironmentName}, env); err == nil {
			if env.Status.State == environmentsv1.EnvironmentStateSnapping {
				targetState := req.Status.PreviousEnvironmentState
				if targetState == "" || targetState == environmentsv1.EnvironmentStateSnapping {
					targetState = environmentsv1.EnvironmentStateActive
				}
				env.Status.State = targetState
				env.Status.Message = "Checkpoint request cancelled"
				if err := r.Status().Update(ctx, env); err != nil {
					logger.Warn("Failed to restore environment state on deletion", zap.Error(err))
				}
			}
		}
	}

	controllerutil.RemoveFinalizer(req, envCheckpointFinalizer)
	if err := r.Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentCheckpointReconciler) setFailed(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	message string,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Checkpoint request failed", zap.String("message", message))

	req.Status.Phase = environmentsv1.EnvironmentCheckpointPhaseFailed
	req.Status.Message = message
	req.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, req); err != nil {
		return reconcile.Result{}, err
	}

	// Try to restore environment state
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Spec.EnvironmentNamespace, Name: req.Spec.EnvironmentName}, env); err == nil {
		if env.Status.State == environmentsv1.EnvironmentStateSnapping {
			targetState := req.Status.PreviousEnvironmentState
			if targetState == "" || targetState == environmentsv1.EnvironmentStateSnapping {
				targetState = environmentsv1.EnvironmentStateActive
			}
			env.Status.State = targetState
			env.Status.Message = fmt.Sprintf("Checkpoint failed: %s", message)
			if err := r.Status().Update(ctx, env); err != nil {
				logger.Warn("Failed to restore environment state after failure", zap.Error(err))
			}
		}
	}

	return reconcile.Result{}, nil
}

// captureArtifacts captures K8s resources into the EnvironmentCheckpoint status
func (r *EnvironmentCheckpointReconciler) captureArtifacts(
	ctx context.Context,
	req *environmentsv1.EnvironmentCheckpoint,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) error {
	artifacts := &environmentsv1.CapturedArtifacts{}
	namespace := env.Spec.TargetNamespace

	// Capture Environment spec
	envSpecCopy := env.Spec.DeepCopy()
	envSpecCopy.TargetNamespace = ""
	envSpecCopy.WorkMachineName = ""
	envSpecCopy.NodeName = ""

	envSpecData, err := json.Marshal(envSpecCopy)
	if err != nil {
		logger.Warn("Failed to marshal environment spec", zap.Error(err))
	} else {
		artifacts.EnvironmentSpec = base64.StdEncoding.EncodeToString(envSpecData)
	}

	// Capture ConfigMaps (excluding system ones)
	configMaps := &corev1.ConfigMapList{}
	if err := pagination.ListAll(ctx, r, configMaps, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list configmaps: %w", err)
	}

	var userConfigMaps []corev1.ConfigMap
	for _, cm := range configMaps.Items {
		if cm.Name == "kube-root-ca.crt" {
			continue
		}
		userConfigMaps = append(userConfigMaps, corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
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
		artifacts.ConfigMaps = base64.StdEncoding.EncodeToString(data)
		artifacts.ConfigMapCount = int32(len(userConfigMaps))
		logger.Info("Captured configmaps", zap.Int("count", len(userConfigMaps)))
	}

	// Capture Secrets (excluding service account tokens and system secrets)
	secrets := &corev1.SecretList{}
	if err := pagination.ListAll(ctx, r, secrets, client.InNamespace(namespace)); err != nil {
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
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
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
		artifacts.Secrets = base64.StdEncoding.EncodeToString(data)
		artifacts.SecretCount = int32(len(userSecrets))
		logger.Info("Captured secrets", zap.Int("count", len(userSecrets)))
	}

	req.Status.Artifacts = artifacts
	return nil
}

// getNodeForWorkMachine finds the k8s node for a workmachine by label
func getNodeForWorkMachine(ctx context.Context, c client.Client, workmachineName string) (string, error) {
	var nodes corev1.NodeList
	if err := pagination.ListAll(ctx, c, &nodes, client.MatchingLabels{
		"kloudlite.io/workmachine": workmachineName,
	}); err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no node found for workmachine %s", workmachineName)
	}
	return nodes.Items[0].Name, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *EnvironmentCheckpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.EnvironmentCheckpoint{}).
		Complete(r)
}
