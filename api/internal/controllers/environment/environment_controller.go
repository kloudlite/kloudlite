package environment

import (
	"context"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	environmentFinalizer = "environments.kloudlite.io/finalizer"
	// Kind constants for owner references
	environmentKind = "Environment"
	workMachineKind = "WorkMachine"
)

// EnvironmentReconciler reconciles Environment objects and creates namespaces
type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles Environment events and ensures namespace exists
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("environment", req.Name),
	)

	logger.Info("Reconciling Environment")

	// Fetch the Environment instance (cluster-scoped)
	environment := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name}, environment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Environment has been deleted, nothing to do
			logger.Info("Environment not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Environment", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if environment is being deleted
	if environment.DeletionTimestamp != nil {
		logger.Info("Environment is being deleted, starting cleanup")
		return r.handleDeletion(ctx, environment, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(environment, environmentFinalizer) {
		logger.Info("Adding finalizer to environment")
		controllerutil.AddFinalizer(environment, environmentFinalizer)
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Set WorkMachine as owner if WorkMachineName is specified and owner reference not yet set
	if environment.Spec.WorkMachineName != "" {
		needsOwnerUpdate := true
		for _, ownerRef := range environment.OwnerReferences {
			if ownerRef.Kind == "WorkMachine" && ownerRef.Name == environment.Spec.WorkMachineName {
				needsOwnerUpdate = false
				break
			}
		}

		if needsOwnerUpdate {
			logger.Info("Setting WorkMachine as owner of Environment",
				zap.String("workmachine", environment.Spec.WorkMachineName))

			// Fetch WorkMachine to set as owner
			workmachine := &workmachinevl.WorkMachine{}
			if err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.WorkMachineName}, workmachine); err != nil {
				logger.Error("Failed to get WorkMachine for ownership",
					zap.String("workmachine", environment.Spec.WorkMachineName),
					zap.Error(err))
				// Don't fail reconciliation, just log the error
				// The ownership will be set on next reconciliation
			} else {
				// Set WorkMachine as owner for cascading deletion (without blockOwnerDeletion)
				// Note: TypeMeta isn't populated by controller-runtime's Get method,
				// so we set them explicitly using the GroupVersion constant
				blockOwnerDeletion := false
				ownerRef := metav1.OwnerReference{
					APIVersion:         workmachinevl.GroupVersion.String(),
					Kind:               workMachineKind,
					Name:               workmachine.Name,
					UID:                workmachine.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				}
				environment.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

				if err := r.Update(ctx, environment); err != nil {
					logger.Error("Failed to update Environment with owner reference", zap.Error(err))
					return reconcile.Result{}, err
				}
				logger.Info("Successfully set WorkMachine as owner of Environment")
				return reconcile.Result{Requeue: true}, nil
			}
		}
	}

	// Handle environment creation from snapshot if fromSnapshot is set
	// Snapshot restore takes precedence over normal environment reconciliation
	if environment.Spec.FromSnapshot != nil {
		logger.Info("Environment has fromSnapshot set, handling snapshot restore",
			zap.String("snapshotName", environment.Spec.FromSnapshot.SnapshotName))
		return r.handleSnapshotRestore(ctx, environment, logger)
	}

	// Note: TargetNamespace is always set by the mutation webhook.
	// The webhook generates it as "env-{name}" if not provided by the user.
	// The webhook also validates that the namespace doesn't already exist.
	// Controller's responsibility is to create the actual Kubernetes namespace resource.

	// Update hash and subdomain in status (computed from envName-owner)
	if err := r.updateHashAndSubdomain(ctx, environment, logger); err != nil {
		logger.Warn("Failed to update hash/subdomain in status", zap.Error(err))
		// Don't fail reconciliation, these are informational fields
	}

	// Check if namespace already exists and handle it
	namespaceExists, err := r.ensureNamespaceExists(ctx, environment, logger)
	if err != nil {
		return reconcile.Result{}, err
	}

	if namespaceExists {
		// Ensure NetworkPolicy is correctly configured based on visibility
		if err := r.ensureNetworkPolicy(ctx, environment, logger); err != nil {
			logger.Error("Failed to ensure network policy", zap.Error(err))
			// Don't fail reconciliation for network policy errors
		}

		// Reconcile compose deployment if spec.Compose is set
		if _, err := r.reconcileCompose(ctx, environment, logger); err != nil {
			logger.Error("Failed to reconcile compose", zap.Error(err))
			// Don't fail reconciliation for compose errors - status is updated
		}

		// Update environment status based on activation state
		desiredState := environmentsv1.EnvironmentStateInactive
		if environment.Spec.Activated {
			desiredState = environmentsv1.EnvironmentStateActive
		}

		// Detect activation/deactivation transitions
		currentState := environment.Status.State
		wasActive := currentState == environmentsv1.EnvironmentStateActive
		wasInactive := currentState == environmentsv1.EnvironmentStateInactive
		willBeActive := desiredState == environmentsv1.EnvironmentStateActive
		willBeInactive := desiredState == environmentsv1.EnvironmentStateInactive

		// Handle snapping state separately - scale down but don't go to inactive
		// Snapshot controller will set state back to active when done
		isSnapping := currentState == environmentsv1.EnvironmentStateSnapping
		if isSnapping {
			// Check if there are any in-progress snapshot requests for this environment
			// If not, transition back to active state (handles crash/restart scenarios)
			hasActiveSnapshotOp, err := r.hasActiveSnapshotOperation(ctx, environment.Name)
			if err != nil {
				logger.Warn("Failed to check for active snapshot operations", zap.Error(err))
			} else if !hasActiveSnapshotOp {
				logger.Info("No active snapshot operations, transitioning back to active state")
				if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateActive, "Ready", logger); err != nil {
					logger.Error("Failed to update environment state to active", zap.Error(err))
					return reconcile.Result{}, err
				}
				return reconcile.Result{Requeue: true}, nil
			}

			// Scale down all deployments for snapshot
			logger.Info("Scaling down deployments for snapshot")
			if err := r.suspendEnvironment(ctx, environment, logger); err != nil {
				logger.Error("Failed to scale down deployments", zap.Error(err))
				// Continue - pods may already be scaled down
			}

			// Wait for all pods to terminate
			if !r.waitForPodsTerminated(ctx, environment.Spec.TargetNamespace, logger) {
				logger.Info("Waiting for pods to terminate for snapshot")
				if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateSnapping, "Taking snapshot, waiting for pods to terminate...", logger); err != nil {
					logger.Warn("Failed to update snapping message", zap.Error(err))
				}
				return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
			}

			// Pods terminated - stay in snapping state, snapshot controller will handle the rest
			logger.Info("All pods terminated, ready for snapshot")
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateSnapping, "Ready for snapshot", logger); err != nil {
				logger.Warn("Failed to update snapping message", zap.Error(err))
			}
			return reconcile.Result{}, nil
		}

		// Handle deactivation transition
		isDeactivating := currentState == environmentsv1.EnvironmentStateDeactivating
		if (wasActive && willBeInactive) || isDeactivating {
			// Set to deactivating state first
			if !isDeactivating {
				if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateDeactivating, "Environment is being deactivated", logger); err != nil {
					logger.Error("Failed to update status to deactivating", zap.Error(err))
				}
				return reconcile.Result{Requeue: true}, nil
			}

			// Environment is being deactivated - disconnect workspaces and remove service intercepts
			logger.Info("Environment is being deactivated, cleaning up connections")
			if err := r.handleEnvironmentDeactivation(ctx, environment, logger); err != nil {
				logger.Error("Failed to complete environment deactivation cleanup", zap.Error(err))
				// Continue with scaling down even if cleanup partially fails
			}

			// Scale down all deployments in the environment
			logger.Info("Scaling down deployments for environment deactivation")
			if err := r.suspendEnvironment(ctx, environment, logger); err != nil {
				logger.Error("Failed to scale down deployments", zap.Error(err))
				// Continue - pods may already be scaled down
			}

			// Wait for all pods to terminate before marking as inactive
			// This ensures databases have time to checkpoint and WAL is flushed
			if !r.waitForPodsTerminated(ctx, environment.Spec.TargetNamespace, logger) {
				logger.Info("Waiting for pods to terminate before marking environment inactive")
				if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateDeactivating, "Waiting for pods to terminate...", logger); err != nil {
					logger.Warn("Failed to update deactivating message", zap.Error(err))
				}
				return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
			}

			// All pods terminated - now set to inactive
			logger.Info("All pods terminated, marking environment as inactive")
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateInactive, "Environment is inactive", logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
			return reconcile.Result{}, nil
		}

		// Handle activation transition
		if wasInactive && willBeActive {
			// Set to activating state first
			if currentState != environmentsv1.EnvironmentStateActivating {
				if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateActivating, "Environment is being activated", logger); err != nil {
					logger.Error("Failed to update status to activating", zap.Error(err))
				}
			}

			// Now set to active
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateActive, "Environment is active", logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
			return reconcile.Result{}, nil
		}

		// Only update status if state actually changed (no transition states)
		if currentState != desiredState {
			message := "Environment is inactive"
			if desiredState == environmentsv1.EnvironmentStateActive {
				message = "Environment is active"
			}

			if err := r.updateEnvironmentStatus(ctx, environment, desiredState, message, logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
		} else {
			logger.Debug("Environment status unchanged, skipping status update")
		}

		return reconcile.Result{}, nil
	}

	// Namespace was just created, update status
	desiredState := environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		desiredState = environmentsv1.EnvironmentStateActive
	}

	// Update status with retry logic
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = desiredState
		environment.Status.Message = "Namespace created successfully"

		// Add condition for namespace creation
		r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionNamespaceCreated, metav1.ConditionTrue, "NamespaceCreated", "Namespace has been created successfully")

		return nil
	}, logger); err != nil {
		logger.Error("Failed to update environment status after retries", zap.Error(err))
		// Don't fail the reconciliation for status update failures
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Environment{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Namespace{}).           // Watch Namespaces owned by Environments
		Owns(&networkingv1.NetworkPolicy{}). // Watch NetworkPolicies owned by Environments
		Watches(
			&appsv1.Deployment{},
			handler.EnqueueRequestsFromMapFunc(r.findEnvironmentForComposeResource),
		).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findEnvironmentForComposeResource),
		).
		Complete(r)
	// Note: We don't watch WorkMachine here because Environment references WorkMachine by name
	// The Environment controller will handle WorkMachine ownership during reconciliation
}

// findEnvironmentForComposeResource finds the environment that owns a compose resource
func (r *EnvironmentReconciler) findEnvironmentForComposeResource(ctx context.Context, obj client.Object) []reconcile.Request {
	// Check if this resource has the docker-composition label
	labels := obj.GetLabels()
	if labels == nil {
		return nil
	}

	envName, ok := labels[dockerCompositionLabel]
	if !ok {
		return nil
	}

	// Return a reconcile request for the environment
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name: envName,
			},
		},
	}
}

// waitForPodsTerminated waits for all pods in a namespace to be fully deleted
// This ensures databases have time to checkpoint and flush WAL before the
// environment is marked as inactive. Only Succeeded pods (completed Jobs) are ignored.
func (r *EnvironmentReconciler) waitForPodsTerminated(ctx context.Context, namespace string, logger *zap.Logger) bool {
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, client.InNamespace(namespace)); err != nil {
		logger.Warn("Failed to list pods", zap.Error(err))
		return false
	}

	// Wait for ALL pods to be deleted, except completed Job pods (Succeeded phase)
	for _, pod := range pods.Items {
		// Skip completed Job pods - they're finished and won't write to disk
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}
		// Any other pod (Running, Pending, Failed, Unknown) means we should wait
		logger.Debug("Pod still exists", zap.String("pod", pod.Name), zap.String("phase", string(pod.Status.Phase)))
		return false
	}
	return true
}

// hasActiveSnapshotOperation checks if there are any in-progress snapshot operations
// (EnvironmentSnapshotRequest or EnvironmentSnapshotRestore) for this environment
func (r *EnvironmentReconciler) hasActiveSnapshotOperation(ctx context.Context, environmentName string) (bool, error) {
	// Check for active EnvironmentSnapshotRequests
	snapshotRequests := &environmentsv1.EnvironmentSnapshotRequestList{}
	if err := r.List(ctx, snapshotRequests); err != nil {
		return false, err
	}

	for _, req := range snapshotRequests.Items {
		if req.Spec.EnvironmentName != environmentName {
			continue
		}
		// Check if request is in-progress (not completed or failed)
		if req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseCompleted &&
			req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseFailed {
			return true, nil
		}
	}

	// Check for active EnvironmentSnapshotRestores
	snapshotRestores := &environmentsv1.EnvironmentSnapshotRestoreList{}
	if err := r.List(ctx, snapshotRestores); err != nil {
		return false, err
	}

	for _, restore := range snapshotRestores.Items {
		if restore.Spec.EnvironmentName != environmentName {
			continue
		}
		// Check if restore is in-progress (not completed or failed)
		if restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseCompleted &&
			restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseFailed {
			return true, nil
		}
	}

	return false, nil
}
