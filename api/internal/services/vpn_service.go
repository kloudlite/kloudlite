package services

import (
	"context"
	"fmt"
	"strings"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HostEntry represents a single hosts file entry
type HostEntry struct {
	Hostname string
	IP       string
}

// TunnelEndpointInfo contains the tunnel endpoint hostname and IP
type TunnelEndpointInfo struct {
	Hostname string // vpn-connect.{subdomain}.{domain}
	IP       string // Public IP of the WorkMachine
}

// VPNService provides business logic for VPN operations
type VPNService interface {
	// GetCACert retrieves the CA certificate
	GetCACert(ctx context.Context, username string) (string, error)

	// GetHosts retrieves the hosts list for a user's environment
	GetHosts(ctx context.Context, username string) ([]HostEntry, error)

	// GetTunnelEndpoint retrieves the tunnel server endpoint for a user's WorkMachine
	// Returns hostname (vpn-connect.{subdomain}.{domain}) and IP for /etc/hosts
	GetTunnelEndpoint(ctx context.Context, username string) (*TunnelEndpointInfo, error)
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

// buildHostEntries creates host entries from Services in the namespace
func (s *vpnService) buildHostEntries(ctx context.Context, namespace string) ([]HostEntry, error) {
	var ingressList networkingv1.IngressList
	if err := s.k8sClient.List(ctx, &ingressList); err != nil {
		return nil, err
	}

	var routerSvc corev1.Service
	if err := s.k8sClient.Get(ctx, fn.NN(namespace, "wm-ingress-controller"), &routerSvc); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	hosts := make([]HostEntry, 0, len(ingressList.Items))

	for i := range ingressList.Items {
		for j := range ingressList.Items[i].Spec.Rules {
			hosts = append(hosts, HostEntry{
				Hostname: ingressList.Items[i].Spec.Rules[j].Host,
				IP:       routerSvc.Spec.ClusterIP,
			})
		}
	}

	return hosts, nil
}

// GetCACert retrieves the CA certificate
func (s *vpnService) GetCACert(ctx context.Context, username string) (string, error) {
	s.logger.Info("VPN: GetCACert", zap.String("username", username))

	// Fetch CA certificate from kloudlite-ingress secret
	caSecret := &corev1.Secret{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{
		Name:      "kloudlite-wildcard-cert-tls",
		Namespace: "kloudlite-ingress",
	}, caSecret); err != nil {
		return "", fmt.Errorf("failed to fetch CA certificate: %w", err)
	}

	caCert := string(caSecret.Data["ca.crt"])
	if caCert == "" {
		return "", fmt.Errorf("ca.crt not found in secret")
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
	var workMachineList workmachinev1.WorkMachineList
	if err := s.k8sClient.List(ctx, &workMachineList); err != nil {
		return "", fmt.Errorf("failed to list work machines: %w", err)
	}

	for i := range workMachineList.Items {
		if workMachineList.Items[i].Spec.OwnedBy == username {
			return workMachineList.Items[i].Spec.TargetNamespace, nil
		}
	}

	return "", fmt.Errorf("no work machine found for user")
}

// findUserWorkMachine finds the WorkMachine for a user
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

// GetTunnelEndpoint retrieves the tunnel server endpoint for a user's WorkMachine
func (s *vpnService) GetTunnelEndpoint(ctx context.Context, username string) (*TunnelEndpointInfo, error) {
	wm, err := s.findUserWorkMachine(ctx, username)
	if err != nil {
		return nil, err
	}

	publicIP := wm.Status.PublicIP
	if publicIP == "" {
		return nil, fmt.Errorf("work machine has no public IP (may not be running)")
	}

	// Check if tunnel server is ready before returning endpoint
	ready, err := s.isTunnelServerReady(ctx, wm.Spec.TargetNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to check tunnel server status: %w", err)
	}
	if !ready {
		return nil, fmt.Errorf("tunnel server is not ready yet (WorkMachine may still be starting)")
	}

	// Get subdomain and domain from DomainRequest
	subdomain, domain, err := s.getDomainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info: %w", err)
	}

	// Build vpn-connect hostname: vpn-connect.{subdomain}.{domain}
	hostname := fmt.Sprintf("vpn-connect.%s.%s", subdomain, domain)

	return &TunnelEndpointInfo{
		Hostname: hostname,
		IP:       publicIP,
	}, nil
}

// getDomainInfo retrieves subdomain and domain from DomainRequest CR
func (s *vpnService) getDomainInfo(ctx context.Context) (subdomain, domain string, err error) {
	var domainRequestList domainrequestv1.DomainRequestList
	if err := s.k8sClient.List(ctx, &domainRequestList); err != nil {
		return "", "", fmt.Errorf("failed to list DomainRequests: %w", err)
	}

	if len(domainRequestList.Items) == 0 {
		return "", "", fmt.Errorf("no DomainRequest found")
	}

	// Use the first DomainRequest's spec.domainRoutes
	dr := domainRequestList.Items[0]
	if len(dr.Spec.DomainRoutes) == 0 {
		return "", "", fmt.Errorf("DomainRequest has no domain routes")
	}

	// Parse the full domain (e.g., "beanbag.khost.dev") into subdomain and domain
	fullDomain := dr.Spec.DomainRoutes[0].Domain
	parts := strings.SplitN(fullDomain, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid domain format: %s (expected subdomain.domain)", fullDomain)
	}

	return parts[0], parts[1], nil
}

// isTunnelServerReady checks if the tunnel-server pod is ready in the given namespace
func (s *vpnService) isTunnelServerReady(ctx context.Context, namespace string) (bool, error) {
	// Get tunnel-server pod (StatefulSet creates pod with name: tunnel-server-0)
	var pod corev1.Pod
	if err := s.k8sClient.Get(ctx, client.ObjectKey{
		Name:      "tunnel-server-0",
		Namespace: namespace,
	}, &pod); err != nil {
		if apiErrors.IsNotFound(err) {
			return false, nil // Pod doesn't exist yet
		}
		return false, fmt.Errorf("failed to get tunnel-server pod: %w", err)
	}

	// Check if pod is ready
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}
