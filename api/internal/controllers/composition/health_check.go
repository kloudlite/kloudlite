package composition

import (
	"context"
	"fmt"
	"strings"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeploymentHealthResult contains the health check result for a composition
type DeploymentHealthResult struct {
	State         compositionsv1.CompositionState
	Message       string
	Services      []compositionsv1.ServiceStatus
	RunningCount  int32
	ServicesCount int32
}

// checkDeploymentHealth checks the actual health of all deployments and pods for a composition
func (r *CompositionReconciler) checkDeploymentHealth(ctx context.Context, composition *compositionsv1.Composition, logger *zap.Logger) (*DeploymentHealthResult, error) {
	if composition.Status.DeployedResources == nil || len(composition.Status.DeployedResources.Deployments) == 0 {
		return &DeploymentHealthResult{
			State:   compositionsv1.CompositionStateRunning,
			Message: "No deployments to check",
		}, nil
	}

	result := &DeploymentHealthResult{
		Services:      make([]compositionsv1.ServiceStatus, 0),
		ServicesCount: int32(len(composition.Status.DeployedResources.Deployments)),
	}

	var failedServices []string
	var degradedServices []string
	var pendingServices []string

	for _, deploymentName := range composition.Status.DeployedResources.Deployments {
		deployment := &appsv1.Deployment{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: composition.Namespace,
			Name:      deploymentName,
		}, deployment)
		if err != nil {
			logger.Error("Failed to get deployment for health check",
				zap.String("deployment", deploymentName),
				zap.Error(err))
			continue
		}

		// Get the corresponding Kubernetes Service to extract port information
		var servicePorts []int32
		svc := &corev1.Service{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: composition.Namespace,
			Name:      deploymentName,
		}, svc); err == nil {
			for _, port := range svc.Spec.Ports {
				servicePorts = append(servicePorts, port.Port)
			}
		}

		// Check deployment status
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
				strings.Contains(serviceStatus.Message, "CrashLoopBackOff") ||
				strings.Contains(serviceStatus.Message, "CreateContainerConfigError") ||
				strings.Contains(serviceStatus.Message, "InvalidImageName")) {
				// These are error states that should be reported as failed
				failedServices = append(failedServices, fmt.Sprintf("%s: %s", serviceStatus.Name, serviceStatus.Message))
			} else if serviceStatus.Message != "" && strings.Contains(serviceStatus.Message, "not ready") {
				degradedServices = append(degradedServices, serviceStatus.Name)
			} else {
				pendingServices = append(pendingServices, serviceStatus.Name)
			}
		}
	}

	// Determine overall composition state
	if len(failedServices) > 0 {
		result.State = compositionsv1.CompositionStateFailed
		result.Message = fmt.Sprintf("Service errors: %s", strings.Join(failedServices, "; "))
	} else if len(degradedServices) > 0 {
		result.State = compositionsv1.CompositionStateDegraded
		result.Message = fmt.Sprintf("Services degraded: %s", strings.Join(degradedServices, ", "))
	} else if len(pendingServices) > 0 {
		result.State = compositionsv1.CompositionStateDeploying
		result.Message = fmt.Sprintf("Services starting: %s", strings.Join(pendingServices, ", "))
	} else if result.RunningCount == result.ServicesCount {
		result.State = compositionsv1.CompositionStateRunning
		result.Message = "All services running"
	} else {
		result.State = compositionsv1.CompositionStateDegraded
		result.Message = fmt.Sprintf("Only %d of %d services running", result.RunningCount, result.ServicesCount)
	}

	return result, nil
}

// checkSingleDeploymentHealth checks the health of a single deployment and its pods
func (r *CompositionReconciler) checkSingleDeploymentHealth(ctx context.Context, deployment *appsv1.Deployment, ports []int32, logger *zap.Logger) compositionsv1.ServiceStatus {
	status := compositionsv1.ServiceStatus{
		Name:     deployment.Name,
		State:    "pending",
		Replicas: 0,
		Ports:    ports,
	}

	if deployment.Spec.Replicas != nil {
		status.Replicas = *deployment.Spec.Replicas
	}
	status.ReadyReplicas = deployment.Status.ReadyReplicas

	// Get image from deployment
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		status.Image = deployment.Spec.Template.Spec.Containers[0].Image
	}

	// If replicas is 0 (environment inactive), mark as stopped
	if status.Replicas == 0 {
		status.State = "stopped"
		status.Message = "Scaled to 0 (environment inactive)"
		return status
	}

	// Check deployment conditions for failures
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

	// ALWAYS check ALL pods for errors first, including during rolling updates
	// During rolling updates, old pods may be running while new pods fail
	// We need to detect this and report the error
	podList := &corev1.PodList{}
	matchLabels := deployment.Spec.Selector.MatchLabels
	if err := r.List(ctx, podList,
		client.InNamespace(deployment.Namespace),
		client.MatchingLabels(matchLabels),
	); err != nil {
		logger.Error("Failed to list pods for deployment",
			zap.String("deployment", deployment.Name),
			zap.Error(err))
		// If we can't list pods, fall back to basic ready check
		if deployment.Status.ReadyReplicas >= status.Replicas && status.Replicas > 0 {
			status.State = "running"
			status.Message = "All replicas ready (pod status unavailable)"
			return status
		}
		status.State = "starting"
		status.Message = "Unable to check pod status"
		return status
	}

	// Check each pod for errors - this catches ImagePullBackOff during rolling updates
	for _, pod := range podList.Items {
		// Check pod phase
		switch pod.Status.Phase {
		case corev1.PodFailed:
			status.State = "failed"
			status.Message = fmt.Sprintf("Pod %s failed: %s", pod.Name, pod.Status.Message)
			return status
		case corev1.PodPending:
			// Check container statuses for detailed error
			errorMsg := r.getPodErrorMessage(&pod)
			if errorMsg != "" {
				status.State = "failed"
				status.Message = errorMsg
				return status
			}
		case corev1.PodRunning:
			// Check if containers are actually ready or have errors
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					reason := containerStatus.State.Waiting.Reason
					message := containerStatus.State.Waiting.Message
					if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" ||
						reason == "ErrImagePull" || reason == "CreateContainerConfigError" ||
						reason == "InvalidImageName" {
						status.State = "failed"
						status.Message = fmt.Sprintf("%s: %s", reason, message)
						return status
					}
				}
				if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
					status.State = "failed"
					status.Message = fmt.Sprintf("Container %s exited with code %d: %s",
						containerStatus.Name,
						containerStatus.State.Terminated.ExitCode,
						containerStatus.State.Terminated.Reason)
					return status
				}
			}
		}
	}

	// No errors found - check if all replicas are ready
	if deployment.Status.ReadyReplicas >= status.Replicas && status.Replicas > 0 {
		status.State = "running"
		status.Message = "All replicas ready"
		return status
	}

	// Still starting up
	status.State = "starting"
	status.Message = fmt.Sprintf("%d of %d replicas ready", deployment.Status.ReadyReplicas, status.Replicas)
	return status
}

// getPodErrorMessage extracts error message from pod waiting containers
func (r *CompositionReconciler) getPodErrorMessage(pod *corev1.Pod) string {
	// Check init containers first
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			message := containerStatus.State.Waiting.Message
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" ||
				reason == "CrashLoopBackOff" || reason == "CreateContainerConfigError" ||
				reason == "InvalidImageName" {
				if message != "" {
					return fmt.Sprintf("%s: %s", reason, message)
				}
				return reason
			}
		}
		if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
			return fmt.Sprintf("Init container %s failed with exit code %d",
				containerStatus.Name, containerStatus.State.Terminated.ExitCode)
		}
	}

	// Check regular containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			message := containerStatus.State.Waiting.Message
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" ||
				reason == "CrashLoopBackOff" || reason == "CreateContainerConfigError" ||
				reason == "InvalidImageName" {
				if message != "" {
					return fmt.Sprintf("%s: %s", reason, message)
				}
				return reason
			}
		}
	}

	// Check pod conditions
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
			return fmt.Sprintf("Pod not scheduled: %s", condition.Message)
		}
	}

	return ""
}
