package composition

import (
	"context"
	"time"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleDeletion handles cleanup when Composition is being deleted
func (r *CompositionReconciler) handleDeletion(ctx context.Context, composition *compositionsv1.Composition, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Cleaning up Composition resources")

	// Update status to deleting if not already set
	if composition.Status.State != compositionsv1.CompositionStateDeleting {
		composition.Status.State = compositionsv1.CompositionStateDeleting
		composition.Status.Message = "Deleting composition and associated resources"
		if err := r.Status().Update(ctx, composition); err != nil {
			logger.Error("Failed to update status to deleting", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

	// Common label selector for all resources created by this composition
	labelSelector := client.MatchingLabels{dockerCompositionLabel: composition.Name}
	namespaceOpt := client.InNamespace(composition.Namespace)

	// Define cleanup order: Deployments -> Services -> ConfigMaps -> Secrets -> PVCs
	cleanupSteps := []cleanupStep{
		{resourceType: "Deployment", deleteFunc: r.deleteDeployments},
		{resourceType: "Service", deleteFunc: r.deleteServices},
		{resourceType: "ConfigMap", deleteFunc: r.deleteConfigMaps},
		{resourceType: "Secret", deleteFunc: r.deleteSecrets},
		{resourceType: "PVC", deleteFunc: r.deletePVCs},
	}

	resourceCounts := make(map[string]int)
	resourcesRemaining := false

	// Execute cleanup steps
	for _, step := range cleanupSteps {
		count, err := step.deleteFunc(ctx, namespaceOpt, labelSelector, logger)
		if err != nil {
			logger.Error("Failed to cleanup resources",
				zap.String("resourceType", step.resourceType),
				zap.Error(err))
			return reconcile.Result{}, err
		}
		resourceCounts[step.resourceType] = count
		if count > 0 {
			resourcesRemaining = true
		}
	}

	// If any resources are still being deleted, requeue
	if resourcesRemaining {
		logger.Info("Waiting for resources to be deleted", zap.Any("resourceCounts", resourceCounts))
		return reconcile.Result{RequeueAfter: time.Duration(deletionRequeueInterval) * time.Nanosecond}, nil
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

// cleanupStep defines a cleanup operation for a specific resource type
type cleanupStep struct {
	resourceType string
	deleteFunc   func(context.Context, client.InNamespace, client.MatchingLabels, *zap.Logger) (int, error)
}

// deleteDeployments deletes all deployments matching the label selector
func (r *CompositionReconciler) deleteDeployments(ctx context.Context, namespaceOpt client.InNamespace, labelSelector client.MatchingLabels, logger *zap.Logger) (int, error) {
	deploymentList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deploymentList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list deployments", zap.Error(err))
		return 0, err
	}

	for _, deployment := range deploymentList.Items {
		if deployment.DeletionTimestamp == nil {
			logger.Info("Deleting deployment", zap.String("name", deployment.Name))
			if err := r.Delete(ctx, &deployment); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete deployment", zap.String("name", deployment.Name), zap.Error(err))
				return 0, err
			}
		}
	}
	return len(deploymentList.Items), nil
}

// deleteServices deletes all services matching the label selector
func (r *CompositionReconciler) deleteServices(ctx context.Context, namespaceOpt client.InNamespace, labelSelector client.MatchingLabels, logger *zap.Logger) (int, error) {
	serviceList := &corev1.ServiceList{}
	if err := r.List(ctx, serviceList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list services", zap.Error(err))
		return 0, err
	}

	for _, service := range serviceList.Items {
		if service.DeletionTimestamp == nil {
			logger.Info("Deleting service", zap.String("name", service.Name))
			if err := r.Delete(ctx, &service); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete service", zap.String("name", service.Name), zap.Error(err))
				return 0, err
			}
		}
	}
	return len(serviceList.Items), nil
}

// deleteConfigMaps deletes all configmaps matching the label selector
func (r *CompositionReconciler) deleteConfigMaps(ctx context.Context, namespaceOpt client.InNamespace, labelSelector client.MatchingLabels, logger *zap.Logger) (int, error) {
	configMapList := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMapList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list configmaps", zap.Error(err))
		return 0, err
	}

	for _, cm := range configMapList.Items {
		if cm.DeletionTimestamp == nil {
			logger.Info("Deleting configmap", zap.String("name", cm.Name))
			if err := r.Delete(ctx, &cm); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete configmap", zap.String("name", cm.Name), zap.Error(err))
				return 0, err
			}
		}
	}
	return len(configMapList.Items), nil
}

// deleteSecrets deletes all secrets matching the label selector
func (r *CompositionReconciler) deleteSecrets(ctx context.Context, namespaceOpt client.InNamespace, labelSelector client.MatchingLabels, logger *zap.Logger) (int, error) {
	secretList := &corev1.SecretList{}
	if err := r.List(ctx, secretList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list secrets", zap.Error(err))
		return 0, err
	}

	for _, secret := range secretList.Items {
		if secret.DeletionTimestamp == nil {
			logger.Info("Deleting secret", zap.String("name", secret.Name))
			if err := r.Delete(ctx, &secret); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete secret", zap.String("name", secret.Name), zap.Error(err))
				return 0, err
			}
		}
	}
	return len(secretList.Items), nil
}

// deletePVCs deletes all PVCs matching the label selector
func (r *CompositionReconciler) deletePVCs(ctx context.Context, namespaceOpt client.InNamespace, labelSelector client.MatchingLabels, logger *zap.Logger) (int, error) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, namespaceOpt, labelSelector); err != nil {
		logger.Error("Failed to list PVCs", zap.Error(err))
		return 0, err
	}

	for _, pvc := range pvcList.Items {
		if pvc.DeletionTimestamp == nil {
			logger.Info("Deleting PVC", zap.String("name", pvc.Name))
			if err := r.Delete(ctx, &pvc); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete PVC", zap.String("name", pvc.Name), zap.Error(err))
				return 0, err
			}
		}
	}
	return len(pvcList.Items), nil
}

// cleanupRemovedResources deletes resources that were deployed before but are no longer in the compose file
func (r *CompositionReconciler) cleanupRemovedResources(ctx context.Context, composition *compositionsv1.Composition, oldDeployedResources *compositionsv1.DeployedResources, currentDeployments, currentServices []string, logger *zap.Logger) error {
	// Skip cleanup if this is the first deployment (no previous resources)
	if oldDeployedResources == nil {
		logger.Info("First deployment, skipping cleanup")
		return nil
	}

	// Convert current resources to sets for efficient lookup
	currentDeploymentSet := makeSet(currentDeployments)
	currentServiceSet := makeSet(currentServices)

	// Find deployments to delete
	for _, oldDeployment := range oldDeployedResources.Deployments {
		if !currentDeploymentSet[oldDeployment] {
			logger.Info("Deleting removed deployment", zap.String("name", oldDeployment))
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      oldDeployment,
					Namespace: composition.Namespace,
				},
			}
			if err := r.Delete(ctx, deployment); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete deployment", zap.String("name", oldDeployment), zap.Error(err))
				return err
			}
		}
	}

	// Find services to delete
	for _, oldService := range oldDeployedResources.Services {
		if !currentServiceSet[oldService] {
			logger.Info("Deleting removed service", zap.String("name", oldService))
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      oldService,
					Namespace: composition.Namespace,
				},
			}
			if err := r.Delete(ctx, service); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete service", zap.String("name", oldService), zap.Error(err))
				return err
			}
		}
	}

	return nil
}

// makeSet creates a set from a string slice for efficient lookup
func makeSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}
