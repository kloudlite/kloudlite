package wireguarddevice

import (
	"context"
	"fmt"
	"net"

	wireguarddevicev1 "github.com/kloudlite/kloudlite/api/internal/controllers/wireguarddevice/v1"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	deviceFinalizer = "vpn.kloudlite.io/wireguard-device-finalizer"
)

// WireGuardDeviceReconciler reconciles a WireGuardDevice object
type WireGuardDeviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vpn.kloudlite.io,resources=wireguarddevices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vpn.kloudlite.io,resources=wireguarddevices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vpn.kloudlite.io,resources=wireguarddevices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop
func (r *WireGuardDeviceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &wireguarddevicev1.WireGuardDevice{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return reconciler.ReconcileSteps(req, []reconciler.Step[*wireguarddevicev1.WireGuardDevice]{
		{
			Name:     "validate-workmachine",
			Title:    "Validate WorkMachine reference",
			OnCreate: r.validateWorkMachine,
		},
		{
			Name:     "allocate-ip",
			Title:    "Allocate IP address",
			OnCreate: r.allocateIP,
		},
		{
			Name:     "generate-keys",
			Title:    "Generate WireGuard keys",
			OnCreate: r.generateKeys,
		},
		{
			Name:     "create-device-secret",
			Title:    "Create device configuration secret",
			OnCreate: r.createDeviceSecret,
		},
		{
			Name:     "update-tunnel-server",
			Title:    "Update tunnel server configuration",
			OnCreate: r.updateTunnelServer,
			OnDelete: r.removePeerFromTunnelServer,
		},
		{
			Name:     "finalize-status",
			Title:    "Update device status to Ready",
			OnCreate: r.finalizeStatus,
		},
	})
}

// validateWorkMachine verifies that the referenced WorkMachine exists
func (r *WireGuardDeviceReconciler) validateWorkMachine(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	ctx := check.Context()

	// Fetch WorkMachine
	workMachine := &workmachinev1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Spec.WorkMachineRef}, workMachine); err != nil {
		if errors.IsNotFound(err) {
			obj.Status.Phase = "Error"
			obj.Status.Message = fmt.Sprintf("WorkMachine %s not found", obj.Spec.WorkMachineRef)
			return check.Failed(err)
		}
		return check.Failed(err)
	}

	// Verify ownership
	if workMachine.Spec.OwnedBy != obj.Spec.UserRef {
		obj.Status.Phase = "Error"
		obj.Status.Message = "WorkMachine owner mismatch"
		return check.Failed(fmt.Errorf("work machine owned by %s, not %s", workMachine.Spec.OwnedBy, obj.Spec.UserRef))
	}

	obj.Status.Phase = "Provisioning"
	return check.Passed()
}

// allocateIP assigns an IP address to the device
func (r *WireGuardDeviceReconciler) allocateIP(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	ctx := check.Context()

	// If IP already assigned, skip
	if obj.Status.AssignedIP != "" {
		return check.Passed()
	}

	// List all WireGuardDevices in namespace to find highest IP
	var deviceList wireguarddevicev1.WireGuardDeviceList
	if err := r.List(ctx, &deviceList, client.InNamespace(obj.Namespace)); err != nil {
		return check.Failed(err)
	}

	// Start from 10.17.0.2 (10.17.0.1 is server)
	baseIP := net.ParseIP("10.17.0.2")
	highestIP := 1 // Start at .1, next will be .2

	// Find highest allocated IP
	for _, device := range deviceList.Items {
		if device.Status.AssignedIP != "" {
			ip := net.ParseIP(device.Status.AssignedIP)
			if ip != nil && ip.To4() != nil {
				lastOctet := int(ip.To4()[3])
				if lastOctet > highestIP {
					highestIP = lastOctet
				}
			}
		}
	}

	// Allocate next IP
	nextIP := make(net.IP, len(baseIP))
	copy(nextIP, baseIP)
	nextIP[len(nextIP)-1] = byte(highestIP + 1)

	obj.Status.AssignedIP = nextIP.String()
	return check.Passed()
}

// generateKeys generates WireGuard key pair for the device
func (r *WireGuardDeviceReconciler) generateKeys(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	// If public key already exists, skip
	if obj.Status.PublicKey != "" {
		return check.Passed()
	}

	// Generate WireGuard key pair
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return check.Failed(err)
	}

	publicKey := privateKey.PublicKey()
	obj.Status.PublicKey = publicKey.String()

	// Store private key in object annotations for use in secret creation
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations["vpn.kloudlite.io/private-key"] = privateKey.String()

	return check.Passed()
}

// createDeviceSecret creates a secret with device configuration
func (r *WireGuardDeviceReconciler) createDeviceSecret(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	ctx := check.Context()

	secretName := fmt.Sprintf("wg-device-%s", obj.Spec.DeviceID)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: obj.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.ObjectMeta.OwnerReferences == nil {
			controllerutil.SetControllerReference(obj, secret, r.Scheme)
		}

		// Get server public key from tunnel-server secret
		serverSecret := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{Name: "tunnel-server", Namespace: obj.Namespace}, serverSecret); err != nil {
			return fmt.Errorf("failed to get tunnel-server secret: %w", err)
		}

		serverPublicKey := string(serverSecret.Data["server-public-key"])
		privateKey := obj.Annotations["vpn.kloudlite.io/private-key"]

		// Create IPC format config
		ipcConfig := fmt.Sprintf(`private_key=%s
listen_port=51820
public_key=%s
allowed_ip=10.17.0.0/24
allowed_ip=10.43.0.0/16
endpoint=127.0.0.1:51821
`, privateKey, serverPublicKey)

		// Create INI format config
		iniConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/24
ListenPort = 51820

[Peer]
PublicKey = %s
AllowedIPs = 10.17.0.0/24, 10.43.0.0/16
Endpoint = 127.0.0.1:51821
`, privateKey, obj.Status.AssignedIP, serverPublicKey)

		secret.StringData = map[string]string{
			"peer.ipc":  ipcConfig,
			"peer.conf": iniConfig,
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// updateTunnelServer updates the tunnel-server configuration to include this device
func (r *WireGuardDeviceReconciler) updateTunnelServer(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	// TODO: Implement multi-peer tunnel-server update
	// For now, mark as completed to allow testing of device creation
	obj.Status.ConfigGeneration++
	return check.Passed()
}

// removePeerFromTunnelServer removes the device peer from tunnel-server on deletion
func (r *WireGuardDeviceReconciler) removePeerFromTunnelServer(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	// TODO: Implement peer removal from tunnel-server
	return check.Passed()
}

// finalizeStatus sets the device status to Ready
func (r *WireGuardDeviceReconciler) finalizeStatus(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	obj.Status.Phase = "Ready"
	obj.Status.Message = "Device provisioned successfully"
	now := metav1.Now()
	obj.Status.LastSeen = &now
	return check.Passed()
}

// SetupWithManager sets up the controller with the Manager
func (r *WireGuardDeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wireguarddevicev1.WireGuardDevice{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
