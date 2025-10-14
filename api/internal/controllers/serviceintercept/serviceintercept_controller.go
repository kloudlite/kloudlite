package serviceintercept

import (
	"context"
	"fmt"
	"time"

	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	// Handle based on status
	if intercept.Spec.Status == "active" {
		return r.handleActive(ctx, intercept, logger)
	} else {
		return r.handleInactive(ctx, intercept, logger)
	}
}

// handleActive activates the service intercept
func (r *ServiceInterceptReconciler) handleActive(ctx context.Context, intercept *interceptsv1.ServiceIntercept, logger *zap.Logger) (reconcile.Result, error) {
	// Validate workspace
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
		intercept.Status.Phase = "Failed"
		intercept.Status.Message = fmt.Sprintf("Workspace '%s' not found", intercept.Spec.WorkspaceRef.Name)
		if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
			logger.Warn("Failed to update status", zap.Error(updateErr))
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	if workspace.Status.Phase != "Running" {
		logger.Warn("Workspace is not running", zap.String("phase", workspace.Status.Phase))
		intercept.Status.Phase = "Failed"
		intercept.Status.Message = fmt.Sprintf("Workspace is not running (phase: %s)", workspace.Status.Phase)
		if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
			logger.Warn("Failed to update status", zap.Error(updateErr))
		}
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Get the service
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
		intercept.Status.Phase = "Failed"
		intercept.Status.Message = fmt.Sprintf("Service '%s' not found", intercept.Spec.ServiceRef.Name)
		if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
			logger.Warn("Failed to update status", zap.Error(updateErr))
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Store original service selector if not already stored
	if intercept.Status.OriginalServiceSelector == nil {
		logger.Info("Storing original service selector")
		intercept.Status.OriginalServiceSelector = make(map[string]string)
		for k, v := range service.Spec.Selector {
			intercept.Status.OriginalServiceSelector[k] = v
		}
	}

	// Find pods matching the original service selector
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
				if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; !isWorkspace {
					affectedPods = append(affectedPods, pod.Name)
				}
			}
			intercept.Status.AffectedPodNames = affectedPods
			logger.Info("Found affected pods", zap.Int("count", len(affectedPods)))
		}
	}

	// Store workspace pod info
	intercept.Status.WorkspacePodName = workspace.Status.PodName
	intercept.Status.WorkspacePodIP = workspace.Status.PodIP

	// Step 1: Update workspace service with intercept ports
	workspaceServiceName := fmt.Sprintf("workspace-%s", workspace.Name)
	workspaceService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      workspaceServiceName,
		Namespace: workspaceNamespace,
	}, workspaceService)

	if err != nil {
		logger.Error("Failed to get workspace service", zap.Error(err))
		intercept.Status.Phase = "Failed"
		intercept.Status.Message = fmt.Sprintf("Workspace service '%s' not found", workspaceServiceName)
		if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
			logger.Warn("Failed to update status", zap.Error(updateErr))
		}
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
	}

	// Add ports to workspace service based on port mappings
	workspaceServiceUpdated := false
	for _, mapping := range intercept.Spec.PortMappings {
		portExists := false
		portName := fmt.Sprintf("intercept-%d", mapping.ServicePort)

		// Check if port already exists
		for _, existingPort := range workspaceService.Spec.Ports {
			if existingPort.Port == mapping.WorkspacePort {
				portExists = true
				break
			}
		}

		if !portExists {
			protocol := mapping.Protocol
			if protocol == "" {
				protocol = corev1.ProtocolTCP
			}
			workspaceService.Spec.Ports = append(workspaceService.Spec.Ports, corev1.ServicePort{
				Name:       portName,
				Port:       mapping.WorkspacePort,
				TargetPort: intstr.FromInt(int(mapping.WorkspacePort)),
				Protocol:   protocol,
			})
			workspaceServiceUpdated = true
			logger.Info("Adding port to workspace service",
				zap.String("portName", portName),
				zap.Int32("port", mapping.WorkspacePort))
		}
	}

	if workspaceServiceUpdated {
		if err := r.Update(ctx, workspaceService); err != nil {
			logger.Error("Failed to update workspace service", zap.Error(err))
			intercept.Status.Phase = "Failed"
			intercept.Status.Message = fmt.Sprintf("Failed to update workspace service: %v", err)
			if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
				logger.Warn("Failed to update status", zap.Error(updateErr))
			}
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
		logger.Info("Successfully updated workspace service with intercept ports")
	}

	// Step 2: Clear intercepted service selector
	service.Spec.Selector = nil
	logger.Info("Cleared intercepted service selector for manual Endpoints")

	// Add annotation to track interception
	if service.Annotations == nil {
		service.Annotations = make(map[string]string)
	}
	service.Annotations["intercepts.kloudlite.io/intercepted-by"] = intercept.Name

	// Update the intercepted service
	if err := r.Update(ctx, service); err != nil {
		logger.Error("Failed to update intercepted service", zap.Error(err))
		intercept.Status.Phase = "Failed"
		intercept.Status.Message = fmt.Sprintf("Failed to update service: %v", err)
		if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
			logger.Warn("Failed to update status", zap.Error(updateErr))
		}
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
	}

	// Step 3: Create manual Endpoints object
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: serviceNamespace,
		},
	}

	// Build endpoint ports from port mappings
	var endpointPorts []corev1.EndpointPort
	for _, mapping := range intercept.Spec.PortMappings {
		protocol := mapping.Protocol
		if protocol == "" {
			protocol = corev1.ProtocolTCP
		}

		// Find the service port name for this mapping
		var portName string
		for _, svcPort := range service.Spec.Ports {
			if svcPort.Port == mapping.ServicePort {
				portName = svcPort.Name
				break
			}
		}

		endpointPorts = append(endpointPorts, corev1.EndpointPort{
			Name:     portName,
			Port:     mapping.WorkspacePort,
			Protocol: protocol,
		})
	}

	endpoints.Subsets = []corev1.EndpointSubset{
		{
			Addresses: []corev1.EndpointAddress{
				{
					IP: workspace.Status.PodIP,
				},
			},
			Ports: endpointPorts,
		},
	}

	// Try to create Endpoints, or update if it already exists
	existingEndpoints := &corev1.Endpoints{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      service.Name,
		Namespace: serviceNamespace,
	}, existingEndpoints)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new Endpoints
			if err := r.Create(ctx, endpoints); err != nil {
				logger.Error("Failed to create Endpoints", zap.Error(err))
				intercept.Status.Phase = "Failed"
				intercept.Status.Message = fmt.Sprintf("Failed to create Endpoints: %v", err)
				if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
					logger.Warn("Failed to update status", zap.Error(updateErr))
				}
				return reconcile.Result{RequeueAfter: 10 * time.Second}, err
			}
			logger.Info("Successfully created manual Endpoints pointing to workspace pod",
				zap.String("podIP", workspace.Status.PodIP))
		} else {
			logger.Error("Failed to get Endpoints", zap.Error(err))
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	} else {
		// Update existing Endpoints
		existingEndpoints.Subsets = endpoints.Subsets
		if err := r.Update(ctx, existingEndpoints); err != nil {
			logger.Error("Failed to update Endpoints", zap.Error(err))
			intercept.Status.Phase = "Failed"
			intercept.Status.Message = fmt.Sprintf("Failed to update Endpoints: %v", err)
			if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
				logger.Warn("Failed to update status", zap.Error(updateErr))
			}
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
		logger.Info("Successfully updated manual Endpoints pointing to workspace pod",
			zap.String("podIP", workspace.Status.PodIP))
	}

	// Delete existing Running pods that match the original service selector
	// This ensures old pods don't keep running while new ones are held in Pending
	// We only delete Running pods to avoid infinite reconciliation loops
	var runningPods []corev1.Pod
	for _, pod := range podList.Items {
		// Skip pods that are already terminating
		if pod.DeletionTimestamp != nil {
			continue
		}

		// Skip workspace pods
		if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; isWorkspace {
			continue
		}

		// Only delete Running pods (not Pending ones that are already held by webhook)
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

	logger.Info("Successfully activated service intercept")

	// Update status to Active
	if intercept.Status.Phase != "Active" {
		intercept.Status.Phase = "Active"
		intercept.Status.Message = fmt.Sprintf("Service '%s' is being intercepted by workspace '%s'",
			service.Name, workspace.Name)
		now := metav1.Now()
		intercept.Status.InterceptStartTime = &now
	}

	if err := r.Status().Update(ctx, intercept); err != nil {
		logger.Warn("Failed to update status", zap.Error(err))
	}

	return reconcile.Result{}, nil
}

// handleInactive deactivates the service intercept
func (r *ServiceInterceptReconciler) handleInactive(ctx context.Context, intercept *interceptsv1.ServiceIntercept, logger *zap.Logger) (reconcile.Result, error) {
	// Check if we have the original selector to restore
	if intercept.Status.OriginalServiceSelector == nil {
		// Nothing to restore
		intercept.Status.Phase = "Inactive"
		intercept.Status.Message = "Service intercept is inactive"
		if err := r.Status().Update(ctx, intercept); err != nil {
			logger.Warn("Failed to update status", zap.Error(err))
		}
		return reconcile.Result{}, nil
	}

	// Get the service
	service := &corev1.Service{}
	serviceNamespace := intercept.Spec.ServiceRef.Namespace
	if serviceNamespace == "" {
		serviceNamespace = intercept.Namespace
	}

	err := r.Get(ctx, client.ObjectKey{
		Name:      intercept.Spec.ServiceRef.Name,
		Namespace: serviceNamespace,
	}, service)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Service no longer exists, consider this as successfully deactivated
			logger.Info("Service not found, considering intercept as deactivated")
			intercept.Status.Phase = "Inactive"
			intercept.Status.Message = "Service no longer exists"
			if updateErr := r.Status().Update(ctx, intercept); updateErr != nil {
				logger.Warn("Failed to update status", zap.Error(updateErr))
			}
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get service", zap.Error(err))
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
	}

	// Step 1: Delete manual Endpoints first
	endpoints := &corev1.Endpoints{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      intercept.Spec.ServiceRef.Name,
		Namespace: serviceNamespace,
	}, endpoints)

	if err == nil {
		if err := r.Delete(ctx, endpoints); err != nil {
			logger.Error("Failed to delete Endpoints", zap.Error(err))
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
		logger.Info("Deleted manual Endpoints")
	} else if !apierrors.IsNotFound(err) {
		logger.Error("Failed to get Endpoints", zap.Error(err))
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
	}

	// Step 2: Restore original service selector
	service.Spec.Selector = intercept.Status.OriginalServiceSelector

	// Remove intercept annotation
	if service.Annotations != nil {
		delete(service.Annotations, "intercepts.kloudlite.io/intercepted-by")
	}

	// Update the service
	if err := r.Update(ctx, service); err != nil {
		logger.Error("Failed to restore service", zap.Error(err))
		return reconcile.Result{RequeueAfter: 10 * time.Second}, err
	}

	logger.Info("Successfully deactivated service intercept and restored service")

	// Update status
	intercept.Status.Phase = "Inactive"
	intercept.Status.Message = "Service intercept deactivated, service restored"
	now := metav1.Now()
	intercept.Status.InterceptEndTime = &now

	if err := r.Status().Update(ctx, intercept); err != nil {
		logger.Warn("Failed to update status", zap.Error(err))
	}

	return reconcile.Result{}, nil
}

// handleDeletion handles the deletion of a ServiceIntercept and cleanup
func (r *ServiceInterceptReconciler) handleDeletion(ctx context.Context, intercept *interceptsv1.ServiceIntercept, logger *zap.Logger) (reconcile.Result, error) {
	// First deactivate the intercept to restore the service
	if intercept.Status.OriginalServiceSelector != nil {
		service := &corev1.Service{}
		serviceNamespace := intercept.Spec.ServiceRef.Namespace
		if serviceNamespace == "" {
			serviceNamespace = intercept.Namespace
		}

		err := r.Get(ctx, client.ObjectKey{
			Name:      intercept.Spec.ServiceRef.Name,
			Namespace: serviceNamespace,
		}, service)

		if err == nil {
			// Step 1: Delete manual Endpoints first
			endpoints := &corev1.Endpoints{}
			err := r.Get(ctx, client.ObjectKey{
				Name:      intercept.Spec.ServiceRef.Name,
				Namespace: serviceNamespace,
			}, endpoints)

			if err == nil {
				if err := r.Delete(ctx, endpoints); err != nil {
					logger.Error("Failed to delete Endpoints during deletion", zap.Error(err))
					return reconcile.Result{RequeueAfter: 5 * time.Second}, err
				}
				logger.Info("Deleted manual Endpoints during deletion")
			} else if !apierrors.IsNotFound(err) {
				logger.Error("Failed to get Endpoints during deletion", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			}

			// Step 2: Restore original service selector
			service.Spec.Selector = intercept.Status.OriginalServiceSelector

			// Remove intercept annotation
			if service.Annotations != nil {
				delete(service.Annotations, "intercepts.kloudlite.io/intercepted-by")
			}

			if err := r.Update(ctx, service); err != nil {
				logger.Error("Failed to restore service during deletion", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			}

			logger.Info("Restored service during deletion")

			// Wait for replacement pods to be Running before completing deletion
			podList := &corev1.PodList{}
			err = r.List(ctx, podList,
				client.InNamespace(serviceNamespace),
				client.MatchingLabels(intercept.Status.OriginalServiceSelector))

			if err != nil {
				logger.Error("Failed to list pods during deletion", zap.Error(err))
				return reconcile.Result{RequeueAfter: 2 * time.Second}, err
			}

			// Delete any Pending pods with hold annotation to allow fresh pods to be created
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
		} else if !apierrors.IsNotFound(err) {
			logger.Error("Failed to get service during deletion", zap.Error(err))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, err
		}
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
		Owns(&corev1.Service{}). // Watch Services (though we don't technically own them)
		Complete(r)
}
