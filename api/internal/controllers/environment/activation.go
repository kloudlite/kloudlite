package environment

import (
	"context"
	"fmt"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/pagination"
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
	var errors []error

	// 1. Find and disconnect all workspaces connected to this environment
	// Workspaces are cluster-scoped, so list without namespace filter
	logger.Info("Finding workspaces connected to this environment")
	workspaceList := &workspacev1.WorkspaceList{}
	if err := pagination.ListAll(ctx, r, workspaceList); err != nil {
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
				// Collect error but continue with other workspaces
				errors = append(errors, fmt.Errorf("workspace %s: %w", workspace.Name, err))
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

	// Return aggregated error if any workspace disconnect failed
	if len(errors) > 0 {
		return fmt.Errorf("failed to disconnect %d workspaces: %w", len(errors), joinErrors(errors))
	}

	return nil
}

// suspendEnvironment scales down all StatefulSets in the environment
// It stores the original replica count in annotations for later resumption
func (r *EnvironmentReconciler) suspendEnvironment(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	namespace := environment.Spec.TargetNamespace
	const originalReplicasAnnotation = "kloudlite.io/original-replicas"

	// Scale down StatefulSets using pagination
	statefulSets := &appsv1.StatefulSetList{}
	if err := pagination.ListAll(ctx, r, statefulSets, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list StatefulSets: %w", err)
	}

	var errors []error
	for _, sts := range statefulSets.Items {
		if sts.Spec.Replicas != nil && *sts.Spec.Replicas > 0 {
			// Store original replica count in annotation
			if sts.Annotations == nil {
				sts.Annotations = make(map[string]string)
			}
			if _, exists := sts.Annotations[originalReplicasAnnotation]; !exists {
				sts.Annotations[originalReplicasAnnotation] = fmt.Sprintf("%d", *sts.Spec.Replicas)
			}

			zero := int32(0)
			sts.Spec.Replicas = &zero
			if err := r.Update(ctx, &sts); err != nil {
				logger.Error("Failed to scale down StatefulSet", zap.String("statefulset", sts.Name), zap.Error(err))
				errors = append(errors, fmt.Errorf("StatefulSet %s: %w", sts.Name, err))
			} else {
				logger.Debug("Scaled down StatefulSet", zap.String("statefulset", sts.Name))
			}
		}
	}

	// Return aggregated error if any StatefulSet scale down failed
	if len(errors) > 0 {
		return fmt.Errorf("failed to scale down %d StatefulSets: %w", len(errors), joinErrors(errors))
	}

	return nil
}
