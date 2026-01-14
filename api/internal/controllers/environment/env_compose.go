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
	environmentNamespaceLabel  = "kloudlite.io/environment-namespace"
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

	// Check if environment should have active workloads
	// We need to check both spec.activated AND status.state because:
	// - spec.activated indicates user intent
	// - status.state indicates current lifecycle state (snapping, deactivating, etc.)
	environmentActivated := environment.Spec.Activated

	// Environment should be suspended (replicas=0) if:
	// 1. Not activated (spec.activated=false)
	// 2. Currently snapping (taking a snapshot)
	// 3. Currently deactivating (transitioning to inactive)
	// 4. Snapshot restore is in progress
	shouldSuspend := !environmentActivated ||
		environment.Status.State == environmentsv1.EnvironmentStateSnapping ||
		environment.Status.State == environmentsv1.EnvironmentStateDeactivating

	// Also suspend if snapshot restore is in progress
	if environment.Status.SnapshotRestoreStatus != nil {
		phase := environment.Status.SnapshotRestoreStatus.Phase
		if phase != "" && phase != environmentsv1.SnapshotRestorePhaseCompleted {
			shouldSuspend = true
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
		zap.Int("statefulsets", len(resources.StatefulSets)),
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

	// Apply StatefulSets
	deployedStatefulSets := make([]string, 0)
	for _, statefulSet := range resources.StatefulSets {
		// Apply nodeName from WorkMachine
		if environment.Spec.WorkMachineName != "" {
			wm, err := r.getWorkMachine(ctx, environment.Spec.WorkMachineName)
			if err != nil {
				logger.Warn("Failed to get WorkMachine for node assignment",
					zap.String("workmachine", environment.Spec.WorkMachineName),
					zap.Error(err))
			} else {
				if statefulSet.Spec.Template.Spec.NodeSelector == nil {
					statefulSet.Spec.Template.Spec.NodeSelector = make(map[string]string)
				}
				statefulSet.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] = wm.Name
				statefulSet.Spec.Template.Spec.Tolerations = []corev1.Toleration{
					{
						Key:      "kloudlite.io/workmachine",
						Operator: corev1.TolerationOpEqual,
						Value:    wm.Name,
						Effect:   corev1.TaintEffectNoSchedule,
					},
				}
			}
		}

		// Fetch existing StatefulSet to check for original-replicas annotation
		existingSts := &appsv1.StatefulSet{}
		existsInCluster := true
		if err := r.Get(ctx, client.ObjectKey{Name: statefulSet.Name, Namespace: statefulSet.Namespace}, existingSts); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Warn("Failed to fetch existing StatefulSet for replica check",
					zap.String("name", statefulSet.Name),
					zap.Error(err))
			}
			existsInCluster = false
		}

		// Scale to 0 if environment should be suspended
		if shouldSuspend {
			if statefulSet.Spec.Replicas != nil && *statefulSet.Spec.Replicas > 0 {
				if statefulSet.Annotations == nil {
					statefulSet.Annotations = make(map[string]string)
				}
				if _, exists := statefulSet.Annotations[originalReplicasAnnotation]; !exists {
					statefulSet.Annotations[originalReplicasAnnotation] = fmt.Sprintf("%d", *statefulSet.Spec.Replicas)
				}
				zero := int32(0)
				statefulSet.Spec.Replicas = &zero
			}
		} else {
			// Environment is active and not in transitional state - restore original replicas
			if existsInCluster && existingSts.Annotations != nil {
				if originalReplicas, exists := existingSts.Annotations[originalReplicasAnnotation]; exists {
					if replicas, err := strconv.ParseInt(originalReplicas, 10, 32); err == nil && replicas > 0 {
						r := int32(replicas)
						statefulSet.Spec.Replicas = &r
						// Mark annotation for deletion by ensuring it's not in the new object
						// The applyComposeResource function will handle the actual annotation removal
						if statefulSet.Annotations == nil {
							statefulSet.Annotations = make(map[string]string)
						}
						// Don't copy the original-replicas annotation to the new object
						// This will cause it to be removed during update
					}
				}
			}
		}

		if err := r.applyComposeResource(ctx, statefulSet, environment, logger); err != nil {
			environment.Status.ComposeStatus.State = environmentsv1.CompositionStateFailed
			environment.Status.ComposeStatus.Message = fmt.Sprintf("Failed to apply StatefulSet %s: %v", statefulSet.Name, err)
			return true, nil
		}
		deployedStatefulSets = append(deployedStatefulSets, statefulSet.Name)
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
	if err := r.cleanupRemovedComposeResources(ctx, environment, oldDeployedResources, deployedStatefulSets, deployedServices, deployedPVCs, logger); err != nil {
		logger.Warn("Failed to cleanup removed resources", zap.Error(err))
	}

	// Update deployed resources in status
	environment.Status.ComposeStatus.DeployedResources = &environmentsv1.DeployedResources{
		StatefulSets: deployedStatefulSets,
		Services:     deployedServices,
		PVCs:         deployedPVCs,
	}
	environment.Status.ComposeStatus.ServicesCount = int32(len(resources.ServiceNames))

	// Check StatefulSet health
	healthResult, err := r.checkComposeStatefulSetHealth(ctx, environment, logger)
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
		zap.Int("statefulsets", len(deployedStatefulSets)),
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
	// Note: We don't set ownerReferences on compose resources because they're in a different
	// namespace (target namespace) than the Environment (workmachine namespace).
	// Kubernetes doesn't support cross-namespace owner references.
	// Instead, we use labels to track ownership and rely on the Environment's finalizer
	// to clean up the target namespace (which cascades to all resources in it).

	// Ensure the docker-composition and environment-namespace labels are set
	labels := resource.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[dockerCompositionLabel] = environment.Name
	labels[environmentNamespaceLabel] = environment.Namespace
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

	// Handle StatefulSet updates with retry
	if sts, ok := resource.(*appsv1.StatefulSet); ok {
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			existingSts := &appsv1.StatefulSet{}
			if err := r.Get(ctx, client.ObjectKeyFromObject(sts), existingSts); err != nil {
				return err
			}

			// Preserve existing kloudlite annotations, except for annotations that need to be removed
			if existingSts.Annotations != nil && sts.Annotations == nil {
				sts.Annotations = make(map[string]string)
			}
			for k, v := range existingSts.Annotations {
				if _, exists := sts.Annotations[k]; !exists && strings.HasPrefix(k, "kloudlite.io/") {
					// Skip original-replicas annotation if it's not in the new object
					// This allows the reconciler to remove it after restoring replicas
					if k == originalReplicasAnnotation {
						continue
					}
					sts.Annotations[k] = v
				}
			}

			if equality.Semantic.DeepEqual(sts.Spec, existingSts.Spec) &&
				equality.Semantic.DeepEqual(sts.Annotations, existingSts.Annotations) &&
				equality.Semantic.DeepEqual(sts.Labels, existingSts.Labels) {
				return nil
			}

			sts.SetResourceVersion(existingSts.GetResourceVersion())
			return r.Update(ctx, sts)
		})
	}

	resource.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, resource)
}

// cleanupRemovedComposeResources deletes resources that are no longer in the compose file
func (r *EnvironmentReconciler) cleanupRemovedComposeResources(ctx context.Context, environment *environmentsv1.Environment, oldResources *environmentsv1.DeployedResources, currentStatefulSets, currentServices, currentPVCs []string, logger *zap.Logger) error {
	if oldResources == nil {
		return nil
	}

	namespace := environment.Spec.TargetNamespace
	currentStatefulSetSet := makeStringSet(currentStatefulSets)
	currentServiceSet := makeStringSet(currentServices)
	currentPVCSet := makeStringSet(currentPVCs)

	// Delete removed StatefulSets
	for _, name := range oldResources.StatefulSets {
		if !currentStatefulSetSet[name] {
			logger.Info("Deleting removed StatefulSet", zap.String("name", name))
			if err := r.Delete(ctx, &appsv1.StatefulSet{
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

// checkComposeStatefulSetHealth checks the health of compose StatefulSets
func (r *EnvironmentReconciler) checkComposeStatefulSetHealth(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (*ComposeHealthResult, error) {
	if environment.Status.ComposeStatus == nil ||
		environment.Status.ComposeStatus.DeployedResources == nil ||
		len(environment.Status.ComposeStatus.DeployedResources.StatefulSets) == 0 {
		return &ComposeHealthResult{
			State:   environmentsv1.CompositionStateRunning,
			Message: "No StatefulSets to check",
		}, nil
	}

	result := &ComposeHealthResult{
		Services:      make([]environmentsv1.ServiceStatus, 0),
		ServicesCount: int32(len(environment.Status.ComposeStatus.DeployedResources.StatefulSets)),
	}

	var failedServices []string
	var degradedServices []string
	var pendingServices []string

	for _, stsName := range environment.Status.ComposeStatus.DeployedResources.StatefulSets {
		sts := &appsv1.StatefulSet{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: environment.Spec.TargetNamespace,
			Name:      stsName,
		}, sts)
		if err != nil {
			continue
		}

		// Get service ports
		var servicePorts []int32
		svc := &corev1.Service{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: environment.Spec.TargetNamespace,
			Name:      stsName,
		}, svc); err == nil {
			for _, port := range svc.Spec.Ports {
				servicePorts = append(servicePorts, port.Port)
			}
		}

		serviceStatus := r.checkSingleStatefulSetHealth(ctx, sts, servicePorts, logger)
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

// checkSingleStatefulSetHealth checks health of a single StatefulSet
func (r *EnvironmentReconciler) checkSingleStatefulSetHealth(ctx context.Context, sts *appsv1.StatefulSet, ports []int32, logger *zap.Logger) environmentsv1.ServiceStatus {
	status := environmentsv1.ServiceStatus{
		Name:     sts.Name,
		State:    "pending",
		Replicas: 0,
		Ports:    ports,
	}

	if sts.Spec.Replicas != nil {
		status.Replicas = *sts.Spec.Replicas
	}
	status.ReadyReplicas = sts.Status.ReadyReplicas

	if len(sts.Spec.Template.Spec.Containers) > 0 {
		status.Image = sts.Spec.Template.Spec.Containers[0].Image
	}

	// If replicas is 0, mark as stopped
	if status.Replicas == 0 {
		status.State = "stopped"
		status.Message = "Scaled to 0 (environment inactive)"
		return status
	}

	// Check pod status
	podList := &corev1.PodList{}
	matchLabels := sts.Spec.Selector.MatchLabels
	if err := r.List(ctx, podList,
		client.InNamespace(sts.Namespace),
		client.MatchingLabels(matchLabels),
	); err != nil {
		if sts.Status.ReadyReplicas >= status.Replicas && status.Replicas > 0 {
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
	if sts.Status.ReadyReplicas >= status.Replicas && status.Replicas > 0 {
		status.State = "running"
		status.Message = "All replicas ready"
		return status
	}

	status.State = "starting"
	status.Message = fmt.Sprintf("%d of %d replicas ready", sts.Status.ReadyReplicas, status.Replicas)
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

	// Delete StatefulSets
	stsList := &appsv1.StatefulSetList{}
	if err := r.List(ctx, stsList, client.InNamespace(namespace), labelSelector); err == nil {
		for _, s := range stsList.Items {
			if err := r.Delete(ctx, &s); err != nil && !apierrors.IsNotFound(err) {
				logger.Warn("Failed to delete StatefulSet", zap.String("name", s.Name), zap.Error(err))
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

	// Delete PVCs (including those created by VolumeClaimTemplates)
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
