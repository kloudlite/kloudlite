package composition

import (
	"context"
	"fmt"
	"strconv"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// deployComposition deploys the docker compose to Kubernetes
func (r *CompositionReconciler) deployComposition(ctx context.Context, composition *compositionsv1.Composition, logger *zap.Logger) error {
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
	var oldDeployedResources *compositionsv1.DeployedResources
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
	resources, err := ConvertComposeToK8s(project, composition, composition.Namespace, envData, environment)
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
		// Apply nodeSelector and tolerations from environment
		if environment != nil {
			deployment.Spec.Template.Spec.NodeSelector = environment.Spec.NodeSelector
			deployment.Spec.Template.Spec.Tolerations = environment.Spec.Tolerations
		}

		// If environment is not activated, scale deployment to 0 replicas
		if !environmentActivated {
			if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 {
				// Store original replica count in annotation (only if not already set)
				if deployment.Annotations == nil {
					deployment.Annotations = make(map[string]string)
				}
				// Only set the annotation if it doesn't already exist
				// This prevents re-adding it every reconciliation
				if _, exists := deployment.Annotations[originalReplicasAnnotation]; !exists {
					deployment.Annotations[originalReplicasAnnotation] = fmt.Sprintf("%d", *deployment.Spec.Replicas)
					logger.Info("Storing original replica count",
						zap.String("deployment", deployment.Name),
						zap.Int32("replicas", *deployment.Spec.Replicas))
				}

				// Scale to 0
				zero := int32(0)
				deployment.Spec.Replicas = &zero
				logger.Info("Scaling deployment to 0 (environment inactive)",
					zap.String("deployment", deployment.Name))
			}
		} else {
			// Environment is active - restore original replicas if they exist
			if deployment.Annotations != nil {
				if originalReplicas, exists := deployment.Annotations[originalReplicasAnnotation]; exists {
					if replicas, err := strconv.ParseInt(originalReplicas, 10, 32); err == nil && replicas > 0 {
						r := int32(replicas)
						deployment.Spec.Replicas = &r
						// Remove the annotation since we've restored the value
						delete(deployment.Annotations, originalReplicasAnnotation)
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
	composition.Status.DeployedResources = &compositionsv1.DeployedResources{
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
func (r *CompositionReconciler) applyResource(ctx context.Context, resource client.Object, composition *compositionsv1.Composition, logger *zap.Logger) error {
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

	// Handle Service updates specially to preserve immutable fields
	if svc, ok := resource.(*corev1.Service); ok {
		existingSvc := existing.(*corev1.Service)

		// Check if the service needs to transition between headless and non-headless
		// ClusterIP is immutable in Kubernetes, so we need to delete and recreate
		needsRecreate := false

		// Case 1: Existing service has ClusterIP assigned, new service should be headless
		if existingSvc.Spec.ClusterIP != "" && existingSvc.Spec.ClusterIP != "None" && svc.Spec.ClusterIP == "None" {
			needsRecreate = true
			logger.Info("Service needs to transition to headless - will delete and recreate",
				zap.String("name", svc.Name),
				zap.String("existingClusterIP", existingSvc.Spec.ClusterIP),
				zap.String("newClusterIP", svc.Spec.ClusterIP))
		}

		// Case 2: Existing service is headless, new service should have ClusterIP assigned
		if existingSvc.Spec.ClusterIP == "None" && svc.Spec.ClusterIP != "None" {
			needsRecreate = true
			logger.Info("Service needs to transition from headless - will delete and recreate",
				zap.String("name", svc.Name),
				zap.String("existingClusterIP", existingSvc.Spec.ClusterIP),
				zap.String("newClusterIP", svc.Spec.ClusterIP))
		}

		if needsRecreate {
			// Delete the existing service
			logger.Info("Deleting existing Service for recreation",
				zap.String("name", svc.Name))
			if err := r.Delete(ctx, existingSvc); err != nil {
				return fmt.Errorf("failed to delete service for recreation: %w", err)
			}

			// Create the new service
			logger.Info("Creating new Service",
				zap.String("name", svc.Name),
				zap.Int("ports", len(svc.Spec.Ports)),
				zap.String("clusterIP", svc.Spec.ClusterIP))
			return r.Create(ctx, resource)
		}

		// Normal update - preserve ClusterIP and other immutable fields
		svc.Spec.ClusterIP = existingSvc.Spec.ClusterIP
		svc.Spec.ClusterIPs = existingSvc.Spec.ClusterIPs
		svc.Spec.IPFamilies = existingSvc.Spec.IPFamilies
		svc.Spec.IPFamilyPolicy = existingSvc.Spec.IPFamilyPolicy

		// Check if specs are equal - if so, skip the update
		if equality.Semantic.DeepEqual(svc.Spec, existingSvc.Spec) {
			logger.Info("Service spec unchanged, skipping update",
				zap.String("name", svc.Name))
			return nil
		}

		logger.Info("Updating Service spec (specs differ)",
			zap.String("name", svc.Name),
			zap.Int("ports", len(svc.Spec.Ports)),
			zap.String("type", string(svc.Spec.Type)))
	}

	// Handle Deployment updates - check if spec has changed
	if deploy, ok := resource.(*appsv1.Deployment); ok {
		existingDeploy := existing.(*appsv1.Deployment)

		// Preserve existing annotations that are not in the new deployment
		// This prevents annotation changes from triggering unnecessary updates
		if existingDeploy.Annotations != nil && deploy.Annotations == nil {
			deploy.Annotations = make(map[string]string)
		}
		for k, v := range existingDeploy.Annotations {
			// Only preserve kloudlite.io annotations that don't exist in new deployment
			if _, exists := deploy.Annotations[k]; !exists && len(k) >= len("kloudlite.io/") && k[:len("kloudlite.io/")] == "kloudlite.io/" {
				deploy.Annotations[k] = v
			}
		}

		specsEqual := equality.Semantic.DeepEqual(deploy.Spec, existingDeploy.Spec)
		annotationsEqual := equality.Semantic.DeepEqual(deploy.Annotations, existingDeploy.Annotations)

		// Check if specs are equal - if so, skip the update
		if specsEqual && annotationsEqual {
			logger.Info("Deployment spec and annotations unchanged, skipping update",
				zap.String("name", deploy.Name))
			return nil
		}

		logger.Info("Deployment will be updated",
			zap.String("name", deploy.Name),
			zap.Bool("specsEqual", specsEqual),
			zap.Bool("annotationsEqual", annotationsEqual),
			zap.Int("newAnnotations", len(deploy.Annotations)),
			zap.Int("existingAnnotations", len(existingDeploy.Annotations)),
			zap.Any("newReplicas", deploy.Spec.Replicas),
			zap.Any("existingReplicas", existingDeploy.Spec.Replicas),
			zap.Any("newAnnotationsMap", deploy.Annotations),
			zap.Any("existingAnnotationsMap", existingDeploy.Annotations))
	}

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
