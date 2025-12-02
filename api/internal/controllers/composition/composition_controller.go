package composition

import (
	"context"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CompositionReconciler reconciles Composition objects
type CompositionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles Composition events
func (r *CompositionReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx).WithValues("dockercomposition", req.NamespacedName)

	zapLogger := r.Logger.With(
		zap.String("dockercomposition", req.Name),
		zap.String("namespace", req.Namespace),
	)

	zapLogger.Info("Reconciling Composition")

	// Fetch the Composition instance
	composition := &compositionsv1.Composition{}
	err := r.Get(ctx, req.NamespacedName, composition)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Composition not found, likely deleted")
			return reconcile.Result{}, nil
		}
		zapLogger.Error("Failed to get Composition", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if composition is being deleted
	if composition.DeletionTimestamp != nil {
		zapLogger.Info("Composition is being deleted, starting cleanup")
		return r.handleDeletion(ctx, composition, zapLogger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(composition, compositionFinalizer) {
		zapLogger.Info("Adding finalizer to Composition")
		controllerutil.AddFinalizer(composition, compositionFinalizer)
		if err := r.Update(ctx, composition); err != nil {
			zapLogger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Get the environment to check activation state
	environment, err := r.getEnvironmentForNamespace(ctx, composition.Namespace, zapLogger)
	if err != nil {
		zapLogger.Error("Failed to get environment for activation check", zap.Error(err))
		// Continue with deployment even if we can't get environment
	}

	// Check if reconciliation is actually needed
	// This prevents unnecessary reconciliation loops
	shouldReconcile := false

	// Case 1: Composition spec changed
	if composition.Status.ObservedGeneration != composition.Generation {
		zapLogger.Info("Reconciling: spec changed",
			zap.Int64("observedGeneration", composition.Status.ObservedGeneration),
			zap.Int64("generation", composition.Generation))
		shouldReconcile = true
	}

	// Case 2: Environment activation state changed
	if environment != nil && composition.Status.EnvironmentActivated != environment.Spec.Activated {
		zapLogger.Info("Reconciling: environment activation changed",
			zap.Bool("statusActivated", composition.Status.EnvironmentActivated),
			zap.Bool("envActivated", environment.Spec.Activated))
		shouldReconcile = true
	}

	// Case 3: Status is not running (need to deploy/fix) or is in an error state (need to re-check)
	if composition.Status.State != compositionsv1.CompositionStateRunning {
		zapLogger.Info("Reconciling: status not running",
			zap.String("currentState", string(composition.Status.State)))
		shouldReconcile = true
	}

	// Case 3b: Status is failed/degraded - re-check health in case pods recovered
	if composition.Status.State == compositionsv1.CompositionStateFailed ||
		composition.Status.State == compositionsv1.CompositionStateDegraded {
		zapLogger.Info("Reconciling: re-checking failed/degraded composition",
			zap.String("currentState", string(composition.Status.State)))
		shouldReconcile = true
	}

	// Case 4: Drift detection - check if deployed resources still exist
	// This handles manual deletions or resources being removed outside of the controller
	if !shouldReconcile && composition.Status.DeployedResources != nil {
		drifted, err := r.checkResourceDrift(ctx, composition, zapLogger)
		if err != nil {
			zapLogger.Error("Failed to check resource drift", zap.Error(err))
			// Continue with reconciliation if we can't determine drift
			shouldReconcile = true
		} else if drifted {
			zapLogger.Info("Reconciling: resource drift detected")
			shouldReconcile = true
		}
	}

	// Skip reconciliation if nothing changed
	// This prevents infinite reconciliation loops
	if !shouldReconcile {
		zapLogger.Debug("Skipping reconciliation: nothing changed")
		return reconcile.Result{}, nil
	}

	// Deploy the composition
	if err := r.deployComposition(ctx, composition, zapLogger); err != nil {
		zapLogger.Error("Failed to deploy composition", zap.Error(err))
		return r.updateStatus(ctx, composition, environment, compositionsv1.CompositionStateFailed, err.Error(), zapLogger)
	}

	// Check actual deployment/pod health status
	healthResult, err := r.checkDeploymentHealth(ctx, composition, zapLogger)
	if err != nil {
		zapLogger.Error("Failed to check deployment health", zap.Error(err))
		// Fall back to running state if health check fails
		return r.updateStatus(ctx, composition, environment, compositionsv1.CompositionStateRunning, "Deployed (health check unavailable)", zapLogger)
	}

	// Update service status in composition
	composition.Status.Services = healthResult.Services
	composition.Status.RunningCount = healthResult.RunningCount

	zapLogger.Info("Deployment health check result",
		zap.String("state", string(healthResult.State)),
		zap.Int32("runningCount", healthResult.RunningCount),
		zap.Int32("servicesCount", healthResult.ServicesCount),
		zap.String("message", healthResult.Message))

	// Update status based on health check
	return r.updateStatus(ctx, composition, environment, healthResult.State, healthResult.Message, zapLogger)
}

// checkResourceDrift checks if deployed resources still exist in the cluster
func (r *CompositionReconciler) checkResourceDrift(ctx context.Context, composition *compositionsv1.Composition, logger *zap.Logger) (bool, error) {
	if composition.Status.DeployedResources == nil {
		return false, nil
	}

	// Check deployments
	for _, deploymentName := range composition.Status.DeployedResources.Deployments {
		deployment := &appsv1.Deployment{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: composition.Namespace,
			Name:      deploymentName,
		}, deployment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Drift detected: deployment not found",
					zap.String("deployment", deploymentName))
				return true, nil
			}
			return false, err
		}
	}

	// Check services
	for _, serviceName := range composition.Status.DeployedResources.Services {
		service := &corev1.Service{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: composition.Namespace,
			Name:      serviceName,
		}, service)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Drift detected: service not found",
					zap.String("service", serviceName))
				return true, nil
			}
			return false, err
		}
	}

	return false, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *CompositionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compositionsv1.Composition{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&appsv1.Deployment{}). // Watch deployments owned by Composition
		Owns(&corev1.Service{}).    // Watch services owned by Composition
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findCompositionsForConfigMap),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findCompositionsForSecret),
		).
		Watches(
			&compositionsv1.Environment{},
			handler.EnqueueRequestsFromMapFunc(r.findCompositionsForEnvironment),
		).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findCompositionsForPod),
		).
		Complete(r)
}
