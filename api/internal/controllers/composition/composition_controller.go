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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

	// Check if composition is already in running state and no changes needed
	if composition.Status.State == compositionsv1.CompositionStateRunning &&
	   composition.Status.ObservedGeneration == composition.Generation {
		zapLogger.Debug("Composition already running and up to date, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	// Deploy the composition (reconcile on Composition changes OR env-config/env-secret changes)
	if err := r.deployComposition(ctx, composition, zapLogger); err != nil {
		zapLogger.Error("Failed to deploy composition", zap.Error(err))
		return r.updateStatus(ctx, composition, compositionsv1.CompositionStateFailed, err.Error(), zapLogger)
	}

	// Update status to running
	return r.updateStatus(ctx, composition, compositionsv1.CompositionStateRunning, "Composition deployed successfully", zapLogger)
}







// SetupWithManager sets up the controller with the Manager
func (r *CompositionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compositionsv1.Composition{}).
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
		Complete(r)
}

