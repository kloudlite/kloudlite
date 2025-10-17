package workmachine

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/aws"
	wmerrors "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/types"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// WorkMachineReconciler reconciles a WorkMachine object
type WorkMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	K3sAgentToken string

	// Cached AWS provider (thread-safe, single instance since all nodes are in same region)
	awsProvider *aws.Provider
}

const WorkMachineFinalizerName = "workmachine.machines.kloudlite.io/cleanup"

// SSH Configuration Constants
const (
	// SSHUserUID is the UID for the SSH server user
	SSHUserUID = "1000"
	// SSHUserGID is the GID for the SSH server user
	SSHUserGID = "1000"
	// SSHUserName is the username for the SSH server
	SSHUserName = "kloudlite"
)

// updateWorkMachineStatus updates the WorkMachine status with retry logic
func (r *WorkMachineReconciler) updateWorkMachineStatus(
	ctx context.Context,
	workMachine *machinesv1.WorkMachine,
	updateFunc func() error,
	logger logr.Logger,
) error {
	// Use a no-op zap logger since we already have logr for this controller
	// The retry logic is more important than the specific logger implementation
	zapLogger := zap.NewNop()
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, workMachine, updateFunc, zapLogger)
}

// SSH Jump Host Architecture
//
// This controller implements a secure SSH jump host (bastion) pattern for accessing workspaces:
//
// Authentication Flow:
// 1. User authenticates to jump host using their public key (from WorkMachine.Spec.SSHPublicKeys)
// 2. Jump host forwards TCP connection to workspace
// 3. Workspace authenticates jump host using SSH proxy key pair
//
// Jump Host (workmachine-host-manager):
// - Runs OpenSSH server on port 2222
// - Authorizes users via ssh-authorized-keys ConfigMap (user keys only)
// - Has TCP forwarding enabled (AllowTcpForwarding yes)
// - Does NOT provide shell access (PermitTTY no, ForceCommand denies shells)
// - Does NOT authenticate to workspaces (jump hosts work by TCP forwarding)
// - Password authentication disabled for security
// - Works like GitHub's SSH: authenticates users but only allows port forwarding
//
// Workspaces:
// - Run OpenSSH servers that authorize the jump host's SSH proxy public key
// - Jump host uses SSH proxy private key to authenticate to workspaces
//
// Security:
// - All SSH keys are validated using ssh.ParseAuthorizedKey() before use
// - Password authentication is disabled (PasswordAuthentication no)
// - Only public key authentication is allowed
// - Jump host runs as non-privileged user (UID/GID 1000)
//
// ConfigMaps:
// - ssh-authorized-keys: Contains user public keys for jump host authentication
// - sshd-config: Contains OpenSSH server configuration with security hardening

// getOrCreateAWSProvider returns a cached AWS provider for the given WorkMachine
// Creates a new provider if one doesn't exist (single instance since all nodes are in same region)
// The context from the first call is used to initialize the AWS client
func (r *WorkMachineReconciler) getOrCreateAWSProvider(
	ctx context.Context, obj *machinesv1.WorkMachine,
) (*aws.Provider, error) {
	if obj.Spec.AWSProvider == nil {
		return nil, NewInvalidConfigurationError("aws", "AWS provider configuration is required")
	}

	// Initialize provider once using sync.Once
	// The context from the first reconciliation is used to create the AWS client
	r.awsProviderOnce.Do(func() {
		// Create provider with K3s token
	})

	if r.awsProviderErr != nil {
		return nil, r.awsProviderErr
	}

	return r.awsProvider, nil
}

// Reconcile handles WorkMachine CR reconciliation
func (r *WorkMachineReconciler) Reconcile(
	ctx context.Context, request reconcile.Request,
) (reconcile.Result, error) {
	// Use operator toolkit pattern for both K8s and cloud provider WorkMachines
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &machinesv1.WorkMachine{})
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*machinesv1.WorkMachine]{
		{
			Name:     "ensure-namespace",
			Title:    "Ensure target namespace exists",
			OnCreate: r.ensureNamespaceStep,
			OnDelete: r.deleteNamespaceStep,
		},
		{
			Name:     "ensure-rbac",
			Title:    "Ensure RBAC resources for workmachine-node-manager",
			OnCreate: r.ensurePackageManagerRBACStep,
			OnDelete: nil,
		},
		{
			Name:     "ensure-ssh-host-keys",
			Title:    "Ensure SSH host keys secret",
			OnCreate: r.ensureSSHHostKeysSecretStep,
			OnDelete: nil,
		},
		{
			Name:     "ensure-ssh-authorized-keys",
			Title:    "Ensure SSH authorized_keys ConfigMap",
			OnCreate: r.ensureSSHAuthorizedKeysConfigMapStep,
			OnDelete: nil,
		},
		{
			Name:     "ensure-sshd-config",
			Title:    "Ensure sshd_config ConfigMap",
			OnCreate: r.ensureSSHDConfigMapStep,
			OnDelete: nil,
		},
		{
			Name:     "ensure-workspace-sshd-config",
			Title:    "Ensure workspace-sshd-config ConfigMap",
			OnCreate: r.ensureWorkspaceSSHDConfigMapStep,
			OnDelete: nil,
		},
		{
			Name:     "ensure-deployment",
			Title:    "Ensure workmachine-host-manager deployment",
			OnCreate: r.ensurePackageManagerDeploymentStep,
			OnDelete: nil,
		},
		{
			Name:     "handle-state-transitions",
			Title:    "Handle WorkMachine state transitions",
			OnCreate: r.handleStateTransitionsStep,
			OnDelete: nil,
		},
		{
			Name:     "validate-cloud-provider-permissions",
			Title:    "Validate cloud provider permissions",
			OnCreate: r.validateCloudProviderPermissionsStep,
			OnDelete: nil,
		},
		{
			Name:     "initialize-cloud-provider-status",
			Title:    "Initialize cloud provider WorkMachine status",
			OnCreate: r.initializeCloudProviderStatusStep,
			OnDelete: nil,
		},
		{
			Name:     "reconcile-cloud-instance-state",
			Title:    "Reconcile cloud instance state",
			OnCreate: r.reconcileCloudInstanceStateStep,
			OnDelete: r.deleteCloudInstanceStep,
		},
		{
			Name:     "ensure-cloud-dns",
			Title:    "Ensure DNS configuration for cloud instance",
			OnCreate: r.ensureCloudDNSStep,
			OnDelete: r.deleteCloudDNSStep,
		},
		{
			Name:     "check-auto-shutdown",
			Title:    "Check auto-shutdown conditions",
			OnCreate: r.checkAutoShutdownStep,
			OnDelete: nil,
		},
	})
}

// getK8sReconciliationSteps returns the reconciliation steps for K8s-based WorkMachines
// Step handlers for K8s-based WorkMachine reconciliation

// ensureNamespaceStep ensures the target namespace exists with finalizer
func (r *WorkMachineReconciler) ensureNamespaceStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines (they don't need K8s namespace)
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := &corev1.Namespace{}
	err := r.Get(check.Context(), client.ObjectKey{Name: obj.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace already exists, ensure it has the finalizer
		if !controllerutil.ContainsFinalizer(namespace, WorkMachineFinalizerName) {
			controllerutil.AddFinalizer(namespace, WorkMachineFinalizerName)
			if err := r.Update(check.Context(), namespace); err != nil {
				return check.Failed(err)
			}
			return check.UpdateMsg("Added finalizer to existing namespace")
		}
		return check.Passed()
	}

	if !errors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Create the namespace with finalizer
	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: obj.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/managed":     "true",
				"kloudlite.io/workmachine": "true",
			},
			Finalizers: []string{WorkMachineFinalizerName},
		},
	}

	if err := r.Create(check.Context(), namespace); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// deleteNamespaceStep handles namespace deletion when WorkMachine is being deleted
func (r *WorkMachineReconciler) deleteNamespaceStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespaceName := obj.Spec.TargetNamespace

	// Check for active Workspaces in the target namespace
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList, client.InNamespace(namespaceName)); err != nil {
		if !errors.IsNotFound(err) {
			return check.Errored(err)
		}
	}

	// Block deletion if there are active workspaces
	if len(workspaceList.Items) > 0 {
		workspaceNames := make([]string, len(workspaceList.Items))
		for i, ws := range workspaceList.Items {
			workspaceNames[i] = ws.Name
		}

		return check.UpdateMsg(fmt.Sprintf("Deletion blocked: %d active workspaces exist (%s)",
			len(workspaceList.Items), strings.Join(workspaceNames, ", ")))
	}

	// No workspaces, proceed with namespace deletion
	namespace := &corev1.Namespace{}
	err := r.Get(check.Context(), client.ObjectKey{Name: namespaceName}, namespace)
	if err == nil {
		// Namespace still exists
		if namespace.DeletionTimestamp != nil {
			// Namespace is being deleted - remove our finalizer to allow it to complete
			if controllerutil.ContainsFinalizer(namespace, WorkMachineFinalizerName) {
				controllerutil.RemoveFinalizer(namespace, WorkMachineFinalizerName)
				if err := r.Update(check.Context(), namespace); err != nil {
					return check.Failed(err)
				}
			}
			return check.UpdateMsg("Namespace is being deleted, waiting for completion")
		}

		// Delete the namespace
		if err := r.Delete(check.Context(), namespace); err != nil && !errors.IsNotFound(err) {
			return check.Failed(err)
		}

		return check.UpdateMsg("Namespace deletion initiated, waiting for completion")
	}

	if !errors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Namespace is deleted
	return check.Passed()
}

// ensurePackageManagerRBACStep ensures RBAC resources for workmachine-node-manager
func (r *WorkMachineReconciler) ensurePackageManagerRBACStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := obj.Spec.TargetNamespace

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
	}

	if err := r.Get(check.Context(), client.ObjectKey{Name: sa.Name, Namespace: namespace}, sa); err != nil {
		if !errors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Create service account
		if err := r.Create(check.Context(), sa); err != nil && !errors.IsAlreadyExists(err) {
			return check.Failed(err)
		}
	}

	// Create Role with PackageRequest permissions
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	existingRole := &rbacv1.Role{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: role.Name, Namespace: namespace}, existingRole); err != nil {
		if !errors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Create role
		if err := r.Create(check.Context(), role); err != nil && !errors.IsAlreadyExists(err) {
			return check.Failed(err)
		}
	}

	// Create RoleBinding
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "workmachine-node-manager",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "workmachine-node-manager",
		},
	}

	existingRB := &rbacv1.RoleBinding{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: rb.Name, Namespace: namespace}, existingRB); err != nil {
		if !errors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Create role binding
		if err := r.Create(check.Context(), rb); err != nil && !errors.IsAlreadyExists(err) {
			return check.Failed(err)
		}
	}

	return check.Passed()
}

// ensureSSHHostKeysSecretStep ensures the SSH host keys secret exists
func (r *WorkMachineReconciler) ensureSSHHostKeysSecretStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := obj.Spec.TargetNamespace
	secretName := "ssh-host-keys"
	secret := &corev1.Secret{}
	err := r.Get(check.Context(), client.ObjectKey{Name: secretName, Namespace: namespace}, secret)

	if err == nil {
		// Secret already exists
		return check.Passed()
	}

	if !errors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Generate RSA 2048-bit key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to generate RSA key: %w", err))
	}

	// Marshal RSA key
	rsaPrivateBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})

	rsaSSHPublicKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to create RSA SSH public key: %w", err))
	}
	rsaPublicBytes := ssh.MarshalAuthorizedKey(rsaSSHPublicKey)

	// Create secret with all host keys
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/ssh-host-keys": "true",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"ssh_host_rsa_key":     rsaPrivateBytes,
			"ssh_host_rsa_key.pub": rsaPublicBytes,
		},
	}

	if err := r.Create(check.Context(), secret); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// ensureSSHAuthorizedKeysConfigMapStep ensures the SSH authorized_keys ConfigMap exists
func (r *WorkMachineReconciler) ensureSSHAuthorizedKeysConfigMapStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := obj.Spec.TargetNamespace
	configMapName := "ssh-authorized-keys"

	// Build authorized_keys content with user keys from WorkMachine spec
	// Validate each SSH key before adding to authorized_keys
	var authorizedKeys strings.Builder
	for _, key := range obj.Spec.SSHPublicKeys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}

		// Validate SSH key format using golang.org/x/crypto/ssh
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmedKey)); err != nil {
			// Skip invalid keys but don't fail the entire reconciliation
			continue
		}

		authorizedKeys.WriteString(trimmedKey)
		authorizedKeys.WriteString("\n")
	}

	cfgMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: namespace}}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, cfgMap, func() error {
		cfgMap.SetLabels(fn.MapMerge(cfgMap.GetLabels(), map[string]string{
			"kloudlite.io/ssh-config": "true",
		}))
		if cfgMap.Data == nil {
			cfgMap.Data = make(map[string]string, 1)
		}

		cfgMap.Data["authorized_keys"] = authorizedKeys.String()
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// ensureSSHDConfigMapStep ensures the sshd_config ConfigMap exists
func (r *WorkMachineReconciler) ensureSSHDConfigMapStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := obj.Spec.TargetNamespace
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

// ensureWorkspaceSSHDConfigMapStep ensures the workspace-sshd-config ConfigMap exists
func (r *WorkMachineReconciler) ensureWorkspaceSSHDConfigMapStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := obj.Spec.TargetNamespace
	configMapName := "workspace-sshd-config"

	// SSHD drop-in config to override AuthorizedKeysFile location
	sshdConfigOverride := `# Kloudlite Workspace SSH Configuration
# Override authorized keys location to use mounted ConfigMap
AuthorizedKeysFile /etc/ssh/kl-authorized-keys/authorized_keys
# Disable StrictModes to allow ConfigMap-mounted directories (owned by root)
StrictModes no
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

		cfgMap.Data["99-kl-authorized-keys.conf"] = sshdConfigOverride
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// ensurePackageManagerDeploymentStep ensures the workmachine-host-manager deployment exists
func (r *WorkMachineReconciler) ensurePackageManagerDeploymentStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	namespace := obj.Spec.TargetNamespace
	deploymentName := "workmachine-host-manager"
	deployment := &appsv1.Deployment{}
	err := r.Get(check.Context(), client.ObjectKey{Name: deploymentName, Namespace: namespace}, deployment)

	if err == nil {
		// Deployment already exists
		return check.Passed()
	}

	if !errors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Create the Deployment (same content as before, just using step result)
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	privileged := true
	replicas := int32(1)

	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                       "workmachine-host-manager",
				"kloudlite.io/package-mgmt": "true",
				"kloudlite.io/workmachine":  obj.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "workmachine-host-manager",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "workmachine-host-manager",
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					DNSPolicy:   corev1.DNSNone,
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{"10.43.0.10"},
						Searches:    []string{namespace + ".svc.cluster.local", "svc.cluster.local", "cluster.local"},
						Options: []corev1.PodDNSConfigOption{
							{Name: "ndots", Value: func() *string { v := "5"; return &v }()},
						},
					},
					ServiceAccountName: "workmachine-node-manager",
					NodeSelector:       workMachine.Spec.NodeSelector,
					InitContainers: []corev1.Container{
						{
							Name:            "setup-nix",
							Image:           "kloudlite/workmachine-node-manager:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Command: []string{"sh", "-c"},
							Args: []string{
								`#!/bin/sh
set -e

echo "Checking if Nix is already on shared volume..."

# Check if Nix is already on the shared volume
if [ -f /nix-shared/var/nix/profiles/default/etc/profile.d/nix.sh ]; then
  echo "Nix already exists on shared volume, skipping copy"
else
  echo "Copying Nix from image to shared volume..."

  # Copy the entire /nix directory from this container's image to the shared volume
  # The kloudlite/workmachine-node-manager image already has Nix installed at /nix
  # We need to copy it to the hostPath so it's available to other containers
  if [ -d /nix ]; then
    # Create target directory structure
    mkdir -p /nix-shared
    # Copy everything from /nix to /nix-shared
    cp -a /nix/* /nix-shared/
    echo "Nix copied successfully to shared volume"
  else
    echo "ERROR: /nix not found in image"
    exit 1
  fi
fi

# Always ensure profile directory exists (idempotent - safe to run multiple times)
# This is required for nix-env to work properly with user profiles
echo "Ensuring profile directory exists..."
mkdir -p /nix-shared/profiles/per-user/root
echo "Profile directory ready"
`,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nix-store",
									MountPath: "/nix-shared",
								},
							},
						},
						{
							Name:            "setup-ssh-key",
							Image:           "busybox:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								fmt.Sprintf("cp /ssh-key-source/private-key /ssh-key-target/id_ed25519 && chown %s:%s /ssh-key-target/id_ed25519 && chmod 600 /ssh-key-target/id_ed25519", SSHUserUID, SSHUserGID),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "ssh-proxy-key",
									MountPath: "/ssh-key-source",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-key-volume",
									MountPath: "/ssh-key-target",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "workmachine-node-manager",
							Image:           "kloudlite/workmachine-node-manager:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: namespace,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nix-store",
									MountPath: "/nix",
								},
								{
									Name:      "workspace-homes",
									MountPath: "/var/lib/kloudlite/workspace-homes",
								},
								{
									Name:      "ssh-config",
									MountPath: "/var/lib/kloudlite/ssh-config",
								},
							},
						},
						{
							Name:            "ssh-server",
							Image:           "linuxserver/openssh-server:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "PUID",
									Value: SSHUserUID,
								},
								{
									Name:  "PGID",
									Value: SSHUserGID,
								},
								{
									Name:  "PASSWORD_ACCESS",
									Value: "false",
								},
								{
									Name:  "USER_PASSWORD",
									Value: "kloudlite123",
								},
								{
									Name:  "USER_NAME",
									Value: SSHUserName,
								},
								{
									Name:  "SUDO_ACCESS",
									Value: "false",
								},
								{
									Name:  "TCP_FORWARDING",
									Value: "true",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "ssh",
									ContainerPort: 2222,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "ssh-key-volume",
									MountPath: "/config/.ssh/id_ed25519",
									SubPath:   "id_ed25519",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-config",
									MountPath: "/var/lib/kloudlite/ssh-config",
									ReadOnly:  false,
								},
								{
									Name:      "sshd-config",
									MountPath: "/etc/ssh/sshd_config",
									SubPath:   "sshd_config",
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
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "nix-store",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/nix-store",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "workspace-homes",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/workspace-homes",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "ssh-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/ssh-config",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "ssh-proxy-key",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "ssh-proxy-key",
								},
							},
						},
						{
							Name: "ssh-key-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
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
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "ssh-host-keys",
									DefaultMode: func() *int32 { m := int32(0o600); return &m }(),
								},
							},
						},
					},
				},
			},
		},
	}

	if err := r.Create(check.Context(), deployment); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// handleStateTransitionsStep handles WorkMachine state transitions
func (r *WorkMachineReconciler) handleStateTransitionsStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for cloud provider WorkMachines (they have their own state handling)
	if obj.Spec.Provider != "" {
		return check.Passed()
	}

	// Initialize status if it doesn't exist
	if obj.Status.State == "" {
		// First time - set current state to desired state
		obj.Status.State = obj.Spec.DesiredState
		now := metav1.Now()
		if obj.Spec.DesiredState == machinesv1.MachineStateRunning {
			obj.Status.StartedAt = &now
		}

		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg(fmt.Sprintf("Initialized WorkMachine status to %s", obj.Status.State))
	}

	// Check if state transition is needed
	currentState := obj.Status.State
	desiredState := obj.Spec.DesiredState

	if currentState == desiredState {
		// No transition needed, machine is in desired state
		return check.Passed()
	}

	// Handle state transitions
	switch currentState {
	case machinesv1.MachineStateStopped:
		if desiredState == machinesv1.MachineStateRunning {
			// Transition to starting
			obj.Status.State = machinesv1.MachineStateStarting
			if err := r.Status().Update(check.Context(), obj); err != nil {
				return check.Failed(err)
			}
			return check.UpdateMsg("Machine transitioning to starting")
		}

	case machinesv1.MachineStateStarting:
		// Transition to running
		obj.Status.State = machinesv1.MachineStateRunning
		now := metav1.Now()
		obj.Status.StartedAt = &now
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg("Machine is now running")

	case machinesv1.MachineStateRunning:
		if desiredState == machinesv1.MachineStateStopped {
			// Transition to stopping
			obj.Status.State = machinesv1.MachineStateStopping
			if err := r.Status().Update(check.Context(), obj); err != nil {
				return check.Failed(err)
			}
			return check.UpdateMsg("Machine transitioning to stopping")
		}

	case machinesv1.MachineStateStopping:
		// Transition to stopped
		obj.Status.State = machinesv1.MachineStateStopped
		now := metav1.Now()
		obj.Status.StoppedAt = &now
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg("Machine is now stopped")
	}

	return check.Passed()
}

// Step handlers for cloud provider-based WorkMachine reconciliation

// validateCloudProviderPermissionsStep validates cloud provider permissions on first reconciliation
func (r *WorkMachineReconciler) validateCloudProviderPermissionsStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only validate on first reconciliation (when state is empty)
	if obj.Status.State != "" {
		return check.Passed()
	}

	// Only AWS is supported for now
	if obj.Spec.Provider != machinesv1.AWS {
		return check.Failed(fmt.Errorf("unsupported provider: %s", obj.Spec.Provider))
	}

	provider, err := r.getOrCreateAWSProvider(check.Context(), obj)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get AWS provider: %w", err))
	}

	if err := provider.ValidatePermissions(check.Context()); err != nil {
		obj.Status.State = machinesv1.MachineStateError
		obj.Status.Message = fmt.Sprintf("Permission validation failed: %v", err)
		if updateErr := r.Status().Update(check.Context(), obj); updateErr != nil {
			return check.Errored(updateErr)
		}
		return check.Failed(err)
	}

	return check.Passed()
}

// initializeCloudProviderStatusStep initializes status for cloud provider WorkMachine
func (r *WorkMachineReconciler) initializeCloudProviderStatusStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only initialize if status is empty
	if obj.Status.State != "" {
		return check.Passed()
	}

	obj.Status.State = obj.Spec.DesiredState
	now := metav1.Now()
	if obj.Spec.DesiredState == machinesv1.MachineStateRunning {
		obj.Status.StartedAt = &now
	}

	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// reconcileCloudInstanceStateStep reconciles the cloud instance to the desired state
func (r *WorkMachineReconciler) reconcileCloudInstanceStateStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only AWS is supported for now
	if obj.Spec.Provider != machinesv1.AWS {
		return check.Failed(fmt.Errorf("unsupported provider: %s", obj.Spec.Provider))
	}

	provider, err := r.getOrCreateAWSProvider(check.Context(), obj)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get AWS provider: %w", err))
	}

	currentState := obj.Status.State
	desiredState := obj.Spec.DesiredState

	// Handle state transitions based on current state
	switch currentState {
	case machinesv1.MachineStateStopped:
		if desiredState == machinesv1.MachineStateRunning {
			return r.handleInstanceStart(check, obj, provider)
		}
		return check.Passed()

	case machinesv1.MachineStateStarting:
		return r.handleInstanceStarting(check, obj, provider)

	case machinesv1.MachineStateRunning:
		if desiredState == machinesv1.MachineStateStopped {
			return r.handleInstanceStop(check, obj, provider)
		}
		return r.handleInstanceRunning(check, obj, provider)

	case machinesv1.MachineStateStopping:
		return r.handleInstanceStopping(check, obj, provider)

	case machinesv1.MachineStateError:
		// Try to recover by checking current instance status
		if obj.Status.InstanceID == "" {
			return check.UpdateMsg("In error state, no instance ID to check")
		}
		info, err := provider.GetInstance(check.Context(), obj.Status.InstanceID)
		if err != nil {
			return check.UpdateMsg(fmt.Sprintf("In error state, failed to get instance status: %v", err))
		}
		obj.Status.State = mapInstanceStateToMachineState(info.State)
		obj.Status.Message = "Recovered from error state"
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg("Recovered from error state")
	}

	return check.Passed()
}

// handleInstanceStart handles starting a stopped cloud instance
func (r *WorkMachineReconciler) handleInstanceStart(
	check *reconciler.Check[*machinesv1.WorkMachine],
	obj *machinesv1.WorkMachine,
	provider CloudProviderInterface,
) reconciler.StepResult {
	// Check if instance exists
	if obj.Status.InstanceID == "" {
		// No instance exists, create it (orchestrated)
		info, err := r.ensureCloudInstance(check.Context(), obj, provider)
		if err != nil {
			obj.Status.State = machinesv1.MachineStateError
			obj.Status.Message = fmt.Sprintf("Failed to create instance: %v", err)
			if updateErr := r.Status().Update(check.Context(), obj); updateErr != nil {
				return check.Errored(updateErr)
			}
			return check.Failed(err)
		}

		// Update status with instance information
		r.updateStatusFromInstanceInfo(obj, info)
		obj.Status.State = machinesv1.MachineStateStarting
		obj.Status.Message = "Instance created, starting"
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}

		return check.UpdateMsg("Instance created, starting")
	}

	// Instance exists, start it
	if err := provider.StartInstance(check.Context(), obj.Status.InstanceID); err != nil {
		obj.Status.State = machinesv1.MachineStateError
		obj.Status.Message = fmt.Sprintf("Failed to start instance: %v", err)
		if updateErr := r.Status().Update(check.Context(), obj); updateErr != nil {
			return check.Errored(updateErr)
		}
		return check.Failed(err)
	}

	obj.Status.State = machinesv1.MachineStateStarting
	obj.Status.Message = "Instance starting"
	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.UpdateMsg("Instance starting")
}

// handleInstanceStarting handles a starting cloud instance
func (r *WorkMachineReconciler) handleInstanceStarting(
	check *reconciler.Check[*machinesv1.WorkMachine],
	obj *machinesv1.WorkMachine,
	provider CloudProviderInterface,
) reconciler.StepResult {
	info, err := provider.GetInstance(check.Context(), obj.Status.InstanceID)
	if err != nil {
		return check.UpdateMsg(fmt.Sprintf("Failed to get instance status: %v", err))
	}

	r.updateStatusFromInstanceInfo(obj, info)

	if info.State == types.InstanceStateRunning {
		obj.Status.State = machinesv1.MachineStateRunning
		now := metav1.Now()
		obj.Status.StartedAt = &now
		obj.Status.Message = "Instance running"

		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg("Instance is now running")
	}

	// Still starting
	obj.Status.Message = fmt.Sprintf("Instance starting: %s", info.Message)
	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.UpdateMsg(fmt.Sprintf("Instance starting: %s", info.Message))
}

// handleInstanceRunning handles a running cloud instance
func (r *WorkMachineReconciler) handleInstanceRunning(
	check *reconciler.Check[*machinesv1.WorkMachine],
	obj *machinesv1.WorkMachine,
	provider CloudProviderInterface,
) reconciler.StepResult {
	// Get current instance status
	info, err := provider.GetInstance(check.Context(), obj.Status.InstanceID)
	if err != nil {
		return check.UpdateMsg(fmt.Sprintf("Failed to get instance status: %v", err))
	}

	r.updateStatusFromInstanceInfo(obj, info)

	// Check if instance is still running
	if info.State != types.InstanceStateRunning {
		obj.Status.State = mapInstanceStateToMachineState(info.State)
		obj.Status.Message = info.Message
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg(fmt.Sprintf("Instance state changed to %s", info.State))
	}

	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// handleInstanceStop handles stopping a running cloud instance
func (r *WorkMachineReconciler) handleInstanceStop(
	check *reconciler.Check[*machinesv1.WorkMachine],
	obj *machinesv1.WorkMachine,
	provider CloudProviderInterface,
) reconciler.StepResult {
	if err := provider.StopInstance(check.Context(), obj.Status.InstanceID); err != nil {
		obj.Status.State = machinesv1.MachineStateError
		obj.Status.Message = fmt.Sprintf("Failed to stop instance: %v", err)
		if updateErr := r.Status().Update(check.Context(), obj); updateErr != nil {
			return check.Errored(updateErr)
		}
		return check.Failed(err)
	}

	obj.Status.State = machinesv1.MachineStateStopping
	obj.Status.Message = "Instance stopping"
	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.UpdateMsg("Instance stopping")
}

// handleInstanceStopping handles a stopping cloud instance
func (r *WorkMachineReconciler) handleInstanceStopping(
	check *reconciler.Check[*machinesv1.WorkMachine],
	obj *machinesv1.WorkMachine,
	provider CloudProviderInterface,
) reconciler.StepResult {
	info, err := provider.GetInstance(check.Context(), obj.Status.InstanceID)
	if err != nil {
		return check.UpdateMsg(fmt.Sprintf("Failed to get instance status: %v", err))
	}

	r.updateStatusFromInstanceInfo(obj, info)

	if info.State == types.InstanceStateStopped {
		obj.Status.State = machinesv1.MachineStateStopped
		now := metav1.Now()
		obj.Status.StoppedAt = &now
		obj.Status.Message = "Instance stopped"
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg("Instance is now stopped")
	}

	// Still stopping
	obj.Status.Message = fmt.Sprintf("Instance stopping: %s", info.Message)
	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.UpdateMsg(fmt.Sprintf("Instance stopping: %s", info.Message))
}

// ensureCloudDNSStep ensures DNS configuration for running cloud instances
func (r *WorkMachineReconciler) ensureCloudDNSStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only configure DNS when instance is running and has a public IP
	if obj.Status.State != machinesv1.MachineStateRunning || obj.Status.PublicIP == "" {
		return check.Passed()
	}

	// Skip if DNS is already configured
	if obj.Status.Route53RecordSet != "" {
		return check.Passed()
	}

	// Only AWS is supported for now
	if obj.Spec.Provider != machinesv1.AWS {
		return check.Failed(fmt.Errorf("unsupported provider: %s", obj.Spec.Provider))
	}

	provider, err := r.getOrCreateAWSProvider(check.Context(), obj)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get AWS provider: %w", err))
	}

	// Build FQDN from WorkMachine name and domain (orchestration logic)
	fqdn := fmt.Sprintf("%s.%s", obj.Name, obj.Spec.AWSProvider.DomainName)

	// Upsert DNS record
	if err := provider.UpsertDNSRecord(check.Context(), fqdn, obj.Status.PublicIP); err != nil {
		obj.Status.Message = fmt.Sprintf("Instance running but DNS configuration failed: %v", err)
		if updateErr := r.Status().Update(check.Context(), obj); updateErr != nil {
			return check.Errored(updateErr)
		}
		return check.Failed(err)
	}

	obj.Status.Route53RecordSet = fqdn
	obj.Status.AccessURL = fmt.Sprintf("https://%s", fqdn)
	obj.Status.Message = fmt.Sprintf("Instance running and accessible at %s", fqdn)

	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// deleteCloudDNSStep deletes DNS records when WorkMachine is being deleted
func (r *WorkMachineReconciler) deleteCloudDNSStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only AWS is supported for now
	if obj.Spec.Provider != machinesv1.AWS {
		return check.Passed()
	}

	provider, err := r.getOrCreateAWSProvider(check.Context(), obj)
	if err != nil {
		// If we can't create provider, just pass - instance might not exist anymore
		return check.Passed()
	}

	// Delete DNS record if it exists
	if obj.Status.Route53RecordSet != "" && obj.Status.PublicIP != "" {
		if err := provider.DeleteDNSRecord(check.Context(), obj.Status.Route53RecordSet, obj.Status.PublicIP); err != nil {
			// Log error but don't fail - continue with deletion
			return check.UpdateMsg(fmt.Sprintf("Failed to delete DNS record: %v (continuing)", err))
		}
	}

	return check.Passed()
}

// deleteCloudInstanceStep deletes the cloud instance when WorkMachine is being deleted
func (r *WorkMachineReconciler) deleteCloudInstanceStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only AWS is supported for now
	if obj.Spec.Provider != machinesv1.AWS {
		return check.Passed()
	}

	provider, err := r.getOrCreateAWSProvider(check.Context(), obj)
	if err != nil {
		// If we can't create provider, just pass - instance might not exist anymore
		return check.Passed()
	}

	// Delete instance if it exists
	if obj.Status.InstanceID != "" {
		if err := provider.DeleteInstance(check.Context(), obj.Status.InstanceID); err != nil {
			return check.Failed(err)
		}
	}

	// Delete security group if it exists
	if obj.Status.SecurityGroupID != "" {
		if err := provider.DeleteSecurityGroup(check.Context(), obj.Status.SecurityGroupID); err != nil {
			// Log error but don't fail - SG might be in use or already deleted
			return check.UpdateMsg(fmt.Sprintf("Failed to delete security group: %v (continuing)", err))
		}
	}

	return check.Passed()
}

// checkAutoShutdownStep checks if auto-shutdown should trigger
func (r *WorkMachineReconciler) checkAutoShutdownStep(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	// Skip for K8s-based WorkMachines
	if obj.Spec.Provider == "" {
		return check.Passed()
	}

	// Only check when instance is running and auto-shutdown is enabled
	if obj.Status.State != machinesv1.MachineStateRunning {
		return check.Passed()
	}

	if obj.Spec.AutoShutdown == nil || !obj.Spec.AutoShutdown.Enabled {
		return check.Passed()
	}

	// List all workspaces for this user
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList, client.MatchingFields{"spec.owner": obj.Spec.OwnedBy}); err != nil {
		return check.UpdateMsg(fmt.Sprintf("Failed to list workspaces: %v", err))
	}

	// Count active workspaces
	activeCount := int32(0)
	var lastActivity *metav1.Time

	for _, ws := range workspaceList.Items {
		// Check if workspace is active (not suspended or archived)
		if ws.Spec.Status == "active" && ws.Status.Phase == "Running" {
			activeCount++

			// Track most recent activity
			if ws.Status.LastActivityTime != nil {
				if lastActivity == nil || ws.Status.LastActivityTime.After(lastActivity.Time) {
					lastActivity = ws.Status.LastActivityTime
				}
			}
		}
	}

	// Update status with workspace activity
	obj.Status.ActiveWorkspaceCount = activeCount
	obj.Status.LastWorkspaceActivity = lastActivity

	// If there are active workspaces, don't shutdown
	if activeCount > 0 {
		if err := r.Status().Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.Passed()
	}

	// No active workspaces - check idle threshold
	if lastActivity == nil {
		// No activity recorded yet, use started time
		lastActivity = obj.Status.StartedAt
	}

	if lastActivity == nil {
		// No reference time, skip auto-shutdown
		return check.Passed()
	}

	idleThreshold := time.Duration(obj.Spec.AutoShutdown.IdleThresholdMinutes) * time.Minute
	idleDuration := time.Since(lastActivity.Time)

	if idleDuration >= idleThreshold {
		// Trigger auto-shutdown
		obj.Spec.DesiredState = machinesv1.MachineStateStopped
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.UpdateMsg(fmt.Sprintf("Auto-shutdown triggered after %v idle", idleDuration))
	}

	if err := r.Status().Update(check.Context(), obj); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// Helper functions for cloud provider reconciliation

// ensureCloudInstance orchestrates security group and instance creation
// This is the orchestration logic that was previously embedded in EnsureInstance
func (r *WorkMachineReconciler) ensureCloudInstance(
	ctx context.Context, wm *machinesv1.WorkMachine, provider CloudProviderInterface,
) (*types.InstanceInfo, error) {
	// Step 1: Ensure security group exists
	sgName := fmt.Sprintf("workmachine-%s", wm.Name)
	sgID, err := provider.GetSecurityGroup(ctx, sgName)
	if err != nil {
		return nil, fmt.Errorf("failed to check security group: %w", err)
	}

	// If security group doesn't exist, create it
	if sgID == "" {
		sgID, err = provider.CreateSecurityGroup(ctx, wm)
		if err != nil {
			// Check if it's an "already exists" error (race condition)
			if _, ok := err.(*wmerrors.ResourceAlreadyExistsError); ok {
				// Try to get it again
				sgID, err = provider.GetSecurityGroup(ctx, sgName)
				if err != nil {
					return nil, fmt.Errorf("failed to get security group after creation race: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create security group: %w", err)
			}
		}
	}

	// Update status with security group ID
	wm.Status.SecurityGroupID = sgID

	// Step 2: Create the instance
	info, err := provider.CreateInstance(ctx, wm)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return info, nil
}

// updateStatusFromInstanceInfo updates WorkMachine status from InstanceInfo
func (r *WorkMachineReconciler) updateStatusFromInstanceInfo(
	wm *machinesv1.WorkMachine, info *types.InstanceInfo,
) {
	wm.Status.InstanceID = info.InstanceID
	wm.Status.PublicIP = info.PublicIP
	wm.Status.PrivateIP = info.PrivateIP
	wm.Status.Region = info.Region
	wm.Status.AvailabilityZone = info.AvailabilityZone
	wm.Status.SecurityGroupID = info.SecurityGroupID
	wm.Status.K3sJoinStatus = info.K3sJoinStatus
}

// mapInstanceStateToMachineState maps cloud provider InstanceState to MachineState
func mapInstanceStateToMachineState(state types.InstanceState) machinesv1.MachineState {
	switch state {
	case types.InstanceStatePending:
		return machinesv1.MachineStateStarting
	case types.InstanceStateRunning:
		return machinesv1.MachineStateRunning
	case types.InstanceStateStopping:
		return machinesv1.MachineStateStopping
	case types.InstanceStateStopped:
		return machinesv1.MachineStateStopped
	case types.InstanceStateTerminating, types.InstanceStateTerminated:
		return machinesv1.MachineStateStopped
	case types.InstanceStateError, types.InstanceStateNotFound:
		return machinesv1.MachineStateError
	default:
		return machinesv1.MachineStateError
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.K3sAgentToken != "" {
		return fmt.Errorf("reconciler.K3sAgentToken must be set")
	}

	var err error
	r.awsProvider, err := aws.NewProvider(context.TODO(), obj.Spec.AWSProvider, r.K3sAgentToken)

	return ctrl.NewControllerManagedBy(mgr).
		For(&machinesv1.WorkMachine{}).
		Complete(r)
}
