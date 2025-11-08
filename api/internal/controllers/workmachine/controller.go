package workmachine

import (
	"context"
	"fmt"
	"time"

	"github.com/codingconcepts/env"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud/aws"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/kubectl"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
const hostManagerNamespace = "kloudlite-hostmanager"

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
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.WorkMachine{})
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.WorkMachine]{
		{
			Name:     "setup-namespace",
			Title:    "Setup a kubernetes namespace for workmachine resources",
			OnCreate: r.createNamespace,
			OnDelete: r.deleteNamespace,
		},
		{
			Name:     "setup-host-manager-RBAC",
			Title:    "Setup RBAC resources for workmachine-node-manager (host manager)",
			OnCreate: r.createHostManagerRBAC,
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
			Name:     "handle-machine-type-change",
			Title:    "Handle machine type changes",
			OnCreate: r.handleMachineTypeChange,
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

// createWorkspaceRBAC is a placeholder step for workspace RBAC setup
// Actual workspace-specific RBAC is created by the Workspace controller
func (r *WorkMachineReconciler) createWorkspaceRBAC(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
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

	// Watch for Nodes to trigger reconciliation when node joins/updates
	builder.Watches(
		&corev1.Node{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			node, ok := obj.(*corev1.Node)
			if !ok {
				return nil
			}

			// Node name matches WorkMachine name
			// Trigger reconciliation to update WorkMachine status
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
