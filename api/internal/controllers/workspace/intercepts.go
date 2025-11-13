package workspace

import (
	"context"
	"fmt"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// reconcileServiceIntercepts updates Composition.spec.intercepts to match the workspace's desired intercepts
func (r *WorkspaceReconciler) reconcileServiceIntercepts(ctx context.Context, workspace *workspacev1.Workspace, env *environmentv1.Environment, logger *zap.Logger) error {
	// Get desired intercepts from workspace spec
	var desiredIntercepts []workspacev1.InterceptSpec
	if workspace.Spec.EnvironmentConnection != nil {
		desiredIntercepts = workspace.Spec.EnvironmentConnection.Intercepts
	}

	// List all compositions in the environment's target namespace
	compositionList := &environmentv1.CompositionList{}
	if err := r.List(ctx, compositionList, client.InNamespace(env.Spec.TargetNamespace)); err != nil {
		return fmt.Errorf("failed to list compositions in environment namespace: %w", err)
	}

	// Build a map of service name -> desired intercept spec for quick lookup
	desiredInterceptMap := make(map[string]workspacev1.InterceptSpec)
	for _, interceptSpec := range desiredIntercepts {
		desiredInterceptMap[interceptSpec.ServiceName] = interceptSpec
	}

	// Update each composition to add/update/remove this workspace's intercepts
	for i := range compositionList.Items {
		composition := &compositionList.Items[i]

		if err := r.updateCompositionIntercepts(ctx, workspace, composition, desiredInterceptMap, logger); err != nil {
			logger.Error("Failed to update composition intercepts",
				zap.String("composition", composition.Name),
				zap.Error(err),
			)
			// Continue with other compositions
		}
	}

	return nil
}

// updateCompositionIntercepts updates a single composition's intercepts for this workspace
func (r *WorkspaceReconciler) updateCompositionIntercepts(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	composition *environmentv1.Composition,
	desiredInterceptMap map[string]workspacev1.InterceptSpec,
	logger *zap.Logger,
) error {
	modified := false

	// Build a map of current intercepts for this workspace
	var newIntercepts []environmentv1.ServiceInterceptConfig

	// First, preserve intercepts from other workspaces
	for _, existingIntercept := range composition.Spec.Intercepts {
		// Check if this intercept belongs to our workspace
		if existingIntercept.WorkspaceRef != nil && existingIntercept.WorkspaceRef.Name == workspace.Name {
			// This intercept belongs to our workspace - we'll handle it below
			continue
		}
		// Keep intercepts from other workspaces
		newIntercepts = append(newIntercepts, existingIntercept)
	}

	// Now add/update intercepts for this workspace
	for serviceName, desiredSpec := range desiredInterceptMap {
		// Check if this service is part of this composition
		// We'll verify the service exists in the composition's namespace
		service := &corev1.Service{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      serviceName,
			Namespace: composition.Namespace,
		}, service)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Debug("Service not found in composition namespace, skipping",
					zap.String("service", serviceName),
					zap.String("composition", composition.Name),
					zap.String("namespace", composition.Namespace),
				)
				// Service doesn't exist in this composition's namespace
				continue
			}
			return fmt.Errorf("failed to check service '%s': %w", serviceName, err)
		}

		// Convert workspace PortMapping to composition PortMapping
		portMappings := make([]environmentv1.PortMapping, len(desiredSpec.PortMappings))
		for i, pm := range desiredSpec.PortMappings {
			portMappings[i] = environmentv1.PortMapping{
				ServicePort:   pm.ServicePort,
				WorkspacePort: pm.WorkspacePort,
				Protocol:      pm.Protocol,
			}
		}

		// Add the intercept for this workspace
		interceptConfig := environmentv1.ServiceInterceptConfig{
			ServiceName:  serviceName,
			PortMappings: portMappings,
			Enabled:      true,
			WorkspaceRef: &corev1.ObjectReference{
				Name: workspace.Name,
			},
		}
		newIntercepts = append(newIntercepts, interceptConfig)
		modified = true

		logger.Info("Adding/updating intercept in composition",
			zap.String("composition", composition.Name),
			zap.String("service", serviceName),
			zap.String("workspace", workspace.Name),
		)
	}

	// Check if we removed any intercepts for this workspace
	if len(composition.Spec.Intercepts) != len(newIntercepts) {
		modified = true
	}

	// Update composition if modified
	if modified {
		composition.Spec.Intercepts = newIntercepts
		if err := r.Update(ctx, composition); err != nil {
			return fmt.Errorf("failed to update composition: %w", err)
		}
		logger.Info("Updated composition intercepts",
			zap.String("composition", composition.Name),
			zap.String("workspace", workspace.Name),
			zap.Int("totalIntercepts", len(newIntercepts)),
		)
	}

	return nil
}

// cleanupServiceIntercepts removes this workspace's intercepts from all compositions
func (r *WorkspaceReconciler) cleanupServiceIntercepts(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// We need to find all compositions that might have this workspace's intercepts
	// Since we don't know which environment namespaces they're in, we'll list all compositions
	compositionList := &environmentv1.CompositionList{}
	if err := r.List(ctx, compositionList); err != nil {
		return fmt.Errorf("failed to list compositions: %w", err)
	}

	for i := range compositionList.Items {
		composition := &compositionList.Items[i]

		// Check if this composition has any intercepts from our workspace
		hasWorkspaceIntercepts := false
		for _, intercept := range composition.Spec.Intercepts {
			if intercept.WorkspaceRef != nil && intercept.WorkspaceRef.Name == workspace.Name {
				hasWorkspaceIntercepts = true
				break
			}
		}

		if !hasWorkspaceIntercepts {
			continue
		}

		// Remove this workspace's intercepts
		var newIntercepts []environmentv1.ServiceInterceptConfig
		for _, intercept := range composition.Spec.Intercepts {
			// Keep intercepts from other workspaces
			if intercept.WorkspaceRef == nil || intercept.WorkspaceRef.Name != workspace.Name {
				newIntercepts = append(newIntercepts, intercept)
			} else {
				logger.Info("Removing workspace intercept from composition",
					zap.String("composition", composition.Name),
					zap.String("service", intercept.ServiceName),
					zap.String("workspace", workspace.Name),
				)
			}
		}

		// Update composition
		composition.Spec.Intercepts = newIntercepts
		if err := r.Update(ctx, composition); err != nil {
			logger.Error("Failed to remove workspace intercepts from composition",
				zap.String("composition", composition.Name),
				zap.Error(err),
			)
			// Continue with other compositions
		} else {
			logger.Info("Removed workspace intercepts from composition",
				zap.String("composition", composition.Name),
				zap.String("workspace", workspace.Name),
			)
		}
	}

	return nil
}

// portMappingsEqual compares two slices of port mappings for equality
func portMappingsEqual(a, b []workspacev1.PortMapping) bool {
	if len(a) != len(b) {
		return false
	}

	// Build map of port mappings from a
	aMap := make(map[int32]workspacev1.PortMapping)
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
