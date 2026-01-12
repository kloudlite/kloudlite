package environment

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kloudlite/kloudlite/api/internal/controllers/composition"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Labels and annotations for compose resources
	dockerCompositionLabel     = "kloudlite.io/docker-composition"
	originalReplicasAnnotation = "kloudlite.io/original-replicas"
)

// reconcileCompose handles compose deployment for the environment
// Returns true if compose was reconciled, false if no compose spec is set
func (r *EnvironmentReconciler) reconcileCompose(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (bool, error) {
	// Skip if no compose spec
	if environment.Spec.Compose == nil || environment.Spec.Compose.ComposeContent == "" {
		// Clean up compose status if it exists
		if environment.Status.ComposeStatus != nil {
			environment.Status.ComposeStatus = nil
		}
		return false, nil
	}

	logger.Info("Reconciling compose deployment")

	// Initialize compose status if needed
	if environment.Status.ComposeStatus == nil {
		environment.Status.ComposeStatus = &environmentsv1.CompositionStatus{}
	}

	// Check if environment is activated
	environmentActivated := environment.Spec.Activated

	// Check if snapshot restore is in progress
	snapshotRestoreInProgress := false
	if environment.Status.State == environmentsv1.EnvironmentStateSnapping {
		snapshotRestoreInProgress = true
	}
	if environment.Status.SnapshotRestoreStatus != nil {
		phase := environment.Status.SnapshotRestoreStatus.Phase
		if phase != "" && phase != environmentsv1.SnapshotRestorePhaseCompleted {
			snapshotRestoreInProgress = true
		}
	}

	// Save old deployed resources for cleanup comparison
	var oldDeployedResources *environmentsv1.DeployedResources
	if environment.Status.ComposeStatus.DeployedResources != nil {
		oldDeployedResources = environment.Status.ComposeStatus.DeployedResources.DeepCopy()
	}

	// Fetch environment data
	envData, err := r.fetchEnvironmentData(ctx, environment.Spec.TargetNamespace, logger)
	if err != nil {
		logger.Warn("Failed to fetch environment data", zap.Error(err))
		envData = &composition.EnvironmentData{
			EnvVars:     make(map[string]string),
			Secrets:     make(map[string]string),
			ConfigFiles: make(map[string]string),
		}
	}

	// Parse the compose file
	project, err := composition.ParseComposeFile(environment.Spec.Compose.ComposeContent, environment.Name, envData)
	if err != nil {
		logger.Error("Failed to parse compose file", zap.Error(err))
		environment.Status.ComposeStatus.State = environmentsv1.CompositionStateFailed
		environment.Status.ComposeStatus.Message = fmt.Sprintf("Parse error: %v", err)
		return true, nil
	}

	logger.Info("Parsed compose file",
		zap.Int("services", len(project.Services)),
		zap.Int("named_volumes", len(project.Volumes)))

	// Create a temporary Composition object for the converter
	tempComposition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      environment.Name,
			Namespace: environment.Spec.TargetNamespace,
		},
		Spec: *environment.Spec.Compose,
	}

	// Convert to Kubernetes resources
	resources, err := composition.ConvertComposeToK8s(project, tempComposition, environment.Spec.TargetNamespace, envData, environment)
	if err != nil {
		logger.Error("Failed to convert to Kubernetes resources", zap.Error(err))
		environment.Status.ComposeStatus.State = environmentsv1.CompositionStateFailed
		environment.Status.ComposeStatus.Message = fmt.Sprintf("Conversion error: %v", err)
		return true, nil
	}

	logger.Info("Converted to Kubernetes resources",
		zap.Int("deployments", len(resources.Deployments)),
		zap.Int("services", len(resources.Services)),
		zap.Int("pvcs", len(resources.PVCs)))

	// Apply PVCs first
	for _, pvc := range resources.PVCs {
		if err := r.applyComposeResource(ctx, pvc, environment, logger); err != nil {
			environment.Status.ComposeStatus.State = environmentsv1.CompositionStateFailed
			environment.Status.ComposeStatus.Message = fmt.Sprintf("Failed to apply PVC %s: %v", pvc.Name, err)
			return true, nil
		}
	}

	// Apply Deployments
	deployedDeployments := make([]string, 0)
	for _, deployment := range resources.Deployments {
		// Apply nodeName from WorkMachine
		if environment.Spec.WorkMachineName != "" {
			wm, err := r.getWorkMachine(ctx, environment.Spec.WorkMachineName)
			if err != nil {
				logger.Warn("Failed to get WorkMachine for node assignment",
					zap.String("workmachine", environment.Spec.WorkMachineName),
					zap.Error(err))
			} else {
				if deployment.Spec.Template.Spec.NodeSelector == nil {
					deployment.Spec.Template.Spec.NodeSelector = make(map[string]string)
				}
				deployment.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] = wm.Name
				deployment.Spec.Template.Spec.Tolerations = []corev1.Toleration{
					{
						Key:      "kloudlite.io/workmachine",
						Operator: corev1.TolerationOpEqual,
						Value:    wm.Name,
						Effect:   corev1.TaintEffectNoSchedule,
					},
				}
			}
		}

		// Scale to 0 if environment is inactive or snapshot restore in progress
		if !environmentActivated || snapshotRestoreInProgress {
			if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 {
				if deployment.Annotations == nil {
					deployment.Annotations = make(map[string]string)
				}
				if _, exists := deployment.Annotations[originalReplicasAnnotation]; !exists {
					deployment.Annotations[originalReplicasAnnotation] = fmt.Sprintf("%d", *deployment.Spec.Replicas)
				}
				zero := int32(0)
				deployment.Spec.Replicas = &zero
			}
		} else {
			// Environment is active - restore original replicas
			if deployment.Annotations != nil {
				if originalReplicas, exists := deployment.Annotations[originalReplicasAnnotation]; exists {
					if replicas, err := strconv.ParseInt(originalReplicas, 10, 32); err == nil && replicas > 0 {
						r := int32(replicas)
						deployment.Spec.Replicas = &r
						delete(deployment.Annotations, originalReplicasAnnotation)
					}
				}
			}
		}

		if err := r.applyComposeResource(ctx, deployment, environment, logger); err != nil {
			environment.Status.ComposeStatus.State = environmentsv1.CompositionStateFailed
			environment.Status.ComposeStatus.Message = fmt.Sprintf("Failed to apply Deployment %s: %v", deployment.Name, err)
			return true, nil
		}
		deployedDeployments = append(deployedDeployments, deployment.Name)
	}

	// Apply Services
	deployedServices := make([]string, 0)
	for _, service := range resources.Services {
		if err := r.applyComposeResource(ctx, service, environment, logger); err != nil {
			environment.Status.ComposeStatus.State = environmentsv1.CompositionStateFailed
			environment.Status.ComposeStatus.Message = fmt.Sprintf("Failed to apply Service %s: %v", service.Name, err)
			return true, nil
		}
		deployedServices = append(deployedServices, service.Name)
	}

	// Get current PVC names
	deployedPVCs := make([]string, len(resources.PVCs))
	for i, pvc := range resources.PVCs {
		deployedPVCs[i] = pvc.Name
	}

	// Cleanup removed resources
	if err := r.cleanupRemovedComposeResources(ctx, environment, oldDeployedResources, deployedDeployments, deployedServices, deployedPVCs, logger); err != nil {
		logger.Warn("Failed to cleanup removed resources", zap.Error(err))
	}

	// Update deployed resources in status
	environment.Status.ComposeStatus.DeployedResources = &environmentsv1.DeployedResources{
		Deployments: deployedDeployments,
		Services:    deployedServices,
		PVCs:        deployedPVCs,
	}
	environment.Status.ComposeStatus.ServicesCount = int32(len(resources.ServiceNames))

	// Check deployment health
	healthResult, err := r.checkComposeDeploymentHealth(ctx, environment, logger)
	if err != nil {
		environment.Status.ComposeStatus.State = environmentsv1.CompositionStateRunning
		environment.Status.ComposeStatus.Message = "Deployed (health check unavailable)"
	} else {
		environment.Status.ComposeStatus.State = healthResult.State
		environment.Status.ComposeStatus.Message = healthResult.Message
		environment.Status.ComposeStatus.Services = healthResult.Services
		environment.Status.ComposeStatus.RunningCount = healthResult.RunningCount
	}

	logger.Info("Compose deployment completed",
		zap.Int("deployments", len(deployedDeployments)),
		zap.Int("services", len(deployedServices)),
		zap.String("state", string(environment.Status.ComposeStatus.State)))

	// Persist ComposeStatus immediately to prevent loss during subsequent status updates
	// This is necessary because updateEnvironmentStatus may refetch on conflict, overwriting
	// in-memory ComposeStatus changes
	composeStatus := environment.Status.ComposeStatus.DeepCopy()
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Refetch to get latest version
		if err := r.Get(ctx, client.ObjectKeyFromObject(environment), environment); err != nil {
			return err
		}
		environment.Status.ComposeStatus = composeStatus
		return r.Status().Update(ctx, environment)
	}); err != nil {
		logger.Warn("Failed to persist compose status", zap.Error(err))
		// Don't fail - the status will be updated on next reconcile
	}

	return true, nil
}

// fetchEnvironmentData fetches environment variables and secrets from the namespace
func (r *EnvironmentReconciler) fetchEnvironmentData(ctx context.Context, namespace string, logger *zap.Logger) (*composition.EnvironmentData, error) {
	data := &composition.EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	// Fetch env-config ConfigMap
	configMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: "env-config"}, configMap); err == nil {
		for k, v := range configMap.Data {
			data.EnvVars[k] = v
		}
	}

	// Fetch env-secret Secret
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: "env-secret"}, secret); err == nil {
		for k, v := range secret.Data {
			data.Secrets[k] = string(v)
		}
	}

	return data, nil
}

// applyComposeResource creates or updates a Kubernetes resource
func (r *EnvironmentReconciler) applyComposeResource(ctx context.Context, resource client.Object, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Set environment as owner (cluster-scoped owner for namespaced resources)
	// We use owner references that don't block deletion
	// Note: TypeMeta (APIVersion, Kind) isn't populated by controller-runtime's Get method,
	// so we set them explicitly using the SchemeGroupVersion constant
	blockOwnerDeletion := false
	ownerRef := metav1.OwnerReference{
		APIVersion:         environmentsv1.SchemeGroupVersion.String(),
		Kind:               environmentKind,
		Name:               environment.Name,
		UID:                environment.UID,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
	resource.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

	// Ensure the docker-composition label is set
	labels := resource.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[dockerCompositionLabel] = environment.Name
	resource.SetLabels(labels)

	// Try to get existing resource
	existing := resource.DeepCopyObject().(client.Object)
	err := r.Get(ctx, client.ObjectKeyFromObject(resource), existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Creating resource",
				zap.String("name", resource.GetName()),
				zap.String("namespace", resource.GetNamespace()))
			return r.Create(ctx, resource)
		}
		return err
	}

	// PVCs are immutable - skip update
	if _, ok := resource.(*corev1.PersistentVolumeClaim); ok {
		return nil
	}

	// Handle Service updates
	if svc, ok := resource.(*corev1.Service); ok {
		existingSvc := existing.(*corev1.Service)

		// Check if we need to recreate for ClusterIP changes
		needsRecreate := false
		if existingSvc.Spec.ClusterIP != "" && existingSvc.Spec.ClusterIP != "None" && svc.Spec.ClusterIP == "None" {
			needsRecreate = true
		}
		if existingSvc.Spec.ClusterIP == "None" && svc.Spec.ClusterIP != "None" {
			needsRecreate = true
		}

		if needsRecreate {
			if err := r.Delete(ctx, existingSvc); err != nil {
				return fmt.Errorf("failed to delete service for recreation: %w", err)
			}
			return r.Create(ctx, resource)
		}

		// Preserve immutable fields
		svc.Spec.ClusterIP = existingSvc.Spec.ClusterIP
		svc.Spec.ClusterIPs = existingSvc.Spec.ClusterIPs
		svc.Spec.IPFamilies = existingSvc.Spec.IPFamilies
		svc.Spec.IPFamilyPolicy = existingSvc.Spec.IPFamilyPolicy

		if equality.Semantic.DeepEqual(svc.Spec, existingSvc.Spec) {
			return nil
		}
	}

	// Handle Deployment updates with retry
	if deploy, ok := resource.(*appsv1.Deployment); ok {
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			existingDeploy := &appsv1.Deployment{}
			if err := r.Get(ctx, client.ObjectKeyFromObject(deploy), existingDeploy); err != nil {
				return err
			}

			// Preserve existing kloudlite annotations
			if existingDeploy.Annotations != nil && deploy.Annotations == nil {
				deploy.Annotations = make(map[string]string)
			}
			for k, v := range existingDeploy.Annotations {
				if _, exists := deploy.Annotations[k]; !exists && strings.HasPrefix(k, "kloudlite.io/") {
					deploy.Annotations[k] = v
				}
			}

			if equality.Semantic.DeepEqual(deploy.Spec, existingDeploy.Spec) &&
				equality.Semantic.DeepEqual(deploy.Annotations, existingDeploy.Annotations) {
				return nil
			}

			deploy.SetResourceVersion(existingDeploy.GetResourceVersion())
			return r.Update(ctx, deploy)
		})
	}

	resource.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, resource)
}

// cleanupRemovedComposeResources deletes resources that are no longer in the compose file
func (r *EnvironmentReconciler) cleanupRemovedComposeResources(ctx context.Context, environment *environmentsv1.Environment, oldResources *environmentsv1.DeployedResources, currentDeployments, currentServices, currentPVCs []string, logger *zap.Logger) error {
	if oldResources == nil {
		return nil
	}

	namespace := environment.Spec.TargetNamespace
	currentDeploymentSet := makeStringSet(currentDeployments)
	currentServiceSet := makeStringSet(currentServices)
	currentPVCSet := makeStringSet(currentPVCs)

	// Delete removed deployments
	for _, name := range oldResources.Deployments {
		if !currentDeploymentSet[name] {
			logger.Info("Deleting removed deployment", zap.String("name", name))
			if err := r.Delete(ctx, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			}); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
	}

	// Delete removed services
	for _, name := range oldResources.Services {
		if !currentServiceSet[name] {
			logger.Info("Deleting removed service", zap.String("name", name))
			if err := r.Delete(ctx, &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			}); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
	}

	// Delete removed PVCs
	for _, name := range oldResources.PVCs {
		if !currentPVCSet[name] {
			logger.Info("Deleting removed PVC", zap.String("name", name))
			if err := r.Delete(ctx, &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			}); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

// checkComposeDeploymentHealth checks the health of compose deployments
func (r *EnvironmentReconciler) checkComposeDeploymentHealth(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (*ComposeHealthResult, error) {
	if environment.Status.ComposeStatus == nil ||
		environment.Status.ComposeStatus.DeployedResources == nil ||
		len(environment.Status.ComposeStatus.DeployedResources.Deployments) == 0 {
		return &ComposeHealthResult{
			State:   environmentsv1.CompositionStateRunning,
			Message: "No deployments to check",
		}, nil
	}

	result := &ComposeHealthResult{
		Services:      make([]environmentsv1.ServiceStatus, 0),
		ServicesCount: int32(len(environment.Status.ComposeStatus.DeployedResources.Deployments)),
	}

	var failedServices []string
	var degradedServices []string
	var pendingServices []string

	for _, deploymentName := range environment.Status.ComposeStatus.DeployedResources.Deployments {
		deployment := &appsv1.Deployment{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: environment.Spec.TargetNamespace,
			Name:      deploymentName,
		}, deployment)
		if err != nil {
			continue
		}

		// Get service ports
		var servicePorts []int32
		svc := &corev1.Service{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: environment.Spec.TargetNamespace,
			Name:      deploymentName,
		}, svc); err == nil {
			for _, port := range svc.Spec.Ports {
				servicePorts = append(servicePorts, port.Port)
			}
		}

		serviceStatus := r.checkSingleDeploymentHealth(ctx, deployment, servicePorts, logger)
		result.Services = append(result.Services, serviceStatus)

		switch serviceStatus.State {
		case "running":
			result.RunningCount++
		case "failed":
			failedServices = append(failedServices, fmt.Sprintf("%s: %s", serviceStatus.Name, serviceStatus.Message))
		case "starting", "pending":
			if serviceStatus.Message != "" && (strings.Contains(serviceStatus.Message, "ImagePullBackOff") ||
				strings.Contains(serviceStatus.Message, "ErrImagePull") ||
				strings.Contains(serviceStatus.Message, "CrashLoopBackOff")) {
				failedServices = append(failedServices, fmt.Sprintf("%s: %s", serviceStatus.Name, serviceStatus.Message))
			} else if strings.Contains(serviceStatus.Message, "not ready") {
				degradedServices = append(degradedServices, serviceStatus.Name)
			} else {
				pendingServices = append(pendingServices, serviceStatus.Name)
			}
		}
	}

	// Determine overall state
	if len(failedServices) > 0 {
		result.State = environmentsv1.CompositionStateFailed
		result.Message = fmt.Sprintf("Service errors: %s", strings.Join(failedServices, "; "))
	} else if len(degradedServices) > 0 {
		result.State = environmentsv1.CompositionStateDegraded
		result.Message = fmt.Sprintf("Services degraded: %s", strings.Join(degradedServices, ", "))
	} else if len(pendingServices) > 0 {
		result.State = environmentsv1.CompositionStateDeploying
		result.Message = fmt.Sprintf("Services starting: %s", strings.Join(pendingServices, ", "))
	} else if result.RunningCount == result.ServicesCount {
		result.State = environmentsv1.CompositionStateRunning
		result.Message = "All services running"
	} else {
		result.State = environmentsv1.CompositionStateDegraded
		result.Message = fmt.Sprintf("Only %d of %d services running", result.RunningCount, result.ServicesCount)
	}

	return result, nil
}

// ComposeHealthResult contains health check results
type ComposeHealthResult struct {
	State         environmentsv1.CompositionState
	Message       string
	Services      []environmentsv1.ServiceStatus
	RunningCount  int32
	ServicesCount int32
}

// checkSingleDeploymentHealth checks health of a single deployment
func (r *EnvironmentReconciler) checkSingleDeploymentHealth(ctx context.Context, deployment *appsv1.Deployment, ports []int32, logger *zap.Logger) environmentsv1.ServiceStatus {
	status := environmentsv1.ServiceStatus{
		Name:     deployment.Name,
		State:    "pending",
		Replicas: 0,
		Ports:    ports,
	}

	if deployment.Spec.Replicas != nil {
		status.Replicas = *deployment.Spec.Replicas
	}
	status.ReadyReplicas = deployment.Status.ReadyReplicas

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		status.Image = deployment.Spec.Template.Spec.Containers[0].Image
	}

	// If replicas is 0, mark as stopped
	if status.Replicas == 0 {
		status.State = "stopped"
		status.Message = "Scaled to 0 (environment inactive)"
		return status
	}

	// Check deployment conditions
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing && condition.Reason == "ProgressDeadlineExceeded" {
			status.State = "failed"
			status.Message = "Deployment progress deadline exceeded"
			return status
		}
		if condition.Type == appsv1.DeploymentReplicaFailure && condition.Status == corev1.ConditionTrue {
			status.State = "failed"
			status.Message = condition.Message
			return status
		}
	}

	// Check pod status
	podList := &corev1.PodList{}
	matchLabels := deployment.Spec.Selector.MatchLabels
	if err := r.List(ctx, podList,
		client.InNamespace(deployment.Namespace),
		client.MatchingLabels(matchLabels),
	); err != nil {
		if deployment.Status.ReadyReplicas >= status.Replicas && status.Replicas > 0 {
			status.State = "running"
			status.Message = "All replicas ready"
			return status
		}
		status.State = "starting"
		status.Message = "Unable to check pod status"
		return status
	}

	// Check each pod for errors
	for _, pod := range podList.Items {
		switch pod.Status.Phase {
		case corev1.PodFailed:
			status.State = "failed"
			status.Message = fmt.Sprintf("Pod %s failed: %s", pod.Name, pod.Status.Message)
			return status
		case corev1.PodPending:
			errorMsg := r.getPodErrorMessage(&pod)
			if errorMsg != "" {
				status.State = "failed"
				status.Message = errorMsg
				return status
			}
		case corev1.PodRunning:
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					reason := containerStatus.State.Waiting.Reason
					message := containerStatus.State.Waiting.Message
					if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" ||
						reason == "ErrImagePull" || reason == "CreateContainerConfigError" {
						status.State = "failed"
						status.Message = fmt.Sprintf("%s: %s", reason, message)
						return status
					}
				}
			}
		}
	}

	// No errors - check if ready
	if deployment.Status.ReadyReplicas >= status.Replicas && status.Replicas > 0 {
		status.State = "running"
		status.Message = "All replicas ready"
		return status
	}

	status.State = "starting"
	status.Message = fmt.Sprintf("%d of %d replicas ready", deployment.Status.ReadyReplicas, status.Replicas)
	return status
}

// getPodErrorMessage extracts error message from pod
func (r *EnvironmentReconciler) getPodErrorMessage(pod *corev1.Pod) string {
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" ||
				reason == "CrashLoopBackOff" || reason == "CreateContainerConfigError" {
				return fmt.Sprintf("%s: %s", reason, containerStatus.State.Waiting.Message)
			}
		}
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" ||
				reason == "CrashLoopBackOff" || reason == "CreateContainerConfigError" {
				return fmt.Sprintf("%s: %s", reason, containerStatus.State.Waiting.Message)
			}
		}
	}

	return ""
}

// getWorkMachine fetches a WorkMachine by name
func (r *EnvironmentReconciler) getWorkMachine(ctx context.Context, name string) (*workmachinev1.WorkMachine, error) {
	wm := &workmachinev1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: name}, wm); err != nil {
		return nil, err
	}
	return wm, nil
}

// cleanupComposeResources removes all compose resources for an environment
func (r *EnvironmentReconciler) cleanupComposeResources(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	namespace := environment.Spec.TargetNamespace
	labelSelector := client.MatchingLabels{dockerCompositionLabel: environment.Name}

	// Delete deployments
	deploymentList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deploymentList, client.InNamespace(namespace), labelSelector); err == nil {
		for _, d := range deploymentList.Items {
			if err := r.Delete(ctx, &d); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete deployment", zap.String("name", d.Name), zap.Error(err))
			}
		}
	}

	// Delete services
	serviceList := &corev1.ServiceList{}
	if err := r.List(ctx, serviceList, client.InNamespace(namespace), labelSelector); err == nil {
		for _, s := range serviceList.Items {
			if err := r.Delete(ctx, &s); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete service", zap.String("name", s.Name), zap.Error(err))
			}
		}
	}

	// Delete PVCs
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcList, client.InNamespace(namespace), labelSelector); err == nil {
		for _, p := range pvcList.Items {
			if err := r.Delete(ctx, &p); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete PVC", zap.String("name", p.Name), zap.Error(err))
			}
		}
	}

	return nil
}

func makeStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}
