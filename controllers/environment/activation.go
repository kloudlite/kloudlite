package environment

import (
	"context"
	"fmt"

	"github.com/kloudlite/kloudlite/pkg/pagination"
	"github.com/kloudlite/kloudlite/pkg/statusutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *EnvironmentReconciler) reconcileActivationState(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (ctrl.Result, error) {
	// Update environment status based on activation state
	desiredState := environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		desiredState = environmentsv1.EnvironmentStateActive
	}

	// Detect activation/deactivation transitions
	currentState := environment.Status.State
	wasActive := currentState == environmentsv1.EnvironmentStateActive
	wasInactive := currentState == environmentsv1.EnvironmentStateInactive
	willBeActive := desiredState == environmentsv1.EnvironmentStateActive
	willBeInactive := desiredState == environmentsv1.EnvironmentStateInactive

	// Handle snapping state separately - scale down but don't go to inactive
	// Snapshot controller will set state back to active when done
	isSnapping := currentState == environmentsv1.EnvironmentStateSnapping
	if isSnapping {
		// Check if there are any in-progress snapshot requests for this environment
		// If not, transition back to active state (handles crash/restart scenarios)
		hasActiveSnapshotOp, err := r.hasActiveSnapshotOperation(ctx, environment)
		if err != nil {
			logger.Warn("Failed to check for active snapshot operations", zap.Error(err))
		} else if !hasActiveSnapshotOp {
			logger.Info("No active snapshot operations, transitioning back to active state")
			setEnvironmentStatus(environment, environmentsv1.EnvironmentStateActive, "Ready")
			return ctrl.Result{Requeue: true}, nil
		}

		// Scale down all deployments for snapshot
		logger.Info("Scaling down deployments for snapshot")
		if err := r.suspendEnvironment(ctx, environment, logger); err != nil {
			logger.Error("Failed to scale down deployments", zap.Error(err))
			// Continue - pods may already be scaled down
		}

		// Wait for all pods to terminate
		if !r.waitForPodsTerminated(ctx, environment.Spec.TargetNamespace, logger) {
			logger.Info("Waiting for pods to terminate for snapshot")
			setEnvironmentStatus(environment, environmentsv1.EnvironmentStateSnapping, "Taking snapshot, waiting for pods to terminate...")
			return ctrl.Result{RequeueAfter: r.Cfg.Environment.PodTerminationRetryInterval}, nil
		}

		// Pods terminated - stay in snapping state, snapshot controller will handle the rest
		logger.Info("All pods terminated, ready for snapshot")
		setEnvironmentStatus(environment, environmentsv1.EnvironmentStateSnapping, "Ready for snapshot")
		return ctrl.Result{}, nil
	}

	// Handle deactivation transition
	isDeactivating := currentState == environmentsv1.EnvironmentStateDeactivating
	if (wasActive && willBeInactive) || isDeactivating {
		// Set to deactivating state first
		if !isDeactivating {
			setEnvironmentStatus(environment, environmentsv1.EnvironmentStateDeactivating, "Environment is being deactivated")
			return ctrl.Result{Requeue: true}, nil
		}

		// Environment is being deactivated - disconnect workspaces and remove service intercepts
		logger.Info("Environment is being deactivated, cleaning up connections")
		if err := r.handleEnvironmentDeactivation(ctx, environment, logger); err != nil {
			logger.Error("Failed to complete environment deactivation cleanup", zap.Error(err))
			// Continue with scaling down even if cleanup partially fails
		}

		// Scale down all deployments in the environment
		logger.Info("Scaling down deployments for environment deactivation")
		if err := r.suspendEnvironment(ctx, environment, logger); err != nil {
			logger.Error("Failed to scale down deployments", zap.Error(err))
			// Continue - pods may already be scaled down
		}

		// Wait for all pods to terminate before marking as inactive
		// This ensures databases have time to checkpoint and WAL is flushed
		if !r.waitForPodsTerminated(ctx, environment.Spec.TargetNamespace, logger) {
			logger.Info("Waiting for pods to terminate before marking environment inactive")
			setEnvironmentStatus(environment, environmentsv1.EnvironmentStateDeactivating, "Waiting for pods to terminate...")
			return ctrl.Result{RequeueAfter: r.Cfg.Environment.PodTerminationRetryInterval}, nil
		}

		// All pods terminated - now set to inactive
		logger.Info("All pods terminated, marking environment as inactive")
		setEnvironmentStatus(environment, environmentsv1.EnvironmentStateInactive, "Environment is inactive")
		return ctrl.Result{}, nil
	}

	// Handle activation transition
	if wasInactive && willBeActive {
		// Set to activating state first
		if currentState != environmentsv1.EnvironmentStateActivating {
			setEnvironmentStatus(environment, environmentsv1.EnvironmentStateActivating, "Environment is being activated")
		}

		// Now set to active
		setEnvironmentStatus(environment, environmentsv1.EnvironmentStateActive, "Environment is active")
		return ctrl.Result{}, nil
	}

	// Only update status if state actually changed (no transition states)
	if currentState != desiredState {
		message := "Environment is inactive"
		if desiredState == environmentsv1.EnvironmentStateActive {
			message = "Environment is active"
		}

		setEnvironmentStatus(environment, desiredState, message)
	} else {
		logger.Debug("Environment status unchanged, skipping status update")
	}

	return ctrl.Result{}, nil
}

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
		if workspace.Status.ConnectedEnvironment != nil &&
			workspace.Status.ConnectedEnvironment.Name == environment.Name &&
			workspace.Status.ConnectedEnvironment.TargetNamespace == environment.Spec.TargetNamespace {
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
