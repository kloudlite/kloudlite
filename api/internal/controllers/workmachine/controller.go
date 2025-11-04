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

	"github.com/codingconcepts/env"
	environmentV1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud/aws"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/kubectl"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	"golang.org/x/crypto/ssh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

type Env struct {
	KloudliteInstallationID string `env:"INSTALLATION_KEY" required:"true"`

	K3sVersion    string `env:"K3S_VERSION" required:"true"`
	K3sServerURL  string `env:"K3S_SERVER_URL" required:"true"`
	K3sAgentToken string `env:"K3S_AGENT_TOKEN" required:"true"`

	CloudProvider v1.CloudProvider `env:"CLOUD_PROVIDER" required:"true"`
}

type awsProviderEnv struct {
	AWS_AMI_ID            string `env:"AWS_AMI_ID" required:"true"`
	AWS_VPC_ID            string `env:"AWS_VPC_ID" required:"true"`
	AWS_SECURITY_GROUP_ID string `env:"AWS_SECURITY_GROUP_ID" required:"true"`
	AWS_REGION            string `env:"AWS_REGION" required:"true"`
}

// WorkMachineReconciler reconciles a WorkMachine object
type WorkMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	YAMLClient kubectl.YAMLClient

	// filled post initialization
	env              Env
	cloudProviderAPI cloud.Provider
}

const WorkMachineFinalizerName = "workmachine.machines.kloudlite.io/cleanup"

// SSH Configuration Constants
const (
	// SSHUserName is the username for the SSH server
	SSHUserName = "kloudlite"
)

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

// Reconcile handles WorkMachine CR reconciliation
func (r *WorkMachineReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// Use operator toolkit pattern for both K8s and cloud provider WorkMachines
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.WorkMachine{})
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.WorkMachine]{
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
			Name:     "setup cloud machine",
			Title:    "Setup Cloud Machine",
			OnCreate: r.setupCloudMachine,
			OnDelete: r.cleanupCloudMachine,
		},
		{
			Name:     "ensure-deployment",
			Title:    "Ensure workmachine-host-manager deployment",
			OnCreate: r.ensurePackageManagerDeploymentStep,
			OnDelete: nil,
		},
	})
}

// getK8sReconciliationSteps returns the reconciliation steps for K8s-based WorkMachines
// Step handlers for K8s-based WorkMachine reconciliation

// createNamespace ensures the target namespace exists with finalizer
func (r *WorkMachineReconciler) createNamespace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, ns, func() error {
		ns.SetLabels(fn.MapMerge(ns.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		if !fn.IsOwner(ns, obj) {
			ns.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		if !controllerutil.ContainsFinalizer(ns, WorkMachineFinalizerName) {
			ns.SetFinalizers(append(ns.GetFinalizers(), WorkMachineFinalizerName))
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// deleteNamespace handles namespace deletion when WorkMachine is being deleted
func (r *WorkMachineReconciler) deleteNamespace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespaceName := obj.Spec.TargetNamespace

	// Check for active Workspaces in the target namespace
	var envList environmentV1.EnvironmentList
	if err := r.List(check.Context(), &envList); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}
	}

	// Check for active Workspaces in the target namespace
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList, client.InNamespace(namespaceName)); err != nil {
		if !apiErrors.IsNotFound(err) {
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
			if controllerutil.RemoveFinalizer(namespace, WorkMachineFinalizerName) {
				if err := r.Update(check.Context(), namespace); err != nil {
					return check.Failed(err)
				}
			}
			return check.UpdateMsg("Namespace is being deleted, waiting for completion")
		}

		// Delete the namespace
		if err := r.Delete(check.Context(), namespace); err != nil && !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		return check.UpdateMsg("Namespace deletion initiated, waiting for completion")
	}

	if !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Namespace is deleted
	return check.Passed()
}

// createPackageManagerRBAC ensures RBAC resources for workmachine-node-manager
func (r *WorkMachineReconciler) createPackageManagerRBAC(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
	}

	if err := r.Get(check.Context(), client.ObjectKey{Name: sa.Name, Namespace: namespace}, sa); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Create service account
		if err := r.Create(check.Context(), sa); err != nil && !apiErrors.IsAlreadyExists(err) {
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
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
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
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Create role
		if err := r.Create(check.Context(), role); err != nil && !apiErrors.IsAlreadyExists(err) {
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
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Create role binding
		if err := r.Create(check.Context(), rb); err != nil && !apiErrors.IsAlreadyExists(err) {
			return check.Failed(err)
		}
	}

	return check.Passed()
}

// createSSHHostKeysSecret ensures the SSH host keys secret exists
func (r *WorkMachineReconciler) createSSHHostKeysSecret(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace
	secretName := "ssh-host-keys"
	secret := &corev1.Secret{}
	err := r.Get(check.Context(), client.ObjectKey{Name: secretName, Namespace: namespace}, secret)

	if err == nil {
		// Secret already exists
		return check.Passed()
	}

	if !apiErrors.IsNotFound(err) {
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
func (r *WorkMachineReconciler) createSSHAuthorizedKeysConfig(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
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
func (r *WorkMachineReconciler) ensureSSHDConfigMapStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Skip for cloud provider WorkMachines
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
func (r *WorkMachineReconciler) ensureWorkspaceSSHDConfigMapStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
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
func (r *WorkMachineReconciler) ensurePackageManagerDeploymentStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace
	deploymentName := "workmachine-host-manager"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, deployment, func() error {
		// Render deployment from template
		b, err := templates.WorkMachineHostManagerDeployment.Render(
			templates.WorkspaceHostManagerValues{
				Namespace:       namespace,
				WorkMachineName: obj.Name,
				SSHUsername:     SSHUserName,
				NodeSelector:    obj.Status.NodeLabels,
				Tolerations:     obj.Status.PodTolerations,
			},
		)
		if err != nil {
			return errors.Wrap("failed to render workmachine host manager deployment template", err)
		}

		deployment.SetLabels(fn.MapMerge(deployment.GetLabels(), map[string]string{
			"app":                       deployment.Name,
			"kloudlite.io/package-mgmt": "true",
			"kloudlite.io/workmachine":  obj.Name,
		}))

		if err := yaml.Unmarshal(b, &deployment); err != nil {
			return errors.Wrap("failed to unmarshal into deployment", err)
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// deployment := &appsv1.Deployment{}
	// err := r.Get(check.Context(), client.ObjectKey{Name: deploymentName, Namespace: namespace}, deployment)
	//
	// if err == nil {
	// 	// Deployment already exists
	// 	return check.Passed()
	// }
	//
	// if !apiErrors.IsNotFound(err) {
	// 	return check.Errored(err)
	// }
	//
	// if _, err := r.YAMLClient.ApplyYAML(check.Context(), b); err != nil {
	// 	return check.Failed(err)
	// }
	//
	return check.Passed()
}

func (r *WorkMachineReconciler) setupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Status.MachineID == "" {
		mi, err := r.cloudProviderAPI.CreateMachine(check.Context(), obj)
		if err != nil {
			return check.Failed(err)
		}

		obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
		obj.Status.MachineInfo = *mi
		return check.UpdateMsg("created cloud machine").RequeueAfter(2 * time.Second)
	}

	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(check.Context(), obj.Status.MachineID)
	if err != nil {
		return check.Failed(err)
	}

	// Handle desired state transitions
	currentState := machineInfo.State

	// Start machine if desired state is running but machine is stopped
	if obj.Spec.State == v1.MachineStateRunning && currentState == v1.MachineStateStopped {
		if err := r.cloudProviderAPI.StartMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to start machine: %w", err))
		}
		obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
		return check.UpdateMsg("Starting machine").RequeueAfter(10 * time.Second)
	}

	// Stop machine if desired state is stopped but machine is running
	if obj.Spec.State == v1.MachineStateStopped && currentState == v1.MachineStateRunning {
		if err := r.cloudProviderAPI.StopMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to stop machine: %w", err))
		}

		obj.Status.StoppedAt = &metav1.Time{Time: time.Now()}
		return check.UpdateMsg("Stopping Machine").RequeueAfter(10 * time.Second)
	}

	if currentState != obj.Spec.State {
		return check.UpdateMsg("waiting for machine status to change").RequeueAfter(5 * time.Second)
	}

	obj.Status.State = machineInfo.State

	specVolume := fn.ValueOf(obj.Spec.VolumeSize)

	if specVolume > obj.Status.RootVolumeSize {
		check.Logger().Info("increasing volume size", "from", obj.Status.RootVolumeSize, "to", obj.Spec.VolumeSize)

		if err := r.cloudProviderAPI.IncreaseVolumeSize(check.Context(), obj.Status.MachineID, specVolume); err != nil {
			return check.Failed(errors.Wrap("failed to increase volume size", err))
		}

		obj.Status.RootVolumeSize = specVolume
		if err := r.cloudProviderAPI.RebootMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Errored(errors.Wrap(fmt.Sprintf("failed to reboot machine(ID: %s)", obj.Status.MachineID), err))
		}

		return check.UpdateMsg("waiting for volume size to be increased").RequeueAfter(10 * time.Second)
	}

	// Update status with current machine info
	obj.Status.State = machineInfo.State
	obj.Status.Message = machineInfo.Message
	obj.Status.PublicIP = machineInfo.PublicIP
	obj.Status.PrivateIP = machineInfo.PrivateIP
	obj.Status.RootVolumeSize = specVolume

	// Fetch node information to populate nodeLabels and nodeTaints
	node := &corev1.Node{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}
		// Node not yet registered, that's ok
		return check.UpdateMsg("Waiting for node to register").RequeueAfter(10 * time.Second)
	}

	// Filter and update only kloudlite.io/ prefixed labels
	obj.Status.NodeLabels = fn.MapFilter(node.Labels, func(k, v string) bool {
		return strings.HasPrefix(k, "kloudlite.io/")
	})

	// Filter and update only kloudlite.io/ prefixed taints
	var podTolerations []corev1.Toleration
	for _, taint := range node.Spec.Taints {
		if strings.HasPrefix(taint.Key, "kloudlite.io/") {
			podTolerations = append(podTolerations, corev1.Toleration{
				Key:      taint.Key,
				Operator: corev1.TolerationOpEqual,
				Value:    taint.Value,
				Effect:   taint.Effect,
			})
		}
	}
	obj.Status.PodTolerations = podTolerations

	return check.Passed()
}

func (r *WorkMachineReconciler) cleanupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	if err := r.cloudProviderAPI.DeleteMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to delete AWS machine: %w", err))
	}

	obj.Status.MachineInfo = v1.MachineInfo{}
	return check.Passed()
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := env.Set(&r.env); err != nil {
		return errors.Wrap("failed to load env vars", err)
	}

	switch r.env.CloudProvider {
	case v1.AWS:
		{

			var awsEnv awsProviderEnv
			if err := env.Set(&awsEnv); err != nil {
				return errors.Wrap("failed to load env vars", err)
			}

			ctx, cf := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cf()
			p, err := aws.NewProvider(ctx, aws.ProviderArgs{
				AMI:             awsEnv.AWS_AMI_ID,
				Region:          awsEnv.AWS_REGION,
				VPC:             awsEnv.AWS_VPC_ID,
				SecurityGroupID: awsEnv.AWS_SECURITY_GROUP_ID,
				ResourceTags: []aws.Tag{
					{
						Key:   "kloudlite.io/installation-id",
						Value: r.env.KloudliteInstallationID,
					},
				},

				K3sVersion: r.env.K3sVersion,
				K3sURL:     r.env.K3sServerURL,
				K3sToken:   r.env.K3sAgentToken,
			})
			if err != nil {
				return errors.Wrap("failed to create aws provider client", err)
			}

			if err := p.ValidatePermissions(ctx); err != nil {
				return err
			}

			r.cloudProviderAPI = p
		}
	default:
		{
			return errors.New(fmt.Sprintf("unsupported cloud provider (%s)", r.env.CloudProvider))
		}
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.WorkMachine{}).Named("workmachine")
	builder.Owns(&corev1.Namespace{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor("workmachine")))
	return builder.Complete(r)
}
