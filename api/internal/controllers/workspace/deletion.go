package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
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

// cleanupWorkspaceIntercepts removes all service intercepts referencing the workspace being deleted.
// This iterates through namespaces where the workspace has environment connections and removes
// any intercepts from Compositions that reference this workspace.
// Errors are logged but do not block workspace deletion.
func (r *WorkspaceReconciler) cleanupWorkspaceIntercepts(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) {
	logger.Info("Cleaning up service intercepts for workspace")

	// Collect namespaces to check based on environment connections
	namespacesToCheck := make(map[string]bool)

	// Check current connected environment
	if workspace.Status.ConnectedEnvironment != nil && workspace.Status.ConnectedEnvironment.TargetNamespace != "" {
		namespacesToCheck[workspace.Status.ConnectedEnvironment.TargetNamespace] = true
	}

	// Also check the environment connection spec in case status hasn't been updated
	if workspace.Spec.EnvironmentConnection != nil {
		env := &environmentv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{
			Name: workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
		}, env); err == nil {
			if env.Spec.TargetNamespace != "" {
				namespacesToCheck[env.Spec.TargetNamespace] = true
			}
		}
	}

	if len(namespacesToCheck) == 0 {
		logger.Info("No environment namespaces found to check for intercepts")
		return
	}

	interceptsCleaned := 0
	for namespace := range namespacesToCheck {
		compList := &environmentv1.CompositionList{}
		if err := r.List(ctx, compList, client.InNamespace(namespace)); err != nil {
			logger.Warn("Failed to list Compositions in namespace",
				zap.String("namespace", namespace),
				zap.Error(err))
			continue
		}

		for i := range compList.Items {
			comp := &compList.Items[i]
			if len(comp.Spec.Intercepts) == 0 {
				continue
			}

			// Check if any intercepts reference this workspace
			interceptsToKeep := make([]environmentv1.ServiceInterceptConfig, 0, len(comp.Spec.Intercepts))
			foundMatch := false

			for _, intercept := range comp.Spec.Intercepts {
				if intercept.WorkspaceRef != nil &&
					intercept.WorkspaceRef.Name == workspace.Name &&
					intercept.WorkspaceRef.Namespace == workspace.Namespace {
					// This intercept references the workspace being deleted, skip it
					foundMatch = true
					logger.Info("Removing intercept from Composition",
						zap.String("composition", comp.Name),
						zap.String("namespace", comp.Namespace),
						zap.String("serviceName", intercept.ServiceName))
				} else {
					// Keep this intercept
					interceptsToKeep = append(interceptsToKeep, intercept)
				}
			}

			if foundMatch {
				// Update the Composition to remove the intercepts
				comp.Spec.Intercepts = interceptsToKeep
				if err := r.Update(ctx, comp); err != nil {
					logger.Warn("Failed to update Composition to remove intercepts",
						zap.String("composition", comp.Name),
						zap.String("namespace", comp.Namespace),
						zap.Error(err))
				} else {
					interceptsCleaned++
					logger.Info("Successfully removed intercepts from Composition",
						zap.String("composition", comp.Name),
						zap.String("namespace", comp.Namespace))
				}
			}
		}
	}

	logger.Info("Completed intercept cleanup",
		zap.Int("interceptsCleaned", interceptsCleaned),
		zap.Int("namespacesChecked", len(namespacesToCheck)))
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
		if err := r.updateStatus(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update status to Terminating", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

	// Clean up service intercepts referencing this workspace
	// This is done early in the deletion flow so the Composition controller
	// can restore original deployments while the workspace pod is still available
	r.cleanupWorkspaceIntercepts(ctx, workspace, logger)

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
		podName := getWorkspacePodName(workspace)
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
