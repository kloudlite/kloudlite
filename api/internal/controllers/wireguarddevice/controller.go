package wireguarddevice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sort"

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
			Name:     "create-device-secret",
			Title:    "Create device configuration secret and generate keys",
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

// createDeviceSecret creates a secret with device configuration and generates WireGuard keys
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

		// Extract server public key (hex format required for IPC protocol)
		serverPublicKeyHex := string(serverSecret.Data["server-public-key"])
		if serverPublicKeyHex == "" {
			return fmt.Errorf("server-public-key not found in tunnel-server secret")
		}

		// Generate private key for IPC config
		// Note: Private key is only used to build peer.ipc and is not stored
		// for security reasons (private keys should only exist on client devices)
		privKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return fmt.Errorf("failed to generate private key: %w", err)
		}
		// WireGuard IPC protocol requires lowercase hex encoding
		privateKeyHex := hex.EncodeToString(privKey[:])

		// Update public key in status if it changed
		pubKey := privKey.PublicKey()
		if obj.Status.PublicKey != pubKey.String() {
			obj.Status.PublicKey = pubKey.String()
		}

		// Create IPC format config (uses hex-encoded keys)
		// This is the only config format used by clients
		ipcConfig := fmt.Sprintf(`private_key=%s
listen_port=51820
public_key=%s
allowed_ip=10.17.0.0/24
allowed_ip=10.43.0.0/16
endpoint=127.0.0.1:51821
`, privateKeyHex, serverPublicKeyHex)

		// Store only the IPC config that clients actually use
		// Clear any existing data to remove legacy fields (peer.conf, private-key)
		secret.Data = map[string][]byte{
			"peer.ipc": []byte(ipcConfig),
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// computePeersHash computes a hash of all WireGuardDevice peers for a WorkMachine
func (r *WireGuardDeviceReconciler) computePeersHash(ctx context.Context, namespace string, workMachineRef string) (string, error) {
	var deviceList wireguarddevicev1.WireGuardDeviceList
	if err := r.List(ctx, &deviceList, client.InNamespace(namespace)); err != nil {
		return "", fmt.Errorf("failed to list WireGuardDevices: %w", err)
	}

	// Collect all Ready devices for this WorkMachine (matching buildTunnelServerConfig logic)
	var peerConfigs []string
	for _, device := range deviceList.Items {
		// Only include Ready devices with valid public keys (same as buildTunnelServerConfig)
		if device.Spec.WorkMachineRef == workMachineRef &&
			device.DeletionTimestamp == nil &&
			device.Status.Phase == "Ready" &&
			device.Status.PublicKey != "" &&
			device.Status.AssignedIP != "" {
			// Create a stable string representation of this peer
			peerConfig := fmt.Sprintf("%s:%s:%s", device.Spec.DeviceID, device.Status.PublicKey, device.Status.AssignedIP)
			peerConfigs = append(peerConfigs, peerConfig)
		}
	}

	// Sort for deterministic hash
	sort.Strings(peerConfigs)

	// Compute SHA256 hash
	hash := sha256.New()
	for _, cfg := range peerConfigs {
		hash.Write([]byte(cfg))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// updateTunnelServer updates the tunnel-server configuration to include this device
func (r *WireGuardDeviceReconciler) updateTunnelServer(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	ctx := check.Context()

	// Fetch the WorkMachine to trigger its reconciliation
	workMachine := &workmachinev1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Spec.WorkMachineRef}, workMachine); err != nil {
		return check.Failed(fmt.Errorf("failed to get WorkMachine: %w", err))
	}

	// Compute current peers hash
	currentHash, err := r.computePeersHash(ctx, obj.Namespace, obj.Spec.WorkMachineRef)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to compute peers hash: %w", err))
	}

	// Update annotation only if hash changed
	if workMachine.Annotations == nil {
		workMachine.Annotations = make(map[string]string)
	}

	existingHash := workMachine.Annotations["vpn.kloudlite.io/peers-config-hash"]
	if existingHash != currentHash {
		// Hash changed, update the annotation to trigger WorkMachine reconciliation
		workMachine.Annotations["vpn.kloudlite.io/peers-config-hash"] = currentHash

		if err := r.Update(ctx, workMachine); err != nil {
			// Use Errored instead of Failed to allow retries on transient errors
			// (e.g., webhook endpoint unavailable, network issues, pod restarts)
			return check.Errored(fmt.Errorf("failed to update WorkMachine peers hash: %w", err))
		}
	}

	obj.Status.ConfigGeneration++
	return check.Passed()
}

// removePeerFromTunnelServer removes the device peer from tunnel-server on deletion
func (r *WireGuardDeviceReconciler) removePeerFromTunnelServer(check *reconciler.Check[*wireguarddevicev1.WireGuardDevice], obj *wireguarddevicev1.WireGuardDevice) reconciler.StepResult {
	ctx := check.Context()

	// Fetch the WorkMachine to trigger its reconciliation
	workMachine := &workmachinev1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Spec.WorkMachineRef}, workMachine); err != nil {
		// WorkMachine might already be deleted, which is fine
		if errors.IsNotFound(err) {
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to get WorkMachine: %w", err))
	}

	// Compute current peers hash (excluding this device since it's being deleted)
	currentHash, err := r.computePeersHash(ctx, obj.Namespace, obj.Spec.WorkMachineRef)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to compute peers hash: %w", err))
	}

	// Update annotation only if hash changed
	if workMachine.Annotations == nil {
		workMachine.Annotations = make(map[string]string)
	}

	existingHash := workMachine.Annotations["vpn.kloudlite.io/peers-config-hash"]
	if existingHash != currentHash {
		// Hash changed, update the annotation to trigger WorkMachine reconciliation
		workMachine.Annotations["vpn.kloudlite.io/peers-config-hash"] = currentHash

		if err := r.Update(ctx, workMachine); err != nil {
			// Use Errored instead of Failed to allow retries on transient errors
			// (e.g., webhook endpoint unavailable, network issues, pod restarts)
			return check.Errored(fmt.Errorf("failed to update WorkMachine peers hash: %w", err))
		}
	}

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
		WithEventFilter(reconciler.ReconcileFilter()).
		Complete(r)
}
