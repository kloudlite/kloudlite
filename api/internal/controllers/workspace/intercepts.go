package workspace

import (
	"context"
	"fmt"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// reconcileServiceIntercepts ensures the ServiceIntercept CRs match the workspace spec
func (r *WorkspaceReconciler) reconcileServiceIntercepts(ctx context.Context, workspace *workspacev1.Workspace, env *environmentv1.Environment, logger *zap.Logger) error {
	// Get desired intercepts from workspace spec
	var desiredIntercepts []workspacev1.InterceptSpec
	if workspace.Spec.EnvironmentConnection != nil {
		desiredIntercepts = workspace.Spec.EnvironmentConnection.Intercepts
	}

	// Get current intercepts owned by this workspace in the environment namespace
	currentIntercepts, err := r.getServiceInterceptsForWorkspace(ctx, workspace, env.Spec.TargetNamespace)
	if err != nil {
		return fmt.Errorf("failed to list current intercepts: %w", err)
	}

	// Build map of current intercepts by service name
	currentInterceptMap := make(map[string]*interceptsv1.ServiceIntercept)
	for i := range currentIntercepts {
		intercept := &currentIntercepts[i]
		currentInterceptMap[intercept.Spec.ServiceRef.Name] = intercept
	}

	// Create or update desired intercepts
	for _, desiredSpec := range desiredIntercepts {
		existingIntercept := currentInterceptMap[desiredSpec.ServiceName]

		if existingIntercept == nil {
			// Create new intercept
			logger.Info("Creating service intercept",
				zap.String("workspace", workspace.Name),
				zap.String("service", desiredSpec.ServiceName),
			)
			if err := r.createServiceIntercept(ctx, workspace, env, desiredSpec, logger); err != nil {
				logger.Error("Failed to create service intercept",
					zap.String("service", desiredSpec.ServiceName),
					zap.Error(err),
				)
				// Don't fail reconciliation, just log error
				continue
			}
		} else {
			// Mark as still desired (don't delete)
			delete(currentInterceptMap, desiredSpec.ServiceName)

			// Check if port mappings changed
			if !portMappingsEqual(existingIntercept.Spec.PortMappings, desiredSpec.PortMappings) {
				logger.Info("Updating service intercept port mappings",
					zap.String("workspace", workspace.Name),
					zap.String("service", desiredSpec.ServiceName),
				)
				existingIntercept.Spec.PortMappings = desiredSpec.PortMappings
				if err := r.Update(ctx, existingIntercept); err != nil {
					logger.Error("Failed to update service intercept",
						zap.String("service", desiredSpec.ServiceName),
						zap.Error(err),
					)
					// Don't fail reconciliation, just log error
				}
			}
		}
	}

	// Delete intercepts that are no longer desired
	for serviceName, intercept := range currentInterceptMap {
		logger.Info("Deleting obsolete service intercept",
			zap.String("workspace", workspace.Name),
			zap.String("service", serviceName),
		)
		if err := r.Delete(ctx, intercept); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to delete service intercept",
				zap.String("service", serviceName),
				zap.Error(err),
			)
			// Don't fail reconciliation, just log error
		}
	}

	return nil
}

// createServiceIntercept creates a ServiceIntercept CR for the workspace
func (r *WorkspaceReconciler) createServiceIntercept(ctx context.Context, workspace *workspacev1.Workspace, env *environmentv1.Environment, interceptSpec workspacev1.InterceptSpec, logger *zap.Logger) error {
	// Verify service exists in environment namespace
	service := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      interceptSpec.ServiceName,
		Namespace: env.Spec.TargetNamespace,
	}, service)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("service '%s' not found in environment namespace '%s'",
				interceptSpec.ServiceName, env.Spec.TargetNamespace)
		}
		return fmt.Errorf("failed to get service '%s': %w", interceptSpec.ServiceName, err)
	}

	// Build ServiceIntercept CR (cluster-scoped resource)
	serviceIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", interceptSpec.ServiceName, workspace.Name),
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name": workspace.Name,
				"intercepts.kloudlite.io/service-name":   interceptSpec.ServiceName,
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			WorkspaceRef: corev1.ObjectReference{
				Name: workspace.Name,
			},
			ServiceRef: corev1.ObjectReference{
				Name:      interceptSpec.ServiceName,
				Namespace: env.Spec.TargetNamespace,
			},
			PortMappings: interceptSpec.PortMappings,
		},
	}

	// Note: We cannot use SetControllerReference for cross-namespace ownership
	// The workspace controller is responsible for cleaning up these intercepts

	// Create the ServiceIntercept
	if err := r.Create(ctx, serviceIntercept); err != nil {
		return fmt.Errorf("failed to create ServiceIntercept CR: %w", err)
	}

	logger.Info("Created ServiceIntercept",
		zap.String("workspace", workspace.Name),
		zap.String("service", interceptSpec.ServiceName),
		zap.String("intercept", serviceIntercept.Name),
	)

	return nil
}

// cleanupServiceIntercepts deletes all ServiceIntercepts owned by the workspace
func (r *WorkspaceReconciler) cleanupServiceIntercepts(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// List all ServiceIntercepts that reference this workspace (cluster-scoped)
	interceptList := &interceptsv1.ServiceInterceptList{}
	if err := r.List(ctx, interceptList, client.MatchingLabels{
		"workspaces.kloudlite.io/workspace-name": workspace.Name,
	}); err != nil {
		return fmt.Errorf("failed to list service intercepts: %w", err)
	}

	for i := range interceptList.Items {
		intercept := &interceptList.Items[i]
		logger.Info("Deleting service intercept during cleanup",
			zap.String("workspace", workspace.Name),
			zap.String("intercept", intercept.Name),
		)
		if err := r.Delete(ctx, intercept); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to delete service intercept",
				zap.String("intercept", intercept.Name),
				zap.Error(err),
			)
			// Continue with other intercepts
		}
	}

	return nil
}

// getServiceInterceptsForWorkspace lists all ServiceIntercepts owned by the workspace
// ServiceIntercepts are now cluster-scoped, so we filter by workspace name label
// and check if they target services in the specified environment namespace
func (r *WorkspaceReconciler) getServiceInterceptsForWorkspace(ctx context.Context, workspace *workspacev1.Workspace, envTargetNamespace string) ([]interceptsv1.ServiceIntercept, error) {
	interceptList := &interceptsv1.ServiceInterceptList{}

	// List all ServiceIntercepts with workspace label (cluster-scoped)
	listOpts := []client.ListOption{
		client.MatchingLabels{
			"workspaces.kloudlite.io/workspace-name": workspace.Name,
		},
	}

	if err := r.List(ctx, interceptList, listOpts...); err != nil {
		return nil, err
	}

	// Filter intercepts that target services in the specified environment namespace
	var filtered []interceptsv1.ServiceIntercept
	for _, intercept := range interceptList.Items {
		if intercept.Spec.ServiceRef.Namespace == envTargetNamespace {
			filtered = append(filtered, intercept)
		}
	}

	return filtered, nil
}

// portMappingsEqual compares two slices of port mappings for equality
func portMappingsEqual(a, b []interceptsv1.PortMapping) bool {
	if len(a) != len(b) {
		return false
	}

	// Build map of port mappings from a
	aMap := make(map[int32]interceptsv1.PortMapping)
	for _, pm := range a {
		aMap[pm.ServicePort] = pm
	}

	// Check all mappings in b exist in a with same values
	for _, pm := range b {
		aPm, exists := aMap[pm.ServicePort]
		if !exists {
			return false
		}
		if aPm.WorkspacePort != pm.WorkspacePort || aPm.Protocol != pm.Protocol {
			return false
		}
	}

	return true
}
