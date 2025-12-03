package composition

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// originalContainerSpec stores the original container configuration for restoration
type originalContainerSpec struct {
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

const (
	interceptAnnotation          = "environments.kloudlite.io/intercepted"
	originalContainerAnnotation  = "environments.kloudlite.io/original-container"
	interceptWorkspaceAnnotation = "environments.kloudlite.io/intercept-workspace"
)

// reconcileSingleIntercept sets up a single service intercept by replacing deployment image
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
			fmt.Sprintf("Workspace '%s' not found", intercept.WorkspaceRef.Name), workspace.Name, workspace.Namespace, nil, nil)
		return err
	}

	if workspace.Status.Phase != "Running" {
		logger.Warn("Workspace is not running", zap.String("phase", workspace.Status.Phase))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Workspace is not running (phase: %s)", workspace.Status.Phase), workspace.Name, workspace.Namespace, nil, nil)
		return fmt.Errorf("workspace is not running")
	}

	serviceName := intercept.ServiceName
	serviceNamespace := composition.Namespace

	// Step 2: Get workspace headless service for SOCAT target
	workmachine := &workmachinevl.WorkMachine{}
	err = r.Get(ctx, client.ObjectKey{Name: workspace.Spec.WorkmachineName}, workmachine)
	if err != nil {
		logger.Error("Failed to get WorkMachine", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Failed to get WorkMachine: %v", err), workspace.Name, workspace.Namespace, nil, nil)
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
			fmt.Sprintf("Workspace headless service not found: %v", err), workspace.Name, workspace.Namespace, nil, nil)
		return err
	}

	logger.Info("Found workspace headless service", zap.String("service", workspaceHeadlessSvcName))

	// Step 3: Get the deployment for this service
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      serviceName,
		Namespace: serviceNamespace,
	}, deployment)

	if err != nil {
		logger.Error("Failed to get deployment", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Deployment '%s' not found", serviceName), workspace.Name, workspace.Namespace, nil, nil)
		return err
	}

	// Step 4: Build SOCAT command for port forwarding
	var socatCommands []string
	for _, mapping := range intercept.PortMappings {
		workspaceTarget := fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			workspaceHeadlessSvcName, workspaceTargetNamespace, mapping.WorkspacePort)

		socatCmd := fmt.Sprintf("socat TCP-LISTEN:%d,fork,reuseaddr TCP:%s",
			mapping.ServicePort, workspaceTarget)
		socatCommands = append(socatCommands, socatCmd+" &")
	}
	socatCommands = append(socatCommands, "wait")
	socatCommand := strings.Join(socatCommands, "\n")

	// Step 5: Replace deployment container with SOCAT forwarder
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}

	// Check if already intercepted
	if deployment.Annotations[interceptAnnotation] == "true" {
		// Already intercepted, check if it's the same workspace
		if deployment.Annotations[interceptWorkspaceAnnotation] == workspace.Name {
			logger.Info("Deployment already intercepted by this workspace")
			// Update status and return
			now := metav1.Now()
			r.updateInterceptStatus(composition, intercept.ServiceName, "active",
				fmt.Sprintf("Service '%s' is being intercepted by workspace '%s'", serviceName, workspace.Name),
				workspace.Name, workspace.Namespace, &now, nil)
			return nil
		}
		// Different workspace - this shouldn't happen, log warning
		logger.Warn("Deployment already intercepted by different workspace",
			zap.String("currentWorkspace", deployment.Annotations[interceptWorkspaceAnnotation]))
	}

	// Store original container spec if not already stored
	if _, exists := deployment.Annotations[originalContainerAnnotation]; !exists {
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			container := deployment.Spec.Template.Spec.Containers[0]
			originalSpec := originalContainerSpec{
				Image:   container.Image,
				Command: container.Command,
				Args:    container.Args,
			}
			originalJSON, err := json.Marshal(originalSpec)
			if err != nil {
				logger.Error("Failed to marshal original container spec", zap.Error(err))
				return err
			}
			deployment.Annotations[originalContainerAnnotation] = string(originalJSON)
		}
	}

	// Mark as intercepted
	deployment.Annotations[interceptAnnotation] = "true"
	deployment.Annotations[interceptWorkspaceAnnotation] = workspace.Name

	// Replace container with SOCAT forwarder
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Image = "alpine/socat:latest"
		deployment.Spec.Template.Spec.Containers[0].Command = []string{"sh", "-c", socatCommand}
		deployment.Spec.Template.Spec.Containers[0].Args = nil
		// Clear probes that won't work with socat
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = nil
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = nil
		deployment.Spec.Template.Spec.Containers[0].StartupProbe = nil
	}

	// Update deployment
	if err := r.Update(ctx, deployment); err != nil {
		logger.Error("Failed to update deployment for intercept", zap.Error(err))
		r.updateInterceptStatus(composition, intercept.ServiceName, "failed",
			fmt.Sprintf("Failed to update deployment: %v", err), workspace.Name, workspace.Namespace, nil, nil)
		return err
	}

	logger.Info("Successfully replaced deployment image with SOCAT forwarder",
		zap.String("deployment", serviceName),
		zap.String("workspace", workspace.Name))

	// Step 6: Update status to Active
	now := metav1.Now()
	r.updateInterceptStatus(composition, intercept.ServiceName, "active",
		fmt.Sprintf("Service '%s' is being intercepted by workspace '%s'", serviceName, workspace.Name),
		workspace.Name, workspace.Namespace, &now, nil)

	return nil
}

// cleanupSingleIntercept restores the original deployment image for a specific intercept
func (r *CompositionReconciler) cleanupSingleIntercept(ctx context.Context, composition *v1.Composition, intercept *v1.ServiceInterceptConfig, logger *zap.Logger) error {
	logger = logger.With(zap.String("service", intercept.ServiceName))

	serviceName := intercept.ServiceName
	serviceNamespace := composition.Namespace

	// Get the deployment
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      serviceName,
		Namespace: serviceNamespace,
	}, deployment)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Deployment doesn't exist, nothing to cleanup
			r.removeInterceptStatus(composition, intercept.ServiceName)
			return nil
		}
		logger.Error("Failed to get deployment during cleanup", zap.Error(err))
		return err
	}

	// Check if deployment is intercepted
	if deployment.Annotations == nil || deployment.Annotations[interceptAnnotation] != "true" {
		// Not intercepted, remove status if any
		r.removeInterceptStatus(composition, intercept.ServiceName)
		return nil
	}

	// Restore original container spec
	if originalJSON, exists := deployment.Annotations[originalContainerAnnotation]; exists {
		var originalSpec originalContainerSpec
		if err := json.Unmarshal([]byte(originalJSON), &originalSpec); err != nil {
			logger.Error("Failed to unmarshal original container spec", zap.Error(err))
			return err
		}

		// Restore container
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			deployment.Spec.Template.Spec.Containers[0].Image = originalSpec.Image
			deployment.Spec.Template.Spec.Containers[0].Command = originalSpec.Command
			deployment.Spec.Template.Spec.Containers[0].Args = originalSpec.Args
		}

		// Remove intercept annotations
		delete(deployment.Annotations, interceptAnnotation)
		delete(deployment.Annotations, originalContainerAnnotation)
		delete(deployment.Annotations, interceptWorkspaceAnnotation)

		// Update deployment
		if err := r.Update(ctx, deployment); err != nil {
			logger.Error("Failed to restore deployment", zap.Error(err))
			return err
		}

		logger.Info("Restored original deployment image",
			zap.String("deployment", serviceName),
			zap.String("image", originalSpec.Image))
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

	// Restore all intercepted deployments
	for _, status := range composition.Status.ActiveIntercepts {
		serviceName := status.ServiceName
		serviceNamespace := composition.Namespace

		deployment := &appsv1.Deployment{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      serviceName,
			Namespace: serviceNamespace,
		}, deployment)

		if err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to get deployment during cleanup",
					zap.String("deployment", serviceName),
					zap.Error(err))
			}
			continue
		}

		// Check if deployment is intercepted
		if deployment.Annotations == nil || deployment.Annotations[interceptAnnotation] != "true" {
			continue
		}

		// Restore original container spec
		if originalJSON, exists := deployment.Annotations[originalContainerAnnotation]; exists {
			var originalSpec originalContainerSpec
			if err := json.Unmarshal([]byte(originalJSON), &originalSpec); err != nil {
				logger.Error("Failed to unmarshal original container spec",
					zap.String("deployment", serviceName),
					zap.Error(err))
				continue
			}

			// Restore container
			if len(deployment.Spec.Template.Spec.Containers) > 0 {
				deployment.Spec.Template.Spec.Containers[0].Image = originalSpec.Image
				deployment.Spec.Template.Spec.Containers[0].Command = originalSpec.Command
				deployment.Spec.Template.Spec.Containers[0].Args = originalSpec.Args
			}

			// Remove intercept annotations
			delete(deployment.Annotations, interceptAnnotation)
			delete(deployment.Annotations, originalContainerAnnotation)
			delete(deployment.Annotations, interceptWorkspaceAnnotation)

			// Update deployment
			if err := r.Update(ctx, deployment); err != nil {
				logger.Error("Failed to restore deployment during cleanup",
					zap.String("deployment", serviceName),
					zap.Error(err))
				continue
			}

			logger.Info("Restored deployment during cleanup",
				zap.String("deployment", serviceName),
				zap.String("image", originalSpec.Image))
		}
	}

	// Clear all intercept statuses
	composition.Status.ActiveIntercepts = nil

	return nil
}

// updateInterceptStatus updates or adds an intercept status entry
func (r *CompositionReconciler) updateInterceptStatus(composition *v1.Composition, serviceName, phase, message, workspaceName, workspaceNamespace string, startTime *metav1.Time, originalSelector map[string]string) {
	// Find existing status
	for i := range composition.Status.ActiveIntercepts {
		if composition.Status.ActiveIntercepts[i].ServiceName == serviceName {
			// Update existing
			composition.Status.ActiveIntercepts[i].Phase = phase
			composition.Status.ActiveIntercepts[i].Message = message
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
