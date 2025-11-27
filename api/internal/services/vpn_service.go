package services

import (
	"context"
	"fmt"

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

// VPNService provides business logic for VPN operations
type VPNService interface {
	// GetCACert retrieves the CA certificate
	GetCACert(ctx context.Context, username string) (string, error)

	// GetHosts retrieves the hosts list for a user's environment
	GetHosts(ctx context.Context, username string) ([]HostEntry, error)

	// GetTunnelEndpoint retrieves the tunnel server endpoint for a user's WorkMachine
	GetTunnelEndpoint(ctx context.Context, username string) (string, error)
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
func (s *vpnService) GetTunnelEndpoint(ctx context.Context, username string) (string, error) {
	wm, err := s.findUserWorkMachine(ctx, username)
	if err != nil {
		return "", err
	}

	publicIP := wm.Status.PublicIP
	if publicIP == "" {
		return "", fmt.Errorf("work machine has no public IP (may not be running)")
	}

	// Return the tunnel endpoint with port 443 (tunnel server listens on 443)
	return fmt.Sprintf("%s:443", publicIP), nil
}
