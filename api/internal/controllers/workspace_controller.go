package controllers

import (
	"context"
	"fmt"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	workspaceFinalizer = "workspaces.kloudlite.io/finalizer"
)

// Port mappings for different server types
var serverPorts = map[string]int32{
	"code-server": 8080,
	"jupyter":     8888,
	"ttyd":        7681,
	"code-web":    3000,
}

// WorkspaceReconciler reconciles Workspace objects and manages VS Code server pods
type WorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// updateStatusPreservingPackages updates workspace status while preserving package-related fields
func (r *WorkspaceReconciler) updateStatusPreservingPackages(ctx context.Context, workspace *workspacesv1.Workspace) error {
	// Refetch to get the latest version with all status fields
	latest := &workspacesv1.Workspace{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(workspace), latest); err != nil {
		return err
	}

	// Preserve package-related fields from the latest version
	workspace.Status.InstalledPackages = latest.Status.InstalledPackages
	workspace.Status.FailedPackages = latest.Status.FailedPackages
	workspace.Status.PackageInstallationMessage = latest.Status.PackageInstallationMessage

	// Also update the resource version to ensure optimistic locking works
	workspace.ResourceVersion = latest.ResourceVersion

	return r.Status().Update(ctx, workspace)
}

// Reconcile handles Workspace events and ensures the workspace pod exists
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("workspace", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling Workspace")

	// Fetch the Workspace instance
	workspace := &workspacesv1.Workspace{}
	err := r.Get(ctx, req.NamespacedName, workspace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Workspace not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Workspace", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if workspace is being deleted
	if workspace.DeletionTimestamp != nil {
		logger.Info("Workspace is being deleted, starting cleanup")
		return r.handleDeletion(ctx, workspace, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
		logger.Info("Adding finalizer to workspace")
		controllerutil.AddFinalizer(workspace, workspaceFinalizer)
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Set default workspace path if not provided
	if workspace.Spec.WorkspacePath == "" {
		workspace.Spec.WorkspacePath = "/workspace"
	}

	// Set default VS Code version if not provided
	if workspace.Spec.VSCodeVersion == "" {
		workspace.Spec.VSCodeVersion = "latest"
	}

	// Set default server type if not provided
	if workspace.Spec.ServerType == "" {
		workspace.Spec.ServerType = "code-server"
	}

	// Handle workspace based on its status
	switch workspace.Spec.Status {
	case "active":
		return r.handleActiveWorkspace(ctx, workspace, logger)
	case "suspended", "archived":
		return r.handleSuspendedWorkspace(ctx, workspace, logger)
	default:
		// Default to active if status is not set
		workspace.Spec.Status = "active"
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to update workspace status", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}
}

// ensurePackageRequest creates or updates PackageRequest for the workspace
func (r *WorkspaceReconciler) ensurePackageRequest(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) error {
	// Skip if no packages defined
	if len(workspace.Spec.Packages) == 0 {
		logger.Info("No packages defined for workspace, skipping PackageRequest creation")
		return nil
	}

	pkgReqName := fmt.Sprintf("%s-packages", workspace.Name)

	// Check if PackageRequest exists
	pkgReq := &packagesv1.PackageRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: pkgReqName, Namespace: workspace.Namespace}, pkgReq)

	if apierrors.IsNotFound(err) {
		// Create new PackageRequest
		pkgReq = &packagesv1.PackageRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pkgReqName,
				Namespace: workspace.Namespace,
			},
			Spec: packagesv1.PackageRequestSpec{
				WorkspaceRef: workspace.Name,
				Packages:     workspace.Spec.Packages,
				ProfileName:  fmt.Sprintf("workspace-%s-packages", workspace.Name),
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, pkgReq, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		if err := r.Create(ctx, pkgReq); err != nil {
			return fmt.Errorf("failed to create PackageRequest: %w", err)
		}
		logger.Info("Created PackageRequest", zap.String("name", pkgReqName))
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get PackageRequest: %w", err)
	}

	// PackageRequest exists, check if packages changed
	packagesChanged := false
	if len(pkgReq.Spec.Packages) != len(workspace.Spec.Packages) {
		packagesChanged = true
	} else {
		for i, pkg := range workspace.Spec.Packages {
			if pkgReq.Spec.Packages[i].Name != pkg.Name || pkgReq.Spec.Packages[i].Version != pkg.Version {
				packagesChanged = true
				break
			}
		}
	}

	if packagesChanged {
		pkgReq.Spec.Packages = workspace.Spec.Packages
		if err := r.Update(ctx, pkgReq); err != nil {
			return fmt.Errorf("failed to update PackageRequest: %w", err)
		}

		// Reset status to Pending to trigger new reconciliation
		pkgReq.Status.Phase = "Pending"
		pkgReq.Status.Message = "Package list updated, waiting for installation"
		pkgReq.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, pkgReq); err != nil {
			return fmt.Errorf("failed to reset PackageRequest status: %w", err)
		}

		logger.Info("Updated PackageRequest with new packages", zap.String("name", pkgReqName))
	}

	return nil
}

// ensurePVC creates or ensures the PersistentVolumeClaim exists for the workspace
func (r *WorkspaceReconciler) ensurePVC(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) error {
	pvcName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Check if PVC exists
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, client.ObjectKey{Name: pvcName, Namespace: workspace.Namespace}, pvc)

	if apierrors.IsNotFound(err) {
		// Set default storage size if not specified
		storageSize := workspace.Spec.StorageSize
		if storageSize == "" {
			storageSize = "10Gi"
		}

		// Create new PVC
		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: workspace.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(storageSize),
					},
				},
			},
		}

		// Set storage class if specified
		if workspace.Spec.StorageClassName != nil {
			pvc.Spec.StorageClassName = workspace.Spec.StorageClassName
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, pvc, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		if err := r.Create(ctx, pvc); err != nil {
			return fmt.Errorf("failed to create PVC: %w", err)
		}
		logger.Info("Created PVC", zap.String("name", pvcName), zap.String("size", storageSize))
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get PVC: %w", err)
	}

	// PVC exists
	logger.Info("PVC already exists", zap.String("name", pvcName))
	return nil
}

// handleActiveWorkspace ensures the workspace pod is running
func (r *WorkspaceReconciler) handleActiveWorkspace(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
	// Ensure PackageRequest is created if packages are defined
	if err := r.ensurePackageRequest(ctx, workspace, logger); err != nil {
		logger.Error("Failed to ensure PackageRequest", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create PackageRequest: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace)
		return reconcile.Result{}, err
	}

	// Ensure PVC exists
	if err := r.ensurePVC(ctx, workspace, logger); err != nil {
		logger.Error("Failed to ensure PVC", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create PVC: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace)
		return reconcile.Result{}, err
	}

	// Check if pod already exists
	podName := fmt.Sprintf("workspace-%s", workspace.Name)
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)

	if err == nil {
		// Pod exists, update workspace status
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
		r.updateStatusPreservingPackages(ctx, workspace)
		return reconcile.Result{}, err
	}

	if err := r.Create(ctx, pod); err != nil {
		logger.Error("Failed to create workspace pod", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create pod: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace)
		return reconcile.Result{}, err
	}

	// Update workspace status
	workspace.Status.Phase = "Creating"
	workspace.Status.Message = "Workspace pod is being created"
	workspace.Status.PodName = podName
	now := metav1.Now()
	workspace.Status.StartTime = &now
	workspace.Status.LastActivityTime = &now

	if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
	}

	logger.Info("Workspace pod created successfully")
	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// handleSuspendedWorkspace ensures the workspace pod is stopped
func (r *WorkspaceReconciler) handleSuspendedWorkspace(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
	// Check if pod exists
	podName := fmt.Sprintf("workspace-%s", workspace.Name)
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)

	if apierrors.IsNotFound(err) {
		// Pod doesn't exist, workspace is already stopped
		workspace.Status.Phase = "Stopped"
		workspace.Status.Message = "Workspace is stopped"
		workspace.Status.PodName = ""
		workspace.Status.PodIP = ""
		workspace.Status.NodeName = ""
		now := metav1.Now()
		workspace.Status.StopTime = &now

		if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
			logger.Warn("Failed to update workspace status", zap.Error(err))
		}
		return reconcile.Result{}, nil
	}

	if err != nil {
		logger.Error("Failed to check existing pod", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Pod exists, delete it
	logger.Info("Deleting workspace pod for suspended workspace", zap.String("pod", podName))
	if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
		logger.Error("Failed to delete workspace pod", zap.Error(err))
		return reconcile.Result{}, err
	}

	workspace.Status.Phase = "Stopping"
	workspace.Status.Message = "Workspace is being stopped"
	if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
	}

	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

// getServerImage returns the Docker image for the given server type
func getServerImage(serverType string) string {
	imageMap := map[string]string{
		"code-server": "kloudlite/workspace-code-server:latest",
		"jupyter":     "kloudlite/workspace-jupyter:latest",
		"ttyd":        "kloudlite/workspace-ttyd:latest",
		"code-web":    "kloudlite/workspace-code-web:latest",
	}

	if image, ok := imageMap[serverType]; ok {
		return image
	}
	// Default to code-server if unknown type
	return imageMap["code-server"]
}

// getServerPort returns the port for the given server type
func getServerPort(serverType string) int32 {
	if port, ok := serverPorts[serverType]; ok {
		return port
	}
	// Default to code-server port if unknown type
	return serverPorts["code-server"]
}

// getServerArgs returns the command arguments for the given server type
func getServerArgs(serverType string) []string {
	switch serverType {
	case "code-server":
		return []string{"code-server", "--auth", "none", "--bind-addr", "0.0.0.0:8080"}
	case "jupyter":
		return []string{"jupyter", "lab", "--ip=0.0.0.0", "--port=8888", "--no-browser", "--allow-root"}
	case "ttyd":
		return []string{"ttyd", "-p", "7681", "bash"}
	case "code-web":
		return []string{"code-web"}
	default:
		return []string{"code-server", "--auth", "none", "--bind-addr", "0.0.0.0:8080"}
	}
}

// createWorkspacePod creates a pod with the selected server and host volume mount
func (r *WorkspaceReconciler) createWorkspacePod(workspace *workspacesv1.Workspace) (*corev1.Pod, error) {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get server image and port based on server type
	serverImage := getServerImage(workspace.Spec.ServerType)
	serverPort := getServerPort(workspace.Spec.ServerType)

	// Build resource requirements
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
	}

	// Override with custom resource quota if provided
	if workspace.Spec.ResourceQuota != nil {
		if workspace.Spec.ResourceQuota.CPU != "" {
			resources.Limits[corev1.ResourceCPU] = resource.MustParse(workspace.Spec.ResourceQuota.CPU)
		}
		if workspace.Spec.ResourceQuota.Memory != "" {
			resources.Limits[corev1.ResourceMemory] = resource.MustParse(workspace.Spec.ResourceQuota.Memory)
		}
	}

	// Build environment variables
	envVars := []corev1.EnvVar{
		{
			Name:  "WORKSPACE_NAME",
			Value: workspace.Name,
		},
		{
			Name:  "WORKSPACE_OWNER",
			Value: workspace.Spec.Owner,
		},
	}

	// Add PATH to include nix binaries from workspace-specific profile
	profilePath := fmt.Sprintf("/nix/var/nix/profiles/per-user/root/workspace-%s-packages", workspace.Name)
	envVars = append(envVars, corev1.EnvVar{
		Name:  "PATH",
		Value: fmt.Sprintf("%s/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", profilePath),
	})

	// Add startup script from settings if provided
	if workspace.Spec.Settings != nil && workspace.Spec.Settings.StartupScript != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "STARTUP_SCRIPT",
			Value: workspace.Spec.Settings.StartupScript,
		})
	}

	// Add custom environment variables from settings
	if workspace.Spec.Settings != nil && workspace.Spec.Settings.EnvironmentVariables != nil {
		for key, value := range workspace.Spec.Settings.EnvironmentVariables {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: workspace.Namespace,
			Labels: map[string]string{
				"app":       "workspace",
				"workspace": workspace.Name,
			},
			Annotations: map[string]string{
				"kloudlite.io/workspace-display-name": workspace.Spec.DisplayName,
				"kloudlite.io/workspace-owner":        workspace.Spec.Owner,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            workspace.Spec.ServerType,
					Image:           serverImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources:       resources,
					Env:             envVars,
					Args:            getServerArgs(workspace.Spec.ServerType),
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: serverPort,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "workspace-data",
							MountPath: workspace.Spec.WorkspacePath,
						},
						{
							Name:      "nix-store",
							MountPath: "/nix",
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt(int(serverPort)),
							},
						},
						InitialDelaySeconds: 30,
						PeriodSeconds:       10,
						TimeoutSeconds:      5,
						FailureThreshold:    3,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt(int(serverPort)),
							},
						},
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
						FailureThreshold:    3,
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "workspace-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: fmt.Sprintf("workspace-%s", workspace.Name),
						},
					},
				},
				{
					Name: "nix-store",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/nix-store",
							Type: func() *corev1.HostPathType {
								t := corev1.HostPathDirectoryOrCreate
								return &t
							}(),
						},
					},
				},

			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}
	// Set owner reference
	if err := controllerutil.SetControllerReference(workspace, pod, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	return pod, nil
}

// updateWorkspaceStatus updates the workspace status based on pod state
func (r *WorkspaceReconciler) updateWorkspaceStatus(ctx context.Context, workspace *workspacesv1.Workspace, pod *corev1.Pod, phase, message string, logger *zap.Logger) (reconcile.Result, error) {
	workspace.Status.Phase = phase
	workspace.Status.Message = message
	workspace.Status.PodName = pod.Name
	workspace.Status.PodIP = pod.Status.PodIP
	workspace.Status.NodeName = pod.Spec.NodeName

	// Build access URL if pod is running
	if pod.Status.PodIP != "" && phase == "Running" {
		serverPort := getServerPort(workspace.Spec.ServerType)
		workspace.Status.AccessURL = fmt.Sprintf("http://%s:%d", pod.Status.PodIP, serverPort)
	}

	// Update last activity time
	now := metav1.Now()
	workspace.Status.LastActivityTime = &now

	if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// handleDeletion cleans up workspace resources when being deleted
func (r *WorkspaceReconciler) handleDeletion(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
		return reconcile.Result{}, nil
	}

	// Delete the workspace pod if it exists
	podName := fmt.Sprintf("workspace-%s", workspace.Name)
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)

	if err == nil {
		logger.Info("Deleting workspace pod", zap.String("pod", podName))
		if err := r.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to delete workspace pod", zap.Error(err))
			return reconcile.Result{}, err
		}
	} else if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check workspace pod", zap.Error(err))
		return reconcile.Result{}, err
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

// SetupWithManager sets up the controller with the Manager
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacesv1.Workspace{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&packagesv1.PackageRequest{}).
		Complete(r)
}
