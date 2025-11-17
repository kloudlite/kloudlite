package environment

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	environmentFinalizer = "environments.kloudlite.io/finalizer"
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
				blockOwnerDeletion := false
				ownerRef := metav1.OwnerReference{
					APIVersion:         workmachine.APIVersion,
					Kind:               workmachine.Kind,
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

	// Check if cloning is requested
	if environment.Spec.CloneFrom != "" {
		logger.Info("Cloning requested from source environment",
			zap.String("source", environment.Spec.CloneFrom))
		return r.handleCloning(ctx, environment, logger)
	}

	// Note: TargetNamespace is always set by the mutation webhook.
	// The webhook generates it as "env-{name}" if not provided by the user.
	// The webhook also validates that the namespace doesn't already exist.
	// Controller's responsibility is to create the actual Kubernetes namespace resource.

	// Check if namespace already exists and handle it
	namespaceExists, err := r.ensureNamespaceExists(ctx, environment, logger)
	if err != nil {
		return reconcile.Result{}, err
	}

	if namespaceExists {
		// Update environment status based on activation state
		// The actual scaling is handled by the composition controller
		desiredState := environmentsv1.EnvironmentStateInactive
		if environment.Spec.Activated {
			desiredState = environmentsv1.EnvironmentStateActive
		}

		// Detect deactivation transition (from active to inactive)
		wasActive := environment.Status.State == environmentsv1.EnvironmentStateActive
		willBeInactive := desiredState == environmentsv1.EnvironmentStateInactive

		if wasActive && willBeInactive {
			// Environment is being deactivated - disconnect workspaces and remove service intercepts
			logger.Info("Environment is being deactivated, cleaning up connections")
			if err := r.handleEnvironmentDeactivation(ctx, environment, logger); err != nil {
				logger.Error("Failed to complete environment deactivation cleanup", zap.Error(err))
				// Continue with status update even if cleanup partially fails
			}
		}

		// Only update status if state actually changed
		if environment.Status.State != desiredState {
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
		Owns(&corev1.Namespace{}). // Watch Namespaces owned by Environments
		Complete(r)
	// Note: We don't watch WorkMachine here because Environment references WorkMachine by name
	// The Environment controller will handle WorkMachine ownership during reconciliation
}
