package controllers

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	workspaceFinalizer = "workspaces.kloudlite.io/finalizer"

	// Default idle timeout if not specified in workspace settings (30 minutes)
	defaultIdleTimeoutMinutes = 30
)

// WorkspaceReconciler reconciles Workspace objects and manages VS Code server pods
type WorkspaceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Logger    *zap.Logger
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

// hasActiveConnections checks if there are active SSH or web connections to the workspace
// by examining active TCP connections in the pod
// Returns: hasConnections bool, connectionCount int, error
func (r *WorkspaceReconciler) hasActiveConnections(ctx context.Context, workspace *workspacesv1.Workspace) (bool, int, error) {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get pod: %w", err)
	}

	// Check pod IP and if it's accessible
	if pod.Status.PodIP == "" {
		return false, 0, nil
	}

	// Check if pod is not ready yet (still initializing)
	if pod.Status.Phase != corev1.PodRunning {
		return true, 0, nil // Consider as active while starting
	}

	// If pod was just started (within last 2 minutes), consider it as having connections
	// This gives time for the user to connect after starting the workspace
	if pod.Status.StartTime != nil {
		timeSinceStart := time.Since(pod.Status.StartTime.Time)
		if timeSinceStart < 2*time.Minute {
			return true, 0, nil
		}
	}

	// Check for actual active network connections
	// We check /proc/net/tcp for ESTABLISHED connections (state 01)
	// Important ports: SSH (22=0016), ttyd (7681=1E01), code-server (8080=1F90), vscode-tunnel (8000=1F40)
	// Connection state 01 = ESTABLISHED, 0A = LISTEN

	// Get the main container name (usually the first container)
	if len(pod.Spec.Containers) == 0 {
		return false, 0, nil
	}

	// Count ESTABLISHED connections by checking /proc/net/tcp
	// Format: awk '$4 == "01"' /proc/net/tcp /proc/net/tcp6 2>/dev/null | wc -l
	// This counts all ESTABLISHED TCP connections (excluding LISTEN sockets)
	command := []string{"sh", "-c", "awk '$4 == \"01\"' /proc/net/tcp /proc/net/tcp6 2>/dev/null | wc -l"}

	output, err := r.execInPod(ctx, pod, pod.Spec.Containers[0].Name, command)
	if err != nil {
		// If we can't check connections, assume there might be connections (fail-safe)
		r.Logger.Warn("Failed to check active connections, assuming active",
			zap.String("workspace", workspace.Name),
			zap.Error(err),
		)
		return true, 0, nil
	}

	// Parse the connection count
	connectionCount := 0
	fmt.Sscanf(strings.TrimSpace(output), "%d", &connectionCount)

	// Log the connection count for debugging
	r.Logger.Info("Active connection check",
		zap.String("workspace", workspace.Name),
		zap.Int("connectionCount", connectionCount),
		zap.Bool("hasConnections", connectionCount > 0),
	)

	return connectionCount > 0, connectionCount, nil
}

// execInPod executes a command in a pod container and returns the stdout output
func (r *WorkspaceReconciler) execInPod(ctx context.Context, pod *corev1.Pod, containerName string, command []string) (string, error) {
	req := r.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
		Stdin:     false,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(r.Config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", fmt.Errorf("failed to exec command: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// reconcileEnvironmentConnection dynamically updates DNS configuration in running pods
// when the workspace's environment connection changes
func (r *WorkspaceReconciler) reconcileEnvironmentConnection(ctx context.Context, workspace *workspacesv1.Workspace, pod *corev1.Pod, logger *zap.Logger) error {
	// Only reconcile if pod is running
	if pod.Status.Phase != corev1.PodRunning {
		return nil
	}

	// Determine desired state from spec
	var desiredEnvName, desiredTargetNs string
	var shouldBeConnected bool

	if workspace.Spec.EnvironmentRef != nil {
		// Fetch environment to get target namespace
		env := &environmentsv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      workspace.Spec.EnvironmentRef.Name,
			Namespace: workspace.Namespace,
		}, env)
		if err == nil && env.Spec.Activated {
			desiredEnvName = env.Name
			desiredTargetNs = env.Spec.TargetNamespace
			shouldBeConnected = true
		} else if err != nil {
			logger.Warn("Failed to fetch environment",
				zap.String("environment", workspace.Spec.EnvironmentRef.Name),
				zap.Error(err),
			)
		}
	}

	// Determine current state from status
	currentConnected := workspace.Status.ConnectedEnvironment != nil &&
		workspace.Status.ConnectedEnvironment.Connected

	// Always update DNS configuration to ensure it's correct
	// This is idempotent and ensures DNS is correct even if status was updated but file wasn't
	if shouldBeConnected {
		logger.Info("Updating DNS configuration for environment connection",
			zap.String("environment", desiredEnvName),
			zap.String("targetNamespace", desiredTargetNs),
		)
	} else if workspace.Spec.EnvironmentRef == nil && currentConnected {
		logger.Info("Resetting DNS configuration: no environment connection")
	}

	// Build search domains
	searchDomains := "svc.cluster.local cluster.local"
	if shouldBeConnected && desiredTargetNs != "" {
		searchDomains = fmt.Sprintf("%s.svc.cluster.local svc.cluster.local cluster.local", desiredTargetNs)
	}

	// Update /etc/resolv.conf using execInPod
	command := []string{"sh", "-c", fmt.Sprintf("cat > /etc/resolv.conf << 'EOF'\nnameserver 10.43.0.10\nsearch %s\noptions ndots:5\nEOF", searchDomains)}

	_, err := r.execInPod(ctx, pod, "workspace", command)
	if err != nil {
		logger.Warn("Failed to update DNS configuration", zap.Error(err))
		return err
	}

	logger.Info("Successfully updated DNS configuration",
		zap.String("searchDomains", searchDomains),
		zap.Bool("connected", shouldBeConnected),
	)

	return nil
}

// isWorkspaceIdle checks if a workspace has been idle by checking for active connections
// A workspace is considered idle ONLY if there are no active connections (SSH, ttyd, vscode, code-server)
// Returns: isIdle bool, connectionCount int, error
func (r *WorkspaceReconciler) isWorkspaceIdle(ctx context.Context, workspace *workspacesv1.Workspace) (bool, int, error) {
	// Check for active connections - this is the ONLY factor that matters
	hasConnections, connectionCount, err := r.hasActiveConnections(ctx, workspace)
	if err != nil {
		// If we can't check connections, assume workspace is active (fail-safe)
		r.Logger.Warn("Failed to check active connections, assuming workspace is active",
			zap.String("workspace", workspace.Name),
			zap.Error(err),
		)
		return false, 0, nil
	}

	// Workspace is idle if there are NO active connections
	isIdle := !hasConnections

	// Log activity status for debugging
	r.Logger.Info("Workspace activity check",
		zap.String("workspace", workspace.Name),
		zap.Bool("hasConnections", hasConnections),
		zap.Int("connectionCount", connectionCount),
		zap.Bool("isIdle", isIdle),
	)

	return isIdle, connectionCount, nil
}

// checkAndSuspendIdleWorkspace checks if a workspace should be auto-suspended and suspends it if needed
func (r *WorkspaceReconciler) checkAndSuspendIdleWorkspace(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) error {
	// Skip if auto-stop is not enabled
	if workspace.Spec.Settings == nil || !workspace.Spec.Settings.AutoStop {
		return nil
	}

	// Skip if workspace is not active
	if workspace.Spec.Status != "active" {
		return nil
	}

	// Get idle timeout from workspace settings or use default
	idleTimeout := defaultIdleTimeoutMinutes
	if workspace.Spec.Settings.IdleTimeout > 0 {
		idleTimeout = int(workspace.Spec.Settings.IdleTimeout)
	}

	// Check if workspace is idle
	isIdle, connectionCount, err := r.isWorkspaceIdle(ctx, workspace)
	if err != nil {
		logger.Warn("Failed to check workspace idle state", zap.Error(err))
		return nil // Don't fail reconciliation on metrics errors
	}

	// Update active connections count in workspace status
	workspace.Status.ActiveConnections = connectionCount

	if !isIdle {
		// Workspace is active, update last activity time
		now := metav1.Now()
		if workspace.Status.LastActivityTime == nil ||
		   time.Since(workspace.Status.LastActivityTime.Time) > 30*time.Second {
			workspace.Status.LastActivityTime = &now
			if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
				logger.Warn("Failed to update last activity time", zap.Error(err))
			}
		}
		return nil
	}

	// Workspace is idle, check if idle timeout has been reached
	if workspace.Status.LastActivityTime == nil {
		// No last activity time set, initialize it
		now := metav1.Now()
		workspace.Status.LastActivityTime = &now
		if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
			logger.Warn("Failed to initialize last activity time", zap.Error(err))
		}
		return nil
	}

	// Calculate idle duration
	idleDuration := time.Since(workspace.Status.LastActivityTime.Time)
	idleTimeoutDuration := time.Duration(idleTimeout) * time.Minute

	// Log idle duration for debugging
	logger.Info("Checking idle timeout",
		zap.String("workspace", workspace.Name),
		zap.Duration("idleDuration", idleDuration),
		zap.Duration("idleTimeout", idleTimeoutDuration),
		zap.Bool("willSuspend", idleDuration >= idleTimeoutDuration),
	)

	if idleDuration >= idleTimeoutDuration {
		// Idle timeout reached, suspend workspace
		logger.Info("Auto-suspending idle workspace",
			zap.String("workspace", workspace.Name),
			zap.Duration("idleDuration", idleDuration),
			zap.Duration("idleTimeout", idleTimeoutDuration),
		)

		// Fetch the latest version to avoid conflict errors
		latest := &workspacesv1.Workspace{}
		if err := r.Get(ctx, client.ObjectKey{Name: workspace.Name, Namespace: workspace.Namespace}, latest); err != nil {
			return fmt.Errorf("failed to fetch latest workspace: %w", err)
		}

		latest.Spec.Status = "suspended"
		if err := r.Update(ctx, latest); err != nil {
			return fmt.Errorf("failed to suspend idle workspace: %w", err)
		}
	}

	return nil
}

// updateStatusPreservingPackages updates workspace status while preserving package-related fields
func (r *WorkspaceReconciler) updateStatusPreservingPackages(ctx context.Context, workspace *workspacesv1.Workspace) error {
	// Refetch to get the latest version for optimistic locking
	latest := &workspacesv1.Workspace{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(workspace), latest); err != nil {
		return err
	}

	// Preserve package-related fields from the current workspace object
	// (these may have been updated by syncPackageStatus)
	installedPackages := workspace.Status.InstalledPackages
	failedPackages := workspace.Status.FailedPackages
	packageMessage := workspace.Status.PackageInstallationMessage

	// Copy all status fields to the latest version
	latest.Status = workspace.Status

	// Ensure package fields are preserved
	latest.Status.InstalledPackages = installedPackages
	latest.Status.FailedPackages = failedPackages
	latest.Status.PackageInstallationMessage = packageMessage

	return r.Status().Update(ctx, latest)
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

	// Handle workspace based on its status
	var result reconcile.Result

	switch workspace.Spec.Status {
	case "active":
		result, err = r.handleActiveWorkspace(ctx, workspace, logger)
	case "suspended", "archived":
		result, err = r.handleSuspendedWorkspace(ctx, workspace, logger)
	default:
		// Default to active if status is not set
		workspace.Spec.Status = "active"
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to update workspace status", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Requeue after 1 minute to check idle status periodically
	if workspace.Spec.Status == "active" && workspace.Spec.Settings != nil && workspace.Spec.Settings.AutoStop {
		if result.RequeueAfter == 0 && !result.Requeue {
			result.RequeueAfter = 1 * time.Minute
		}
	}

	return result, err
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
			if pkgReq.Spec.Packages[i].Name != pkg.Name ||
				pkgReq.Spec.Packages[i].Channel != pkg.Channel ||
				pkgReq.Spec.Packages[i].NixpkgsCommit != pkg.NixpkgsCommit {
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

		// PackageManagerReconciler will be triggered by the spec change
		// and will handle the status update (don't update status here to avoid race conditions)
		logger.Info("Updated PackageRequest with new packages", zap.String("name", pkgReqName))
	}

	return nil
}


// syncPackageStatus syncs package installation status from PackageRequest to Workspace
func (r *WorkspaceReconciler) syncPackageStatus(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) error {
	// Only sync if packages are defined
	if len(workspace.Spec.Packages) == 0 {
		return nil
	}

	// Get the PackageRequest
	pkgReqName := fmt.Sprintf("%s-packages", workspace.Name)
	pkgReq := &packagesv1.PackageRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: pkgReqName, Namespace: workspace.Namespace}, pkgReq)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// PackageRequest not yet created
			return nil
		}
		return fmt.Errorf("failed to get PackageRequest: %w", err)
	}

	// Copy package status from PackageRequest to Workspace
	workspace.Status.InstalledPackages = pkgReq.Status.InstalledPackages
	workspace.Status.FailedPackages = pkgReq.Status.FailedPackages
	workspace.Status.PackageInstallationMessage = pkgReq.Status.Message

	return nil
}

// ensureWorkspaceService ensures a Service is created for the workspace
func (r *WorkspaceReconciler) ensureWorkspaceService(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) error {
	serviceName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Check if Service exists
	svc := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: workspace.Namespace}, svc)

	if apierrors.IsNotFound(err) {
		// Create new Service
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: workspace.Namespace,
				Labels: map[string]string{
					"app":       "workspace",
					"workspace": workspace.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app":       "workspace",
					"workspace": workspace.Name,
				},
				Ports: []corev1.ServicePort{
					{
						Name:       "ssh",
						Protocol:   corev1.ProtocolTCP,
						Port:       22,
						TargetPort: intstr.FromInt(22),
					},
					{
						Name:       "code-server",
						Protocol:   corev1.ProtocolTCP,
						Port:       8080,
						TargetPort: intstr.FromInt(8080),
					},
					{
						Name:       "ttyd",
						Protocol:   corev1.ProtocolTCP,
						Port:       7681,
						TargetPort: intstr.FromInt(7681),
					},
					{
						Name:       "vscode-tunnel",
						Protocol:   corev1.ProtocolTCP,
						Port:       8000,
						TargetPort: intstr.FromInt(8000),
					},
				},
				Type: corev1.ServiceTypeClusterIP,
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, svc, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		if err := r.Create(ctx, svc); err != nil {
			return fmt.Errorf("failed to create Service: %w", err)
		}
		logger.Info("Created Service", zap.String("name", serviceName))
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get Service: %w", err)
	}

	// Service exists, no need to update
	logger.Info("Service already exists", zap.String("name", serviceName))
	return nil
}

// handleActiveWorkspace ensures the workspace pod is running
func (r *WorkspaceReconciler) handleActiveWorkspace(ctx context.Context, workspace *workspacesv1.Workspace, logger *zap.Logger) (reconcile.Result, error) {
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
		r.updateStatusPreservingPackages(ctx, workspace)
		return reconcile.Result{}, err
	}

	// Ensure Service is created for workspace SSHD
	if err := r.ensureWorkspaceService(ctx, workspace, logger); err != nil {
		logger.Error("Failed to ensure Service", zap.Error(err))
		workspace.Status.Phase = "Failed"
		workspace.Status.Message = fmt.Sprintf("Failed to create Service: %v", err)
		r.updateStatusPreservingPackages(ctx, workspace)
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
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)

	if err == nil {
		// Pod exists, check its actual status
		logger.Info("Workspace pod already exists", zap.String("pod", podName))

		// Reconcile environment connection for running pods (happens on every reconciliation)
		if pod.Status.Phase == corev1.PodRunning {
			if err := r.reconcileEnvironmentConnection(ctx, workspace, pod, logger); err != nil {
				logger.Warn("Failed to reconcile environment connection", zap.Error(err))
				// Don't fail reconciliation on DNS update errors
			}
		}

		// Determine actual phase based on pod status
		var phase, message string
		switch pod.Status.Phase {
		case corev1.PodPending:
			// Check if it's stuck due to mount failures or other issues
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
					phase = "Failed"
					message = fmt.Sprintf("Pod scheduling failed: %s", condition.Message)
					logger.Warn("Workspace pod scheduling failed", zap.String("reason", condition.Reason), zap.String("message", condition.Message))
					return r.updateWorkspaceStatus(ctx, workspace, pod, phase, message, logger)
				}
			}

			// Check for container status errors
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
					phase = "Failed"
					message = fmt.Sprintf("Container %s is crash looping: %s", containerStatus.Name, containerStatus.State.Waiting.Message)
					logger.Warn("Workspace container crash looping", zap.String("container", containerStatus.Name))
					return r.updateWorkspaceStatus(ctx, workspace, pod, phase, message, logger)
				}
			}

			// Check init container status
			for _, initStatus := range pod.Status.InitContainerStatuses {
				if initStatus.State.Waiting != nil {
					phase = "Creating"
					message = fmt.Sprintf("Initializing: %s - %s", initStatus.State.Waiting.Reason, initStatus.State.Waiting.Message)
					logger.Info("Workspace pod initializing", zap.String("reason", initStatus.State.Waiting.Reason))
					return r.updateWorkspaceStatus(ctx, workspace, pod, phase, message, logger)
				}
				if initStatus.State.Terminated != nil && initStatus.State.Terminated.ExitCode != 0 {
					phase = "Failed"
					message = fmt.Sprintf("Init container failed: %s", initStatus.State.Terminated.Reason)
					logger.Error("Workspace init container failed", zap.String("reason", initStatus.State.Terminated.Reason))
					return r.updateWorkspaceStatus(ctx, workspace, pod, phase, message, logger)
				}
			}

			phase = "Creating"
			message = "Workspace pod is starting"

		case corev1.PodRunning:
			// Check if all containers are ready
			allReady := true
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					allReady = false
					if containerStatus.State.Waiting != nil {
						message = fmt.Sprintf("Container %s not ready: %s", containerStatus.Name, containerStatus.State.Waiting.Reason)
					} else {
						message = fmt.Sprintf("Container %s not ready", containerStatus.Name)
					}
					break
				}
			}

			if allReady {
				phase = "Running"
				message = "Workspace is running"

				// Reconcile environment connection (update DNS if needed)
				if err := r.reconcileEnvironmentConnection(ctx, workspace, pod, logger); err != nil {
					logger.Warn("Failed to reconcile environment connection", zap.Error(err))
					// Don't fail reconciliation on DNS update errors
				}
			} else {
				phase = "Creating"
				if message == "" {
					message = "Workspace containers are starting"
				}
			}

		case corev1.PodSucceeded:
			phase = "Stopped"
			message = "Workspace pod has completed"

		case corev1.PodFailed:
			phase = "Failed"
			message = fmt.Sprintf("Workspace pod failed: %s", pod.Status.Reason)
			logger.Error("Workspace pod failed", zap.String("reason", pod.Status.Reason), zap.String("message", pod.Status.Message))

		case corev1.PodUnknown:
			phase = "Failed"
			message = "Workspace pod status unknown"
			logger.Warn("Workspace pod status unknown")

		default:
			phase = "Pending"
			message = fmt.Sprintf("Pod in unexpected phase: %s", pod.Status.Phase)
		}

		return r.updateWorkspaceStatus(ctx, workspace, pod, phase, message, logger)
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

// createWorkspacePod creates a pod with multiple containers for different access methods
func (r *WorkspaceReconciler) createWorkspacePod(workspace *workspacesv1.Workspace) (*corev1.Pod, error) {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Check if workspace has an environment connection and get target namespace
	var envTargetNamespace string
	if workspace.Spec.EnvironmentRef != nil {
		env := &environmentsv1.Environment{}
		err := r.Get(context.Background(), client.ObjectKey{
			Name:      workspace.Spec.EnvironmentRef.Name,
			Namespace: workspace.Namespace,
		}, env)
		if err == nil && env.Spec.Activated {
			envTargetNamespace = env.Spec.TargetNamespace
			r.Logger.Info("Workspace has environment connection",
				zap.String("workspace", workspace.Name),
				zap.String("environment", env.Name),
				zap.String("targetNamespace", envTargetNamespace),
			)
		}
	}

	// Default resource requirements per container
	defaultResources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}

	// Override with custom resource quota if provided (divided among containers)
	if workspace.Spec.ResourceQuota != nil {
		if workspace.Spec.ResourceQuota.CPU != "" {
			defaultResources.Limits[corev1.ResourceCPU] = resource.MustParse(workspace.Spec.ResourceQuota.CPU)
		}
		if workspace.Spec.ResourceQuota.Memory != "" {
			defaultResources.Limits[corev1.ResourceMemory] = resource.MustParse(workspace.Spec.ResourceQuota.Memory)
		}
	}

	// Build environment variables
	envVars := []corev1.EnvVar{
		{
			Name:  "WORKSPACE_NAME",
			Value: workspace.Name,
		},
		{
			Name:  "WORKSPACE_NAMESPACE",
			Value: workspace.Namespace,
		},
		{
			Name:  "WORKSPACE_OWNER",
			Value: workspace.Spec.Owner,
		},
	}

	// Set PATH for container environment (kubectl exec, running services, etc.)
	// This is also set in /etc/environment for SSH sessions via PAM
	// Include /kloudlite/bin for kl CLI and other kloudlite tools
	envVars = append(envVars, corev1.EnvVar{
		Name:  "PATH",
		Value: fmt.Sprintf("/kloudlite/bin:/nix/profiles/per-user/root/workspace-%s-packages/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", workspace.Name),
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
			InitContainers: []corev1.Container{
				{
					Name:  "init-workspace-dir",
					Image: "alpine:latest",
					Command: []string{
						"sh",
						"-c",
						func() string {
							// Build search domains based on whether workspace has an environment connection
							searchDomains := "svc.cluster.local cluster.local"
							if envTargetNamespace != "" {
								// Include environment namespace first for priority
								searchDomains = fmt.Sprintf("%s.svc.cluster.local svc.cluster.local cluster.local", envTargetNamespace)
							}

							return fmt.Sprintf(`
# Create workspace directory
mkdir -p /home/kl/workspaces/%s
chown -R 1001:1001 /home/kl/workspaces

# Create /etc/environment with PATH and Kubernetes service env vars for PAM
# This will be read by PAM on SSH login (both interactive and non-interactive)
# The Kubernetes env vars are needed for kl binary to work with in-cluster config
# Include /kloudlite/bin for kl CLI and other kloudlite tools
cat > /etc-writable/environment << 'EOF'
PATH=/kloudlite/bin:/nix/profiles/per-user/root/workspace-%s-packages/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
KUBERNETES_SERVICE_HOST=10.43.0.1
KUBERNETES_SERVICE_PORT=443
WORKSPACE_NAME=%s
WORKSPACE_NAMESPACE=%s
EOF
chmod 644 /etc-writable/environment

# Create initial /etc/resolv.conf with DNS configuration
# This is written to the shared /etc directory (emptyDir volume)
# If workspace is connected to an environment, include that namespace in search domains
cat > /etc/resolv.conf << 'EOFR'
nameserver 10.43.0.10
search %s
options ndots:5
EOFR
`, workspace.Name, workspace.Name, workspace.Name, workspace.Namespace, searchDomains)
						}(),
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "kl-home",
							MountPath: "/home/kl",
						},
						{
							Name:      "etc-environment",
							MountPath: "/etc-writable",
						},
						{
							Name:      "etc-dir",
							MountPath: "/etc",
						},
					},
				},
			},
			Containers: []corev1.Container{
				// Comprehensive workspace container with all services
				{
					Name:            "workspace",
					Image:           "kloudlite/workspace-comprehensive:latest",
					ImagePullPolicy: corev1.PullNever,
					Resources:       defaultResources,
					Env:             envVars,
					Ports: []corev1.ContainerPort{
						{
							Name:          "ssh",
							ContainerPort: 22,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "code-server",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "ttyd",
							ContainerPort: 7681,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "vscode-tunnel",
							ContainerPort: 8000,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					WorkingDir: fmt.Sprintf("/home/kl/workspaces/%s", workspace.Name),
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "nix-store",
							MountPath: "/nix",
						},
						{
							Name:      "kl-home",
							MountPath: "/home/kl",
						},
						{
							Name:      "ssh-authorized-keys",
							MountPath: "/etc/ssh/kl-authorized-keys",
							ReadOnly:  true,
						},
						{
							Name:      "sshd-config",
							MountPath: "/etc/ssh/sshd_config.d",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/etc/ssh/ssh_host_rsa_key",
							SubPath:   "ssh_host_rsa_key",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/etc/ssh/ssh_host_rsa_key.pub",
							SubPath:   "ssh_host_rsa_key.pub",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/etc/ssh/ssh_host_ecdsa_key",
							SubPath:   "ssh_host_ecdsa_key",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/etc/ssh/ssh_host_ecdsa_key.pub",
							SubPath:   "ssh_host_ecdsa_key.pub",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/etc/ssh/ssh_host_ed25519_key",
							SubPath:   "ssh_host_ed25519_key",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/etc/ssh/ssh_host_ed25519_key.pub",
							SubPath:   "ssh_host_ed25519_key.pub",
							ReadOnly:  true,
						},
						{
							Name:      "etc-environment",
							MountPath: "/etc/environment",
							SubPath:   "environment",
							ReadOnly:  true,
						},
						{
							Name:      "etc-dir",
							MountPath: "/etc/resolv.conf",
							SubPath:   "resolv.conf",
							ReadOnly:  false,
						},
						{
							Name:      "kloudlite-bin",
							MountPath: "/kloudlite/bin",
							ReadOnly:  true,
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt(22),
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
								Port: intstr.FromInt(8080),
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
				{
					Name: "kl-home",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/workspace-homes/kl",
							Type: func() *corev1.HostPathType {
								t := corev1.HostPathDirectoryOrCreate
								return &t
							}(),
						},
					},
				},
				{
					Name: "ssh-authorized-keys",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/ssh-config",
							Type: func() *corev1.HostPathType {
								t := corev1.HostPathDirectoryOrCreate
								return &t
							}(),
						},
					},
				},
				{
					Name: "sshd-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "workspace-sshd-config",
							},
						},
					},
				},
				{
					Name: "ssh-host-keys",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  "ssh-host-keys",
							DefaultMode: func() *int32 { m := int32(0600); return &m }(),
						},
					},
				},
				{
					Name: "etc-environment",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/etc-environment",
							Type: func() *corev1.HostPathType {
								t := corev1.HostPathDirectoryOrCreate
								return &t
							}(),
						},
					},
				},
				{
					Name: "etc-dir",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "kloudlite-bin",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/kloudlite/bin",
							Type: func() *corev1.HostPathType{
								t := corev1.HostPathDirectory
								return &t
							}(),
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}

	// Disable Kubernetes DNS management completely
	// DNS will be managed manually:
	// - Init container creates initial /etc/resolv.conf on emptyDir volume
	// - Controller updates /etc/resolv.conf via exec when environment connection changes
	// We provide minimal DNSConfig (required by K8s when dnsPolicy=None),
	// but /etc/resolv.conf is managed separately.
	pod.Spec.DNSPolicy = corev1.DNSNone
	pod.Spec.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{"10.43.0.10"}, // Required but not used (managed via exec)
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

	// Build access URLs for all services if pod is running
	if pod.Status.PodIP != "" && phase == "Running" {
		accessURLs := make(map[string]string)
		accessURLs["ssh"] = fmt.Sprintf("ssh://%s:22", pod.Status.PodIP)
		accessURLs["code-server"] = fmt.Sprintf("http://%s:8080", pod.Status.PodIP)
		accessURLs["ttyd"] = fmt.Sprintf("http://%s:7681", pod.Status.PodIP)
		accessURLs["vscode-tunnel"] = fmt.Sprintf("http://%s:8000", pod.Status.PodIP)
		workspace.Status.AccessURLs = accessURLs

		// Keep AccessURL for backward compatibility (default to code-server)
		workspace.Status.AccessURL = accessURLs["code-server"]
	}

	// Update ConnectedEnvironment status if EnvironmentRef is set
	if workspace.Spec.EnvironmentRef != nil {
		env := &environmentsv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      workspace.Spec.EnvironmentRef.Name,
			Namespace: workspace.Namespace,
		}, env)

		if err == nil && env.Spec.Activated {
			// Update connected environment status
			workspace.Status.ConnectedEnvironment = &workspacesv1.ConnectedEnvironmentInfo{
				Name:            env.Name,
				TargetNamespace: env.Spec.TargetNamespace,
				Connected:       true,
			}
			logger.Info("Updated ConnectedEnvironment status",
				zap.String("workspace", workspace.Name),
				zap.String("environment", env.Name),
				zap.String("targetNamespace", env.Spec.TargetNamespace),
			)
		} else if err != nil {
			// Environment not found or fetch failed
			workspace.Status.ConnectedEnvironment = &workspacesv1.ConnectedEnvironmentInfo{
				Name:      workspace.Spec.EnvironmentRef.Name,
				Connected: false,
			}
			logger.Warn("Failed to fetch environment for status update",
				zap.String("workspace", workspace.Name),
				zap.String("environment", workspace.Spec.EnvironmentRef.Name),
				zap.Error(err),
			)
		} else {
			// Environment exists but not activated
			workspace.Status.ConnectedEnvironment = &workspacesv1.ConnectedEnvironmentInfo{
				Name:            env.Name,
				TargetNamespace: env.Spec.TargetNamespace,
				Connected:       false,
			}
		}
	} else {
		// No environment reference, clear connected environment status
		workspace.Status.ConnectedEnvironment = nil
	}

	if err := r.updateStatusPreservingPackages(ctx, workspace); err != nil {
		logger.Warn("Failed to update workspace status", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteHostDirectory deletes a directory on the host by creating a temporary pod
// that mounts the host filesystem and removes the directory
func (r *WorkspaceReconciler) deleteHostDirectory(ctx context.Context, hostPath string, logger *zap.Logger) error {
	// Create a privileged pod to delete the directory on the host
	// We use a Job-like approach with a one-off pod
	cleanupPodName := fmt.Sprintf("cleanup-%s", strings.ReplaceAll(hostPath, "/", "-"))
	if len(cleanupPodName) > 63 {
		// Kubernetes name limit is 63 characters
		cleanupPodName = cleanupPodName[:63]
	}

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

	// Wait for the pod to complete (with timeout)
	// We don't need to wait in the reconciliation loop, just log it
	logger.Info("Cleanup pod created, workspace directory deletion scheduled",
		zap.String("pod", cleanupPodName),
		zap.String("hostPath", hostPath),
	)

	// Schedule cleanup pod deletion after a delay (5 minutes) in a goroutine
	go func() {
		time.Sleep(5 * time.Minute)
		if err := r.Delete(context.Background(), cleanupPod); err != nil && !apierrors.IsNotFound(err) {
			logger.Warn("Failed to delete cleanup pod",
				zap.String("pod", cleanupPodName),
				zap.Error(err),
			)
		}
	}()

	return nil
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

	// Delete the workspace directory on the host
	workspaceHostPath := fmt.Sprintf("/home/kl/workspaces/%s", workspace.Name)
	if err := r.deleteHostDirectory(ctx, workspaceHostPath, logger); err != nil {
		logger.Warn("Failed to delete workspace host directory",
			zap.String("path", workspaceHostPath),
			zap.Error(err),
		)
		// Don't fail the deletion if we can't clean up the directory
		// This allows workspace deletion to proceed even if cleanup fails
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
		Owns(&packagesv1.PackageRequest{}).
		Complete(r)
}
