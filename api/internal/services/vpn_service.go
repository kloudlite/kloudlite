package services

import (
	"context"
	"fmt"

	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

// VPNService provides business logic for VPN operations
type VPNService interface {
	// GetVPNConfig retrieves VPN configuration for a user based on their username
	GetVPNConfig(ctx context.Context, tokenID, username string) (*VPNConfig, error)
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
func (s *vpnService) fetchWireGuardConfig(ctx context.Context, namespace string) (string, error) {
	wgSecret := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{
		Name:      "tunnel-server",
		Namespace: namespace,
	}, wgSecret); err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("VPN tunnel not configured. Please ensure your WorkMachine is running")
		}
		return "", err
	}

	wgConfig := string(wgSecret.Data["peer1.conf"])
	if wgConfig == "" {
		return "", fmt.Errorf("peer1.conf not found in tunnel-server secret")
	}

	return wgConfig, nil
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
