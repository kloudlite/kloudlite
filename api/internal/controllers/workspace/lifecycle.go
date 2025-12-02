package workspace

import (
	"context"
	"fmt"
	"time"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// handleActiveWorkspace ensures the workspace pod is running
func (r *WorkspaceReconciler) handleActiveWorkspace(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
	// Check if workspace is being used as a cloning source
	if workspace.Status.SourceCloningStatus != nil {
		logger.Info("Workspace is being used as cloning source, ensuring it's suspended",
			zap.String("phase", string(workspace.Status.SourceCloningStatus.Phase)),
			zap.String("targetWorkspace", workspace.Status.SourceCloningStatus.TargetWorkspaceName))

		// Get target namespace to check and delete pod if needed
		targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
		if err != nil {
			logger.Warn("Failed to get target namespace for source workspace suspension", zap.Error(err))
		} else {
			// Check if pod exists and delete it to ensure data consistency during cloning
			podName := getWorkspacePodName(workspace)
			pod := &corev1.Pod{}
			err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)
			if err == nil {
				// Pod exists, delete it
				logger.Info("Deleting source workspace pod for cloning", zap.String("pod", podName))
				if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
					logger.Warn("Failed to delete source workspace pod", zap.Error(err))
				}
			}
		}

		// Ensure workspace pod is suspended while being cloned
		workspace.Status.Phase = "Suspended"
		workspace.Status.Message = fmt.Sprintf("Workspace suspended for cloning to %s", workspace.Status.SourceCloningStatus.TargetWorkspaceName)
		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update status for source cloning", zap.Error(err))
		}

		// Requeue to check again later
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		logger.Error("Failed to get target namespace", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to get target namespace: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	// Check if connected environment is deactivated - if so, disconnect the workspace
	if workspace.Spec.EnvironmentConnection != nil {
		envName := workspace.Spec.EnvironmentConnection.EnvironmentRef.Name
		logger.Info("Checking environment connection status", zap.String("environmentRef", envName))

		// Fetch environment directly (don't use validateEnvironmentConnection as it returns error for deactivated envs)
		env := &environmentv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
			if apierrors.IsNotFound(err) {
				// Environment doesn't exist - disconnect workspace
				logger.Info("Disconnecting workspace from deleted environment", zap.String("environment", envName))
				workspace.Spec.EnvironmentConnection = nil
				if err := r.Update(ctx, workspace); err != nil {
					logger.Error("Failed to disconnect from deleted environment", zap.Error(err))
					return reconcile.Result{}, fmt.Errorf("failed to disconnect from deleted environment: %w", err)
				}
				return reconcile.Result{Requeue: true}, nil
			}
			logger.Warn("Failed to fetch environment", zap.Error(err))
		} else {
			logger.Info("Environment fetched", zap.String("environment", env.Name), zap.Bool("activated", env.Spec.Activated))
			if !env.Spec.Activated {
				// Environment is deactivated - disconnect workspace
				logger.Info("Disconnecting workspace from deactivated environment", zap.String("environment", env.Name))
				workspace.Spec.EnvironmentConnection = nil
				if err := r.Update(ctx, workspace); err != nil {
					logger.Error("Failed to disconnect from deactivated environment", zap.Error(err))
					return reconcile.Result{}, fmt.Errorf("failed to disconnect from deactivated environment: %w", err)
				}
				return reconcile.Result{Requeue: true}, nil
			}
		}
	}

	// Check and suspend idle workspace if auto-stop is enabled
	if err := r.checkAndSuspendIdleWorkspace(ctx, workspace, logger); err != nil {
		logger.Warn("Failed to check idle workspace", zap.Error(err))
		// Don't fail reconciliation, just log the warning
	}

	// Ensure PackageRequest is created if packages are defined
	if err := r.ensurePackageRequest(ctx, workspace, logger); err != nil {
		logger.Error("Failed to ensure PackageRequest", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create PackageRequest: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	// Ensure Service is created for workspace SSHD
	if err := r.ensureWorkspaceService(ctx, workspace, logger); err != nil {
		logger.Error("Failed to ensure Service", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create Service: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	// Setup Ingress for HTTP services
	if err := r.setupWorkspaceIngress(ctx, workspace, logger); err != nil {
		logger.Error("Failed to setup Ingress", zap.Error(err))
		// Don't fail reconciliation - Ingress is optional if domain isn't ready yet
		// The error is logged and will be retried on next reconciliation
	}

	// Ensure headless Service is created for service intercepts
	if err := r.ensureWorkspaceHeadlessService(ctx, workspace, logger); err != nil {
		logger.Error("Failed to ensure headless Service", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create headless Service: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	// Sync package installation status from PackageRequest
	if err := r.syncPackageStatus(ctx, workspace, logger); err != nil {
		logger.Warn("Failed to sync package status", zap.Error(err))
		// Don't fail the reconciliation, just log the warning
	}

	// Check if pod already exists
	podName := getWorkspacePodName(workspace)
	pod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)

	if err == nil {
		// Pod exists

		// Check if environment connection changed by comparing target namespaces
		// Note: status.ConnectedEnvironment.Name is display format (owner/name) while
		// spec.EnvironmentConnection.EnvironmentRef.Name is the actual env name, so compare using TargetNamespace
		envChanged := false
		if workspace.Spec.EnvironmentConnection != nil {
			// Workspace has environment connection - fetch env to get target namespace
			envName := workspace.Spec.EnvironmentConnection.EnvironmentRef.Name
			connEnv := &environmentv1.Environment{}
			if err := r.Get(ctx, client.ObjectKey{Name: envName}, connEnv); err == nil {
				expectedTargetNs := connEnv.Spec.TargetNamespace
				if workspace.Status.ConnectedEnvironment == nil ||
					workspace.Status.ConnectedEnvironment.TargetNamespace != expectedTargetNs {
					envChanged = true
					logger.Info("Environment connection changed - will update DNS",
						zap.String("newEnvironment", envName),
						zap.String("targetNamespace", expectedTargetNs))
				}
			}
		} else {
			// Workspace has no environment connection
			if workspace.Status.ConnectedEnvironment != nil {
				envChanged = true
				logger.Info("Environment disconnected - will update DNS")
			}
		}

		// Update DNS if environment changed and pod is running
		if envChanged && pod.Status.Phase == corev1.PodRunning {
			if err := r.updateDNSConfigInRunningPod(ctx, workspace, logger); err != nil {
				logger.Warn("Failed to update DNS config in running pod", zap.Error(err))
				// Don't fail reconciliation, just log the warning
				// The DNS will be correct on next pod restart
			}
		}

		// Always update Kloudlite context file when pod is running to ensure consistency
		// This is lightweight (just a pod exec with cat) and ensures the file stays in sync
		// even if it got out of sync somehow (e.g., from before a controller restart)
		if pod.Status.Phase == corev1.PodRunning {
			if err := r.updateKloudliteContextFile(ctx, workspace, logger); err != nil {
				logger.Warn("Failed to update Kloudlite context file", zap.Error(err))
				// Don't fail reconciliation, just log the warning
			}
		}

		// Update workspace status based on pod phase
		logger.Info("Workspace pod already exists", zap.String("pod", podName), zap.String("podPhase", string(pod.Status.Phase)))

		// Set phase and message based on actual pod status
		phase := "Creating"
		message := "Workspace pod is starting"

		switch pod.Status.Phase {
		case corev1.PodRunning:
			phase = "Running"
			message = "Workspace is running"
		case corev1.PodPending:
			phase = "Creating"
			message = "Workspace pod is pending"
		case corev1.PodFailed:
			phase = "Failed"
			message = fmt.Sprintf("Workspace pod failed: %s", pod.Status.Message)
		case corev1.PodSucceeded:
			phase = "Stopped"
			message = "Workspace pod completed"
		default:
			phase = "Creating"
			message = fmt.Sprintf("Workspace pod status: %s", pod.Status.Phase)
		}

		return r.updateWorkspaceStatus(ctx, workspace, pod, phase, message, logger)
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing pod", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Pod doesn't exist, create it
	logger.Info("Creating workspace pod", zap.String("pod", podName))

	// Ensure Docker config Secret exists for image registry authentication
	registryHost := r.getImageRegistryHost(ctx)
	if err := r.ensureDockerConfigSecret(ctx, workspace, registryHost, targetNamespace, logger); err != nil {
		logger.Warn("Failed to create Docker config Secret", zap.Error(err))
		// Don't fail reconciliation - Docker config is optional for registry auth
	}

	pod, err = r.createWorkspacePod(workspace)
	if err != nil {
		logger.Error("Failed to build workspace pod", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to build pod: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	if err := r.Create(ctx, pod); err != nil {
		logger.Error("Failed to create workspace pod", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create pod: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	// Update workspace status
	workspace.Status.Phase = "Creating"
	workspace.Status.Message = "Workspace pod is being created"
	workspace.Status.PodName = podName
	now := metav1.Now()
	workspace.Status.StartTime = &now
	workspace.Status.LastActivityTime = &now

	if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
	}

	logger.Info("Workspace pod created successfully")
	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// handleSuspendedWorkspace ensures the workspace pod is stopped
func (r *WorkspaceReconciler) handleSuspendedWorkspace(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		logger.Warn("Failed to get target namespace during suspension", zap.Error(err))
		// Set to stopped state anyway since we can't manage the pod
		workspace.Status.Phase = "Stopped"
		workspace.Status.Message = fmt.Sprintf("Failed to get target namespace: %v", err)
		workspace.Status.PodName = ""
		workspace.Status.PodIP = ""
		workspace.Status.NodeName = ""
		now := metav1.Now()
		workspace.Status.StopTime = &now
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
	}

	// Check if pod exists
	podName := getWorkspacePodName(workspace)
	pod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)

	if apierrors.IsNotFound(err) {
		// Pod doesn't exist, workspace is already stopped
		workspace.Status.Phase = "Stopped"
		workspace.Status.Message = "Workspace is stopped"
		workspace.Status.PodName = ""
		workspace.Status.PodIP = ""
		workspace.Status.NodeName = ""
		now := metav1.Now()
		workspace.Status.StopTime = &now

		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update workspace status", zap.Error(err))
		}
		return reconcile.Result{}, nil
	}

	if err != nil {
		logger.Error("Failed to check existing pod", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Delete the workspace pod using the helper method
	forceDeleted, err := r.deleteWorkspacePod(ctx, pod, podName, logger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// If force deleted, mark as stopped immediately
	if forceDeleted {
		workspace.Status.Phase = "Stopped"
		workspace.Status.Message = "Workspace stopped (node not ready)"
		workspace.Status.PodName = ""
		workspace.Status.PodIP = ""
		workspace.Status.NodeName = ""
		now := metav1.Now()
		workspace.Status.StopTime = &now
		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update workspace status", zap.Error(err))
		}
		return reconcile.Result{}, nil
	}

	workspace.Status.Phase = "Stopping"
	workspace.Status.Message = "Workspace is being stopped"
	if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
	}

	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}
