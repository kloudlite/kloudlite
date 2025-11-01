package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// hasActiveConnections checks if there are active SSH or web connections to the workspace
// by examining active TCP connections in the pod
// Returns: hasConnections bool, connectionCount int, error
func (r *WorkspaceReconciler) hasActiveConnections(ctx context.Context, workspace *workspacev1.Workspace) (bool, int, error) {
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

// isWorkspaceIdle checks if a workspace has been idle by checking for active connections
// A workspace is considered idle ONLY if there are no active connections (SSH, ttyd, vscode, code-server)
// Returns: isIdle bool, connectionCount int, error
func (r *WorkspaceReconciler) isWorkspaceIdle(ctx context.Context, workspace *workspacev1.Workspace) (bool, int, error) {
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
func (r *WorkspaceReconciler) checkAndSuspendIdleWorkspace(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
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
			if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
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
		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
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
		latest := &workspacev1.Workspace{}
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

// updateDNSConfigInRunningPod updates /etc/resolv.conf in a running workspace pod
// when the environment connection changes
func (r *WorkspaceReconciler) updateDNSConfigInRunningPod(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get the pod
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Only update if pod is running
	if pod.Status.Phase != corev1.PodRunning {
		logger.Info("Skipping DNS update - pod is not running",
			zap.String("phase", string(pod.Status.Phase)))
		return nil
	}

	// Build search domains based on environment connection with validation
	var domains []string
	if workspace.Spec.EnvironmentConnection != nil {
		env := &environmentv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
			Namespace: workspace.Namespace,
		}, env)
		if err == nil && env.Spec.Activated {
			// Validate environment namespace for security
			if err := utils.ValidateKubernetesNamespace(env.Spec.TargetNamespace); err != nil {
				logger.Warn("Invalid environment namespace, skipping DNS update",
					zap.String("environment", env.Name),
					zap.String("targetNamespace", env.Spec.TargetNamespace),
					zap.Error(err))
				domains = []string{"svc.cluster.local", "cluster.local"}
			} else {
				// Include validated environment namespace in search domains
				envDomain := fmt.Sprintf("%s.svc.cluster.local", env.Spec.TargetNamespace)
				domains = []string{envDomain, "svc.cluster.local", "cluster.local"}
				logger.Info("Environment connection detected for DNS update",
					zap.String("environment", env.Name),
					zap.String("targetNamespace", env.Spec.TargetNamespace))
			}
		} else {
			// Environment not found or not activated
			domains = []string{"svc.cluster.local", "cluster.local"}
			logger.Info("Environment reference exists but not active for DNS update")
		}
	} else {
		// No environment connection
		domains = []string{"svc.cluster.local", "cluster.local"}
		logger.Info("No environment connection for DNS update")
	}

	// Sanitize search domains to prevent DNS injection
	searchDomains, err := utils.SanitizeSearchDomains(domains)
	if err != nil {
		logger.Warn("Failed to sanitize search domains, using defaults",
			zap.Strings("domains", domains),
			zap.Error(err))
		searchDomains = "svc.cluster.local cluster.local"
	}

	// Build new resolv.conf content with validated domains
	resolvConf := fmt.Sprintf("nameserver 10.43.0.10\nsearch %s\noptions ndots:5\n", searchDomains)

	// Exec into pod and update /etc/resolv.conf
	// Note: /etc/resolv.conf is mounted from EmptyDir with ReadOnly: false, so it's writable
	command := []string{"sh", "-c", fmt.Sprintf("cat > /etc/resolv.conf << 'EOFR'\n%sEOFR\n", resolvConf)}
	_, err = r.execInPod(ctx, pod, "workspace", command)
	if err != nil {
		return fmt.Errorf("failed to update DNS config: %w", err)
	}

	logger.Info("Successfully updated DNS configuration in running pod",
		zap.String("workspace", workspace.Name),
		zap.String("searchDomains", searchDomains))

	return nil
}

// updateKloudliteContextFile writes the Kloudlite context state to a file in the running pod
// This file is used by kloudlite-context.sh for fast prompt rendering without API calls
func (r *WorkspaceReconciler) updateKloudliteContextFile(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get the pod
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: workspace.Namespace}, pod)
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Only update if pod is running
	if pod.Status.Phase != corev1.PodRunning {
		logger.Info("Skipping context file update - pod is not running",
			zap.String("phase", string(pod.Status.Phase)))
		return nil
	}

	// Get environment name from spec
	envName := ""
	if workspace.Spec.EnvironmentConnection != nil {
		// Fetch environment to validate it exists and is activated
		env := &environmentv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
			Namespace: workspace.Namespace,
		}, env)
		if err == nil && env.Spec.Activated {
			envName = env.Name
		}
	}

	// Get active service intercepts from workspace status
	// The status is populated by the collectActiveIntercepts function during status updates
	intercepts := []string{}
	for _, interceptStatus := range workspace.Status.ActiveIntercepts {
		intercepts = append(intercepts, interceptStatus.ServiceName)
	}

	// Build JSON content
	contextData := map[string]interface{}{
		"environment": envName,
		"intercepts":  intercepts,
	}

	jsonBytes, err := json.Marshal(contextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	// Write to /tmp/kloudlite-context.json in the pod
	command := []string{"sh", "-c", fmt.Sprintf("cat > /tmp/kloudlite-context.json << 'EOF'\n%s\nEOF\n", string(jsonBytes))}
	_, err = r.execInPod(ctx, pod, "workspace", command)
	if err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	logger.Info("Successfully updated Kloudlite context file in running pod",
		zap.String("workspace", workspace.Name),
		zap.String("environment", envName),
		zap.Strings("intercepts", intercepts))

	return nil
}

// getWorkMachineNodeSelector retrieves the nodeSelector from the user's WorkMachine
// Returns nil if WorkMachine doesn't exist or has no nodeSelector configured
func (r *WorkspaceReconciler) getWorkMachineNodeSelector(ctx context.Context, owner string) (map[string]string, error) {
	// Sanitize owner to generate WorkMachine name (same logic as webhook)
	// Replace @ with -at- and . with -
	sanitizedOwner := strings.ReplaceAll(owner, "@", "-at-")
	sanitizedOwner = strings.ReplaceAll(sanitizedOwner, ".", "-")
	workMachineName := fmt.Sprintf("wm-%s", sanitizedOwner)

	// WorkMachine is cluster-scoped, so we don't need a namespace
	workMachine := &machinesv1.WorkMachine{}
	err := r.Get(ctx, client.ObjectKey{Name: workMachineName}, workMachine)
	if err != nil {
		// WorkMachine doesn't exist or error fetching it
		r.Logger.Info("WorkMachine not found for owner",
			zap.String("owner", owner),
			zap.String("workMachineName", workMachineName),
			zap.Error(err),
		)
		return nil, nil // Return nil selector, not an error (WorkMachine might not exist yet)
	}

	return nil, nil

	// Return the nodeSelector from WorkMachine (may be nil if not set)
	// if len(workMachine.Spec.NodeSelector) > 0 {
	// 	r.Logger.Info("Found nodeSelector from WorkMachine",
	// 		zap.String("owner", owner),
	// 		zap.String("workMachineName", workMachineName),
	// 		zap.Any("nodeSelector", workMachine.Spec.NodeSelector),
	// 	)
	// }
	//
	// return workMachine.Spec.NodeSelector, nil
}

// createWorkspacePod creates a pod with multiple containers for different access methods
func (r *WorkspaceReconciler) createWorkspacePod(workspace *workspacev1.Workspace) (*corev1.Pod, error) {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get nodeSelector from the user's WorkMachine to ensure workspace runs on the same node
	// This is important for shared Nix store access via hostPath volumes
	nodeSelector, err := r.getWorkMachineNodeSelector(context.Background(), workspace.Spec.Owner)
	if err != nil {
		r.Logger.Warn("Failed to get WorkMachine nodeSelector, proceeding without it",
			zap.String("workspace", workspace.Name),
			zap.String("owner", workspace.Spec.Owner),
			zap.Error(err),
		)
	}

	// Check if workspace has an environment connection and get target namespace
	var envTargetNamespace string
	if workspace.Spec.EnvironmentConnection != nil {
		env := &environmentv1.Environment{}
		err := r.Get(context.Background(), client.ObjectKey{
			Name:      workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
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
	// /kloudlite/bin has highest priority for kl binary and system tools
	// Include /home/kl/.local/bin for user-installed npm packages like Claude Code
	envVars = append(envVars, corev1.EnvVar{
		Name:  "PATH",
		Value: fmt.Sprintf("/kloudlite/bin:/home/kl/.local/bin:/nix/profiles/per-user/root/workspace-%s-packages/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", workspace.Name),
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
				"app":                                    "workspace",
				"workspace":                              workspace.Name,
				"workspaces.kloudlite.io/workspace-name": workspace.Name,
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
							envName := ""
							if envTargetNamespace != "" {
								// Include environment namespace first for priority
								searchDomains = fmt.Sprintf("%s.svc.cluster.local svc.cluster.local cluster.local", envTargetNamespace)
								// Get environment name from EnvironmentConnection
								if workspace.Spec.EnvironmentConnection != nil {
									envName = workspace.Spec.EnvironmentConnection.EnvironmentRef.Name
								}
							}

							// Build initial Kloudlite context JSON
							// Intercepts will be empty on pod creation (added later via controller updates)
							contextJSON := fmt.Sprintf(`{"environment":"%s","intercepts":[]}`, envName)

							return fmt.Sprintf(`
# Create workspace directory
mkdir -p /home/kl/workspaces/%s
chown -R 1001:1001 /home/kl/workspaces

# Install Claude Code in user scope if not already installed
# This allows users to update it themselves without rebuilding the container
if [ ! -d "/home/kl/.local/lib/node_modules/@anthropic-ai/claude-code" ]; then
  echo "Installing Claude Code in user scope..."
  mkdir -p /home/kl/.local/lib/node_modules
  mkdir -p /home/kl/.local/bin
  # We'll use the container's npm to install in user scope on first run
  # This will be done by supervisord init script instead, skipping here
fi
chown -R 1001:1001 /home/kl/.local

# Create /etc/environment with PATH and Kubernetes service env vars for PAM
# This will be read by PAM on SSH login (both interactive and non-interactive)
# The Kubernetes env vars are needed for kl binary to work with in-cluster config
# /kloudlite/bin has highest priority for kl binary and system tools
# Include user's local bin in PATH for user-installed npm packages like Claude Code
cat > /etc-writable/environment << 'EOF'
PATH=/kloudlite/bin:/home/kl/.local/bin:/nix/profiles/per-user/root/workspace-%s-packages/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
KUBERNETES_SERVICE_HOST=10.43.0.1
KUBERNETES_SERVICE_PORT=443
WORKSPACE_NAME=%s
WORKSPACE_NAMESPACE=%s
EOF
chmod 644 /etc-writable/environment

# Create /etc/resolv.conf with DNS configuration
# If workspace is connected to an environment, include that namespace in search domains
cat > /etc-writable-resolv/resolv.conf << 'EOFR'
nameserver 10.43.0.10
search %s
options ndots:5
EOFR
chown 1001:1001 /etc-writable-resolv/resolv.conf
chmod 666 /etc-writable-resolv/resolv.conf

# Create initial Kloudlite context file for Starship prompt
# This will be updated by the controller when environment connection or intercepts change
cat > /tmp-writable/kloudlite-context.json << 'EOFC'
%s
EOFC
chmod 644 /tmp-writable/kloudlite-context.json
`, workspace.Name, workspace.Name, workspace.Name, workspace.Namespace, searchDomains, contextJSON)
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
							Name:      "etc-resolv",
							MountPath: "/etc-writable-resolv",
						},
						{
							Name:      "tmp-context",
							MountPath: "/tmp-writable",
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
							Name:          "claude-ttyd",
							ContainerPort: 7682,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "opencode-ttyd",
							ContainerPort: 7683,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "codex-ttyd",
							ContainerPort: 7684,
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
							Name:      "etc-resolv",
							MountPath: "/etc/resolv.conf",
							SubPath:   "resolv.conf",
							ReadOnly:  false,
						},
						{
							Name:      "kloudlite-bin",
							MountPath: "/kloudlite/bin",
							ReadOnly:  true,
						},
						{
							Name:      "tmp-context",
							MountPath: "/tmp",
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
							Path: "/home/kl",
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
							DefaultMode: func() *int32 { m := int32(0o600); return &m }(),
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
					Name: "etc-resolv",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "tmp-context",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "kloudlite-bin",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/kloudlite/bin",
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

	// Disable Kubernetes DNS management completely
	// DNS will be managed manually via /etc/resolv.conf written by init container to EmptyDir
	// and configured based on workspace's environment connection. We provide minimal DNSConfig
	// (required by K8s when dnsPolicy=None), but since /etc/resolv.conf is mounted from EmptyDir, this config won't be used.
	pod.Spec.DNSPolicy = corev1.DNSNone
	pod.Spec.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{"10.43.0.10"}, // Required but will be overridden by mounted resolv.conf
	}

	// Apply nodeSelector from WorkMachine to ensure workspace runs on the same node
	// This is critical for shared Nix store access via hostPath volumes
	if len(nodeSelector) > 0 {
		pod.Spec.NodeSelector = nodeSelector
		r.Logger.Info("Applied nodeSelector to workspace pod",
			zap.String("workspace", workspace.Name),
			zap.Any("nodeSelector", nodeSelector),
		)
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(workspace, pod, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	return pod, nil
}
