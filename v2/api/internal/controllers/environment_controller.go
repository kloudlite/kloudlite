package controllers

import (
	"context"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	corev1 "k8s.io/api/core/v1"
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

	// Fetch the Environment instance
	environment := &environmentsv1.Environment{}
	err := r.Get(ctx, req.NamespacedName, environment)
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

	// Check if namespace already exists
	namespace := &corev1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace already exists
		logger.Info("Namespace already exists for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))

		// Update namespace labels if needed
		if namespace.Labels == nil {
			namespace.Labels = make(map[string]string)
		}

		// Add environment labels
		namespace.Labels["kloudlite.io/environment"] = environment.Name

		// Update namespace annotations if needed
		if namespace.Annotations == nil {
			namespace.Annotations = make(map[string]string)
		}

		// Add createdBy to annotations (emails contain invalid label characters)
		namespace.Annotations["kloudlite.io/created-by"] = environment.Spec.CreatedBy

		// Add any custom labels from the environment spec to annotations
		// (they may contain invalid label characters like @, =, etc.)
		if environment.Spec.Labels != nil {
			for k, v := range environment.Spec.Labels {
				namespace.Annotations[k] = v
			}
		}

		// Add any custom annotations from the environment spec
		if environment.Spec.Annotations != nil {
			for k, v := range environment.Spec.Annotations {
				namespace.Annotations[k] = v
			}
		}

		// Update the namespace
		if err := r.Update(ctx, namespace); err != nil {
			logger.Warn("Failed to update namespace labels/annotations", zap.Error(err))
		}

		return reconcile.Result{}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing namespace", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Create namespace for the Environment
	logger.Info("Creating namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: environment.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/environment": environment.Name,
			},
			Annotations: map[string]string{
				"kloudlite.io/environment-uid": string(environment.UID),
				"kloudlite.io/creation-reason": "auto-created-for-environment",
				"kloudlite.io/created-by":      environment.Spec.CreatedBy, // Move email to annotations
			},
		},
	}

	// Add custom labels from environment spec (skip invalid label values)
	if environment.Spec.Labels != nil {
		for k, v := range environment.Spec.Labels {
			// Only add valid label values (no @, =, etc.)
			// Labels with emails or special chars should go to annotations
			namespace.Annotations[k] = v
		}
	}

	// Add custom annotations from environment spec
	if environment.Spec.Annotations != nil {
		for k, v := range environment.Spec.Annotations {
			namespace.Annotations[k] = v
		}
	}

	// Set Environment as the owner of the Namespace using OwnerReferences
	// Note: This might not work for namespace as it's cluster-scoped
	// and Environment is also cluster-scoped. Owner references typically
	// work for namespace-scoped resources owned by cluster-scoped resources.
	if err := controllerutil.SetControllerReference(environment, namespace, r.Scheme); err != nil {
		logger.Warn("Failed to set owner reference on namespace (expected for cluster-scoped resources)", zap.Error(err))
		// Continue anyway, as this is expected behavior for cluster-scoped resources
	}

	// Create the namespace
	if err := r.Create(ctx, namespace); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Another reconciliation might have created it
			logger.Info("Namespace already exists (race condition)")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to create namespace", zap.Error(err))
		// Retry after a delay
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	logger.Info("Successfully created namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	// Update environment status to indicate namespace has been created
	environment.Status.State = environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		environment.Status.State = environmentsv1.EnvironmentStateActive
	}
	environment.Status.Message = "Namespace created successfully"

	// Add condition for namespace creation
	condition := environmentsv1.EnvironmentCondition{
		Type:               environmentsv1.EnvironmentConditionNamespaceCreated,
		Status:            metav1.ConditionTrue,
		LastTransitionTime: &metav1.Time{Time: time.Now()},
		Reason:            "NamespaceCreated",
		Message:           "Namespace has been created successfully",
	}

	// Initialize conditions if nil
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	// Add or update the condition
	found := false
	for i, c := range environment.Status.Conditions {
		if c.Type == environmentsv1.EnvironmentConditionNamespaceCreated {
			environment.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, condition)
	}

	// Update the environment status
	if err := r.Status().Update(ctx, environment); err != nil {
		logger.Warn("Failed to update environment status", zap.Error(err))
		// Don't fail the reconciliation for status update failures
	}

	return reconcile.Result{}, nil
}

// handleDeletion handles the deletion of an environment and its child resources
func (r *EnvironmentReconciler) handleDeletion(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	// Update status to show deletion in progress
	if environment.Status.State != environmentsv1.EnvironmentStateDeleting {
		environment.Status.State = environmentsv1.EnvironmentStateDeleting
		environment.Status.Message = "Deleting environment and cleaning up resources"

		if err := r.Status().Update(ctx, environment); err != nil {
			logger.Warn("Failed to update environment status to deleting", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

	// Check if namespace exists
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace exists, delete it
		logger.Info("Deleting namespace for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))

		if err := r.Delete(ctx, namespace); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete namespace", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			}
		}

		// Requeue to wait for namespace deletion to complete
		logger.Info("Waiting for namespace deletion to complete")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check namespace", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	// Namespace is deleted, remove finalizer
	logger.Info("Namespace deleted, removing finalizer from environment")

	if controllerutil.ContainsFinalizer(environment, environmentFinalizer) {
		controllerutil.RemoveFinalizer(environment, environmentFinalizer)
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to remove finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("Environment cleanup completed successfully")
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Environment{}).
		Owns(&corev1.Namespace{}). // Watch Namespaces owned by Environments
		Complete(r)
}