package environment

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
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

const envCheckpointRestoreFinalizer = "environments.kloudlite.io/checkpoint-restore-finalizer"

// EnvironmentCheckpointRestoreReconciler reconciles EnvironmentCheckpointRestore objects
type EnvironmentCheckpointRestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
	Cfg    *controllerconfig.ControllerConfig
}

func (r *EnvironmentCheckpointRestoreReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(zap.String("envCheckpointRestore", req.Name))

	envRestore := &environmentsv1.EnvironmentCheckpointRestore{}
	if err := r.Get(ctx, req.NamespacedName, envRestore); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Handle deletion
	if envRestore.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, envRestore, logger)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(envRestore, envCheckpointRestoreFinalizer) {
		controllerutil.AddFinalizer(envRestore, envCheckpointRestoreFinalizer)
		if err := r.Update(ctx, envRestore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Skip terminal states
	if envRestore.Status.Phase == environmentsv1.EnvironmentCheckpointRestorePhaseCompleted ||
		envRestore.Status.Phase == environmentsv1.EnvironmentCheckpointRestorePhaseFailed {
		return reconcile.Result{}, nil
	}

	// Get the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: envRestore.Spec.EnvironmentNamespace, Name: envRestore.Spec.EnvironmentName}, env); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, envRestore, nil, "Environment not found", logger)
		}
		return reconcile.Result{}, err
	}

	switch envRestore.Status.Phase {
	case "", environmentsv1.EnvironmentCheckpointRestorePhasePending:
		return r.handlePending(ctx, envRestore, env, logger)
	case environmentsv1.EnvironmentCheckpointRestorePhaseStoppingWorkloads:
		return r.handleStoppingWorkloads(ctx, envRestore, env, logger)
	case environmentsv1.EnvironmentCheckpointRestorePhaseWaitingForPods:
		return r.handleWaitingForPods(ctx, envRestore, env, logger)
	case environmentsv1.EnvironmentCheckpointRestorePhaseDownloading,
		environmentsv1.EnvironmentCheckpointRestorePhaseRestoringData:
		return r.handleRestoreInProgress(ctx, envRestore, env, logger)
	case environmentsv1.EnvironmentCheckpointRestorePhaseApplyingArtifacts:
		return r.handleApplyingArtifacts(ctx, envRestore, env, logger)
	case environmentsv1.EnvironmentCheckpointRestorePhaseActivating:
		return r.handleActivating(ctx, envRestore, env, logger)
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handlePending(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Starting checkpoint restore, validating checkpoint")

	// Verify checkpoint exists and is ready
	checkpoint := &checkpointv1.Checkpoint{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.CheckpointName, Namespace: restore.Spec.SourceNamespace}, checkpoint); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, restore, env, fmt.Sprintf("Checkpoint %s not found in namespace %s", restore.Spec.CheckpointName, restore.Spec.SourceNamespace), logger)
		}
		return reconcile.Result{}, err
	}

	if checkpoint.Status.Phase != checkpointv1.CheckpointPhaseReady {
		restore.Status.Message = fmt.Sprintf("Waiting for checkpoint to be ready (phase: %s)", checkpoint.Status.Phase)
		if err := r.Status().Update(ctx, restore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRestoreRetryInterval}, nil
	}

	// Verify workmachine is ready
	if env.Spec.WorkMachineName == "" {
		return r.setFailed(ctx, restore, env, "Environment has no workmachine assigned", logger)
	}

	if _, err := getNodeForWorkMachine(ctx, r.Client, env.Spec.WorkMachineName); err != nil {
		restore.Status.Message = "Waiting for workmachine to be ready"
		if err := r.Status().Update(ctx, restore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRestoreRetryInterval}, nil
	}

	restore.Status.StartTime = &metav1.Time{Time: time.Now()}
	restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseStoppingWorkloads
	restore.Status.Message = "Stopping environment workloads..."

	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	env.Status.State = environmentsv1.EnvironmentStateSnapping
	env.Status.Message = "Restoring from checkpoint..."
	if err := r.Status().Update(ctx, env); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handleStoppingWorkloads(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Scaling down environment workloads for restore")

	envReconciler := &EnvironmentReconciler{Client: r.Client, Scheme: r.Scheme, Logger: r.Logger}
	if err := envReconciler.suspendEnvironment(ctx, env, logger); err != nil {
		logger.Error("Failed to suspend environment", zap.Error(err))
	}

	restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseWaitingForPods
	restore.Status.Message = "Waiting for pods to terminate..."
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRestoreRetryInterval}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handleWaitingForPods(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	pods := &corev1.PodList{}
	if err := pagination.ListAll(ctx, r, pods, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		return reconcile.Result{}, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRestoreRetryInterval}, nil
		}
	}

	logger.Info("All pods terminated, creating checkpoint restore")

	nodeName, err := getNodeForWorkMachine(ctx, r.Client, env.Spec.WorkMachineName)
	if err != nil {
		return r.setFailed(ctx, restore, env, fmt.Sprintf("Failed to find node: %v", err), logger)
	}

	// Create the CheckpointRestore CR
	cpSuffix := restore.Spec.CheckpointName
	if len(cpSuffix) > 8 {
		cpSuffix = cpSuffix[len(cpSuffix)-8:]
	}
	cpSuffix = strings.Trim(cpSuffix, "-")
	checkpointRestoreName := fmt.Sprintf("env-restore-%s-%s", env.Name, cpSuffix)
	targetPath := fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", env.Spec.TargetNamespace)

	cpRestore := &checkpointv1.CheckpointRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      checkpointRestoreName,
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":                env.Spec.OwnedBy,
				"checkpoints.kloudlite.io/environment": env.Name,
				"checkpoints.kloudlite.io/source":      restore.Spec.CheckpointName,
			},
		},
		Spec: checkpointv1.CheckpointRestoreSpec{
			CheckpointName: restore.Spec.CheckpointName,
			TargetPath:     targetPath,
			NodeName:       nodeName,
		},
	}

	if err := r.Create(ctx, cpRestore); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, restore, env, fmt.Sprintf("Failed to create CheckpointRestore: %v", err), logger)
		}
	}

	restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseDownloading
	restore.Status.Message = "Downloading checkpoint from registry..."
	restore.Status.CheckpointRestoreName = checkpointRestoreName
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRestoreRetryInterval}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handleRestoreInProgress(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	cpRestore := &checkpointv1.CheckpointRestore{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      restore.Status.CheckpointRestoreName,
		Namespace: env.Spec.TargetNamespace,
	}, cpRestore); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, restore, env, "CheckpointRestore not found", logger)
		}
		return reconcile.Result{}, err
	}

	switch cpRestore.Status.Phase {
	case checkpointv1.CheckpointRestorePhaseCompleted:
		logger.Info("Checkpoint data restore completed, applying artifacts")
		restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseApplyingArtifacts
		restore.Status.Message = "Applying K8s artifacts from checkpoint..."
		if err := r.Status().Update(ctx, restore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil

	case checkpointv1.CheckpointRestorePhaseFailed:
		return r.setFailed(ctx, restore, env, fmt.Sprintf("CheckpointRestore failed: %s", cpRestore.Status.Message), logger)

	case checkpointv1.CheckpointRestorePhaseRestoring:
		if restore.Status.Phase != environmentsv1.EnvironmentCheckpointRestorePhaseRestoringData {
			restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseRestoringData
			restore.Status.Message = "Restoring checkpoint data..."
			if err := r.Status().Update(ctx, restore); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.SnapshotRestoreRetryInterval}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handleApplyingArtifacts(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Applying checkpoint artifacts")

	// Read artifacts from the source EnvironmentCheckpoint's status
	envCheckpointList := &environmentsv1.EnvironmentCheckpointList{}
	if err := pagination.ListAll(ctx, r, envCheckpointList, client.InNamespace(restore.Spec.SourceNamespace)); err != nil {
		logger.Warn("Failed to list EnvironmentCheckpoints", zap.Error(err))
	}

	// Find the EnvironmentCheckpoint that created this checkpoint
	var sourceArtifacts *environmentsv1.CapturedArtifacts
	for i := range envCheckpointList.Items {
		ec := &envCheckpointList.Items[i]
		if ec.Spec.CheckpointName == restore.Spec.CheckpointName &&
			ec.Status.Phase == environmentsv1.EnvironmentCheckpointPhaseCompleted {
			sourceArtifacts = ec.Status.Artifacts
			break
		}
	}

	if sourceArtifacts != nil {
		restoredInfo, err := r.applyArtifacts(ctx, sourceArtifacts, env, logger)
		if err != nil {
			logger.Error("Failed to apply artifacts", zap.Error(err))
			return reconcile.Result{}, fmt.Errorf("failed to apply artifacts: %w", err)
		}
		restore.Status.RestoredArtifacts = restoredInfo
	} else {
		logger.Warn("No source EnvironmentCheckpoint found with artifacts, skipping artifact application")
	}

	// Update LastRestoredCheckpoint on environment
	now := metav1.Now()
	env.Status.LastRestoredCheckpoint = &environmentsv1.LastRestoredCheckpointInfo{
		Name:       restore.Spec.CheckpointName,
		RestoredAt: now,
	}

	restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseActivating
	restore.Status.Message = "Activating environment..."
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handleActivating(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Activating environment after restore")

	targetState := environmentsv1.EnvironmentStateInactive
	if restore.Spec.ActivateAfterRestore {
		targetState = environmentsv1.EnvironmentStateActive
	}

	env.Status.State = targetState
	env.Status.Message = fmt.Sprintf("Restored from checkpoint %s", restore.Spec.CheckpointName)

	if err := r.Status().Update(ctx, env); err != nil {
		return reconcile.Result{}, err
	}

	restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseCompleted
	restore.Status.Message = fmt.Sprintf("Successfully restored from checkpoint '%s'", restore.Spec.CheckpointName)
	restore.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	logger.Info("Checkpoint restore completed successfully")
	return reconcile.Result{}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) handleDeletion(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	logger *zap.Logger,
) (reconcile.Result, error) {
	if restore.Status.Phase != environmentsv1.EnvironmentCheckpointRestorePhaseCompleted &&
		restore.Status.Phase != environmentsv1.EnvironmentCheckpointRestorePhaseFailed &&
		restore.Status.Phase != "" {
		env := &environmentsv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Namespace: restore.Spec.EnvironmentNamespace, Name: restore.Spec.EnvironmentName}, env); err == nil {
			if env.Status.State == environmentsv1.EnvironmentStateSnapping {
				env.Status.State = environmentsv1.EnvironmentStateInactive
				env.Status.Message = "Checkpoint restore cancelled"
				if err := r.Status().Update(ctx, env); err != nil {
					logger.Error("Failed to restore environment state on deletion", zap.Error(err))
					return reconcile.Result{}, fmt.Errorf("failed to restore environment state: %w", err)
				}
			}
		}
	}

	controllerutil.RemoveFinalizer(restore, envCheckpointRestoreFinalizer)
	if err := r.Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) setFailed(
	ctx context.Context,
	restore *environmentsv1.EnvironmentCheckpointRestore,
	env *environmentsv1.Environment,
	message string,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Checkpoint restore failed", zap.String("message", message))

	restore.Status.Phase = environmentsv1.EnvironmentCheckpointRestorePhaseFailed
	restore.Status.Message = message
	restore.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	if env != nil {
		if env.Status.State == environmentsv1.EnvironmentStateSnapping {
			env.Status.State = environmentsv1.EnvironmentStateInactive
			env.Status.Message = fmt.Sprintf("Checkpoint restore failed: %s", message)
			if err := r.Status().Update(ctx, env); err != nil {
				logger.Error("Failed to restore environment state after failure", zap.Error(err))
			}
		}
	}

	return reconcile.Result{}, nil
}

// applyArtifacts applies captured K8s artifacts (ConfigMaps, Secrets) to the environment namespace
func (r *EnvironmentCheckpointRestoreReconciler) applyArtifacts(
	ctx context.Context,
	artifacts *environmentsv1.CapturedArtifacts,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (*environmentsv1.RestoredArtifactsInfo, error) {
	info := &environmentsv1.RestoredArtifactsInfo{}
	namespace := env.Spec.TargetNamespace

	// Apply ConfigMaps
	if artifacts.ConfigMaps != "" {
		data, err := base64.StdEncoding.DecodeString(artifacts.ConfigMaps)
		if err != nil {
			return info, fmt.Errorf("failed to decode configmaps: %w", err)
		}

		var configMaps []corev1.ConfigMap
		if err := yaml.Unmarshal(data, &configMaps); err != nil {
			return info, fmt.Errorf("failed to unmarshal configmaps: %w", err)
		}

		for _, cm := range configMaps {
			cm.Namespace = namespace
			existing := &corev1.ConfigMap{}
			if err := r.Get(ctx, client.ObjectKey{Name: cm.Name, Namespace: namespace}, existing); err != nil {
				if apierrors.IsNotFound(err) {
					if err := r.Create(ctx, &cm); err != nil {
						logger.Warn("Failed to create configmap", zap.String("name", cm.Name), zap.Error(err))
						continue
					}
				} else {
					logger.Warn("Failed to get configmap", zap.String("name", cm.Name), zap.Error(err))
					continue
				}
			} else {
				existing.Data = cm.Data
				existing.BinaryData = cm.BinaryData
				if err := r.Update(ctx, existing); err != nil {
					logger.Warn("Failed to update configmap", zap.String("name", cm.Name), zap.Error(err))
					continue
				}
			}
			info.ConfigMapsRestored++
		}
		logger.Info("Restored configmaps", zap.Int32("count", info.ConfigMapsRestored))
	}

	// Apply Secrets
	if artifacts.Secrets != "" {
		data, err := base64.StdEncoding.DecodeString(artifacts.Secrets)
		if err != nil {
			return info, fmt.Errorf("failed to decode secrets: %w", err)
		}

		var secrets []corev1.Secret
		if err := yaml.Unmarshal(data, &secrets); err != nil {
			return info, fmt.Errorf("failed to unmarshal secrets: %w", err)
		}

		for _, secret := range secrets {
			secret.Namespace = namespace
			existing := &corev1.Secret{}
			if err := r.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: namespace}, existing); err != nil {
				if apierrors.IsNotFound(err) {
					if err := r.Create(ctx, &secret); err != nil {
						logger.Warn("Failed to create secret", zap.String("name", secret.Name), zap.Error(err))
						continue
					}
				} else {
					logger.Warn("Failed to get secret", zap.String("name", secret.Name), zap.Error(err))
					continue
				}
			} else {
				existing.Data = secret.Data
				if err := r.Update(ctx, existing); err != nil {
					logger.Warn("Failed to update secret", zap.String("name", secret.Name), zap.Error(err))
					continue
				}
			}
			info.SecretsRestored++
		}
		logger.Info("Restored secrets", zap.Int32("count", info.SecretsRestored))
	}

	return info, nil
}

func (r *EnvironmentCheckpointRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.EnvironmentCheckpointRestore{}).
		Complete(r)
}
