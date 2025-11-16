package workspace

import (
	"context"
	"fmt"
	"time"

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
	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		logger.Error("Failed to get target namespace", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to get target namespace: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace, logger)
		return reconcile.Result{}, err
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
	podName := fmt.Sprintf("workspace-%s", workspace.Name)
	pod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)

	if err == nil {
		// Pod exists

		// Check if environment connection changed
		envChanged := false
		if workspace.Spec.EnvironmentConnection != nil {
			// Workspace has environment connection
			if workspace.Status.ConnectedEnvironment == nil ||
				workspace.Status.ConnectedEnvironment.Name != workspace.Spec.EnvironmentConnection.EnvironmentRef.Name {
				envChanged = true
				logger.Info("Environment connection changed - will update DNS",
					zap.String("newEnvironment", workspace.Spec.EnvironmentConnection.EnvironmentRef.Name))
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

		// Reconcile service intercepts based on environment connection
		if workspace.Spec.EnvironmentConnection != nil && workspace.Status.ConnectedEnvironment != nil {
			// Environment is connected, reconcile intercepts
			env, err := r.validateEnvironmentConnection(ctx, workspace)
			if err == nil && env != nil {
				if err := r.reconcileServiceIntercepts(ctx, workspace, env, logger); err != nil {
					logger.Warn("Failed to reconcile service intercepts", zap.Error(err))
					// Don't fail reconciliation, just log warning
				}
			} else if err != nil {
				logger.Warn("Environment validation failed, skipping intercept reconciliation", zap.Error(err))
			}
		} else {
			// Environment is disconnected, cleanup all intercepts
			if err := r.cleanupServiceIntercepts(ctx, workspace, logger); err != nil {
				logger.Warn("Failed to cleanup service intercepts", zap.Error(err))
				// Don't fail reconciliation, just log warning
			}
		}

		// Update workspace status
		logger.Info("Workspace pod already exists", zap.String("pod", podName))
		return r.updateWorkspaceStatus(ctx, workspace, pod, "Running", "Workspace is running", logger)
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing pod", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Pod doesn't exist, create it
	logger.Info("Creating workspace pod", zap.String("pod", podName))
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
	podName := fmt.Sprintf("workspace-%s", workspace.Name)
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
