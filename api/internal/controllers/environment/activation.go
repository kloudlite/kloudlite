package environment

import (
	"context"
	"fmt"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// handleEnvironmentDeactivation disconnects workspaces and removes service intercepts when environment is deactivated
func (r *EnvironmentReconciler) handleEnvironmentDeactivation(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	logger.Info("Handling environment deactivation cleanup",
		zap.String("environment", environment.Name),
		zap.String("targetNamespace", environment.Spec.TargetNamespace))

	disconnectedWorkspaces := 0

	// 1. Find and disconnect all workspaces connected to this environment
	// Workspaces are cluster-scoped, so list without namespace filter
	logger.Info("Finding workspaces connected to this environment")
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		logger.Error("Failed to list workspaces", zap.Error(err))
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	for i := range workspaceList.Items {
		workspace := &workspaceList.Items[i]

		// Check if this workspace is connected to the environment being deactivated
		if workspace.Status.ConnectedEnvironment != nil && workspace.Status.ConnectedEnvironment.Name == environment.Name {
			logger.Info("Disconnecting workspace from environment",
				zap.String("workspace", workspace.Name),
				zap.String("environment", environment.Name))

			// Clear the connected environment from workspace status
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
				workspace.Status.ConnectedEnvironment = nil
				return nil
			}, logger); err != nil {
				logger.Error("Failed to disconnect workspace",
					zap.String("workspace", workspace.Name),
					zap.Error(err))
				// Continue with other workspaces instead of failing
				continue
			}

			disconnectedWorkspaces++
			logger.Info("Successfully disconnected workspace",
				zap.String("workspace", workspace.Name),
				zap.String("environment", environment.Name))
		}
	}

	// Note: Service intercepts are now managed as part of Composition.spec.intercepts
	// The composition controller handles their lifecycle automatically

	logger.Info("Environment deactivation cleanup completed",
		zap.String("environment", environment.Name),
		zap.Int("disconnectedWorkspaces", disconnectedWorkspaces))

	return nil
}

// suspendEnvironment scales down all deployments in the environment
// It stores the original replica count in annotations for later resumption
func (r *EnvironmentReconciler) suspendEnvironment(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	namespace := environment.Spec.TargetNamespace
	const originalReplicasAnnotation = "kloudlite.io/original-replicas"

	// Scale down deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, dep := range deployments.Items {
		if dep.Spec.Replicas != nil && *dep.Spec.Replicas > 0 {
			// Store original replica count in annotation
			if dep.Annotations == nil {
				dep.Annotations = make(map[string]string)
			}
			if _, exists := dep.Annotations[originalReplicasAnnotation]; !exists {
				dep.Annotations[originalReplicasAnnotation] = fmt.Sprintf("%d", *dep.Spec.Replicas)
			}

			zero := int32(0)
			dep.Spec.Replicas = &zero
			if err := r.Update(ctx, &dep); err != nil {
				logger.Error("Failed to scale down deployment", zap.String("deployment", dep.Name), zap.Error(err))
			} else {
				logger.Debug("Scaled down deployment", zap.String("deployment", dep.Name))
			}
		}
	}

	return nil
}

// resumeEnvironment scales up deployments to their original replica counts
func (r *EnvironmentReconciler) resumeEnvironment(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	namespace := environment.Spec.TargetNamespace
	const originalReplicasAnnotation = "kloudlite.io/original-replicas"

	// Scale up deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, dep := range deployments.Items {
		if dep.Annotations != nil {
			if originalReplicasStr, exists := dep.Annotations[originalReplicasAnnotation]; exists {
				var originalReplicas int32
				if _, err := fmt.Sscanf(originalReplicasStr, "%d", &originalReplicas); err == nil && originalReplicas > 0 {
					dep.Spec.Replicas = &originalReplicas
					// Remove the annotation after restoring
					delete(dep.Annotations, originalReplicasAnnotation)

					if err := r.Update(ctx, &dep); err != nil {
						logger.Error("Failed to scale up deployment", zap.String("deployment", dep.Name), zap.Error(err))
					} else {
						logger.Debug("Scaled up deployment", zap.String("deployment", dep.Name), zap.Int32("replicas", originalReplicas))
					}
				}
			}
		}
	}

	return nil
}
