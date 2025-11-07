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
	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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

	// Check if DomainRequest exists, and if not, clear the check status to force re-run
	// DomainRequest is cluster-scoped
	obj := req.Object
	domainRequest := &domainrequestv1.DomainRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, domainRequest); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// DomainRequest doesn't exist, clear the check status for "setup domain request" step
			if obj.Status.Checks != nil {
				if _, exists := obj.Status.Checks["create/setup domain request"]; exists {
					delete(obj.Status.Checks, "create/setup domain request")
					// Update the status on the API server
					if updateErr := r.Status().Update(ctx, obj); updateErr != nil {
						return reconcile.Result{}, updateErr
					}
				}
			}
		}
	}

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
			Name:     "setup-workspace-RBAC",
			Title:    "Setup RBAC resources for workspace pods",
			OnCreate: r.createWorkspaceRBAC,
			OnDelete: nil,
		},
		{
			Name:     "ensure-ssh-host-keys",
			Title:    "Ensure SSH host keys secret",
			OnCreate: r.createSSHHostKeysSecret,
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
			Name:     "setup domain request",
			Title:    "Sets up Domain Request Settings for the workmachine",
			OnCreate: r.syncDomainRequest,
			OnDelete: r.deleteDomainRequest,
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

	// Delete workspace pods directly (bypass finalizers for faster cleanup)
	// When WorkMachine is being deleted, we don't need graceful workspace cleanup
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}
	}

	// Filter workspaces owned by this WorkMachine
	var ownedWorkspaces []workspacev1.Workspace
	for _, ws := range workspaceList.Items {
		if ws.Spec.WorkmachineName == obj.Name {
			ownedWorkspaces = append(ownedWorkspaces, ws)
		}
	}

	// Directly delete workspace pods to speed up cleanup
	for _, ws := range ownedWorkspaces {
		// Delete the workspace pod directly
		podName := fmt.Sprintf("workspace-%s", ws.Name)
		pod := &corev1.Pod{}
		err := r.Get(check.Context(), client.ObjectKey{Name: podName, Namespace: namespaceName}, pod)
		if err == nil {
			// Pod exists, delete it
			if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete workspace pod %s: %w", podName, err))
			}
		} else if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Remove finalizer from workspace to allow it to be deleted immediately
		if ws.DeletionTimestamp == nil {
			// Workspace not being deleted yet, delete it
			if err := r.Delete(check.Context(), &ws); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete workspace %s: %w", ws.Name, err))
			}
		} else {
			// Workspace is being deleted but stuck on finalizer, remove it
			if controllerutil.ContainsFinalizer(&ws, "workspaces.kloudlite.io/finalizer") {
				controllerutil.RemoveFinalizer(&ws, "workspaces.kloudlite.io/finalizer")
				if err := r.Update(check.Context(), &ws); err != nil && !apiErrors.IsNotFound(err) {
					return check.Failed(fmt.Errorf("failed to remove finalizer from workspace %s: %w", ws.Name, err))
				}
			}
		}
	}

	// Delete environment namespaces directly (bypass finalizers for faster cleanup)
	for _, env := range envList.Items {
		if env.Spec.WorkMachineName == obj.Name {
			// Delete the environment namespace directly if it exists
			if env.Spec.TargetNamespace != "" {
				envNs := &corev1.Namespace{}
				err := r.Get(check.Context(), client.ObjectKey{Name: env.Spec.TargetNamespace}, envNs)
				if err == nil {
					// Namespace exists, delete it
					if err := r.Delete(check.Context(), envNs); err != nil && !apiErrors.IsNotFound(err) {
						return check.Failed(fmt.Errorf("failed to delete environment namespace %s: %w", env.Spec.TargetNamespace, err))
					}
				} else if !apiErrors.IsNotFound(err) {
					return check.Errored(err)
				}
			}

			// Remove finalizer from environment to allow it to be deleted immediately
			if env.DeletionTimestamp == nil {
				// Environment not being deleted yet, delete it
				if err := r.Delete(check.Context(), &env); err != nil && !apiErrors.IsNotFound(err) {
					return check.Failed(fmt.Errorf("failed to delete environment %s: %w", env.Name, err))
				}
			} else {
				// Environment is being deleted but stuck on finalizer, remove it
				if controllerutil.ContainsFinalizer(&env, "environments.kloudlite.io/finalizer") {
					controllerutil.RemoveFinalizer(&env, "environments.kloudlite.io/finalizer")
					if err := r.Update(check.Context(), &env); err != nil && !apiErrors.IsNotFound(err) {
						return check.Failed(fmt.Errorf("failed to remove finalizer from environment %s: %w", env.Name, err))
					}
				}
			}
		}
	}

	// Delete host-manager pod and service in finalizer
	// Cluster-scoped WorkMachine cannot own namespaced resources via owner references
	hostManagerName := fmt.Sprintf("hm-%s", obj.Name)

	// Delete pod
	pod := &corev1.Pod{}
	if err := r.Get(check.Context(), client.ObjectKey{
		Name:      hostManagerName,
		Namespace: hostManagerNamespace,
	}, pod); err == nil {
		if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete host manager pod: %w", err))
		}
	} else if !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Delete service
	service := &corev1.Service{}
	if err := r.Get(check.Context(), client.ObjectKey{
		Name:      hostManagerName,
		Namespace: hostManagerNamespace,
	}, service); err == nil {
		if err := r.Delete(check.Context(), service); err != nil && !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete host manager service: %w", err))
		}
	} else if !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Proceed with namespace deletion
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
	// Create RBAC in both namespaces:
	// 1. WorkMachine target namespace for workspace pods
	// 2. Shared hostmanager namespace for host manager deployment
	namespaces := []string{obj.Spec.TargetNamespace, hostManagerNamespace}

	for _, namespace := range namespaces {
		if err := r.createRBACInNamespace(check.Context(), namespace); err != nil {
			return check.Failed(err)
		}
	}

	// Create ClusterRole and ClusterRoleBinding for cluster-scoped resources
	// This allows workmachine-node-manager to access Workspaces and PackageRequests
	if err := r.createClusterRBAC(check.Context()); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// createRBACInNamespace creates RBAC resources for package management in the given namespace
func (r *WorkMachineReconciler) createRBACInNamespace(ctx context.Context, namespace string) error {
	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
	}

	if err := r.Get(ctx, client.ObjectKey{Name: sa.Name, Namespace: namespace}, sa); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		// Create service account
		if err := r.Create(ctx, sa); err != nil && !apiErrors.IsAlreadyExists(err) {
			return err
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
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	existingRole := &rbacv1.Role{}
	if err := r.Get(ctx, client.ObjectKey{Name: role.Name, Namespace: namespace}, existingRole); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		// Create role
		if err := r.Create(ctx, role); err != nil && !apiErrors.IsAlreadyExists(err) {
			return err
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
	if err := r.Get(ctx, client.ObjectKey{Name: rb.Name, Namespace: namespace}, existingRB); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		// Create role binding
		if err := r.Create(ctx, rb); err != nil && !apiErrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

// createClusterRBAC creates ClusterRole and ClusterRoleBinding for workmachine-node-manager
// to access cluster-scoped resources like Workspaces and PackageRequests
func (r *WorkMachineReconciler) createClusterRBAC(ctx context.Context) error {
	// Create ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workmachine-node-manager",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"workspaces"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"workspaces/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
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
		},
	}

	existingClusterRole := &rbacv1.ClusterRole{}
	if err := r.Get(ctx, client.ObjectKey{Name: clusterRole.Name}, existingClusterRole); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		// Create cluster role
		if err := r.Create(ctx, clusterRole); err != nil && !apiErrors.IsAlreadyExists(err) {
			return err
		}
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workmachine-node-manager",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "workmachine-node-manager",
				Namespace: hostManagerNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "workmachine-node-manager",
		},
	}

	existingCRB := &rbacv1.ClusterRoleBinding{}
	if err := r.Get(ctx, client.ObjectKey{Name: clusterRoleBinding.Name}, existingCRB); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		// Create cluster role binding
		if err := r.Create(ctx, clusterRoleBinding); err != nil && !apiErrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

// createWorkspaceRBAC ensures RBAC resources for workspace pods
func (r *WorkMachineReconciler) createWorkspaceRBAC(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	// Create ServiceAccount for workspace pods
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-user",
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

	// Create Role for future namespace-scoped permissions
	// Note: PackageRequest permissions are now in workspace-specific ClusterRoles
	// Note: Workspace permissions are now in workspace-specific ClusterRoles
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-user",
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, role, func() error {
		// Empty rules for now - reserved for future namespace-scoped resources
		role.Rules = []rbacv1.PolicyRule{}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create RoleBinding
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-user-binding",
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, rb, func() error {
		rb.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "workspace-user",
				Namespace: namespace,
			},
		}

		rb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "workspace-user",
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create ClusterRole for cluster-wide Environment access
	// Note: Workspace permissions are now managed per-workspace by the Workspace controller
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workspace-user-cluster-access",
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRole, func() error {
		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments/status"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "secrets", "services"},
				Verbs:     []string{"get", "list", "watch"},
			},
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create ClusterRoleBinding for this WorkMachine's namespace
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("workspace-user-%s", namespace),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "workspace-user",
				Namespace: namespace,
			},
		}

		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "workspace-user-cluster-access",
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// createSSHHostKeysSecret ensures the SSH host keys secret exists
func (r *WorkMachineReconciler) createSSHHostKeysSecret(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := hostManagerNamespace
	secretName := fmt.Sprintf("ssh-host-keys-%s", obj.Name)

	// Generate RSA 2048-bit key (only used if Secret doesn't exist)
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

const hostManagerNamespace = "kloudlite-hostmanager"

// ensurePackageManagerDeploymentStep ensures the workmachine-host-manager pod exists
// Pod will be recreated by the controller if it crashes
func (r *WorkMachineReconciler) ensurePackageManagerDeploymentStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := hostManagerNamespace
	// Use unique name per WorkMachine since all host managers share the same namespace
	hostManagerName := fmt.Sprintf("hm-%s", obj.Name)

	pod := &corev1.Pod{}
	err := r.Get(check.Context(), client.ObjectKey{Name: hostManagerName, Namespace: hostManagerNamespace}, pod)

	if err != nil && !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// If pod exists, check its status
	if err == nil {
		// Check if pod is in a failed/completed state and needs recreation
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
			// Delete the failed/completed pod so it will be recreated
			if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete failed/completed pod: %w", err))
			}
			// Pod deleted, will be recreated in next reconcile
			return check.UpdateMsg("Recreating failed pod").RequeueAfter(2 * time.Second)
		}

		// Check if pod is ready (all containers ready)
		if pod.Status.Phase != corev1.PodRunning {
			return check.UpdateMsg(fmt.Sprintf("Waiting for host-manager pod to be running (current: %s)", pod.Status.Phase)).RequeueAfter(5 * time.Second)
		}

		// Check all containers are ready
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				return check.UpdateMsg(fmt.Sprintf("Waiting for container %s to be ready", containerStatus.Name)).RequeueAfter(5 * time.Second)
			}
		}

		// Pod is running and all containers are ready
		return check.Passed()
	}

	// Pod doesn't exist, create it
	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: hostManagerNamespace,
		},
	}

	// Render pod from template
	b, err := templates.WorkMachineHostManagerPod.Render(
		templates.WorkspaceHostManagerValues{
			Namespace:       namespace,
			WorkMachineName: obj.Name,
			TargetNamespace: obj.Spec.TargetNamespace,
			SSHUsername:     SSHUserName,
		},
	)
	if err != nil {
		return check.Failed(errors.Wrap("failed to render workmachine host manager pod template", err))
	}

	if err := yaml.Unmarshal(b, &pod); err != nil {
		return check.Failed(errors.Wrap("failed to unmarshal into pod", err))
	}

	// Set labels
	pod.SetLabels(fn.MapMerge(pod.GetLabels(), map[string]string{
		"app":                       hostManagerName,
		"kloudlite.io/package-mgmt": "true",
		"kloudlite.io/workmachine":  obj.Name,
	}))

	if err := r.Create(check.Context(), pod); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return check.Passed()
		}
		return check.Failed(err)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: hostManagerNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svc, func() error {
		svc.SetLabels(fn.MapMerge(svc.GetLabels(), map[string]string{
			"app":                       hostManagerName,
			"kloudlite.io/package-mgmt": "true",
			"kloudlite.io/workmachine":  obj.Name,
		}))

		svc.Spec.Selector = map[string]string{
			"app": hostManagerName,
		}

		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "ssh",
				Protocol:   corev1.ProtocolTCP,
				Port:       22,
				TargetPort: intstr.FromInt32(2222),
			},
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// syncDomainRequest creates or updates the DomainRequest with the latest WorkMachine IP
// This runs on every reconcile to keep the DomainRequest in sync
func (r *WorkMachineReconciler) syncDomainRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// First check if DomainRequest exists and has the correct IP
	domainRequest := &domainrequestv1.DomainRequest{}
	err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, domainRequest)

	if err != nil && !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Always update/create DomainRequest to sync workspace routes
	// CreateOrUpdate will handle both creation and updates efficiently
	return r.createDomainRequest(check, obj)
}

func (r *WorkMachineReconciler) createDomainRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Fetch subdomain from installation DomainRequest instead of env var
	// DomainRequest is cluster-scoped
	installationDR := &domainrequestv1.DomainRequest{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: "installation-domain"}, installationDR); err != nil {
		return check.Errored(fmt.Errorf("failed to get installation DomainRequest: %w", err)).RequeueAfter(5 * time.Second)
	}

	subDomain := installationDR.Status.Subdomain
	if subDomain == "" {
		return check.Errored(fmt.Errorf("installation subdomain not yet configured")).RequeueAfter(5 * time.Second)
	}

	// List all cluster-scoped workspaces and filter by WorkmachineName
	var wsList workspacev1.WorkspaceList
	if err := r.List(check.Context(), &wsList); err != nil {
		return check.Failed(err)
	}

	// Filter workspaces owned by this WorkMachine
	var ownedWorkspaces []workspacev1.Workspace
	for _, ws := range wsList.Items {
		if ws.Spec.WorkmachineName == obj.Name {
			ownedWorkspaces = append(ownedWorkspaces, ws)
		}
	}

	var domainRoutes []domainrequestv1.DomainRoute
	for _, ws := range ownedWorkspaces {
		serviceName := "workspace-" + ws.Name + "-headless"
		domainRoutes = append(domainRoutes,
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("vscode-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      8080,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("tty-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7681,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("claude-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7682,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("opencode-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7683,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("codex-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7684,
			},
		)
	}

	// DomainRequest is cluster-scoped, with workloads running in shared workloadNamespace
	domainRequest := &domainrequestv1.DomainRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: obj.Name,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, domainRequest, func() error {
		// Set WorkMachine as owner for cascading deletion
		// MUST be set inside the mutate function to ensure it's preserved on updates
		blockOwnerDeletion := false
		controller := true
		ownerRef := metav1.OwnerReference{
			APIVersion:         obj.APIVersion,
			Kind:               obj.Kind,
			Name:               obj.Name,
			UID:                obj.UID,
			Controller:         &controller,
			BlockOwnerDeletion: &blockOwnerDeletion,
		}
		domainRequest.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

		hostManagerName := fmt.Sprintf("hm-%s", obj.Name)
		domainRequest.Spec = domainrequestv1.DomainRequestSpec{
			NodeName:          obj.Name,
			WorkloadNamespace: "kloudlite-ingress", // Shared namespace for all DomainRequest workloads
			IPAddress:         obj.Status.PublicIP,
			CertificateScope:  "workmachine",
			OriginCertificateHostnames: []string{
				fmt.Sprintf("%s.%s", obj.Name, subDomain),
				fmt.Sprintf("*.%s.%s", obj.Name, subDomain),
			},
			SSHBackend: &domainrequestv1.IngressBackendConfig{
				ServiceName:      hostManagerName,
				ServiceNamespace: hostManagerNamespace,
				ServicePort:      22,
			},
			DomainRoutes: domainRoutes,
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Wait for DomainRequest to be ready before marking this step as complete
	if domainRequest.Status.State != "Ready" {
		return check.UpdateMsg(fmt.Sprintf("Waiting for DomainRequest to be ready (current state: %s)", domainRequest.Status.State)).RequeueAfter(5 * time.Second)
	}

	// Store the DNS host in WorkMachine status
	obj.Status.DNSHost = domainRequest.Status.Domain

	return check.Passed()
}

func (r *WorkMachineReconciler) deleteDomainRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// DomainRequest is cluster-scoped
	domainRequest := &domainrequestv1.DomainRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: obj.Name,
		},
	}

	if err := r.Delete(check.Context(), domainRequest); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
		// Already deleted, that's fine
	}

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
		// Set state to starting since node hasn't joined yet
		if obj.Spec.State == v1.MachineStateRunning {
			obj.Status.State = v1.MachineStateStarting
			obj.Status.Message = "Cloud machine created, waiting for node to join"
		}
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
		obj.Status.State = v1.MachineStateStarting
		obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
		return check.UpdateMsg("Starting machine").RequeueAfter(10 * time.Second)
	}

	// Stop machine if desired state is stopped but machine is running
	if obj.Spec.State == v1.MachineStateStopped && currentState == v1.MachineStateRunning {
		if err := r.cloudProviderAPI.StopMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to stop machine: %w", err))
		}

		obj.Status.State = v1.MachineStateStopping
		obj.Status.StoppedAt = &metav1.Time{Time: time.Now()}
		return check.UpdateMsg("Stopping Machine").RequeueAfter(10 * time.Second)
	}

	// Check if machine state matches desired state (but we'll verify node readiness below)
	if currentState != obj.Spec.State {
		// Machine is transitioning
		return check.UpdateMsg("waiting for machine status to change").RequeueAfter(5 * time.Second)
	}

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
	obj.Status.PublicIP = machineInfo.PublicIP
	obj.Status.PrivateIP = machineInfo.PrivateIP
	obj.Status.RootVolumeSize = specVolume
	obj.Status.Message = machineInfo.Message

	// Check if node has joined the cluster before marking as running
	if machineInfo.State == v1.MachineStateRunning {
		node := &corev1.Node{}
		if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
			if apiErrors.IsNotFound(err) {
				// Node hasn't joined yet, keep state as "starting"
				obj.Status.State = v1.MachineStateStarting
				obj.Status.Message = "Waiting for node to join cluster"
				return check.UpdateMsg("waiting for node to join cluster").RequeueAfter(10 * time.Second)
			}
			return check.Failed(fmt.Errorf("failed to get node: %w", err))
		}

		// Check if node is ready
		nodeReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				nodeReady = true
				break
			}
		}

		if !nodeReady {
			obj.Status.State = v1.MachineStateStarting
			obj.Status.Message = "Node joined, waiting for node to be ready"
			return check.UpdateMsg("waiting for node to be ready").RequeueAfter(5 * time.Second)
		}

		// Node is ready, mark as running
		obj.Status.State = v1.MachineStateRunning
		obj.Status.Message = "Node is ready"
	} else {
		// For other states (stopped, stopping, etc.), use the cloud provider state directly
		obj.Status.State = machineInfo.State
	}

	return check.Passed()
}

func (r *WorkMachineReconciler) cleanupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	// Step 1: Add NoExecute taint to the node to evict pods
	node := &corev1.Node{}
	nodeExists := false
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if !apiErrors.IsNotFound(err) {
			check.Logger().Warn("failed to get node for tainting", "error", err)
		}
		// Node doesn't exist or failed to get, skip tainting
	} else {
		nodeExists = true
		// Add NoExecute taint if not already present
		taintExists := false
		for _, taint := range node.Spec.Taints {
			if taint.Key == "kloudlite.io/workmachine-deleting" && taint.Effect == corev1.TaintEffectNoExecute {
				taintExists = true
				break
			}
		}

		if !taintExists {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
				Key:    "kloudlite.io/workmachine-deleting",
				Value:  "true",
				Effect: corev1.TaintEffectNoExecute,
			})
			if err := r.Update(check.Context(), node); err != nil {
				check.Logger().Warn("failed to add NoExecute taint to node", "error", err)
			} else {
				check.Logger().Info("added NoExecute taint to node, waiting for pod eviction")
				return check.UpdateMsg("Added NoExecute taint to node").RequeueAfter(2 * time.Second)
			}
		}
	}

	// Step 2: Force delete any remaining pods on this node
	podList := &corev1.PodList{}
	if err := r.List(check.Context(), podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
		check.Logger().Warn("failed to list pods on node", "error", err)
	} else if len(podList.Items) > 0 {
		gracePeriod := int64(0)
		deleteOptions := &client.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			check.Logger().Info("force deleting pod", "pod", pod.Name, "namespace", pod.Namespace)
			if err := r.Delete(check.Context(), pod, deleteOptions); err != nil && !apiErrors.IsNotFound(err) {
				check.Logger().Warn("failed to force delete pod", "pod", pod.Name, "namespace", pod.Namespace, "error", err)
			}
		}
		// Wait for pods to be deleted
		return check.UpdateMsg("Waiting for pods to be deleted").RequeueAfter(2 * time.Second)
	}

	// Step 3: Delete the Kubernetes Node object (only if it exists)
	if nodeExists {
		if err := r.Delete(check.Context(), node); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete Kubernetes node: %w", err))
			}
		}
	}

	// Step 4: Delete the cloud machine (EC2 instance)
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

	// Watch for workspaces and trigger reconciliation of their owning WorkMachine
	builder.Watches(
		&workspacev1.Workspace{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			workspace, ok := obj.(*workspacev1.Workspace)
			if !ok {
				return nil
			}

			// Use the WorkmachineName directly from the workspace spec
			if workspace.Spec.WorkmachineName == "" {
				return nil
			}

			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: workspace.Spec.WorkmachineName}},
			}
		}),
	)

	// Watch for DomainRequests and trigger reconciliation to recreate if deleted
	builder.Watches(
		&domainrequestv1.DomainRequest{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			domainRequest, ok := obj.(*domainrequestv1.DomainRequest)
			if !ok {
				return nil
			}

			// DomainRequest name matches WorkMachine name
			// Trigger reconciliation of the WorkMachine to recreate DomainRequest if needed
			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: domainRequest.Name}},
			}
		}),
	)

	// Watch for Nodes to trigger reconciliation when node joins/updates
	builder.Watches(
		&corev1.Node{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			node, ok := obj.(*corev1.Node)
			if !ok {
				return nil
			}

			// Node name matches WorkMachine name
			// Trigger reconciliation to update DomainRequest with node IP
			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: node.Name}},
			}
		}),
	)

	// Watch for host-manager Pods to recreate them if they crash
	builder.Watches(
		&corev1.Pod{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return nil
			}

			// Only watch pods in the host-manager namespace
			if pod.Namespace != hostManagerNamespace {
				return nil
			}

			// Get WorkMachine name from pod label
			workmachineName, exists := pod.Labels["kloudlite.io/workmachine"]
			if !exists {
				return nil
			}

			// Trigger reconciliation to check and recreate pod if needed
			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: workmachineName}},
			}
		}),
	)

	// Add indexer for pod.spec.nodeName to efficiently query pods by node name
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName", func(obj client.Object) []string {
		pod := obj.(*corev1.Pod)
		return []string{pod.Spec.NodeName}
	}); err != nil {
		return errors.Wrap("failed to setup field indexer for pod.spec.nodeName", err)
	}

	return builder.Complete(r)
}
