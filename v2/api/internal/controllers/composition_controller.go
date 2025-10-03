package controllers

import (
	"context"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	compositionFinalizer = "dockercompositions.environments.kloudlite.io/finalizer"
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
	composition := &environmentsv1.Composition{}
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

	// Skip reconciliation if generation hasn't changed (status-only update)
	if composition.Status.ObservedGeneration == composition.Generation {
		zapLogger.Info("Skipping reconciliation - generation unchanged")
		return reconcile.Result{}, nil
	}

	// Deploy the composition
	if err := r.deployComposition(ctx, composition, zapLogger); err != nil {
		zapLogger.Error("Failed to deploy composition", zap.Error(err))
		return r.updateStatus(ctx, composition, environmentsv1.CompositionStateFailed, err.Error(), zapLogger)
	}

	// Update status to running
	return r.updateStatus(ctx, composition, environmentsv1.CompositionStateRunning, "Composition deployed successfully", zapLogger)
}

// deployComposition deploys the docker compose to Kubernetes
func (r *CompositionReconciler) deployComposition(ctx context.Context, composition *environmentsv1.Composition, logger *zap.Logger) error {
	logger.Info("Deploying Composition")

	// Parse the docker-compose file
	project, err := ParseComposeFile(composition.Spec.ComposeContent, composition.Name)
	if err != nil {
		logger.Error("Failed to parse compose file", zap.Error(err))
		return fmt.Errorf("parse error: %w", err)
	}

	logger.Info("Parsed compose file",
		zap.Int("services", len(project.Services)),
		zap.Int("volumes", len(project.Volumes)))

	// Convert to Kubernetes resources
	resources, err := ConvertComposeToK8s(project, composition, composition.Namespace)
	if err != nil {
		logger.Error("Failed to convert to Kubernetes resources", zap.Error(err))
		return fmt.Errorf("conversion error: %w", err)
	}

	logger.Info("Converted to Kubernetes resources",
		zap.Int("deployments", len(resources.Deployments)),
		zap.Int("services", len(resources.Services)),
		zap.Int("pvcs", len(resources.PVCs)))

	// Apply PVCs first
	for _, pvc := range resources.PVCs {
		if err := r.applyResource(ctx, pvc, composition, logger); err != nil {
			return fmt.Errorf("failed to apply PVC %s: %w", pvc.Name, err)
		}
	}

	// Apply Deployments
	deployedDeployments := make([]string, 0)
	for _, deployment := range resources.Deployments {
		if err := r.applyResource(ctx, deployment, composition, logger); err != nil {
			return fmt.Errorf("failed to apply Deployment %s: %w", deployment.Name, err)
		}
		deployedDeployments = append(deployedDeployments, deployment.Name)
	}

	// Apply Services
	deployedServices := make([]string, 0)
	for _, service := range resources.Services {
		if err := r.applyResource(ctx, service, composition, logger); err != nil {
			return fmt.Errorf("failed to apply Service %s: %w", service.Name, err)
		}
		deployedServices = append(deployedServices, service.Name)
	}

	// Update status with deployed resources
	composition.Status.DeployedResources = &environmentsv1.DeployedResources{
		Deployments: deployedDeployments,
		Services:    deployedServices,
		PVCs:        getPVCNames(resources.PVCs),
	}
	composition.Status.ServicesCount = int32(len(resources.ServiceNames))

	logger.Info("Successfully deployed Composition",
		zap.Int("deployments", len(deployedDeployments)),
		zap.Int("services", len(deployedServices)))

	return nil
}

// applyResource creates or updates a Kubernetes resource
func (r *CompositionReconciler) applyResource(ctx context.Context, resource client.Object, composition *environmentsv1.Composition, logger *zap.Logger) error {
	// Set controller ownership
	if err := controllerutil.SetControllerReference(composition, resource, r.Scheme); err != nil {
		return err
	}

	// Try to get existing resource
	existing := resource.DeepCopyObject().(client.Object)
	err := r.Get(ctx, client.ObjectKeyFromObject(resource), existing)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new resource
			logger.Info("Creating resource",
				zap.String("kind", resource.GetObjectKind().GroupVersionKind().Kind),
				zap.String("name", resource.GetName()))
			return r.Create(ctx, resource)
		}
		return err
	}

	// Update existing resource
	logger.Info("Updating resource",
		zap.String("kind", resource.GetObjectKind().GroupVersionKind().Kind),
		zap.String("name", resource.GetName()))

	// Copy resource version for update
	resource.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, resource)
}

// getPVCNames extracts PVC names from a list
func getPVCNames(pvcs []*corev1.PersistentVolumeClaim) []string {
	names := make([]string, len(pvcs))
	for i, pvc := range pvcs {
		names[i] = pvc.Name
	}
	return names
}

// handleDeletion handles cleanup when Composition is being deleted
func (r *CompositionReconciler) handleDeletion(ctx context.Context, composition *environmentsv1.Composition, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Cleaning up Composition resources")

	// Common label selector for all resources created by this composition
	labelSelector := client.MatchingLabels{
		"kloudlite.io/docker-composition": composition.Name,
	}
	namespaceOpt := client.InNamespace(composition.Namespace)

	resourcesRemaining := false

	// Delete Deployments
	deploymentList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deploymentList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list deployments", zap.Error(err))
		return reconcile.Result{}, err
	}
	for _, deployment := range deploymentList.Items {
		if deployment.DeletionTimestamp == nil {
			logger.Info("Deleting deployment", zap.String("name", deployment.Name))
			if err := r.Delete(ctx, &deployment); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete deployment", zap.String("name", deployment.Name), zap.Error(err))
				return reconcile.Result{}, err
			}
		}
	}
	if len(deploymentList.Items) > 0 {
		resourcesRemaining = true
	}

	// Delete Services
	serviceList := &corev1.ServiceList{}
	if err := r.List(ctx, serviceList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list services", zap.Error(err))
		return reconcile.Result{}, err
	}
	for _, service := range serviceList.Items {
		if service.DeletionTimestamp == nil {
			logger.Info("Deleting service", zap.String("name", service.Name))
			if err := r.Delete(ctx, &service); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete service", zap.String("name", service.Name), zap.Error(err))
				return reconcile.Result{}, err
			}
		}
	}
	if len(serviceList.Items) > 0 {
		resourcesRemaining = true
	}

	// Delete ConfigMaps
	configMapList := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMapList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list configmaps", zap.Error(err))
		return reconcile.Result{}, err
	}
	for _, cm := range configMapList.Items {
		if cm.DeletionTimestamp == nil {
			logger.Info("Deleting configmap", zap.String("name", cm.Name))
			if err := r.Delete(ctx, &cm); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete configmap", zap.String("name", cm.Name), zap.Error(err))
				return reconcile.Result{}, err
			}
		}
	}
	if len(configMapList.Items) > 0 {
		resourcesRemaining = true
	}

	// Delete Secrets
	secretList := &corev1.SecretList{}
	if err := r.List(ctx, secretList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list secrets", zap.Error(err))
		return reconcile.Result{}, err
	}
	for _, secret := range secretList.Items {
		if secret.DeletionTimestamp == nil {
			logger.Info("Deleting secret", zap.String("name", secret.Name))
			if err := r.Delete(ctx, &secret); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete secret", zap.String("name", secret.Name), zap.Error(err))
				return reconcile.Result{}, err
			}
		}
	}
	if len(secretList.Items) > 0 {
		resourcesRemaining = true
	}

	// Delete PVCs (last, after deployments are gone)
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		return reconcile.Result{}, err
	}
	for _, pvc := range pvcList.Items {
		if pvc.DeletionTimestamp == nil {
			logger.Info("Deleting PVC", zap.String("name", pvc.Name))
			if err := r.Delete(ctx, &pvc); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete PVC", zap.String("name", pvc.Name), zap.Error(err))
				return reconcile.Result{}, err
			}
		}
	}
	if len(pvcList.Items) > 0 {
		resourcesRemaining = true
	}

	// If any resources are still being deleted, requeue
	if resourcesRemaining {
		logger.Info("Waiting for resources to be deleted",
			zap.Int("deployments", len(deploymentList.Items)),
			zap.Int("services", len(serviceList.Items)),
			zap.Int("configmaps", len(configMapList.Items)),
			zap.Int("secrets", len(secretList.Items)),
			zap.Int("pvcs", len(pvcList.Items)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// All resources deleted, remove finalizer
	logger.Info("All resources deleted, removing finalizer")
	controllerutil.RemoveFinalizer(composition, compositionFinalizer)
	if err := r.Update(ctx, composition); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Composition cleanup completed successfully")
	return reconcile.Result{}, nil
}

// updateStatus updates the status of the Composition
func (r *CompositionReconciler) updateStatus(ctx context.Context, composition *environmentsv1.Composition, state environmentsv1.CompositionState, message string, logger *zap.Logger) (reconcile.Result, error) {
	composition.Status.State = state
	composition.Status.Message = message
	composition.Status.ObservedGeneration = composition.Generation

	now := metav1.Now()
	if state == environmentsv1.CompositionStateRunning {
		composition.Status.LastDeployedTime = &now
	}

	// Update condition
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: composition.Generation,
		LastTransitionTime: now,
		Reason:             string(state),
		Message:            message,
	}

	if state == environmentsv1.CompositionStateRunning {
		readyCondition.Status = metav1.ConditionTrue
	}

	// Update or add condition
	found := false
	for i, condition := range composition.Status.Conditions {
		if condition.Type == "Ready" {
			composition.Status.Conditions[i] = readyCondition
			found = true
			break
		}
	}
	if !found {
		composition.Status.Conditions = append(composition.Status.Conditions, readyCondition)
	}

	if err := r.Status().Update(ctx, composition); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Updated Composition status",
		zap.String("state", string(state)),
		zap.String("message", message))

	// Requeue to continue monitoring
	if state == environmentsv1.CompositionStateDeploying {
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *CompositionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Composition{}).
		Owns(&appsv1.Deployment{}). // Watch deployments owned by Composition
		Owns(&corev1.Service{}).    // Watch services owned by Composition
		Complete(r)
}
