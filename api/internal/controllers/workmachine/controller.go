package workmachine

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	// NOTE: AWS SDK import commented out - using Job-based approach
	// "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/aws"
	// "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/types"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud/aws"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/kubectl"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
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

	YAMLClient kubectl.YAMLClient

	awsClient cloud.Provider
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
// func (r *WorkMachineReconciler) updateWorkMachineStatus(
// 	ctx context.Context,
// 	workMachine *machinesv1.WorkMachine,
// 	updateFunc func() error,
// 	logger *slog.Logger,
// ) error {
// 	return statusutil.UpdateStatusWithRetry(ctx, r.Client, workMachine, updateFunc, logger)
// }

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

// NOTE: AWS SDK provider removed - using Job-based approach instead
// getOrCreateAWSProvider is commented out as we're using Jobs now
/*
func (r *WorkMachineReconciler) getOrCreateAWSProvider(
	ctx context.Context, obj *machinesv1.WorkMachine,
) (*aws.Provider, error) {
	if obj.Spec.AWSProvider == nil {
		return nil, NewInvalidConfigurationError("aws", "AWS provider configuration is required")
	}
	// ... implementation removed ...
	return nil, nil
}
*/

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
			Name:     "setup-namespace",
			Title:    "Setup a kubernetes namespace for workmachine resources",
			OnCreate: r.createNamespace,
			OnDelete: r.deleteNamespace,
		},
		{
			Name:     "setup-package-manager-RBAC",
			Title:    "Setup RBAC resources for workmachine-node-manager",
			OnCreate: r.createPackageManagerRBAC,
			OnDelete: nil,
		},
		{
			Name:     "ensure-job-rbac",
			Title:    "Ensure RBAC resources for AWS Job runner",
			OnCreate: r.createJobRBAC,
			OnDelete: nil,
		},
		{
			Name:     "ensure-ssh-host-keys",
			Title:    "Ensure SSH host keys secret",
			OnCreate: r.createSSHHostKeysSecret,
			OnDelete: nil,
		},
		{
			Name:     "ensure-ssh-authorized-keys",
			Title:    "Ensure SSH authorized_keys ConfigMap",
			OnCreate: r.createSSHAuthorizedKeysConfig,
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
			Name:     "setup cloud machine",
			Title:    "Setup Cloud Machine",
			OnCreate: r.setupCloudMachine,
			OnDelete: r.cleanupCloudMachine,
		},
	})
}

// getK8sReconciliationSteps returns the reconciliation steps for K8s-based WorkMachines
// Step handlers for K8s-based WorkMachine reconciliation

// createNamespace ensures the target namespace exists with finalizer
func (r *WorkMachineReconciler) createNamespace(
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

// deleteNamespace handles namespace deletion when WorkMachine is being deleted
func (r *WorkMachineReconciler) deleteNamespace(
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

// createPackageManagerRBAC ensures RBAC resources for workmachine-node-manager
func (r *WorkMachineReconciler) createPackageManagerRBAC(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
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

// createJobRBAC ensures RBAC resources for AWS Job runner
func (r *WorkMachineReconciler) createJobRBAC(
	check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine,
) reconciler.StepResult {
	namespace := obj.Namespace

	// Create ServiceAccount for Jobs
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-job-runner",
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

	// Create Role with ConfigMap permissions
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-job-runner",
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
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
			Name:      "workmachine-job-runner",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "workmachine-job-runner",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "workmachine-job-runner",
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

// createSSHHostKeysSecret ensures the SSH host keys secret exists
func (r *WorkMachineReconciler) createSSHHostKeysSecret(
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

// createSSHAuthorizedKeysConfig ensures the SSH authorized_keys ConfigMap exists
func (r *WorkMachineReconciler) createSSHAuthorizedKeysConfig(
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

	// Render deployment from template
	b, err := templates.WorkMachineHostManagerDeployment.Render(map[string]any{
		"Namespace":       namespace,
		"WorkMachineName": obj.Name,
		"SSHUserUID":      SSHUserUID,
		"SSHUserGID":      SSHUserGID,
		"SSHUserName":     SSHUserName,
	})
	if err != nil {
		return check.Failed(fmt.Errorf("failed to render deployment template: %w", err))
	}

	if _, err := r.YAMLClient.ApplyYAML(check.Context(), b); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *WorkMachineReconciler) setupCloudMachine(check *reconciler.Check[*machinesv1.WorkMachine], obj *machinesv1.WorkMachine) reconciler.StepResult {
	switch obj.Spec.Provider {
	case machinesv1.AWS:
		{
			// TODO(claude): implement this
			return check.Passed()
		}
	default:
		return check.Failed(fmt.Errorf("unknown cloud provider %s", obj.Spec.Provider))
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.K3sAgentToken == "" {
		return fmt.Errorf("reconciler.K3sAgentToken must be set")
	}

	p, err := aws.NewProvider(context.TODO(), "", "")
	if err != nil {
		return err
	}

	r.awsClient = p

	builder := ctrl.NewControllerManagedBy(mgr)
	return builder.Complete(r)
}
