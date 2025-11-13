package workmachine

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// createSSHHostKeysSecret ensures the SSH host keys secret exists
func (r *WorkMachineReconciler) createSSHHostKeysSecret(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := hostManagerNamespace
	secretName := fmt.Sprintf("ssh-host-keys-%s", obj.Name)

	// Defer key generation until we know if we need it (performance optimization)
	var rsaPrivateBytes, rsaPublicBytes []byte

	// Build authorized_keys content from WorkMachine spec
	var authorizedKeys strings.Builder
	for _, key := range obj.Spec.SSHPublicKeys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}

		// Validate SSH key format
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmedKey)); err != nil {
			// Skip invalid keys but don't fail the entire reconciliation
			continue
		}

		authorizedKeys.WriteString(trimmedKey)
		authorizedKeys.WriteString("\n")
	}

	// Create or update secret with all host keys and authorized_keys
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, secret, func() error {
		secret.Labels = fn.MapMerge(secret.Labels, map[string]string{
			"kloudlite.io/ssh-host-keys": "true",
			"kloudlite.io/workmachine":   obj.Name,
		})

		// Set owner reference for cascade deletion
		secret.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: obj.APIVersion,
				Kind:       obj.Kind,
				Name:       obj.Name,
				UID:        obj.UID,
			},
		})

		secret.Type = corev1.SecretTypeOpaque
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		// Only set host keys if they don't exist (preserve existing keys)
		if _, exists := secret.Data["ssh_host_rsa_key"]; !exists {
			// Generate keys only when needed (performance optimization)
			if len(rsaPrivateBytes) == 0 {
				rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					return fmt.Errorf("failed to generate RSA key: %w", err)
				}

				// Marshal RSA key
				rsaPrivateBytes = pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
				})

				rsaSSHPublicKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
				if err != nil {
					return fmt.Errorf("failed to create RSA SSH public key: %w", err)
				}
				rsaPublicBytes = ssh.MarshalAuthorizedKey(rsaSSHPublicKey)
			}
			secret.Data["ssh_host_rsa_key"] = rsaPrivateBytes
			secret.Data["ssh_host_rsa_key.pub"] = rsaPublicBytes
		}

		// Always update authorized_keys (can change when user updates SSH keys)
		secret.Data["authorized_keys"] = []byte(authorizedKeys.String())

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// ensureSSHDConfigMapStep ensures the sshd_config ConfigMap exists
func (r *WorkMachineReconciler) ensureSSHDConfigMapStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	namespace := hostManagerNamespace
	configMapName := "sshd-config"

	// Define secure sshd_config content
	sshdConfig := `# OpenSSH Server Configuration for WorkMachine Jump Host
# This configuration enables secure SSH access with jump host (bastion) functionality

# Network Configuration
Port 2222
ListenAddress 0.0.0.0

# Authentication
PermitRootLogin no
PubkeyAuthentication yes
PasswordAuthentication no
PermitEmptyPasswords no
ChallengeResponseAuthentication no
AuthorizedKeysFile /var/lib/kloudlite/ssh-config/authorized_keys

# Host Keys
HostKey /var/lib/kloudlite/ssh-config/ssh_host_rsa_key
HostKeyAlgorithms rsa-sha2-512,rsa-sha2-256

# SSH Jump Host / Bastion Configuration
AllowTcpForwarding yes
GatewayPorts yes
X11Forwarding no

# Deny shell access - only allow port forwarding (like GitHub)
PermitTTY no
AllowAgentForwarding no
PermitOpen any
ForceCommand /bin/echo "You've successfully authenticated, but Kloudlite does not provide shell access. Use port forwarding to access workspaces."

# Security
StrictModes no
MaxAuthTries 3
MaxSessions 10

# Logging
SyslogFacility AUTH
LogLevel VERBOSE

# Environment
AcceptEnv LANG LC_*

# Subsystems
Subsystem sftp /usr/lib/ssh/sftp-server
`

	cfgMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: namespace}}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, cfgMap, func() error {
		cfgMap.SetLabels(fn.MapMerge(cfgMap.GetLabels(), map[string]string{
			"kloudlite.io/ssh-config": "true",
		}))

		// Set owner reference for cascade deletion
		cfgMap.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: obj.APIVersion,
				Kind:       obj.Kind,
				Name:       obj.Name,
				UID:        obj.UID,
			},
		})

		if cfgMap.Data == nil {
			cfgMap.Data = make(map[string]string, 1)
		}

		cfgMap.Data["sshd_config"] = sshdConfig
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// ensureWorkspaceSSHDConfigMapStep ensures the sshd-config ConfigMap exists in target namespace
func (r *WorkMachineReconciler) ensureWorkspaceSSHDConfigMapStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace
	configMapName := "sshd-config"

	// Full sshd_config for workspace pods
	sshdConfig := `# Kloudlite Workspace SSH Configuration
# This configuration provides secure SSH access to workspace containers

# Network Configuration
Port 22
ListenAddress 0.0.0.0

# Authentication
PermitRootLogin no
PubkeyAuthentication yes
PasswordAuthentication no
PermitEmptyPasswords no
ChallengeResponseAuthentication no
AuthorizedKeysFile /var/lib/kloudlite/ssh-config/authorized_keys

# Host Keys
HostKey /var/lib/kloudlite/ssh-config/ssh_host_rsa_key
HostKeyAlgorithms rsa-sha2-512,rsa-sha2-256

# SSH Configuration
AllowTcpForwarding yes
X11Forwarding no
PermitTTY yes
AllowAgentForwarding yes

# Security
StrictModes no
MaxAuthTries 3
MaxSessions 10

# Logging
SyslogFacility AUTH
LogLevel INFO

# Environment
AcceptEnv LANG LC_*
SetEnv TERM=xterm-256color

# Subsystems
Subsystem sftp /usr/lib/ssh/sftp-server
`

	cfgMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: namespace}}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, cfgMap, func() error {
		cfgMap.SetLabels(fn.MapMerge(cfgMap.GetLabels(), map[string]string{
			"kloudlite.io/ssh-config":       "true",
			"kloudlite.io/workspace-config": "true",
		}))
		if cfgMap.Data == nil {
			cfgMap.Data = make(map[string]string, 1)
		}

		cfgMap.Data["sshd_config"] = sshdConfig
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}
