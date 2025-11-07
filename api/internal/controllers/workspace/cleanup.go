package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// deleteHostDirectory deletes a directory on the host by creating a temporary pod
// that mounts the host filesystem and removes the directory
func (r *WorkspaceReconciler) deleteHostDirectory(ctx context.Context, hostPath string, workspaceName string, logger *zap.Logger) error {
	// Validate the host path is safe
	if err := r.validateHostPath(hostPath, workspaceName); err != nil {
		logger.Error("Unsafe host path detected, refusing cleanup",
			zap.String("path", hostPath),
			zap.Error(err))
		return err
	}

	// Create a privileged pod to delete the directory on the host
	// We use a Job-like approach with a one-off pod
	cleanupPodName := fmt.Sprintf("cleanup-%s", strings.ReplaceAll(hostPath, "/", "-"))
	if len(cleanupPodName) > 63 {
		// Kubernetes name limit is 63 characters
		cleanupPodName = cleanupPodName[:63]
	}

	// Calculate TTL - pod will be automatically deleted 5 minutes after completion
	ttlSecondsAfterFinished := int64(300) // 5 minutes

	cleanupPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cleanupPodName,
			Namespace: "default", // Use default namespace for cleanup pods
			Labels: map[string]string{
				"app":  "workspace-cleanup",
				"type": "temporary",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			// Kubernetes will automatically delete this pod 5 minutes after it completes
			// This eliminates the need for a background goroutine
			ActiveDeadlineSeconds: &ttlSecondsAfterFinished,
			Containers: []corev1.Container{
				{
					Name:    "cleanup",
					Image:   "alpine:latest",
					Command: []string{"rm", "-rf", hostPath},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host-home",
							MountPath: "/home/kl",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "host-home",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/home/kl",
						},
					},
				},
			},
		},
	}

	logger.Info("Creating cleanup pod to delete workspace directory",
		zap.String("pod", cleanupPodName),
		zap.String("hostPath", hostPath),
	)

	// Create the cleanup pod
	if err := r.Create(ctx, cleanupPod); err != nil {
		if apierrors.IsAlreadyExists(err) {
			logger.Info("Cleanup pod already exists, deleting old one first")
			if err := r.Delete(ctx, cleanupPod); err != nil && !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed to delete existing cleanup pod: %w", err)
			}
			// Wait a bit and retry
			time.Sleep(2 * time.Second)
			if err := r.Create(ctx, cleanupPod); err != nil {
				return fmt.Errorf("failed to create cleanup pod after retry: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create cleanup pod: %w", err)
		}
	}

	// Kubernetes will automatically clean up the pod after it completes
	// ActiveDeadlineSeconds ensures the pod is terminated and cleaned up
	logger.Info("Cleanup pod created, workspace directory deletion scheduled",
		zap.String("pod", cleanupPodName),
		zap.String("hostPath", hostPath),
		zap.Int64("ttlSeconds", ttlSecondsAfterFinished),
	)

	return nil
}

// handleDeletion cleans up workspace resources when being deleted
func (r *WorkspaceReconciler) handleDeletion(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
		return reconcile.Result{}, nil
	}

	// Check if WorkMachine owner is being deleted
	workMachineBeingDeleted := false
	if workspace.Spec.WorkmachineName != "" {
		wm, err := r.getWorkMachine(ctx, workspace.Spec.WorkmachineName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// WorkMachine already deleted
				workMachineBeingDeleted = true
				logger.Info("WorkMachine not found, removing directory-cleanup finalizer")
			} else {
				logger.Warn("Failed to check WorkMachine status", zap.Error(err))
			}
		} else if wm.DeletionTimestamp != nil {
			// WorkMachine is being deleted
			workMachineBeingDeleted = true
			logger.Info("WorkMachine is being deleted, removing directory-cleanup finalizer")
		}
	}

	// If WorkMachine is being deleted, remove directory-cleanup finalizer
	// since the entire node/VM will be deleted anyway
	if workMachineBeingDeleted && controllerutil.ContainsFinalizer(workspace, "workspaces.kloudlite.io/directory-cleanup") {
		controllerutil.RemoveFinalizer(workspace, "workspaces.kloudlite.io/directory-cleanup")
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to remove directory-cleanup finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Removed directory-cleanup finalizer since WorkMachine is being deleted")
		// Requeue to continue with normal cleanup
		return reconcile.Result{Requeue: true}, nil
	}

	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		logger.Warn("Failed to get target namespace during deletion, skipping pod cleanup", zap.Error(err))
		// Continue with cleanup even if we can't get the namespace
		// The workspace directory cleanup can still proceed
	} else {
		// Delete the workspace pod if it exists
		podName := fmt.Sprintf("workspace-%s", workspace.Name)
		pod := &corev1.Pod{}
		err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)

		if err == nil {
			// Check if node is ready before attempting deletion
			nodeName := pod.Spec.NodeName
			shouldForceDelete := false

			if nodeName != "" {
				node := &corev1.Node{}
				if err := r.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
					if apierrors.IsNotFound(err) {
						logger.Warn("Node not found during deletion, will force delete pod", zap.String("node", nodeName))
						shouldForceDelete = true
					}
				} else {
					nodeReady := false
					for _, condition := range node.Status.Conditions {
						if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
							nodeReady = true
							break
						}
					}
					if !nodeReady {
						logger.Info("Node is not ready during deletion, will force delete pod", zap.String("node", nodeName))
						shouldForceDelete = true
					}
				}
			}

			logger.Info("Deleting workspace pod",
				zap.String("pod", podName),
				zap.Bool("forceDelete", shouldForceDelete))

			if shouldForceDelete {
				gracePeriod := int64(0)
				deleteOptions := &client.DeleteOptions{
					GracePeriodSeconds: &gracePeriod,
				}
				if err := r.Delete(ctx, pod, deleteOptions); err != nil && !apierrors.IsNotFound(err) {
					logger.Error("Failed to force delete workspace pod", zap.Error(err))
					return reconcile.Result{}, err
				}
			} else {
				if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
					logger.Error("Failed to delete workspace pod", zap.Error(err))
					return reconcile.Result{}, err
				}
			}
		} else if !apierrors.IsNotFound(err) {
			logger.Error("Failed to check workspace pod", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Directory cleanup is now handled by workmachine-node-manager via the
	// "workspaces.kloudlite.io/directory-cleanup" finalizer, so we don't need
	// to create cleanup pods anymore

	// Remove finalizer
	controllerutil.RemoveFinalizer(workspace, workspaceFinalizer)
	if err := r.Update(ctx, workspace); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Workspace cleanup completed")
	return reconcile.Result{}, nil
}

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

	// Pod exists, check if node is ready before attempting graceful deletion
	nodeName := pod.Spec.NodeName
	shouldForceDelete := false

	if nodeName != "" {
		// Check if the node is ready
		node := &corev1.Node{}
		if err := r.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Warn("Node not found, will force delete pod", zap.String("node", nodeName))
				shouldForceDelete = true
			} else {
				logger.Warn("Failed to get node status", zap.String("node", nodeName), zap.Error(err))
			}
		} else {
			// Check if node is ready
			nodeReady := false
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					nodeReady = true
					break
				}
			}
			if !nodeReady {
				logger.Info("Node is not ready, will force delete pod", zap.String("node", nodeName))
				shouldForceDelete = true
			}
		}
	}

	// Delete the pod (force delete if node is not ready)
	logger.Info("Deleting workspace pod for suspended workspace",
		zap.String("pod", podName),
		zap.Bool("forceDelete", shouldForceDelete))

	if shouldForceDelete {
		// Force delete the pod with zero grace period
		gracePeriod := int64(0)
		deleteOptions := &client.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
		}
		if err := r.Delete(ctx, pod, deleteOptions); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to force delete workspace pod", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Mark as stopped immediately since we force deleted
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

	// Normal graceful deletion
	if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
		logger.Error("Failed to delete workspace pod", zap.Error(err))
		return reconcile.Result{}, err
	}

	workspace.Status.Phase = "Stopping"
	workspace.Status.Message = "Workspace is being stopped"
	if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
	}

	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}
