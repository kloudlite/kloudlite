package environment

import (
	"context"
	"fmt"
	"strings"
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
)

const (
	envSnapshotRestoreFinalizer = "environments.kloudlite.io/snapshot-restore-finalizer"
)

// EnvironmentSnapshotRestoreReconciler reconciles EnvironmentSnapshotRestore objects
type EnvironmentSnapshotRestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles EnvironmentSnapshotRestore events
func (r *EnvironmentSnapshotRestoreReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(zap.String("envSnapshotRestore", req.Name))
	logger.Info("Reconciling EnvironmentSnapshotRestore")

	// Fetch the EnvironmentSnapshotRestore
	envRestore := &environmentsv1.EnvironmentSnapshotRestore{}
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

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(envRestore, envSnapshotRestoreFinalizer) {
		controllerutil.AddFinalizer(envRestore, envSnapshotRestoreFinalizer)
		if err := r.Update(ctx, envRestore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Skip if already completed or failed
	if envRestore.Status.Phase == environmentsv1.EnvironmentSnapshotRestorePhaseCompleted ||
		envRestore.Status.Phase == environmentsv1.EnvironmentSnapshotRestorePhaseFailed {
		return reconcile.Result{}, nil
	}

	// Get the environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{Name: envRestore.Spec.EnvironmentName}, env); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, envRestore, nil, "Environment not found", logger)
		}
		return reconcile.Result{}, err
	}

	// Process based on current phase
	switch envRestore.Status.Phase {
	case "", environmentsv1.EnvironmentSnapshotRestorePhasePending:
		return r.handlePending(ctx, envRestore, env, logger)

	case environmentsv1.EnvironmentSnapshotRestorePhaseStoppingWorkloads:
		return r.handleStoppingWorkloads(ctx, envRestore, env, logger)

	case environmentsv1.EnvironmentSnapshotRestorePhaseWaitingForPods:
		return r.handleWaitingForPods(ctx, envRestore, env, logger)

	case environmentsv1.EnvironmentSnapshotRestorePhaseDownloading,
		environmentsv1.EnvironmentSnapshotRestorePhaseRestoringData:
		return r.handleRestoreInProgress(ctx, envRestore, env, logger)

	case environmentsv1.EnvironmentSnapshotRestorePhaseApplyingArtifacts:
		return r.handleApplyingArtifacts(ctx, envRestore, env, logger)

	case environmentsv1.EnvironmentSnapshotRestorePhaseActivating:
		return r.handleActivating(ctx, envRestore, env, logger)
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handlePending(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Starting snapshot restore, validating snapshot")

	// Verify snapshot exists and is ready (snapshots are namespaced)
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.SnapshotName, Namespace: restore.Spec.SourceNamespace}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, restore, env, fmt.Sprintf("Snapshot %s not found in namespace %s", restore.Spec.SnapshotName, restore.Spec.SourceNamespace), logger)
		}
		return reconcile.Result{}, err
	}

	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		restore.Status.Message = fmt.Sprintf("Waiting for snapshot to be ready (state: %s)", snapshot.Status.State)
		if err := r.Status().Update(ctx, restore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Verify workmachine is ready
	if env.Spec.WorkMachineName == "" {
		return r.setFailed(ctx, restore, env, "Environment has no workmachine assigned", logger)
	}

	_, err := r.getNodeForWorkMachine(ctx, env.Spec.WorkMachineName)
	if err != nil {
		restore.Status.Message = "Waiting for workmachine to be ready"
		if err := r.Status().Update(ctx, restore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Update status to start stopping workloads
	restore.Status.StartTime = &metav1.Time{Time: time.Now()}
	restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseStoppingWorkloads
	restore.Status.Message = "Stopping environment workloads..."

	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	// Set environment state to indicate restore in progress
	env.Status.State = environmentsv1.EnvironmentStateSnapping // Reuse snapping state for now
	env.Status.Message = "Restoring from snapshot..."
	if err := r.Status().Update(ctx, env); err != nil {
		logger.Error("Failed to update environment state", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handleStoppingWorkloads(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Scaling down environment workloads for restore")

	// Use the existing suspendEnvironment logic
	envReconciler := &EnvironmentReconciler{Client: r.Client, Scheme: r.Scheme, Logger: r.Logger}
	if err := envReconciler.suspendEnvironment(ctx, env, logger); err != nil {
		logger.Error("Failed to suspend environment", zap.Error(err))
		// Continue anyway - some resources might already be scaled down
	}

	// Move to waiting for pods
	restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseWaitingForPods
	restore.Status.Message = "Waiting for pods to terminate..."
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handleWaitingForPods(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	// Check if all pods are terminated
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		return reconcile.Result{}, err
	}

	// Check for running pods (ignore completed/failed jobs)
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			logger.Debug("Pod still running", zap.String("pod", pod.Name), zap.String("phase", string(pod.Status.Phase)))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		}
	}

	logger.Info("All pods terminated, creating snapshot restore")

	// Find the node for this workmachine
	nodeName, err := r.getNodeForWorkMachine(ctx, env.Spec.WorkMachineName)
	if err != nil {
		return r.setFailed(ctx, restore, env, fmt.Sprintf("Failed to find node for WorkMachine: %v", err), logger)
	}

	// Create the SnapshotRestore CR
	// Use last 8 chars of snapshot name for uniqueness (e.g., "snap-v1", "snap-v2")
	// and remove leading/trailing hyphens to ensure valid DNS name
	snapshotSuffix := restore.Spec.SnapshotName
	if len(snapshotSuffix) > 8 {
		snapshotSuffix = snapshotSuffix[len(snapshotSuffix)-8:]
	}
	snapshotSuffix = strings.Trim(snapshotSuffix, "-")
	snapshotRestoreName := fmt.Sprintf("env-restore-%s-%s", env.Name, snapshotSuffix)
	targetPath := fmt.Sprintf("/var/lib/kloudlite/storage/environments/%s", env.Spec.TargetNamespace)

	snapshotRestore := &snapshotv1.SnapshotRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotRestoreName,
			Namespace: env.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/owned-by":              env.Spec.OwnedBy,
				"snapshots.kloudlite.io/environment": env.Name,
				"snapshots.kloudlite.io/source":      restore.Spec.SnapshotName,
			},
		},
		Spec: snapshotv1.SnapshotRestoreSpec{
			SnapshotName: restore.Spec.SnapshotName,
			TargetPath:   targetPath,
			NodeName:     nodeName,
		},
	}

	if err := r.Create(ctx, snapshotRestore); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, restore, env, fmt.Sprintf("Failed to create SnapshotRestore: %v", err), logger)
		}
	}

	// Update status
	restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseDownloading
	restore.Status.Message = "Downloading snapshot from registry..."
	restore.Status.SnapshotRestoreName = snapshotRestoreName
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handleRestoreInProgress(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	// Get the SnapshotRestore status
	snapshotRestore := &snapshotv1.SnapshotRestore{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      restore.Status.SnapshotRestoreName,
		Namespace: env.Spec.TargetNamespace,
	}, snapshotRestore); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, restore, env, "SnapshotRestore not found", logger)
		}
		return reconcile.Result{}, err
	}

	// Check SnapshotRestore state
	switch snapshotRestore.Status.State {
	case snapshotv1.SnapshotRestoreStateCompleted:
		logger.Info("Snapshot data restore completed, applying artifacts")
		restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseApplyingArtifacts
		restore.Status.Message = "Applying K8s artifacts from snapshot..."
		if err := r.Status().Update(ctx, restore); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil

	case snapshotv1.SnapshotRestoreStateFailed:
		return r.setFailed(ctx, restore, env, fmt.Sprintf("SnapshotRestore failed: %s", snapshotRestore.Status.Message), logger)

	case snapshotv1.SnapshotRestoreStateRestoring:
		if restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseRestoringData {
			restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseRestoringData
			restore.Status.Message = "Restoring snapshot data..."
			if err := r.Status().Update(ctx, restore); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handleApplyingArtifacts(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Applying snapshot artifacts")
	sourceNamespace := restore.Spec.SourceNamespace

	// Get the Snapshot to retrieve lineage (snapshots are namespaced)
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.SnapshotName, Namespace: sourceNamespace}, snapshot); err != nil {
		logger.Warn("Failed to get snapshot for lineage", zap.Error(err))
		// Continue anyway
	}

	// Use the existing apply artifacts logic from environment controller
	envReconciler := &EnvironmentReconciler{Client: r.Client, Scheme: r.Scheme, Logger: r.Logger}
	if err := envReconciler.applySnapshotArtifacts(ctx, restore.Spec.SnapshotName, sourceNamespace, env, logger); err != nil {
		logger.Warn("Failed to apply snapshot artifacts", zap.Error(err))
		// Don't fail the restore, just log the warning
	}

	// Track restored artifacts in status
	artifacts := &snapshotv1.SnapshotArtifacts{}
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.SnapshotName, Namespace: sourceNamespace}, artifacts); err == nil {
		restore.Status.RestoredArtifacts = &environmentsv1.RestoredArtifactsInfo{
			ConfigMapsRestored: artifacts.Status.ConfigMapCount,
			SecretsRestored:    artifacts.Status.SecretCount,
		}
	}

	// Build full lineage and clone snapshots to the target environment's namespace
	lineage := append(snapshot.Status.Lineage, restore.Spec.SnapshotName)
	if err := envReconciler.cloneSnapshotsForLineage(ctx, env, sourceNamespace, lineage, logger); err != nil {
		logger.Error("Failed to clone snapshots for lineage", zap.Error(err))
		// Continue anyway
	}

	// Update LastRestoredSnapshot on environment
	now := metav1.Now()
	env.Status.LastRestoredSnapshot = &environmentsv1.LastRestoredSnapshotInfo{
		Name:       restore.Spec.SnapshotName,
		RestoredAt: now,
		Lineage:    lineage,
	}

	// Move to activating phase
	restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseActivating
	restore.Status.Message = "Activating environment..."
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handleActivating(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Activating environment after restore")

	// Set environment state based on spec
	targetState := environmentsv1.EnvironmentStateInactive
	if restore.Spec.ActivateAfterRestore {
		targetState = environmentsv1.EnvironmentStateActive
	}

	env.Status.State = targetState
	env.Status.Message = fmt.Sprintf("Restored from snapshot %s", restore.Spec.SnapshotName)

	if err := r.Status().Update(ctx, env); err != nil {
		logger.Error("Failed to update environment state", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Mark restore as completed
	restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseCompleted
	restore.Status.Message = fmt.Sprintf("Successfully restored from snapshot '%s'", restore.Spec.SnapshotName)
	restore.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot restore completed successfully")
	return reconcile.Result{}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) handleDeletion(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Info("Handling deletion of EnvironmentSnapshotRestore")

	// If restore was in progress, try to restore environment state
	if restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseCompleted &&
		restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseFailed &&
		restore.Status.Phase != "" {
		env := &environmentsv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.EnvironmentName}, env); err == nil {
			if env.Status.State == environmentsv1.EnvironmentStateSnapping {
				env.Status.State = environmentsv1.EnvironmentStateInactive
				env.Status.Message = "Snapshot restore cancelled"
				if err := r.Status().Update(ctx, env); err != nil {
					logger.Warn("Failed to restore environment state on deletion", zap.Error(err))
				}
			}
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(restore, envSnapshotRestoreFinalizer)
	if err := r.Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentSnapshotRestoreReconciler) setFailed(
	ctx context.Context,
	restore *environmentsv1.EnvironmentSnapshotRestore,
	env *environmentsv1.Environment,
	message string,
	logger *zap.Logger,
) (reconcile.Result, error) {
	logger.Error("Snapshot restore failed", zap.String("message", message))

	restore.Status.Phase = environmentsv1.EnvironmentSnapshotRestorePhaseFailed
	restore.Status.Message = message
	restore.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, restore); err != nil {
		return reconcile.Result{}, err
	}

	// Try to restore environment state
	if env != nil {
		if env.Status.State == environmentsv1.EnvironmentStateSnapping {
			env.Status.State = environmentsv1.EnvironmentStateInactive
			env.Status.Message = fmt.Sprintf("Snapshot restore failed: %s", message)
			if err := r.Status().Update(ctx, env); err != nil {
				logger.Warn("Failed to restore environment state after failure", zap.Error(err))
			}
		}
	}

	return reconcile.Result{}, nil
}

// getNodeForWorkMachine finds the k8s node for a workmachine by label
func (r *EnvironmentSnapshotRestoreReconciler) getNodeForWorkMachine(ctx context.Context, workmachineName string) (string, error) {
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
func (r *EnvironmentSnapshotRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.EnvironmentSnapshotRestore{}).
		Complete(r)
}
