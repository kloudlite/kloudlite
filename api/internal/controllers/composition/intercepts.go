package composition

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIntercepts manages service intercepts defined in the composition spec
func (r *CompositionReconciler) reconcileIntercepts(ctx context.Context, composition *v1.Composition, logger *zap.Logger) error {
	if len(composition.Spec.Intercepts) == 0 {
		// No intercepts defined, clean up any existing ones
		return r.cleanupAllIntercepts(ctx, composition, logger)
	}

	// Process each intercept configuration
	for i := range composition.Spec.Intercepts {
		intercept := &composition.Spec.Intercepts[i]

		if intercept.Enabled && intercept.WorkspaceRef != nil {
			// Intercept is enabled and has a workspace reference
			if err := r.reconcileSingleIntercept(ctx, composition, intercept, logger); err != nil {
				logger.Error("Failed to reconcile intercept",
					zap.String("service", intercept.ServiceName),
					zap.Error(err))
				// Continue with other intercepts even if one fails
				continue
			}
		} else {
			// Intercept is disabled or no workspace reference, clean it up
			if err := r.cleanupSingleIntercept(ctx, composition, intercept, logger); err != nil {
				logger.Error("Failed to cleanup intercept",
					zap.String("service", intercept.ServiceName),
					zap.Error(err))
				continue
			}
		}
	}

	return nil
}

// reconcileSingleIntercept sets up a single service intercept with SOCAT pod
func (r *CompositionReconciler) reconcileSingleIntercept(ctx context.Context, composition *v1.Composition, intercept *v1.ServiceInterceptConfig, logger *zap.Logger) error {
	logger = logger.With(zap.String("service", intercept.ServiceName))

	// Step 1: Validate workspace (namespaced)
	workspace := &workspacesv1.Workspace{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      intercept.WorkspaceRef.Name,
		Namespace: intercept.WorkspaceRef.Namespace,
	}, workspace)

	if err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Workspace '%s' not found", intercept.WorkspaceRef.Name), "", intercept.WorkspaceRef.Name, intercept.WorkspaceRef.Namespace, nil, nil)
		return err
	}

	if workspace.Status.Phase != "Running" {
		logger.Warn("Workspace is not running", zap.String("phase", workspace.Status.Phase))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Workspace is not running (phase: %s)", workspace.Status.Phase), "", workspace.Name, workspace.Namespace, nil, nil)
		return fmt.Errorf("workspace is not running")
	}

	// Step 2: Get the service in the composition's namespace
	service := &corev1.Service{}
	serviceName := intercept.ServiceName
	serviceNamespace := composition.Namespace

	err = r.Get(ctx, client.ObjectKey{
		Name:      serviceName,
		Namespace: serviceNamespace,
	}, service)

	if err != nil {
		logger.Error("Failed to get service", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Service '%s' not found", serviceName), "", workspace.Name, workspace.Namespace, nil, nil)
		return err
	}

	// Step 3: Store original service selector if not already stored
	var originalSelector map[string]string
	existingStatus := r.findInterceptStatus(composition, intercept.ServiceName)
	if existingStatus != nil && existingStatus.OriginalServiceSelector != nil {
		originalSelector = existingStatus.OriginalServiceSelector
	} else {
		originalSelector = make(map[string]string)
		for k, v := range service.Spec.Selector {
			originalSelector[k] = v
		}
	}

	// Step 4: Get workspace headless service
	// The headless service runs in the WorkMachine's targetNamespace
	workmachine := &workmachinevl.WorkMachine{}
	err = r.Get(ctx, client.ObjectKey{Name: workspace.Spec.WorkmachineName}, workmachine)
	if err != nil {
		logger.Error("Failed to get WorkMachine", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Failed to get WorkMachine: %v", err), "", workspace.Name, workspace.Namespace, nil, originalSelector)
		return err
	}

	workspaceTargetNamespace := workmachine.Spec.TargetNamespace
	workspaceHeadlessSvcName := fmt.Sprintf("ws-%s-headless", workspace.Name)
	headlessSvc := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      workspaceHeadlessSvcName,
		Namespace: workspaceTargetNamespace,
	}, headlessSvc)

	if err != nil {
		logger.Error("Failed to get workspace headless service", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Workspace headless service not found: %v", err), "", workspace.Name, workspace.Namespace, nil, originalSelector)
		return err
	}

	logger.Info("Found workspace headless service", zap.String("service", workspaceHeadlessSvcName))

	// Step 5: Create SOCAT forwarding pod
	socatPodName := fmt.Sprintf("%s-intercept-%s", serviceName, workspace.Name)

	// Build SOCAT command for all port mappings
	var socatCommands []string
	for _, mapping := range intercept.PortMappings {
		// SOCAT forwards from servicePort to workspace headless service + workspacePort
		workspaceTarget := fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			workspaceHeadlessSvcName, workspaceTargetNamespace, mapping.WorkspacePort)

		socatCmd := fmt.Sprintf("socat TCP-LISTEN:%d,fork,reuseaddr TCP:%s",
			mapping.ServicePort, workspaceTarget)
		socatCommands = append(socatCommands, socatCmd+" &")
	}
	socatCommands = append(socatCommands, "wait") // Keep container running

	socatCommand := strings.Join(socatCommands, "\n")

	// Set short termination grace period for faster cleanup
	terminationGracePeriod := int64(5)

	// Build pod with original service selector labels (so service routes to this pod)
	socatPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      socatPodName,
			Namespace: serviceNamespace,
			Labels:    make(map[string]string),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "socat-forwarder",
					Image:   "alpine/socat:latest",
					Command: []string{"sh", "-c", socatCommand},
					Ports:   []corev1.ContainerPort{},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							Exec: &corev1.ExecAction{
								Command: []string{"sh", "-c", "killall socat"},
							},
						},
					},
				},
			},
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: &terminationGracePeriod,
		},
	}

	// Copy original service selector labels to SOCAT pod
	for k, v := range originalSelector {
		socatPod.Labels[k] = v
	}

	// Add intercept tracking labels
	socatPod.Labels["environments.kloudlite.io/managed"] = "true"
	socatPod.Labels["environments.kloudlite.io/composition"] = composition.Name
	socatPod.Labels["environments.kloudlite.io/service"] = serviceName
	socatPod.Labels["environments.kloudlite.io/workspace"] = workspace.Name

	// Add container ports
	for _, mapping := range intercept.PortMappings {
		protocol := mapping.Protocol
		if protocol == "" {
			protocol = corev1.ProtocolTCP
		}
		socatPod.Spec.Containers[0].Ports = append(socatPod.Spec.Containers[0].Ports, corev1.ContainerPort{
			ContainerPort: mapping.ServicePort,
			Protocol:      protocol,
		})
	}

	// Set owner reference for automatic cleanup
	if err := controllerutil.SetControllerReference(composition, socatPod, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference on SOCAT pod", zap.Error(err))
		return err
	}

	// Create or check SOCAT pod
	existingSocatPod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      socatPodName,
		Namespace: serviceNamespace,
	}, existingSocatPod)

	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Create(ctx, socatPod); err != nil {
				logger.Error("Failed to create SOCAT pod", zap.Error(err))
				r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
					fmt.Sprintf("Failed to create SOCAT pod: %v", err), "", workspace.Name, workspace.Namespace, nil, originalSelector)
				return err
			}
			logger.Info("Created SOCAT forwarding pod",
				zap.String("pod", socatPodName),
				zap.String("workspace", workspace.Name))
		} else {
			logger.Error("Failed to get SOCAT pod", zap.Error(err))
			return err
		}
	} else {
		logger.Info("SOCAT pod already exists", zap.String("pod", socatPodName))
	}

	// Step 6: Delete existing Running pods that match the original service selector
	podList := &corev1.PodList{}
	err = r.List(ctx, podList,
		client.InNamespace(serviceNamespace),
		client.MatchingLabels(originalSelector))

	if err == nil {
		var runningPods []corev1.Pod
		for _, pod := range podList.Items {
			// Skip pods being deleted
			if pod.DeletionTimestamp != nil {
				continue
			}

			// Skip workspace pods
			if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; isWorkspace {
				continue
			}

			// Skip SOCAT pods
			if pod.Labels["environments.kloudlite.io/composition"] == composition.Name {
				continue
			}

			// Only delete Running pods
			if pod.Status.Phase == corev1.PodRunning {
				runningPods = append(runningPods, pod)
			}
		}

		if len(runningPods) > 0 {
			logger.Info("Deleting running pods that match intercepted service", zap.Int("count", len(runningPods)))
			for _, pod := range runningPods {
				logger.Info("Deleting intercepted pod", zap.String("pod", pod.Name))
				if err := r.Delete(ctx, &pod); err != nil {
					logger.Warn("Failed to delete pod", zap.String("pod", pod.Name), zap.Error(err))
				}
			}
		}
	}

	logger.Info("Successfully activated service intercept with SOCAT")

	// Step 7: Update status to Active
	now := metav1.Now()
	r.updateInterceptStatus(composition, intercept.ServiceName, "active",
		fmt.Sprintf("Service '%s' is being intercepted by workspace '%s' via SOCAT pod '%s'",
			serviceName, workspace.Name, socatPodName),
		socatPodName, workspace.Name, workspace.Namespace, &now, originalSelector)

	return nil
}

// cleanupSingleIntercept removes the SOCAT pod for a specific intercept
func (r *CompositionReconciler) cleanupSingleIntercept(ctx context.Context, composition *v1.Composition, intercept *v1.ServiceInterceptConfig, logger *zap.Logger) error {
	logger = logger.With(zap.String("service", intercept.ServiceName))

	// Find the intercept status to get SOCAT pod name
	status := r.findInterceptStatus(composition, intercept.ServiceName)
	if status == nil || status.SOCATPodName == "" {
		// No active intercept
		return nil
	}

	serviceNamespace := composition.Namespace
	socatPodName := status.SOCATPodName

	// Delete SOCAT pod
	socatPod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      socatPodName,
		Namespace: serviceNamespace,
	}, socatPod)

	if err == nil {
		// Check if pod is already terminating
		if socatPod.DeletionTimestamp != nil {
			deletionTime := socatPod.DeletionTimestamp.Time
			timeSinceDeletion := time.Since(deletionTime)

			// If pod has been terminating for more than 30 seconds, force delete it
			if timeSinceDeletion > 30*time.Second {
				logger.Warn("SOCAT pod stuck in terminating state, force deleting",
					zap.String("pod", socatPodName),
					zap.Duration("terminating_for", timeSinceDeletion))

				// Force delete by setting grace period to 0
				gracePeriod := int64(0)
				deleteOptions := client.DeleteOptions{
					GracePeriodSeconds: &gracePeriod,
				}
				if err := r.Delete(ctx, socatPod, &deleteOptions); err != nil && !apierrors.IsNotFound(err) {
					logger.Error("Failed to force delete SOCAT pod", zap.Error(err))
					return err
				}
				logger.Info("Force deleted SOCAT pod")
			}
			return nil
		}

		// Pod exists and not terminating yet, initiate deletion
		if err := r.Delete(ctx, socatPod); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to delete SOCAT pod", zap.Error(err))
			return err
		}
		logger.Info("Initiated deletion of SOCAT pod", zap.String("pod", socatPodName))
	} else if !apierrors.IsNotFound(err) {
		logger.Error("Failed to get SOCAT pod during cleanup", zap.Error(err))
		return err
	}

	// Remove intercept status
	r.removeInterceptStatus(composition, intercept.ServiceName)
	logger.Info("Cleaned up service intercept")

	return nil
}

// cleanupAllIntercepts removes all active intercepts for a composition
func (r *CompositionReconciler) cleanupAllIntercepts(ctx context.Context, composition *v1.Composition, logger *zap.Logger) error {
	if len(composition.Status.ActiveIntercepts) == 0 {
		return nil
	}

	logger.Info("Cleaning up all active intercepts", zap.Int("count", len(composition.Status.ActiveIntercepts)))

	// Delete all SOCAT pods
	for _, status := range composition.Status.ActiveIntercepts {
		if status.SOCATPodName == "" {
			continue
		}

		socatPod := &corev1.Pod{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      status.SOCATPodName,
			Namespace: composition.Namespace,
		}, socatPod)

		if err == nil {
			if err := r.Delete(ctx, socatPod); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete SOCAT pod during cleanup",
					zap.String("pod", status.SOCATPodName),
					zap.Error(err))
			} else {
				logger.Info("Deleted SOCAT pod", zap.String("pod", status.SOCATPodName))
			}
		}
	}

	// Clear all intercept statuses
	composition.Status.ActiveIntercepts = nil

	return nil
}

// updateInterceptStatus updates or adds an intercept status entry
func (r *CompositionReconciler) updateInterceptStatus(composition *v1.Composition, serviceName, phase, message, socatPodName, workspaceName, workspaceNamespace string, startTime *metav1.Time, originalSelector map[string]string) {
	// Find existing status
	for i := range composition.Status.ActiveIntercepts {
		if composition.Status.ActiveIntercepts[i].ServiceName == serviceName {
			// Update existing
			composition.Status.ActiveIntercepts[i].Phase = phase
			composition.Status.ActiveIntercepts[i].Message = message
			composition.Status.ActiveIntercepts[i].SOCATPodName = socatPodName
			composition.Status.ActiveIntercepts[i].WorkspaceName = workspaceName
			composition.Status.ActiveIntercepts[i].WorkspaceNamespace = workspaceNamespace
			if startTime != nil {
				composition.Status.ActiveIntercepts[i].InterceptStartTime = startTime
			}
			if originalSelector != nil {
				composition.Status.ActiveIntercepts[i].OriginalServiceSelector = originalSelector
			}
			return
		}
	}

	// Add new status entry
	status := v1.InterceptStatus{
		ServiceName:             serviceName,
		WorkspaceName:           workspaceName,
		WorkspaceNamespace:      workspaceNamespace,
		Phase:                   phase,
		Message:                 message,
		SOCATPodName:            socatPodName,
		OriginalServiceSelector: originalSelector,
		InterceptStartTime:      startTime,
	}
	composition.Status.ActiveIntercepts = append(composition.Status.ActiveIntercepts, status)
}

// findInterceptStatus finds an intercept status by service name
func (r *CompositionReconciler) findInterceptStatus(composition *v1.Composition, serviceName string) *v1.InterceptStatus {
	for i := range composition.Status.ActiveIntercepts {
		if composition.Status.ActiveIntercepts[i].ServiceName == serviceName {
			return &composition.Status.ActiveIntercepts[i]
		}
	}
	return nil
}

// removeInterceptStatus removes an intercept status entry
func (r *CompositionReconciler) removeInterceptStatus(composition *v1.Composition, serviceName string) {
	for i := range composition.Status.ActiveIntercepts {
		if composition.Status.ActiveIntercepts[i].ServiceName == serviceName {
			// Remove this entry
			composition.Status.ActiveIntercepts = append(
				composition.Status.ActiveIntercepts[:i],
				composition.Status.ActiveIntercepts[i+1:]...,
			)
			return
		}
	}
}
