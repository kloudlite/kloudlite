package workmachine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/codingconcepts/env"
	wireguarddevicev1 "github.com/kloudlite/kloudlite/api/internal/controllers/wireguarddevice/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud/aws"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/kubectl"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	appsv1 "k8s.io/api/apps/v1"
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
)

type Env struct {
	KloudliteInstallationID string `env:"INSTALLATION_KEY" required:"true"`

	K3sVersion    string `env:"K3S_VERSION" required:"true"`
	K3sServerURL  string `env:"K3S_SERVER_URL" required:"true"`
	K3sAgentToken string `env:"K3S_AGENT_TOKEN" required:"true"`

	CloudProvider v1.CloudProvider `env:"CLOUD_PROVIDER" required:"true"`

	HostManagerImage string `env:"HOST_MANAGER_IMAGE" required:"true"`
}

type awsProviderEnv struct {
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

const (
	WorkMachineFinalizerName = "workmachine.machines.kloudlite.io/cleanup"
	hostManagerNamespace     = "kloudlite-hostmanager"
)

// SSH Configuration Constants
const (
	// SSHUserName is the username for the SSH server
	SSHUserName = "kloudlite"

	wireguardTunnelImage     = "ghcr.io/kloudlite/kloudlite/wireguard-server:latest"
	wmIngressControllerImage = "ghcr.io/kloudlite/kloudlite/wm-ingress-controller:development"
)

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
			Name:     "ensure-wm-ingress-controller",
			Title:    "Ensure Workmachine Ingress Controller",
			OnCreate: r.ensureWorkmachineIngressController,
			OnDelete: nil,
		},
		{
			Name:  "when-running/ensure-tunnel-server",
			Title: "Ensure tunnel server is running",
			// ShouldRun: func(obj *v1.WorkMachine) bool {
			// 	return obj.Spec.State == v1.MachineStateRunning
			// },
			OnCreate: r.ensureTunnelServer,
			OnDelete: r.cleanupTunnelServer,
		},
		{
			Name:  "when-stopped/cleanup-tunnel-server",
			Title: "Cleanup tunnel server when machine is not running",
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateStopped ||
					obj.Spec.State == v1.MachineStateStopping ||
					obj.Spec.State == v1.MachineStateDisabled
			},
			OnCreate: r.cleanupTunnelServer,
			OnDelete: nil,
		},
		{
			Name:  "when-running/ensure-host-manager",
			Title: "Ensure host manager pod is running",
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateRunning
			},
			OnCreate: r.ensureHostManagerPod,
			OnDelete: r.cleanupHostManagerPod,
		},
		{
			Name:  "when-stopped/cleanup-host-manager",
			Title: "Cleanup host manager when machine is not running",
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateStopped ||
					obj.Spec.State == v1.MachineStateStopping ||
					obj.Spec.State == v1.MachineStateDisabled
			},
			OnCreate: r.cleanupHostManagerPod,
			OnDelete: nil,
		},
		{
			Name:     "handle-machine-type-change",
			Title:    "Handle machine type changes",
			OnCreate: r.handleMachineTypeChange,
			OnDelete: nil,
		},
		{
			Name:     "handle-node-reboot-request",
			Title:    "Handle node reboot requests for driver installation",
			OnCreate: r.handleNodeRebootRequest,
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

// buildTunnelServerConfig generates tunnel-server.conf with WireGuardDevice peers
func (r *WorkMachineReconciler) buildTunnelServerConfig(ctx context.Context, namespace string, serverPrivateKey string) (string, error) {
	// List all WireGuardDevices in the namespace
	var deviceList wireguarddevicev1.WireGuardDeviceList
	if err := r.List(ctx, &deviceList, client.InNamespace(namespace)); err != nil {
		return "", fmt.Errorf("failed to list WireGuardDevices: %w", err)
	}

	// Build [Interface] section
	config := fmt.Sprintf(`# WireGuard Server Configuration
[Interface]
PrivateKey = %s
Address = 10.17.0.1/24
ListenPort = 51820

PostUp = proxyguard --listen 0.0.0.0:443 --to 127.0.0.1:51820

PostUp = iptables -A FORWARD -i %%i -j ACCEPT;
PostUp = iptables -A FORWARD -o %%i -j ACCEPT;
PostUp = iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE;

PostDown = iptables -D FORWARD -i %%i -j ACCEPT;
PostDown = iptables -D FORWARD -o %%i -j ACCEPT;
PostDown = iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE;

`, serverPrivateKey)

	// Add [Peer] section for each Ready WireGuardDevice
	// Sort devices by name to ensure deterministic config generation
	readyDevices := make([]wireguarddevicev1.WireGuardDevice, 0, len(deviceList.Items))
	for _, device := range deviceList.Items {
		// Only include Ready devices with valid public keys
		if device.Status.Phase != "Ready" || device.Status.PublicKey == "" || device.Status.AssignedIP == "" {
			continue
		}
		readyDevices = append(readyDevices, device)
	}

	// Sort by device name for deterministic ordering
	sort.Slice(readyDevices, func(i, j int) bool {
		return readyDevices[i].Name < readyDevices[j].Name
	})

	for _, device := range readyDevices {
		config += fmt.Sprintf(`[Peer]
PublicKey = %s
AllowedIPs = %s/32

`, device.Status.PublicKey, device.Status.AssignedIP)
	}

	return config, nil
}

func (r *WorkMachineReconciler) ensureWorkmachineIngressController(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	deploymentName := "wm-ingress-controller"
	serviceAccountName := "wm-ingress-controller"
	clusterRoleName := fmt.Sprintf("wm-ingress-controller-%s", obj.Name)
	clusterRoleBindingName := fmt.Sprintf("wm-ingress-controller-%s", obj.Name)

	labels := map[string]string{
		"app":                      "wm-ingress-controller",
		"kloudlite.io/workmachine": obj.Name,
	}

	// Create ServiceAccount
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, serviceAccount, func() error {
		if !fn.IsOwner(serviceAccount, obj) {
			serviceAccount.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		serviceAccount.Labels = labels
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update wm-ingress-controller service account: %w", err))
	}

	// Create ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRole, func() error {
		if !fn.IsOwner(clusterRole, obj) {
			clusterRole.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		clusterRole.Labels = labels
		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		}
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update wm-ingress-controller cluster role: %w", err))
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRoleBinding, func() error {
		if !fn.IsOwner(clusterRoleBinding, obj) {
			clusterRoleBinding.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		clusterRoleBinding.Labels = labels
		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		}
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: obj.Spec.TargetNamespace,
			},
		}
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update wm-ingress-controller cluster role binding: %w", err))
	}

	// Create Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, deployment, func() error {
		if !fn.IsOwner(deployment, obj) {
			deployment.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		deployment.Labels = labels

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: fn.Ptr(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					NodeName:           obj.Name,
					RestartPolicy:      corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:            "wm-ingress-controller",
							Image:           wmIngressControllerImage,
							ImagePullPolicy: "Always",
							Args: []string{
								"--http-port",
								"80",
								"--https-port",
								"443",

								"--health-probe-bind-address",
								":17777",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "https",
									ContainerPort: 443,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "health",
									ContainerPort: 17777,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(17777),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
								TimeoutSeconds:      2,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(17777),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
				},
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update wm-ingress-controller deployment: %w", err))
	}

	// Now ensure the service exists
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wm-ingress-controller",
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, service, func() error {
		if !fn.IsOwner(service, obj) {
			service.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		service.Labels = labels

		service.Spec = corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update wm-ingress-controller service: %w", err))
	}

	return check.Passed()
}

// ensureTunnelServer creates and maintains the WireGuard tunnel server for the workmachine
// This function is called when the WorkMachine is in running state
func (r *WorkMachineReconciler) ensureTunnelServer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Create Secret for WireGuard configuration
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tunnel-server",
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, secret, func() error {
		if !fn.IsOwner(secret, obj) {
			secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		// Get or generate server private key
		var serverPrivateKey string
		if secret.Data != nil {
			if existingConf, exists := secret.Data["tunnel-server.conf"]; exists {
				// Extract existing private key from config
				lines := strings.SplitSeq(string(existingConf), "\n")
				for line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "PrivateKey") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							serverPrivateKey = strings.TrimSpace(parts[1])
							break
						}
					}
				}
			}
		}

		// Generate new server key if we don't have one
		if serverPrivateKey == "" {
			serverPriv, _, err := generateWgKeys()
			if err != nil {
				return err
			}
			serverPrivateKey = serverPriv
		}

		// Build config with WireGuardDevice peers
		newConfig, err := r.buildTunnelServerConfig(check.Context(), obj.Spec.TargetNamespace, serverPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to build tunnel-server config: %w", err)
		}

		// Compute hash of new config
		newHash := fmt.Sprintf("%x", sha256.Sum256([]byte(newConfig)))

		// Initialize Data and Annotations if needed
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}

		// Update config if changed
		secret.Data["tunnel-server.conf"] = []byte(newConfig)
		secret.Annotations["wireguard.kloudlite.io/config-hash"] = newHash

		// Store server public key for WireGuardDevice controller to use
		serverPrivKey, err := wgtypes.ParseKey(serverPrivateKey)
		if err == nil {
			serverPubKey := serverPrivKey.PublicKey()
			secret.Data["server-public-key"] = []byte(hex.EncodeToString(serverPubKey[:]))
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server secret: %w", err))
	}

	// Create StatefulSet for tunnel-server
	labels := map[string]string{
		"app":                      "tunnel-server",
		"kloudlite.io/workmachine": obj.Name,
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tunnel-server",
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, statefulSet, func() error {
		if !fn.IsOwner(statefulSet, obj) {
			statefulSet.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		statefulSet.Labels = labels

		statefulSet.Spec = appsv1.StatefulSetSpec{
			Replicas:            fn.Ptr(int32(1)),
			ServiceName:         "tunnel-server",
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					NodeName:      obj.Name,
					HostNetwork:   true,
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:  "tunnel-server",
							Image: wireguardTunnelImage,
							SecurityContext: &corev1.SecurityContext{
								Privileged: fn.Ptr(true),
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										corev1.Capability("NET_ADMIN"),
										corev1.Capability("SYS_ADMIN"),
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "wg-http-proxy",
									ContainerPort: 443,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Command: []string{
								"sh",
								"-c",
								strings.Join([]string{
									"wg-quick down wg0 || echo starting wireguard",
									"wg-quick up wg0 &",
									"pid=$!",
									"trap 'kill -9 $pid' SIGINT SIGTERM EXIT",
									"wait $pid",
								}, "\n"),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "wireguard-secret",
									MountPath: "/etc/wireguard/wg0.conf",
									SubPath:   "tunnel-server.conf",
									ReadOnly:  true,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"sh", "-c", "wg show wg0 | grep -q interface"},
									},
								},
								InitialDelaySeconds: 2,
								PeriodSeconds:       5,
								TimeoutSeconds:      2,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"sh", "-c", "wg show wg0 | grep -q interface"},
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "wireguard-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secret.Name,
								},
							},
						},
					},
				},
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server statefulset: %w", err))
	}

	return check.Passed()
}

func generateWgKeys() (privateKey, publicKey string, err error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return "", "", errors.Wrap("failed to generate wireguard keys", err)
	}

	return key.String(), key.PublicKey().String(), nil
}

func (r *WorkMachineReconciler) cleanupTunnelServer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Spec.TargetNamespace == "" {
		return check.Failed(fmt.Errorf("target namespace cannot be empty"))
	}

	// Delete StatefulSet (this will cascade delete pods)
	if err := r.Delete(check.Context(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tunnel-server",
			Namespace: obj.Spec.TargetNamespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	}

	// Delete Secret
	if err := r.Delete(check.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tunnel-server",
			Namespace: obj.Spec.TargetNamespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	}

	return check.Passed()
}

// handleNodeRebootRequest checks if the node associated with this WorkMachine has requested a reboot
// (typically for loading NVIDIA drivers after installation) and reboots the instance if needed
func (r *WorkMachineReconciler) handleNodeRebootRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Only handle reboot requests if machine is created and running
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	// Get the node with the same name as the WorkMachine
	var node corev1.Node
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, &node); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Node doesn't exist yet, nothing to do
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to get node: %w", err))
	}

	// Check if node has reboot requested annotation
	rebootRequested, exists := node.Annotations["kloudlite.io/workmachine-reboot-requested"]
	if !exists || rebootRequested != "true" {
		// No reboot requested
		return check.Passed()
	}

	check.Logger().Info("node reboot requested, rebooting instance", "node", node.Name, "machineID", obj.Status.MachineID)

	// Reboot the instance using cloud provider API
	if err := r.cloudProviderAPI.RebootMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to reboot machine: %w", err))
	}

	// Remove the reboot annotation from the node
	delete(node.Annotations, "kloudlite.io/workmachine-reboot-requested")
	if err := r.Update(check.Context(), &node); err != nil {
		return check.Failed(fmt.Errorf("failed to remove reboot annotation from node: %w", err))
	}

	check.Logger().Info("instance rebooted successfully, waiting for node to rejoin", "node", node.Name, "machineID", obj.Status.MachineID)

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

			ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
			defer cf()
			p, err := aws.NewProvider(ctx, aws.ProviderArgs{
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
	builder.Owns(&corev1.Pod{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.ClusterRole{})
	builder.Owns(&rbacv1.ClusterRoleBinding{})
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

	// Watch for WireGuardDevices to update tunnel-server when devices change
	builder.Watches(
		&wireguarddevicev1.WireGuardDevice{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			device, ok := obj.(*wireguarddevicev1.WireGuardDevice)
			if !ok {
				return nil
			}

			// Get the WorkMachine that references this namespace as TargetNamespace
			// The WireGuardDevice.Spec.WorkMachineRef contains the WorkMachine name
			if device.Spec.WorkMachineRef == "" {
				return nil
			}

			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: device.Spec.WorkMachineRef}},
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
