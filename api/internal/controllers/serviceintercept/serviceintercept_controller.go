package serviceintercept

import (
	"context"
	"fmt"
	"strings"
	"time"

	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	serviceInterceptFinalizer = "intercepts.kloudlite.io/finalizer"
)

// ServiceInterceptReconciler reconciles ServiceIntercept objects
type ServiceInterceptReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles ServiceIntercept events and manages service interception
func (r *ServiceInterceptReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("serviceintercept", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling ServiceIntercept")

	// Fetch the ServiceIntercept instance
	intercept := &interceptsv1.ServiceIntercept{}
	err := r.Get(ctx, req.NamespacedName, intercept)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ServiceIntercept not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get ServiceIntercept", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if intercept is being deleted
	if intercept.DeletionTimestamp != nil {
		logger.Info("ServiceIntercept is being deleted, starting cleanup")
		return r.handleDeletion(ctx, intercept, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(intercept, serviceInterceptFinalizer) {
		logger.Info("Adding finalizer to ServiceIntercept")
		controllerutil.AddFinalizer(intercept, serviceInterceptFinalizer)
		if err := r.Update(ctx, intercept); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Reconcile the service intercept
	return r.reconcileIntercept(ctx, intercept, logger)
}

// reconcileIntercept sets up the service intercept by creating SOCAT pod and headless service
func (r *ServiceInterceptReconciler) reconcileIntercept(ctx context.Context, intercept *interceptsv1.ServiceIntercept, logger *zap.Logger) (reconcile.Result, error) {
	// Step 1: Validate workspace
	workspace := &workspacesv1.Workspace{}
	workspaceNamespace := intercept.Spec.WorkspaceRef.Namespace
	if workspaceNamespace == "" {
		workspaceNamespace = intercept.Namespace
	}

	err := r.Get(ctx, client.ObjectKey{
		Name:      intercept.Spec.WorkspaceRef.Name,
		Namespace: workspaceNamespace,
	}, workspace)

	if err != nil {
		logger.Error("Failed to get workspace", zap.Error(err))
		if statusErr := statusutil.UpdateStatusWithRetry(ctx, r.Client, intercept, func() error {
			intercept.Status.Phase = "Failed"
			intercept.Status.Message = fmt.Sprintf("Workspace '%s' not found", intercept.Spec.WorkspaceRef.Name)
			return nil
		}, logger); statusErr != nil {
			logger.Error("Failed to update status after workspace error", zap.Error(statusErr))
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	if workspace.Status.Phase != "Running" {
		logger.Warn("Workspace is not running", zap.String("phase", workspace.Status.Phase))
		if statusErr := statusutil.UpdateStatusWithRetry(ctx, r.Client, intercept, func() error {
			intercept.Status.Phase = "Failed"
			intercept.Status.Message = fmt.Sprintf("Workspace is not running (phase: %s)", workspace.Status.Phase)
			return nil
		}, logger); statusErr != nil {
			logger.Error("Failed to update status after workspace phase check", zap.Error(statusErr))
		}
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Step 2: Get the service
	service := &corev1.Service{}
	serviceNamespace := intercept.Spec.ServiceRef.Namespace
	if serviceNamespace == "" {
		serviceNamespace = intercept.Namespace
	}

	err = r.Get(ctx, client.ObjectKey{
		Name:      intercept.Spec.ServiceRef.Name,
		Namespace: serviceNamespace,
	}, service)

	if err != nil {
		logger.Error("Failed to get service", zap.Error(err))
		if statusErr := statusutil.UpdateStatusWithRetry(ctx, r.Client, intercept, func() error {
			intercept.Status.Phase = "Failed"
			intercept.Status.Message = fmt.Sprintf("Service '%s' not found", intercept.Spec.ServiceRef.Name)
			return nil
		}, logger); statusErr != nil {
			logger.Error("Failed to update status after service error", zap.Error(statusErr))
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Step 3: Store original service configuration if not already stored
	if intercept.Status.OriginalServiceSelector == nil {
		logger.Info("Storing original service selector")
		intercept.Status.OriginalServiceSelector = make(map[string]string)
		for k, v := range service.Spec.Selector {
			intercept.Status.OriginalServiceSelector[k] = v
		}
	}

	// Step 4: Find pods matching the original service selector
	podList := &corev1.PodList{}
	if intercept.Status.OriginalServiceSelector != nil {
		err = r.List(ctx, podList,
			client.InNamespace(serviceNamespace),
			client.MatchingLabels(intercept.Status.OriginalServiceSelector))

		if err != nil {
			logger.Error("Failed to list pods", zap.Error(err))
		} else {
			// Store affected pod names
			var affectedPods []string
			for _, pod := range podList.Items {
				// Skip workspace pods
				if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; isWorkspace {
					continue
				}

				// Skip SOCAT pods (they have the intercept label)
				if pod.Labels["intercepts.kloudlite.io/intercept"] == intercept.Name {
					continue
				}

				affectedPods = append(affectedPods, pod.Name)
			}
			intercept.Status.AffectedPodNames = affectedPods
			logger.Info("Found affected pods", zap.Int("count", len(affectedPods)))
		}
	}

	// Store workspace pod info
	intercept.Status.WorkspacePodName = workspace.Status.PodName
	intercept.Status.WorkspacePodIP = workspace.Status.PodIP

	// Step 5: Get workspace headless service (created by workspace controller)
	workspaceHeadlessSvcName := fmt.Sprintf("workspace-%s-headless", workspace.Name)
	headlessSvc := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      workspaceHeadlessSvcName,
		Namespace: workspaceNamespace,
	}, headlessSvc)

	if err != nil {
		logger.Error("Failed to get workspace headless service", zap.Error(err))
		if statusErr := statusutil.UpdateStatusWithRetry(ctx, r.Client, intercept, func() error {
			intercept.Status.Phase = "Failed"
			intercept.Status.Message = fmt.Sprintf("Workspace headless service not found: %v", err)
			return nil
		}, logger); statusErr != nil {
			logger.Error("Failed to update status after headless service error", zap.Error(statusErr))
		}
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
	}

	logger.Info("Found workspace headless service", zap.String("service", workspaceHeadlessSvcName))

	// Step 6: Create SOCAT forwarding pod
	socatPodName := fmt.Sprintf("%s-intercept-%s", service.Name, workspace.Name)

	// Build SOCAT command for all port mappings
	var socatCommands []string
	for _, mapping := range intercept.Spec.PortMappings {
		// SOCAT forwards from servicePort to workspace headless service + workspacePort
		workspaceTarget := fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			workspaceHeadlessSvcName, workspaceNamespace, mapping.WorkspacePort)

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
					Name:  "socat-forwarder",
					Image: "alpine/socat:latest",
					Command: []string{"sh", "-c", socatCommand},
					Ports: []corev1.ContainerPort{},
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
	for k, v := range intercept.Status.OriginalServiceSelector {
		socatPod.Labels[k] = v
	}

	// Add intercept tracking labels
	socatPod.Labels["intercepts.kloudlite.io/managed"] = "true"
	socatPod.Labels["intercepts.kloudlite.io/service"] = service.Name
	socatPod.Labels["intercepts.kloudlite.io/workspace"] = workspace.Name
	socatPod.Labels["intercepts.kloudlite.io/intercept"] = intercept.Name

	// Add container ports
	for _, mapping := range intercept.Spec.PortMappings {
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
	if err := controllerutil.SetControllerReference(intercept, socatPod, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference on SOCAT pod", zap.Error(err))
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
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
				if statusErr := statusutil.UpdateStatusWithRetry(ctx, r.Client, intercept, func() error {
					intercept.Status.Phase = "Failed"
					intercept.Status.Message = fmt.Sprintf("Failed to create SOCAT pod: %v", err)
					return nil
				}, logger); statusErr != nil {
					logger.Error("Failed to update status after SOCAT pod creation error", zap.Error(statusErr))
				}
				return reconcile.Result{RequeueAfter: 10 * time.Second}, err
			}
			logger.Info("Created SOCAT forwarding pod",
				zap.String("pod", socatPodName),
				zap.String("workspace", workspace.Name))
		} else {
			logger.Error("Failed to get SOCAT pod", zap.Error(err))
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	} else {
		logger.Info("SOCAT pod already exists", zap.String("pod", socatPodName))
	}

	// Store SOCAT pod name in status
	intercept.Status.SOCATPodName = socatPodName

	// Step 7: Delete existing Running pods that match the original service selector
	// This prevents old pods from competing with SOCAT pod
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
		if pod.Labels["intercepts.kloudlite.io/intercept"] == intercept.Name {
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

	logger.Info("Successfully activated service intercept with SOCAT")

	// Step 8: Update status to Active
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, intercept, func() error {
		if intercept.Status.Phase != "Active" {
			intercept.Status.Phase = "Active"
			intercept.Status.Message = fmt.Sprintf("Service '%s' is being intercepted by workspace '%s' via SOCAT pod '%s'",
				service.Name, workspace.Name, socatPodName)
			now := metav1.Now()
			intercept.Status.InterceptStartTime = &now
		}
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status to Active", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	return reconcile.Result{}, nil
}

// handleDeletion handles the deletion of a ServiceIntercept and cleanup
func (r *ServiceInterceptReconciler) handleDeletion(ctx context.Context, intercept *interceptsv1.ServiceIntercept, logger *zap.Logger) (reconcile.Result, error) {
	serviceNamespace := intercept.Spec.ServiceRef.Namespace
	if serviceNamespace == "" {
		serviceNamespace = intercept.Namespace
	}

	workspaceNamespace := intercept.Spec.WorkspaceRef.Namespace
	if workspaceNamespace == "" {
		workspaceNamespace = intercept.Namespace
	}

	// Step 1: Delete SOCAT pod
	if intercept.Status.SOCATPodName != "" {
		socatPod := &corev1.Pod{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      intercept.Status.SOCATPodName,
			Namespace: serviceNamespace,
		}, socatPod)

		if err == nil {
			// Check if pod is already terminating
			if socatPod.DeletionTimestamp != nil {
				// Pod is already being deleted
				deletionTime := socatPod.DeletionTimestamp.Time
				timeSinceDeletion := time.Since(deletionTime)

				// If pod has been terminating for more than 30 seconds, force delete it
				if timeSinceDeletion > 30*time.Second {
					logger.Warn("SOCAT pod stuck in terminating state, force deleting",
						zap.String("pod", intercept.Status.SOCATPodName),
						zap.Duration("terminating_for", timeSinceDeletion))

					// Force delete by setting grace period to 0
					gracePeriod := int64(0)
					deleteOptions := client.DeleteOptions{
						GracePeriodSeconds: &gracePeriod,
					}
					if err := r.Delete(ctx, socatPod, &deleteOptions); err != nil && !apierrors.IsNotFound(err) {
						logger.Error("Failed to force delete SOCAT pod", zap.Error(err))
						return reconcile.Result{RequeueAfter: 2 * time.Second}, err
					}
					logger.Info("Force deleted SOCAT pod")
				} else {
					// Still waiting for graceful termination
					logger.Info("Waiting for SOCAT pod to terminate",
						zap.String("pod", intercept.Status.SOCATPodName),
						zap.Duration("terminating_for", timeSinceDeletion))
				}
				return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
			}

			// Pod exists and not terminating yet, initiate deletion
			if err := r.Delete(ctx, socatPod); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete SOCAT pod", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			}
			logger.Info("Initiated deletion of SOCAT pod", zap.String("pod", intercept.Status.SOCATPodName))
			return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
		} else if !apierrors.IsNotFound(err) {
			logger.Error("Failed to get SOCAT pod during deletion", zap.Error(err))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, err
		}
		// Pod not found, already deleted
		logger.Info("SOCAT pod already deleted", zap.String("pod", intercept.Status.SOCATPodName))
	}

	// Step 2: Workspace headless service is managed by workspace controller - don't delete it
	logger.Info("Workspace headless service is managed by workspace controller, skipping deletion")

	// Step 3: Original service is unchanged - no restoration needed
	// Original pods will automatically be recreated by their controllers (Deployment, StatefulSet, etc.)

	// Step 4: Wait for replacement pods to be Running before completing deletion
	// BUT: Skip this check if the namespace is being deleted (Terminating state)
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: serviceNamespace}, namespace)

	namespaceTerminating := false
	if err == nil && namespace.DeletionTimestamp != nil {
		namespaceTerminating = true
		logger.Info("Namespace is being deleted, skipping wait for replacement pods",
			zap.String("namespace", serviceNamespace))
	}

	if !namespaceTerminating && intercept.Status.OriginalServiceSelector != nil {
		podList := &corev1.PodList{}
		err := r.List(ctx, podList,
			client.InNamespace(serviceNamespace),
			client.MatchingLabels(intercept.Status.OriginalServiceSelector))

		if err == nil {
			// Delete any Pending pods with hold annotation to allow fresh pods
			var pendingPodsToDelete []corev1.Pod
			for _, pod := range podList.Items {
				// Skip workspace pods
				if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; isWorkspace {
					continue
				}

				// Skip pods already being deleted
				if pod.DeletionTimestamp != nil {
					continue
				}

				// Find Pending pods with hold annotation
				if pod.Status.Phase == corev1.PodPending {
					if pod.Annotations != nil {
						if heldBy, hasAnnotation := pod.Annotations["intercepts.kloudlite.io/held-by"]; hasAnnotation && heldBy == intercept.Name {
							pendingPodsToDelete = append(pendingPodsToDelete, pod)
						}
					}
				}
			}

			if len(pendingPodsToDelete) > 0 {
				logger.Info("Deleting Pending pods with hold annotation to allow fresh pods", zap.Int("count", len(pendingPodsToDelete)))
				for _, pod := range pendingPodsToDelete {
					logger.Info("Deleting stuck pending pod", zap.String("pod", pod.Name))
					if err := r.Delete(ctx, &pod); err != nil {
						logger.Warn("Failed to delete pending pod", zap.String("pod", pod.Name), zap.Error(err))
					}
				}
				// Requeue to allow fresh pods to be created
				logger.Info("Requeuing to allow fresh pods to be created")
				return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
			}

			// Check if at least one replacement pod is Running
			var hasRunningPod bool
			for _, pod := range podList.Items {
				// Skip workspace pods
				if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; isWorkspace {
					continue
				}

				// Skip SOCAT pods
				if pod.Labels["intercepts.kloudlite.io/intercept"] == intercept.Name {
					continue
				}

				// Skip pods that are being deleted
				if pod.DeletionTimestamp != nil {
					continue
				}

				// Check if pod is Running
				if pod.Status.Phase == corev1.PodRunning {
					hasRunningPod = true
					break
				}
			}

			if !hasRunningPod {
				logger.Info("Waiting for replacement pods to reach Running state before completing deletion")
				return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
			}

			logger.Info("Replacement pods are running, proceeding with deletion")
		} else {
			logger.Error("Failed to list pods during deletion", zap.Error(err))
			// Continue with deletion even if we can't verify pods
		}
	} else if namespaceTerminating {
		logger.Info("Skipping pod replacement check since namespace is terminating")
	}

	// Remove finalizer
	if controllerutil.ContainsFinalizer(intercept, serviceInterceptFinalizer) {
		controllerutil.RemoveFinalizer(intercept, serviceInterceptFinalizer)
		if err := r.Update(ctx, intercept); err != nil {
			logger.Error("Failed to remove finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("ServiceIntercept cleanup completed successfully")
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *ServiceInterceptReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interceptsv1.ServiceIntercept{}).
		Complete(r)
}
