package workspace

import (
	"context"
	"fmt"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// createWorkspacePod creates a pod with multiple containers for different access methods
func (r *WorkspaceReconciler) createWorkspacePod(workspace *workspacev1.Workspace) (*corev1.Pod, error) {
	podName := workspace.Name

	// Get nodeSelector from the user's WorkMachine to ensure workspace runs on the same node
	// This is important for shared Nix store access via hostPath volumes
	wm, err := r.getWorkMachine(context.Background(), workspace.Spec.WorkmachineName)
	if err != nil {
		r.Logger.Warn("Failed to get WorkMachine nodeSelector, proceeding without it",
			zap.String("workspace", workspace.Name),
			zap.String("owner", workspace.Spec.OwnedBy),
			zap.Error(err),
		)
		return nil, err
	}

	// Check if workspace has an environment connection and get target namespace
	var envTargetNamespace string
	if workspace.Spec.EnvironmentConnection != nil {
		env := &environmentv1.Environment{}
		err := r.Get(context.Background(), client.ObjectKey{
			Name: workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
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
			Value: workspace.Spec.OwnedBy,
		},
	}

	// Set PATH for container environment (kubectl exec, running services, etc.)
	// This is also set in /etc/environment for SSH sessions via PAM
	// /kloudlite/bin has highest priority for kl binary and system tools
	// Include /home/kl/.local/bin for user-installed npm packages like Claude Code
	envVars = append(envVars, corev1.EnvVar{
		Name:  "PATH",
		Value: fmt.Sprintf("/kloudlite/bin:/home/kl/.local/bin:/nix/profiles/per-user/root/%s-packages/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", workspace.Name),
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

	// Get target namespace from WorkMachine to create pod in correct namespace
	targetNamespace := wm.Spec.TargetNamespace

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: targetNamespace,
			Labels: map[string]string{
				"app":                                    "workspace",
				"workspace":                              workspace.Name,
				"workspaces.kloudlite.io/workspace-name": workspace.Name,
			},
			Annotations: map[string]string{
				"kloudlite.io/workspace-display-name": workspace.Spec.DisplayName,
				"kloudlite.io/workspace-owner":        workspace.Spec.OwnedBy,
			},
		},
		Spec: corev1.PodSpec{
			InitContainers: func() []corev1.Container {
				initContainers := []corev1.Container{
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

# Create /etc/environment with all environment variables for PAM
# This will be read by PAM on SSH login (both interactive and non-interactive)
# Dump all env vars to ensure Kubernetes client config works properly
# /kloudlite/bin has highest priority for kl binary and system tools
# Include user's local bin in PATH for user-installed npm packages like Claude Code

# Start with PATH
cat > /etc-writable/environment << 'EOF'
PATH=/kloudlite/bin:/home/kl/.local/bin:/nix/profiles/per-user/root/%s-packages/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
WORKSPACE_NAME=%s
EOF

# Dump all environment variables to /etc/environment
# This ensures Kubernetes env vars like KUBERNETES_PORT_443_TCP_ADDR are available in SSH sessions
env | grep -E '^KUBERNETES_' >> /etc-writable/environment || true

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
`, workspace.Name, workspace.Name, workspace.Name, searchDomains, contextJSON)
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
				}

				// Add git clone init container if git repository is specified
				if workspace.Spec.GitRepository != nil && workspace.Spec.GitRepository.URL != "" {
					workspaceDir := fmt.Sprintf("/home/kl/workspaces/%s", workspace.Name)

					// Build git clone command with SSH config to disable host key checking
					// SSH keys are mounted from /var/lib/kloudlite/ssh-config on host to /root/.ssh
					// Use ssh_host_rsa_key which is the SSH private key available in that directory
					// This is safe because we're cloning from trusted sources specified by the user
					sshCommand := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i /root/.ssh/ssh_host_rsa_key"

					// Build git clone command
					cloneCmd := fmt.Sprintf("GIT_SSH_COMMAND='%s' git clone %s", sshCommand, workspace.Spec.GitRepository.URL)
					if workspace.Spec.GitRepository.Branch != "" {
						cloneCmd = fmt.Sprintf("GIT_SSH_COMMAND='%s' git clone -b %s %s", sshCommand, workspace.Spec.GitRepository.Branch, workspace.Spec.GitRepository.URL)
					}
					cloneCmd += fmt.Sprintf(" %s && chown -R 1001:1001 %s", workspaceDir, workspaceDir)

					// Only clone if workspace directory is empty
					fullCmd := fmt.Sprintf(
						"if [ ! -d %s ] || [ -z \"$(ls -A %s 2>/dev/null)\" ]; then %s; else echo 'Workspace folder is not empty, skipping git clone'; fi",
						workspaceDir,
						workspaceDir,
						cloneCmd,
					)

					initContainers = append(initContainers, corev1.Container{
						Name:  "git-clone",
						Image: "alpine/git:latest",
						Command: []string{
							"sh",
							"-c",
							fullCmd,
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "kl-home",
								MountPath: "/home/kl",
							},
							{
								Name:      "ssh-host-keys",
								MountPath: "/root/.ssh",
								ReadOnly:  true,
							},
						},
					})
				}

				return initContainers
			}(),
			Containers: []corev1.Container{
				// Comprehensive workspace container with all services
				{
					Name:            "workspace",
					Image:           "ghcr.io/kloudlite/kloudlite/workspace-comprehensive:dev",
					ImagePullPolicy: corev1.PullAlways,
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
							MountPath: "/etc/ssh/sshd_config",
							SubPath:   "sshd_config",
							ReadOnly:  true,
						},
						{
							Name:      "ssh-host-keys",
							MountPath: "/var/lib/kloudlite/ssh-config",
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
							Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "kl-home",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/home",
							Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "ssh-authorized-keys",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/ssh-config",
							Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "sshd-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "sshd-config",
							},
						},
					},
				},
				{
					Name: "ssh-host-keys",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/ssh-config",
							Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "etc-environment",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/kloudlite/etc-environment",
							Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
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
							Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
			},
			ServiceAccountName: workspace.Name,
			RestartPolicy:      corev1.RestartPolicyAlways,
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

	// Use node selector and tolerations instead of nodeName
	// This allows the scheduler to properly handle the pod, which is required for
	// WaitForFirstConsumer volume binding to work correctly with PVCs
	// This is critical for shared Nix store access via hostPath volumes
	pod.Spec.NodeSelector = map[string]string{
		"kubernetes.io/hostname": wm.Name,
	}
	pod.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "kloudlite.io/workmachine",
			Operator: corev1.TolerationOpEqual,
			Value:    wm.Name,
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(workspace, pod, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	return pod, nil
}
