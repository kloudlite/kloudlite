package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/pagination"
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
	// Check if another controller is already deleting this pod
	if !podDeletionTracker.TryStartDeletion(pod.UID, podName, pod.Namespace, "workspace") {
		logger.Info("Pod deletion already in progress by another controller, skipping",
			zap.String("pod", podName),
			zap.String("namespace", pod.Namespace))
		return false, nil
	}

	// Ensure we mark the deletion as complete when we exit
	defer podDeletionTracker.CompleteDeletion(pod.UID, podName, pod.Namespace, err)

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
		zap.String("namespace", pod.Namespace),
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

	// Calculate TTL - pod will be automatically deleted after configured duration
	ttlSecondsAfterFinished := cfg.Workspace.CleanupPodTTLSeconds

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
					Image:   cfg.Workspace.AlpineImage,
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
// This updates the connected Environment's Compose.Intercepts to remove any intercepts for this workspace.
func (r *WorkspaceReconciler) cleanupWorkspaceIntercepts(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	logger.Info("Cleaning up service intercepts for workspace")

	// Get the connected environment
	if workspace.Spec.EnvironmentConnection == nil {
		logger.Info("No environment connection, skipping intercept cleanup")
		return nil
	}

	env := &environmentv1.Environment{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: workspace.Spec.EnvironmentConnection.EnvironmentRef.Namespace,
		Name:      workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
	}, env); err != nil {
		logger.Error("Failed to get environment for intercept cleanup", zap.Error(err))
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// Check if environment has compose with intercepts
	if env.Spec.Compose == nil || len(env.Spec.Compose.Intercepts) == 0 {
		logger.Info("No intercepts in environment compose")
		return nil
	}

	// Check if any intercepts reference this workspace
	interceptsToKeep := make([]environmentv1.ServiceInterceptConfig, 0, len(env.Spec.Compose.Intercepts))
	interceptsCleaned := 0

	for _, intercept := range env.Spec.Compose.Intercepts {
		if intercept.WorkspaceRef != nil &&
			intercept.WorkspaceRef.Name == workspace.Name &&
			intercept.WorkspaceRef.Namespace == workspace.Namespace {
			// This intercept references the workspace being deleted, skip it
			interceptsCleaned++
			logger.Info("Removing intercept from Environment",
				zap.String("environment", env.Name),
				zap.String("serviceName", intercept.ServiceName))
		} else {
			// Keep this intercept
			interceptsToKeep = append(interceptsToKeep, intercept)
		}
	}

	if interceptsCleaned > 0 {
		// Update the Environment to remove the intercepts
		env.Spec.Compose.Intercepts = interceptsToKeep
		if err := r.Update(ctx, env); err != nil {
			logger.Error("Failed to update Environment to remove intercepts",
				zap.String("environment", env.Name),
				zap.Error(err))
			return fmt.Errorf("failed to update environment: %w", err)
		}
		logger.Info("Successfully removed intercepts from Environment",
			zap.String("environment", env.Name),
			zap.Int("interceptsCleaned", interceptsCleaned))
	}

	logger.Info("Completed intercept cleanup",
		zap.Int("interceptsCleaned", interceptsCleaned))

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
		if err := r.updateStatus(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update status to Terminating", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

	// Clean up service intercepts referencing this workspace
	// This is done early in the deletion flow so the Environment controller
	// can restore original deployments while the workspace pod is still available
	if err := r.cleanupWorkspaceIntercepts(ctx, workspace, logger); err != nil {
		logger.Error("Failed to cleanup workspace intercepts, will retry", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, fmt.Errorf("failed to cleanup intercepts: %w", err)
	}

	// Clean up snapshots for this workspace
	r.cleanupWorkspaceSnapshots(ctx, workspace, logger)

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
	// IMPORTANT: Must use the same naming convention as setupWorkspaceRBAC
	if err != nil {
		// Previous getWorkspaceTargetNamespace failed, try again
		targetNamespace, err = r.getWorkspaceTargetNamespace(ctx, workspace)
		if err != nil {
			logger.Warn("Failed to get target namespace for ClusterRole cleanup, skipping RBAC cleanup",
				zap.Error(err))
			// Continue with cleanup even if we can't get the namespace
		} else {
			clusterRoleName := fmt.Sprintf("workspace-%s-%s", targetNamespace, workspace.Name)

			// Delete ClusterRole - now returns error on failure
			if err := r.deleteClusterRole(ctx, clusterRoleName, logger); err != nil {
				return reconcile.Result{}, err
			}

			// Delete ClusterRoleBinding - now returns error on failure
			if err := r.deleteClusterRoleBinding(ctx, clusterRoleName, logger); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// targetNamespace was already set earlier in the function
		clusterRoleName := fmt.Sprintf("workspace-%s-%s", targetNamespace, workspace.Name)

		// Delete ClusterRole - now returns error on failure
		if err := r.deleteClusterRole(ctx, clusterRoleName, logger); err != nil {
			return reconcile.Result{}, err
		}

		// Delete ClusterRoleBinding - now returns error on failure
		if err := r.deleteClusterRoleBinding(ctx, clusterRoleName, logger); err != nil {
			return reconcile.Result{}, err
		}
	}

	// NOTE: PackageRequest deletion is handled automatically by Kubernetes garbage collection
	// via owner references since PackageRequest is now namespace-scoped like Workspace

	// Remove finalizer
	controllerutil.RemoveFinalizer(workspace, workspaceFinalizer)
	if err := r.Update(ctx, workspace); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Workspace cleanup completed")
	return reconcile.Result{}, nil
}

// deleteClusterRole deletes a ClusterRole for the workspace
// Returns an error if the deletion fails (excluding NotFound which is expected if already deleted)
func (r *WorkspaceReconciler) deleteClusterRole(ctx context.Context, name string, logger *zap.Logger) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := r.Delete(ctx, clusterRole); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ClusterRole already deleted", zap.String("name", name))
			return nil
		}
		logger.Error("Failed to delete ClusterRole", zap.String("name", name), zap.Error(err))
		return fmt.Errorf("failed to delete ClusterRole %s: %w", name, err)
	}
	logger.Info("Deleted ClusterRole", zap.String("name", name))
	return nil
}

// deleteClusterRoleBinding deletes a ClusterRoleBinding for the workspace
// Returns an error if the deletion fails (excluding NotFound which is expected if already deleted)
func (r *WorkspaceReconciler) deleteClusterRoleBinding(ctx context.Context, name string, logger *zap.Logger) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := r.Delete(ctx, clusterRoleBinding); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ClusterRoleBinding already deleted", zap.String("name", name))
			return nil
		}
		logger.Error("Failed to delete ClusterRoleBinding", zap.String("name", name), zap.Error(err))
		return fmt.Errorf("failed to delete ClusterRoleBinding %s: %w", name, err)
	}
	logger.Info("Deleted ClusterRoleBinding", zap.String("name", name))
	return nil
}

// cleanupWorkspaceSnapshots deletes all snapshots associated with this workspace
func (r *WorkspaceReconciler) cleanupWorkspaceSnapshots(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) {
	// List all snapshots for this workspace using pagination
	snapshotList := &snapshotv1.SnapshotList{}
	if err := pagination.ListAll(ctx, r, snapshotList, client.MatchingLabels{
		"snapshots.kloudlite.io/workspace": workspace.Name,
	}); err != nil {
		logger.Error("Failed to list snapshots for workspace", zap.Error(err))
		return
	}

	if len(snapshotList.Items) == 0 {
		return
	}

	logger.Info("Deleting snapshots for workspace",
		zap.String("workspace", workspace.Name),
		zap.Int("count", len(snapshotList.Items)))

	deletedCount := 0
	for i := range snapshotList.Items {
		snapshot := &snapshotList.Items[i]
		logger.Info("Deleting snapshot",
			zap.String("snapshot", snapshot.Name),
			zap.String("workspace", workspace.Name))

		if err := r.Delete(ctx, snapshot); err != nil {
			logger.Error("Failed to delete snapshot",
				zap.String("snapshot", snapshot.Name),
				zap.Error(err))
			// Continue with other snapshots even if one fails
			continue
		}
		deletedCount++
	}

	logger.Info("Deleted workspace snapshots",
		zap.String("workspace", workspace.Name),
		zap.Int("deleted", deletedCount),
		zap.Int("total", len(snapshotList.Items)))
}

// cleanupOrphanedRBACResources cleans up ClusterRole and ClusterRoleBinding resources
// that reference non-existent workspaces. This is called periodically to prevent
// resource accumulation in the cluster.
//
// Uses pagination with a default page size of 100 items to avoid API server overload
// when cleaning up large numbers of orphaned resources. This is critical for clusters
// with many workspaces that may have leaked RBAC resources.
//
// Returns the number of successfully deleted resources and a slice of errors encountered.
func (r *WorkspaceReconciler) cleanupOrphanedRBACResources(ctx context.Context, logger *zap.Logger) (int, []error) {
	logger.Info("Starting orphaned RBAC resource cleanup with pagination")

	// List all ClusterRoles with workspace labels using pagination
	// Page size of 100 prevents API server overload in large clusters
	var clusterRoles rbacv1.ClusterRoleList
	if err := pagination.ListAll(ctx, r, &clusterRoles, client.MatchingLabels{
		"kloudlite.io/workspace-rbac": "true",
	}); err != nil {
		logger.Error("Failed to list ClusterRoles for cleanup", zap.Error(err))
		return 0, []error{fmt.Errorf("failed to list ClusterRoles: %w", err)}
	}

	logger.Info("Listed ClusterRoles for cleanup", zap.Int("count", len(clusterRoles.Items)))

	orphanedClusterRoles := 0
	var errors []error
	for i := range clusterRoles.Items {
		cr := &clusterRoles.Items[i]
		workspaceName := cr.Labels["kloudlite.io/workspace-name"]
		namespace := cr.Labels["kloudlite.io/workspace-namespace"]

		if workspaceName == "" || namespace == "" {
			logger.Warn("ClusterRole missing workspace labels, skipping",
				zap.String("clusterRole", cr.Name))
			continue
		}

		// Check if the workspace still exists
		workspace := &workspacev1.Workspace{}
		err := r.Get(ctx, client.ObjectKey{Name: workspaceName, Namespace: namespace}, workspace)

		if err != nil {
			if apierrors.IsNotFound(err) {
				// Workspace doesn't exist, delete the orphaned ClusterRole
				logger.Info("Deleting orphaned ClusterRole",
					zap.String("clusterRole", cr.Name),
					zap.String("workspaceName", workspaceName),
					zap.String("namespace", namespace))

				if err := r.Delete(ctx, cr); err != nil && !apierrors.IsNotFound(err) {
					deleteErr := fmt.Errorf("failed to delete orphaned ClusterRole %s: %w", cr.Name, err)
					logger.Error(deleteErr.Error())
					errors = append(errors, deleteErr)
				} else {
					orphanedClusterRoles++
					logger.Info("Successfully deleted orphaned ClusterRole",
						zap.String("clusterRole", cr.Name))
				}
			} else {
				checkErr := fmt.Errorf("failed to check workspace existence for ClusterRole %s: %w", cr.Name, err)
				logger.Warn(checkErr.Error())
				errors = append(errors, checkErr)
			}
		}
	}

	// List all ClusterRoleBindings with workspace labels using pagination
	// Page size of 100 prevents API server overload in large clusters
	var clusterRoleBindings rbacv1.ClusterRoleBindingList
	if err := pagination.ListAll(ctx, r, &clusterRoleBindings, client.MatchingLabels{
		"kloudlite.io/workspace-rbac": "true",
	}); err != nil {
		logger.Error("Failed to list ClusterRoleBindings for cleanup", zap.Error(err))
		return orphanedClusterRoles, append(errors, fmt.Errorf("failed to list ClusterRoleBindings: %w", err))
	}

	logger.Info("Listed ClusterRoleBindings for cleanup", zap.Int("count", len(clusterRoleBindings.Items)))

	orphanedClusterRoleBindings := 0
	for i := range clusterRoleBindings.Items {
		crb := &clusterRoleBindings.Items[i]
		workspaceName := crb.Labels["kloudlite.io/workspace-name"]
		namespace := crb.Labels["kloudlite.io/workspace-namespace"]

		if workspaceName == "" || namespace == "" {
			logger.Warn("ClusterRoleBinding missing workspace labels, skipping",
				zap.String("clusterRoleBinding", crb.Name))
			continue
		}

		// Check if the workspace still exists
		workspace := &workspacev1.Workspace{}
		err := r.Get(ctx, client.ObjectKey{Name: workspaceName, Namespace: namespace}, workspace)

		if err != nil {
			if apierrors.IsNotFound(err) {
				// Workspace doesn't exist, delete the orphaned ClusterRoleBinding
				logger.Info("Deleting orphaned ClusterRoleBinding",
					zap.String("clusterRoleBinding", crb.Name),
					zap.String("workspaceName", workspaceName),
					zap.String("namespace", namespace))

				if err := r.Delete(ctx, crb); err != nil && !apierrors.IsNotFound(err) {
					deleteErr := fmt.Errorf("failed to delete orphaned ClusterRoleBinding %s: %w", crb.Name, err)
					logger.Error(deleteErr.Error())
					errors = append(errors, deleteErr)
				} else {
					orphanedClusterRoleBindings++
					logger.Info("Successfully deleted orphaned ClusterRoleBinding",
						zap.String("clusterRoleBinding", crb.Name))
				}
			} else {
				checkErr := fmt.Errorf("failed to check workspace existence for ClusterRoleBinding %s: %w", crb.Name, err)
				logger.Warn(checkErr.Error())
				errors = append(errors, checkErr)
			}
		}
	}

	totalOrphaned := orphanedClusterRoles + orphanedClusterRoleBindings
	logger.Info("Orphaned RBAC resource cleanup completed",
		zap.Int("orphanedClusterRoles", orphanedClusterRoles),
		zap.Int("orphanedClusterRoleBindings", orphanedClusterRoleBindings),
		zap.Int("totalOrphaned", totalOrphaned),
		zap.Int("errors", len(errors)))

	if len(errors) > 0 {
		logger.Warn("Orphaned RBAC cleanup encountered errors",
			zap.Int("errorCount", len(errors)))
	}

	return totalOrphaned, errors
}

// joinErrors combines multiple errors into a single error message
func joinErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var sb strings.Builder
	for i, err := range errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return fmt.Errorf("%s", sb.String())
}
