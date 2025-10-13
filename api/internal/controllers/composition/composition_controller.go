package composition

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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

	// Deploy the composition (reconcile on Composition changes OR env-config/env-secret changes)
	if err := r.deployComposition(ctx, composition, zapLogger); err != nil {
		zapLogger.Error("Failed to deploy composition", zap.Error(err))
		return r.updateStatus(ctx, composition, environmentsv1.CompositionStateFailed, err.Error(), zapLogger)
	}

	// Update status to running
	return r.updateStatus(ctx, composition, environmentsv1.CompositionStateRunning, "Composition deployed successfully", zapLogger)
}

// getEnvironmentForNamespace finds the environment that owns the given namespace
func (r *CompositionReconciler) getEnvironmentForNamespace(ctx context.Context, namespace string, logger *zap.Logger) (*environmentsv1.Environment, error) {
	envList := &environmentsv1.EnvironmentList{}
	if err := r.List(ctx, envList); err != nil {
		logger.Error("Failed to list environments", zap.Error(err))
		return nil, err
	}

	for _, env := range envList.Items {
		if env.Spec.TargetNamespace == namespace {
			return &env, nil
		}
	}

	logger.Warn("No environment found for namespace", zap.String("namespace", namespace))
	return nil, nil
}

// deployComposition deploys the docker compose to Kubernetes
func (r *CompositionReconciler) deployComposition(ctx context.Context, composition *environmentsv1.Composition, logger *zap.Logger) error {
	logger.Info("Deploying Composition")

	// Get the environment for this composition's namespace
	environment, err := r.getEnvironmentForNamespace(ctx, composition.Namespace, logger)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// Check if environment is activated
	environmentActivated := true
	if environment != nil {
		environmentActivated = environment.Spec.Activated
		logger.Info("Environment activation state",
			zap.String("environment", environment.Name),
			zap.Bool("activated", environmentActivated))
	}

	// Save old deployed resources for cleanup comparison
	var oldDeployedResources *environmentsv1.DeployedResources
	if composition.Status.DeployedResources != nil {
		oldDeployedResources = composition.Status.DeployedResources.DeepCopy()
	}

	// Fetch environment data (envvars and config files) BEFORE parsing
	// This allows the parser to resolve variable references
	envData, err := r.fetchEnvironmentData(ctx, composition.Namespace, logger)
	if err != nil {
		logger.Error("Failed to fetch environment data", zap.Error(err))
		// Don't fail deployment if environment data is not found - just log and continue
		envData = &EnvironmentData{
			EnvVars:     make(map[string]string),
			Secrets:     make(map[string]string),
			ConfigFiles: make(map[string]string),
		}
	}

	// Parse the docker-compose file with environment data
	project, err := ParseComposeFile(composition.Spec.ComposeContent, composition.Name, envData)
	if err != nil {
		logger.Error("Failed to parse compose file", zap.Error(err))
		return fmt.Errorf("parse error: %w", err)
	}

	// Count total service volume mounts
	totalVolumeMounts := 0
	for _, svc := range project.Services {
		totalVolumeMounts += len(svc.Volumes)
	}
	logger.Info("Parsed compose file",
		zap.Int("services", len(project.Services)),
		zap.Int("named_volumes", len(project.Volumes)),
		zap.Int("service_volume_mounts", totalVolumeMounts))

	// Convert to Kubernetes resources
	resources, err := ConvertComposeToK8s(project, composition, composition.Namespace, envData)
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

	// Apply Deployments (scale to 0 if environment is inactive)
	deployedDeployments := make([]string, 0)
	for _, deployment := range resources.Deployments {
		// If environment is not activated, scale deployment to 0 replicas
		if !environmentActivated {
			if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 {
				// Store original replica count in annotation
				if deployment.Annotations == nil {
					deployment.Annotations = make(map[string]string)
				}
				deployment.Annotations["kloudlite.io/original-replicas"] = fmt.Sprintf("%d", *deployment.Spec.Replicas)

				// Scale to 0
				zero := int32(0)
				deployment.Spec.Replicas = &zero
				logger.Info("Scaling deployment to 0 (environment inactive)",
					zap.String("deployment", deployment.Name))
			}
		} else {
			// Environment is active - restore original replicas if they exist
			if deployment.Annotations != nil {
				if originalReplicas, exists := deployment.Annotations["kloudlite.io/original-replicas"]; exists {
					if replicas, err := strconv.ParseInt(originalReplicas, 10, 32); err == nil && replicas > 0 {
						r := int32(replicas)
						deployment.Spec.Replicas = &r
						// Remove the annotation since we've restored the value
						delete(deployment.Annotations, "kloudlite.io/original-replicas")
						logger.Info("Restored deployment replicas (environment active)",
							zap.String("deployment", deployment.Name),
							zap.Int32("replicas", r))
					}
				}
			}
		}

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

	// Cleanup removed resources using OLD deployed resources
	if err := r.cleanupRemovedResources(ctx, composition, oldDeployedResources, deployedDeployments, deployedServices, logger); err != nil {
		return fmt.Errorf("failed to cleanup removed resources: %w", err)
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

	// PVCs are mostly immutable - skip updating them if they already exist
	if _, ok := resource.(*corev1.PersistentVolumeClaim); ok {
		logger.Info("Skipping update for existing PVC (PVC spec is immutable)",
			zap.String("name", resource.GetName()))
		return nil
	}

	// Copy resource version for update
	resource.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, resource)
}

// cleanupRemovedResources deletes resources that were deployed before but are no longer in the compose file
func (r *CompositionReconciler) cleanupRemovedResources(ctx context.Context, composition *environmentsv1.Composition, oldDeployedResources *environmentsv1.DeployedResources, currentDeployments, currentServices []string, logger *zap.Logger) error {
	// Skip cleanup if this is the first deployment (no previous resources)
	if oldDeployedResources == nil {
		logger.Info("First deployment, skipping cleanup")
		return nil
	}

	// Convert current resources to sets for efficient lookup
	currentDeploymentSet := make(map[string]bool)
	for _, name := range currentDeployments {
		currentDeploymentSet[name] = true
	}

	currentServiceSet := make(map[string]bool)
	for _, name := range currentServices {
		currentServiceSet[name] = true
	}

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

	// Update status to deleting if not already set
	if composition.Status.State != environmentsv1.CompositionStateDeleting {
		composition.Status.State = environmentsv1.CompositionStateDeleting
		composition.Status.Message = "Deleting composition and associated resources"
		if err := r.Status().Update(ctx, composition); err != nil {
			logger.Error("Failed to update status to deleting", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

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

// fetchEnvironmentData fetches environment envvars and config files from ConfigMaps and Secrets
func (r *CompositionReconciler) fetchEnvironmentData(ctx context.Context, namespace string, logger *zap.Logger) (*EnvironmentData, error) {
	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	// Fetch environment envvars ConfigMap
	envVarsConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      "env-config",
		Namespace: namespace,
	}, envVarsConfigMap)
	if err == nil && envVarsConfigMap.Data != nil {
		logger.Info("Loaded environment envvars from ConfigMap", zap.Int("count", len(envVarsConfigMap.Data)))
		for k, v := range envVarsConfigMap.Data {
			envData.EnvVars[k] = v
		}
	} else if !apierrors.IsNotFound(err) {
		logger.Error("Failed to fetch environment envvars ConfigMap", zap.Error(err))
	}

	// Fetch environment envvars Secret
	envVarsSecret := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      "env-secret",
		Namespace: namespace,
	}, envVarsSecret)
	if err == nil && envVarsSecret.Data != nil {
		logger.Info("Loaded environment secrets from Secret", zap.Int("count", len(envVarsSecret.Data)))
		for k, v := range envVarsSecret.Data {
			envData.Secrets[k] = string(v)
		}
	} else if !apierrors.IsNotFound(err) {
		logger.Error("Failed to fetch environment envvars Secret", zap.Error(err))
	}

	// Fetch environment config files from individual ConfigMaps (env-file-*)
	configMapList := &corev1.ConfigMapList{}
	err = r.List(ctx, configMapList, client.InNamespace(namespace), client.MatchingLabels{
		"kloudlite.io/file-type": "environment-file",
	})
	if err == nil {
		logger.Info("Found environment config file ConfigMaps", zap.Int("count", len(configMapList.Items)))
		for _, cm := range configMapList.Items {
			// Extract filename from ConfigMap name (remove "env-file-" prefix)
			filename := strings.TrimPrefix(cm.Name, "env-file-")
			// Get the file content (should be a single key in the ConfigMap data)
			for _, content := range cm.Data {
				envData.ConfigFiles[filename] = content
				break // Only use the first data entry
			}
		}
	} else {
		logger.Error("Failed to list environment config file ConfigMaps", zap.Error(err))
	}

	return envData, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *CompositionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Composition{}).
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
			&environmentsv1.Environment{},
			handler.EnqueueRequestsFromMapFunc(r.findCompositionsForEnvironment),
		).
		Complete(r)
}

// findCompositionsForConfigMap triggers reconciliation of all Compositions when env-config changes
func (r *CompositionReconciler) findCompositionsForConfigMap(ctx context.Context, obj client.Object) []reconcile.Request {
	configMap := obj.(*corev1.ConfigMap)

	// Only trigger for env-config ConfigMap
	if configMap.Name != "env-config" {
		return []reconcile.Request{}
	}

	// List all Compositions in the same namespace
	compositionList := &environmentsv1.CompositionList{}
	if err := r.List(ctx, compositionList, client.InNamespace(configMap.Namespace)); err != nil {
		return []reconcile.Request{}
	}

	// Create reconcile requests for all Compositions
	requests := make([]reconcile.Request, len(compositionList.Items))
	for i, composition := range compositionList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      composition.Name,
				Namespace: composition.Namespace,
			},
		}
	}

	return requests
}

// findCompositionsForSecret triggers reconciliation of all Compositions when env-secret changes
func (r *CompositionReconciler) findCompositionsForSecret(ctx context.Context, obj client.Object) []reconcile.Request {
	secret := obj.(*corev1.Secret)

	// Only trigger for env-secret Secret
	if secret.Name != "env-secret" {
		return []reconcile.Request{}
	}

	// List all Compositions in the same namespace
	compositionList := &environmentsv1.CompositionList{}
	if err := r.List(ctx, compositionList, client.InNamespace(secret.Namespace)); err != nil {
		return []reconcile.Request{}
	}

	// Create reconcile requests for all Compositions
	requests := make([]reconcile.Request, len(compositionList.Items))
	for i, composition := range compositionList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      composition.Name,
				Namespace: composition.Namespace,
			},
		}
	}

	return requests
}

// findCompositionsForEnvironment triggers reconciliation of all Compositions when environment activation changes
func (r *CompositionReconciler) findCompositionsForEnvironment(ctx context.Context, obj client.Object) []reconcile.Request {
	environment := obj.(*environmentsv1.Environment)

	// Get the target namespace for this environment
	targetNamespace := environment.Spec.TargetNamespace

	// List all Compositions in the environment's namespace
	compositionList := &environmentsv1.CompositionList{}
	if err := r.List(ctx, compositionList, client.InNamespace(targetNamespace)); err != nil {
		return []reconcile.Request{}
	}

	// Create reconcile requests for all Compositions
	requests := make([]reconcile.Request, len(compositionList.Items))
	for i, composition := range compositionList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      composition.Name,
				Namespace: composition.Namespace,
			},
		}
	}

	return requests
}
