package environment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kloudlite/kloudlite/api/internal/controllerconfig"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	envForkRequestFinalizer = "environments.kloudlite.io/fork-request-finalizer"
)

// EnvironmentForkRequestReconciler reconciles EnvironmentForkRequest objects
type EnvironmentForkRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
	Cfg    *controllerconfig.ControllerConfig // Controller configuration
}

// Reconcile handles EnvironmentForkRequest events
func (r *EnvironmentForkRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(zap.String("envForkRequest", req.Name), zap.String("namespace", req.Namespace))
	logger.Info("Reconciling EnvironmentForkRequest")

	// Fetch the EnvironmentForkRequest
	forkReq := &environmentsv1.EnvironmentForkRequest{}
	if err := r.Get(ctx, req.NamespacedName, forkReq); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Handle deletion
	if forkReq.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, forkReq, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(forkReq, envForkRequestFinalizer) {
		controllerutil.AddFinalizer(forkReq, envForkRequestFinalizer)
		if err := r.Update(ctx, forkReq); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Skip if already completed or failed
	if forkReq.Status.Phase == environmentsv1.EnvironmentForkRequestPhaseCompleted ||
		forkReq.Status.Phase == environmentsv1.EnvironmentForkRequestPhaseFailed {
		return reconcile.Result{}, nil
	}

	// Process based on current phase
	switch forkReq.Status.Phase {
	case "", environmentsv1.EnvironmentForkRequestPhasePending:
		return r.handlePending(ctx, forkReq, logger)

	case environmentsv1.EnvironmentForkRequestPhaseValidating:
		return r.handleValidating(ctx, forkReq, logger)

	case environmentsv1.EnvironmentForkRequestPhaseCreatingEnvironment:
		return r.handleCreatingEnvironment(ctx, forkReq, logger)

	case environmentsv1.EnvironmentForkRequestPhaseWaitingForRestore:
		return r.handleWaitingForRestore(ctx, forkReq, logger)
	}

	return reconcile.Result{}, nil
}

// handlePending starts the fork request processing
func (r *EnvironmentForkRequestReconciler) handlePending(ctx context.Context, forkReq *environmentsv1.EnvironmentForkRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Starting fork request processing",
		zap.String("newEnvName", forkReq.Spec.NewEnvironmentName),
		zap.String("snapshotName", forkReq.Spec.SourceSnapshot.SnapshotName),
		zap.String("sourceNamespace", forkReq.Spec.SourceSnapshot.SourceNamespace))

	now := metav1.Now()
	forkReq.Status.Phase = environmentsv1.EnvironmentForkRequestPhaseValidating
	forkReq.Status.Message = "Validating snapshot and artifacts"
	forkReq.Status.StartTime = &now

	if err := r.Status().Update(ctx, forkReq); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleValidating validates that the snapshot and artifacts exist
func (r *EnvironmentForkRequestReconciler) handleValidating(ctx context.Context, forkReq *environmentsv1.EnvironmentForkRequest, logger *zap.Logger) (reconcile.Result, error) {
	// Check if snapshot exists
	snapshot := &snapshotv1.Snapshot{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      forkReq.Spec.SourceSnapshot.SnapshotName,
		Namespace: forkReq.Spec.SourceSnapshot.SourceNamespace,
	}, snapshot); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Snapshot %q not found in namespace %q",
				forkReq.Spec.SourceSnapshot.SnapshotName, forkReq.Spec.SourceSnapshot.SourceNamespace), logger)
		}
		return reconcile.Result{}, err
	}

	// Check if snapshot is ready
	if snapshot.Status.State != snapshotv1.SnapshotStateReady {
		forkReq.Status.Message = fmt.Sprintf("Waiting for snapshot to be ready (current: %s)", snapshot.Status.State)
		if err := r.Status().Update(ctx, forkReq); err != nil {
			if !apierrors.IsConflict(err) {
				logger.Error("Failed to update status", zap.Error(err))
			}
		}
		return reconcile.Result{RequeueAfter: r.Cfg.Environment.ForkRetryInterval}, nil
	}

	// Check if SnapshotArtifacts exists
	artifacts := &snapshotv1.SnapshotArtifacts{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      forkReq.Spec.SourceSnapshot.SnapshotName,
		Namespace: forkReq.Spec.SourceSnapshot.SourceNamespace,
	}, artifacts); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("SnapshotArtifacts %q not found",
				forkReq.Spec.SourceSnapshot.SnapshotName), logger)
		}
		return reconcile.Result{}, err
	}

	// Check if EnvironmentSpec is present
	if artifacts.Spec.EnvironmentSpec == "" {
		// Fall back to ComposeSpec for backward compatibility
		if artifacts.Spec.ComposeSpec == "" {
			return r.setFailed(ctx, forkReq, "SnapshotArtifacts has no EnvironmentSpec or ComposeSpec", logger)
		}
		logger.Info("Using legacy ComposeSpec (EnvironmentSpec not available)")
	}

	// Check if target environment already exists
	existingEnv := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      forkReq.Spec.NewEnvironmentName,
		Namespace: forkReq.Namespace,
	}, existingEnv); err == nil {
		return r.setFailed(ctx, forkReq, fmt.Sprintf("Environment %q already exists", forkReq.Spec.NewEnvironmentName), logger)
	} else if !apierrors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	// Validation passed, move to creating environment
	forkReq.Status.Phase = environmentsv1.EnvironmentForkRequestPhaseCreatingEnvironment
	forkReq.Status.Message = "Creating new environment from snapshot"

	if err := r.Status().Update(ctx, forkReq); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	logger.Info("Validation passed, proceeding to create environment")
	return reconcile.Result{Requeue: true}, nil
}

// handleCreatingEnvironment creates the new environment from snapshot metadata
func (r *EnvironmentForkRequestReconciler) handleCreatingEnvironment(ctx context.Context, forkReq *environmentsv1.EnvironmentForkRequest, logger *zap.Logger) (reconcile.Result, error) {
	// Get the SnapshotArtifacts
	artifacts := &snapshotv1.SnapshotArtifacts{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      forkReq.Spec.SourceSnapshot.SnapshotName,
		Namespace: forkReq.Spec.SourceSnapshot.SourceNamespace,
	}, artifacts); err != nil {
		return r.setFailed(ctx, forkReq, fmt.Sprintf("Failed to get SnapshotArtifacts: %v", err), logger)
	}

	var newEnvSpec environmentsv1.EnvironmentSpec

	// Try to decode EnvironmentSpec first (preferred)
	if artifacts.Spec.EnvironmentSpec != "" {
		specData, err := base64.StdEncoding.DecodeString(artifacts.Spec.EnvironmentSpec)
		if err != nil {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Failed to decode EnvironmentSpec: %v", err), logger)
		}

		if err := json.Unmarshal(specData, &newEnvSpec); err != nil {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Failed to unmarshal EnvironmentSpec: %v", err), logger)
		}
		logger.Info("Decoded EnvironmentSpec from snapshot")
	} else if artifacts.Spec.ComposeSpec != "" {
		// Fall back to ComposeSpec for backward compatibility
		composeData, err := base64.StdEncoding.DecodeString(artifacts.Spec.ComposeSpec)
		if err != nil {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Failed to decode ComposeSpec: %v", err), logger)
		}

		var composeSpec environmentsv1.CompositionSpec
		if err := json.Unmarshal(composeData, &composeSpec); err != nil {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Failed to unmarshal ComposeSpec: %v", err), logger)
		}

		// Create minimal spec with just compose
		newEnvSpec = environmentsv1.EnvironmentSpec{
			Compose:   &composeSpec,
			Activated: true,
		}
		logger.Info("Using legacy ComposeSpec to create environment")
	} else {
		return r.setFailed(ctx, forkReq, "No EnvironmentSpec or ComposeSpec in artifacts", logger)
	}

	// Clear fields that should be regenerated
	newEnvSpec.TargetNamespace = "" // Will be auto-generated by webhook
	newEnvSpec.FromSnapshot = nil   // Will be set below
	newEnvSpec.WorkMachineName = "" // Will be derived from namespace
	newEnvSpec.NodeName = ""        // Will be derived from workmachine

	// Apply overrides if provided
	if forkReq.Spec.Overrides != nil {
		r.applyOverrides(&newEnvSpec, forkReq.Spec.Overrides)
		logger.Info("Applied spec overrides")
	}

	// Set fromSnapshot to trigger data restore
	newEnvSpec.FromSnapshot = &environmentsv1.FromSnapshotRef{
		SnapshotName:    forkReq.Spec.SourceSnapshot.SnapshotName,
		SourceNamespace: forkReq.Spec.SourceSnapshot.SourceNamespace,
	}

	// Ensure environment is activated
	newEnvSpec.Activated = true

	// Create the new environment
	newEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forkReq.Spec.NewEnvironmentName,
			Namespace: forkReq.Namespace, // Same namespace as fork request (wm-{user})
			Labels: map[string]string{
				"kloudlite.io/forked-from-snapshot": forkReq.Spec.SourceSnapshot.SnapshotName,
			},
			Annotations: map[string]string{
				"kloudlite.io/fork-request": forkReq.Name,
			},
		},
		Spec: newEnvSpec,
	}

	if err := r.Create(ctx, newEnv); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Environment %q already exists", forkReq.Spec.NewEnvironmentName), logger)
		}
		return r.setFailed(ctx, forkReq, fmt.Sprintf("Failed to create environment: %v", err), logger)
	}

	logger.Info("Created new environment from snapshot",
		zap.String("envName", newEnv.Name),
		zap.String("namespace", newEnv.Namespace))

	// Transition to WaitingForRestore
	forkReq.Status.Phase = environmentsv1.EnvironmentForkRequestPhaseWaitingForRestore
	forkReq.Status.Message = "Waiting for snapshot restore to complete"
	forkReq.Status.CreatedEnvironment = newEnv.Name

	if err := r.Status().Update(ctx, forkReq); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.Cfg.Environment.ForkRetryInterval}, nil
}

// handleWaitingForRestore waits for the environment's snapshot restore to complete
func (r *EnvironmentForkRequestReconciler) handleWaitingForRestore(ctx context.Context, forkReq *environmentsv1.EnvironmentForkRequest, logger *zap.Logger) (reconcile.Result, error) {
	// Get the created environment
	env := &environmentsv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      forkReq.Status.CreatedEnvironment,
		Namespace: forkReq.Namespace,
	}, env); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, forkReq, "Created environment was deleted", logger)
		}
		return reconcile.Result{}, err
	}

	// Check snapshot restore status
	if env.Status.SnapshotRestoreStatus != nil {
		switch env.Status.SnapshotRestoreStatus.Phase {
		case environmentsv1.SnapshotRestorePhaseCompleted:
			// Restore completed, mark fork as completed
			now := metav1.Now()
			forkReq.Status.Phase = environmentsv1.EnvironmentForkRequestPhaseCompleted
			forkReq.Status.Message = fmt.Sprintf("Environment %q created and restored successfully", forkReq.Status.CreatedEnvironment)
			forkReq.Status.CompletionTime = &now

			if err := r.Status().Update(ctx, forkReq); err != nil {
				if apierrors.IsConflict(err) {
					return reconcile.Result{Requeue: true}, nil
				}
				return reconcile.Result{}, err
			}

			logger.Info("Fork completed successfully",
				zap.String("createdEnv", forkReq.Status.CreatedEnvironment))
			return reconcile.Result{}, nil

		case environmentsv1.SnapshotRestorePhaseFailed:
			return r.setFailed(ctx, forkReq, fmt.Sprintf("Snapshot restore failed: %s", env.Status.SnapshotRestoreStatus.Message), logger)

		default:
			// Still in progress, update message
			forkReq.Status.Message = fmt.Sprintf("Restore in progress: %s", env.Status.SnapshotRestoreStatus.Phase)
			if err := r.Status().Update(ctx, forkReq); err != nil {
				if !apierrors.IsConflict(err) {
					logger.Warn("Failed to update status", zap.Error(err))
				}
			}
		}
	}

	// Requeue to check again
	return reconcile.Result{RequeueAfter: r.Cfg.Environment.ForkRetryInterval}, nil
}

// applyOverrides applies the override values to the environment spec
func (r *EnvironmentForkRequestReconciler) applyOverrides(spec *environmentsv1.EnvironmentSpec, overrides *environmentsv1.EnvironmentSpecOverrides) {
	if overrides.Visibility != "" {
		spec.Visibility = overrides.Visibility
	}

	if overrides.OwnedBy != "" {
		spec.OwnedBy = overrides.OwnedBy
	}

	if overrides.ResourceQuotas != nil {
		spec.ResourceQuotas = overrides.ResourceQuotas
	}

	// Merge labels
	if len(overrides.Labels) > 0 {
		if spec.Labels == nil {
			spec.Labels = make(map[string]string)
		}
		for k, v := range overrides.Labels {
			spec.Labels[k] = v
		}
	}

	// Merge annotations
	if len(overrides.Annotations) > 0 {
		if spec.Annotations == nil {
			spec.Annotations = make(map[string]string)
		}
		for k, v := range overrides.Annotations {
			spec.Annotations[k] = v
		}
	}
}

// handleDeletion cleans up when the fork request is deleted
func (r *EnvironmentForkRequestReconciler) handleDeletion(ctx context.Context, forkReq *environmentsv1.EnvironmentForkRequest, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(forkReq, envForkRequestFinalizer) {
		return reconcile.Result{}, nil
	}

	// Note: We don't delete the created environment when the fork request is deleted
	// The environment is independent once created
	logger.Info("Fork request being deleted",
		zap.String("createdEnv", forkReq.Status.CreatedEnvironment))

	// Remove finalizer
	controllerutil.RemoveFinalizer(forkReq, envForkRequestFinalizer)
	if err := r.Update(ctx, forkReq); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// setFailed updates the fork request to Failed state
func (r *EnvironmentForkRequestReconciler) setFailed(ctx context.Context, forkReq *environmentsv1.EnvironmentForkRequest, message string, logger *zap.Logger) (reconcile.Result, error) {
	logger.Error("Fork request failed", zap.String("message", message))

	now := metav1.Now()
	forkReq.Status.Phase = environmentsv1.EnvironmentForkRequestPhaseFailed
	forkReq.Status.Message = message
	forkReq.Status.CompletionTime = &now

	if err := r.Status().Update(ctx, forkReq); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentForkRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.EnvironmentForkRequest{}).
		Complete(r)
}
