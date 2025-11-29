package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// deleteWorkspacePod deletes a workspace pod, using force delete if the node is not ready
// Returns true if the pod was force deleted, false otherwise
func (r *WorkspaceReconciler) deleteWorkspacePod(ctx context.Context, pod *corev1.Pod, podName string, logger *zap.Logger) (forceDeleted bool, err error) {
	// Check if node is ready before attempting deletion
	nodeName := pod.Spec.NodeName
	shouldForceDelete := false

	if nodeName != "" {
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
			return true, err
		}
		return true, nil
	}

	// Normal graceful deletion
	if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
		logger.Error("Failed to delete workspace pod", zap.Error(err))
		return false, err
	}
	return false, nil
}

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

	// Update status to show workspace is being deleted
	if workspace.Status.Phase != "Terminating" {
		workspace.Status.Phase = "Terminating"
		workspace.Status.Message = "Workspace is being deleted"
		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update status to Terminating", zap.Error(err))
			// Continue with deletion even if status update fails
		}
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
		podName := workspace.Name
		pod := &corev1.Pod{}
		err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)

		if err == nil {
			// Delete the workspace pod using the helper method
			if _, err := r.deleteWorkspacePod(ctx, pod, podName, logger); err != nil {
				return reconcile.Result{}, err
			}
		} else if !apierrors.IsNotFound(err) {
			logger.Error("Failed to check workspace pod", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Directory cleanup is now handled by workmachine-node-manager via the
	// "workspaces.kloudlite.io/directory-cleanup" finalizer, so we don't need
	// to create cleanup pods anymore

	// Delete ClusterRole and ClusterRoleBinding for environments access
	// These cannot have owner references so must be deleted manually
	clusterRoleName := fmt.Sprintf("workspace-%s-%s", workspace.Namespace, workspace.Name)

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}
	if err := r.Delete(ctx, clusterRole); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete ClusterRole", zap.String("name", clusterRoleName), zap.Error(err))
	} else {
		logger.Info("Deleted ClusterRole", zap.String("name", clusterRoleName))
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}
	if err := r.Delete(ctx, clusterRoleBinding); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete ClusterRoleBinding", zap.String("name", clusterRoleName), zap.Error(err))
	} else {
		logger.Info("Deleted ClusterRoleBinding", zap.String("name", clusterRoleName))
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(workspace, workspaceFinalizer)
	if err := r.Update(ctx, workspace); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Workspace cleanup completed")
	return reconcile.Result{}, nil
}
