package services

import (
	"context"
	"fmt"
	"time"

	wireguarddevicev1 "github.com/kloudlite/kloudlite/api/internal/controllers/wireguarddevice/v1"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VPNConfig represents the VPN configuration returned to clients
type VPNConfig struct {
	CACert   string      `json:"ca_cert"`
	WGConfig string      `json:"wg_config"`
	Hosts    []HostEntry `json:"hosts"`
}

// HostEntry represents a single hosts file entry
type HostEntry struct {
	Hostname string
	IP       string
}

// WireGuardConfigResponse contains WireGuard configuration with metadata
type WireGuardConfigResponse struct {
	Config         string `json:"config"`          // IPC format configuration
	AssignedIP     string `json:"assigned_ip"`     // Device IP address (e.g., "10.17.0.2")
	PublicKey      string `json:"public_key"`      // Device public key
	ServerEndpoint string `json:"server_endpoint"` // WorkMachine endpoint (e.g., "203.0.113.1:443")
}

// VPNService provides business logic for VPN operations
type VPNService interface {
	// GetVPNConfig retrieves VPN configuration for a user based on their username (combined endpoint)
	GetVPNConfig(ctx context.Context, tokenID, username string) (*VPNConfig, error)

	// GetWireGuardConfig retrieves only WireGuard configuration for a user
	GetWireGuardConfig(ctx context.Context, deviceID, username string) (*WireGuardConfigResponse, error)

	// GetCACert retrieves only the CA certificate
	GetCACert(ctx context.Context, username string) (string, error)

	// GetHosts retrieves only the hosts list for a user's environment
	GetHosts(ctx context.Context, username string) ([]HostEntry, error)
}

// vpnService implements VPNService
type vpnService struct {
	k8sClient client.Client
	logger    *zap.Logger
}

// NewVPNService creates a new VPNService
func NewVPNService(k8sClient client.Client, logger *zap.Logger) VPNService {
	return &vpnService{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// GetVPNConfig retrieves VPN configuration for a user by username
// tokenID is no longer used (kept for backwards compatibility but can be removed)
func (s *vpnService) GetVPNConfig(ctx context.Context, tokenID, username string) (*VPNConfig, error) {
	// 1. Find the user's WorkMachine by username (matches WorkMachine.Spec.OwnedBy)
	var workMachineList workmachinev1.WorkMachineList
	if err := s.k8sClient.List(ctx, &workMachineList); err != nil {
		return nil, fmt.Errorf("failed to list work machines: %w", err)
	}

	var userWorkMachine *workmachinev1.WorkMachine
	for i := range workMachineList.Items {
		if workMachineList.Items[i].Spec.OwnedBy == username {
			userWorkMachine = &workMachineList.Items[i]
			break
		}
	}

	if userWorkMachine == nil {
		return nil, fmt.Errorf("no work machine found for user")
	}

	targetNamespace := userWorkMachine.Spec.TargetNamespace
	s.logger.Info("VPN config: Found WorkMachine",
		zap.String("username", username),
		zap.String("workMachine", userWorkMachine.Name),
		zap.String("namespace", targetNamespace))

	// 2. Fetch CA certificate from kloudlite-ingress/kloudlite-wildcard-cert-tls
	caCert, err := s.fetchCACertificate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CA certificate: %w", err)
	}

	// 3. Fetch WireGuard configuration from tunnel-server secret
	wgConfig, err := s.fetchWireGuardConfig(ctx, targetNamespace)
	if err != nil {
		return nil, err // Error already formatted
	}

	// 4. Build host entries from Services in the target namespace
	hosts, err := s.buildHostEntries(ctx, targetNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to build host entries: %w", err)
	}

	s.logger.Info("VPN config: Retrieved successfully",
		zap.String("username", username),
		zap.Int("hostCount", len(hosts)))

	return &VPNConfig{
		CACert:   caCert,
		WGConfig: wgConfig,
		Hosts:    hosts,
	}, nil
}

// fetchCACertificate retrieves the CA certificate from the Kubernetes secret
func (s *vpnService) fetchCACertificate(ctx context.Context) (string, error) {
	caSecret := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{
		Name:      "kloudlite-wildcard-cert-tls",
		Namespace: "kloudlite-ingress",
	}, caSecret); err != nil {
		return "", err
	}

	caCert := string(caSecret.Data["ca.crt"])
	if caCert == "" {
		return "", fmt.Errorf("ca.crt not found in secret")
	}

	return caCert, nil
}

// fetchWireGuardConfig retrieves the WireGuard peer configuration from the tunnel-server secret
// DEPRECATED: This function is for legacy compatibility only. New clients should use GetWireGuardConfig()
// which uses per-device WireGuardDevice CRDs instead of shared peer1.ipc/peer1.conf
func (s *vpnService) fetchWireGuardConfig(ctx context.Context, namespace string) (string, error) {
	return "", fmt.Errorf("legacy peer1.ipc/peer1.conf configuration is no longer supported. Please use the /api/vpn/wireguard-config?device_id=<device_id> endpoint with a unique device ID")
}

// buildHostEntries creates host entries from Services in the namespace
func (s *vpnService) buildHostEntries(ctx context.Context, namespace string) ([]HostEntry, error) {
	var serviceList corev1.ServiceList
	if err := s.k8sClient.List(ctx, &serviceList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	hosts := make([]HostEntry, 0, len(serviceList.Items))
	for _, svc := range serviceList.Items {
		// Only include services with ClusterIP (not headless services)
		if svc.Spec.ClusterIP != "" && svc.Spec.ClusterIP != "None" {
			// Create hostname as {service-name}.{namespace}.svc.cluster.local
			hostname := fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace)
			hosts = append(hosts, HostEntry{
				Hostname: hostname,
				IP:       svc.Spec.ClusterIP,
			})
		}
	}

	return hosts, nil
}

// GetWireGuardConfig retrieves WireGuard configuration for a user by username
// deviceID is the unique identifier for the client device
func (s *vpnService) GetWireGuardConfig(ctx context.Context, deviceID, username string) (*WireGuardConfigResponse, error) {
	// Find the user's WorkMachine
	targetNamespace, workMachineRef, err := s.findUserNamespaceAndWorkMachine(ctx, username)
	if err != nil {
		return nil, err
	}

	// Also get the full WorkMachine to access IP address
	workMachine, err := s.findUserWorkMachine(ctx, username)
	if err != nil {
		return nil, err
	}

	s.logger.Info("VPN: GetWireGuardConfig",
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("namespace", targetNamespace))

	// Get or create WireGuardDevice
	device, err := s.getOrCreateWireGuardDevice(ctx, deviceID, username, workMachineRef, targetNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create WireGuard device: %w", err)
	}

	// Wait for device to become Ready (simple polling, max 10 seconds)
	for i := 0; i < 10; i++ {
		if device.Status.Phase == "Ready" {
			break
		}
		time.Sleep(1 * time.Second)
		if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: device.Name, Namespace: device.Namespace}, device); err != nil {
			return nil, fmt.Errorf("failed to get device status: %w", err)
		}
	}

	if device.Status.Phase != "Ready" {
		return nil, fmt.Errorf("device not ready: %s - %s", device.Status.Phase, device.Status.Message)
	}

	// Fetch device config from secret
	secretName := fmt.Sprintf("wg-device-%s", deviceID)
	deviceSecret := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: secretName, Namespace: targetNamespace}, deviceSecret); err != nil {
		return nil, fmt.Errorf("failed to get device secret: %w", err)
	}

	// Use IPC format (standard WireGuard protocol)
	var config string
	if ipcConfig, ok := deviceSecret.Data["peer.ipc"]; ok {
		config = string(ipcConfig)
	} else {
		return nil, fmt.Errorf("device secret missing peer.ipc configuration")
	}

	// Determine server endpoint from WorkMachine IP
	serverEndpoint := ""
	if workMachine.Status.PublicIP != "" {
		serverEndpoint = fmt.Sprintf("%s:443", workMachine.Status.PublicIP)
	} else if workMachine.Status.PrivateIP != "" {
		serverEndpoint = fmt.Sprintf("%s:443", workMachine.Status.PrivateIP)
	}

	// Return IPC config along with IP and public key
	// IP address must be configured separately as IPC protocol doesn't support it
	return &WireGuardConfigResponse{
		Config:         config,
		AssignedIP:     device.Status.AssignedIP,
		PublicKey:      device.Status.PublicKey,
		ServerEndpoint: serverEndpoint,
	}, nil
}

// GetCACert retrieves the CA certificate
func (s *vpnService) GetCACert(ctx context.Context, username string) (string, error) {
	s.logger.Info("VPN: GetCACert", zap.String("username", username))

	// Fetch CA certificate
	caCert, err := s.fetchCACertificate(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch CA certificate: %w", err)
	}

	return caCert, nil
}

// GetHosts retrieves host entries for a user's environment
func (s *vpnService) GetHosts(ctx context.Context, username string) ([]HostEntry, error) {
	// Find the user's WorkMachine
	targetNamespace, err := s.findUserNamespace(ctx, username)
	if err != nil {
		return nil, err
	}

	s.logger.Info("VPN: GetHosts",
		zap.String("username", username),
		zap.String("namespace", targetNamespace))

	// Build host entries
	hosts, err := s.buildHostEntries(ctx, targetNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to build host entries: %w", err)
	}

	return hosts, nil
}

// findUserNamespace finds the target namespace for a user by looking up their WorkMachine
func (s *vpnService) findUserNamespace(ctx context.Context, username string) (string, error) {
	namespace, _, err := s.findUserNamespaceAndWorkMachine(ctx, username)
	return namespace, err
}

// findUserNamespaceAndWorkMachine finds both namespace and WorkMachine name for a user
func (s *vpnService) findUserNamespaceAndWorkMachine(ctx context.Context, username string) (string, string, error) {
	var workMachineList workmachinev1.WorkMachineList
	if err := s.k8sClient.List(ctx, &workMachineList); err != nil {
		return "", "", fmt.Errorf("failed to list work machines: %w", err)
	}

	for i := range workMachineList.Items {
		if workMachineList.Items[i].Spec.OwnedBy == username {
			return workMachineList.Items[i].Spec.TargetNamespace, workMachineList.Items[i].Name, nil
		}
	}

	return "", "", fmt.Errorf("no work machine found for user")
}

// findUserWorkMachine finds the WorkMachine for a user and returns the full object
func (s *vpnService) findUserWorkMachine(ctx context.Context, username string) (*workmachinev1.WorkMachine, error) {
	var workMachineList workmachinev1.WorkMachineList
	if err := s.k8sClient.List(ctx, &workMachineList); err != nil {
		return nil, fmt.Errorf("failed to list work machines: %w", err)
	}

	for i := range workMachineList.Items {
		if workMachineList.Items[i].Spec.OwnedBy == username {
			return &workMachineList.Items[i], nil
		}
	}

	return nil, fmt.Errorf("no work machine found for user")
}

// getOrCreateWireGuardDevice retrieves or creates a WireGuardDevice for the given deviceID
func (s *vpnService) getOrCreateWireGuardDevice(ctx context.Context, deviceID, username, workMachineRef, namespace string) (*wireguarddevicev1.WireGuardDevice, error) {
	deviceName := fmt.Sprintf("wg-%s", deviceID[:8]) // Use first 8 chars of device UUID

	// Try to get existing device
	device := &wireguarddevicev1.WireGuardDevice{}
	err := s.k8sClient.Get(ctx, client.ObjectKey{Name: deviceName, Namespace: namespace}, device)
	if err == nil {
		// Device exists, update lastSeen
		now := metav1.Now()
		device.Status.LastSeen = &now
		if err := s.k8sClient.Status().Update(ctx, device); err != nil {
			s.logger.Warn("Failed to update device lastSeen", zap.Error(err))
		}
		return device, nil
	}

	if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// Device doesn't exist, create it
	device = &wireguarddevicev1.WireGuardDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deviceName,
			Namespace: namespace,
		},
		Spec: wireguarddevicev1.WireGuardDeviceSpec{
			DeviceID:       deviceID,
			UserRef:        username,
			WorkMachineRef: workMachineRef,
		},
	}

	if err := s.k8sClient.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	s.logger.Info("Created new WireGuardDevice",
		zap.String("deviceID", deviceID),
		zap.String("username", username),
		zap.String("namespace", namespace))

	return device, nil
}
