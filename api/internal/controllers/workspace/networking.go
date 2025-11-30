package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// generateHash generates an 8-character hash from the input string
func generateHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])[:8]
}

// buildServicePorts builds the list of service ports including exposed ports
func (r *WorkspaceReconciler) buildServicePorts(workspace *workspacev1.Workspace) []corev1.ServicePort {
	// Start with default ports
	ports := []corev1.ServicePort{
		{
			Name:       "ssh",
			Protocol:   corev1.ProtocolTCP,
			Port:       22,
			TargetPort: intstr.FromInt(22),
		},
		{
			Name:       "code-server",
			Protocol:   corev1.ProtocolTCP,
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
		},
		{
			Name:       "ttyd",
			Protocol:   corev1.ProtocolTCP,
			Port:       7681,
			TargetPort: intstr.FromInt(7681),
		},
		{
			Name:       "claude-ttyd",
			Protocol:   corev1.ProtocolTCP,
			Port:       7682,
			TargetPort: intstr.FromInt(7682),
		},
		{
			Name:       "opencode-ttyd",
			Protocol:   corev1.ProtocolTCP,
			Port:       7683,
			TargetPort: intstr.FromInt(7683),
		},
		{
			Name:       "codex-ttyd",
			Protocol:   corev1.ProtocolTCP,
			Port:       7684,
			TargetPort: intstr.FromInt(7684),
		},
	}

	// Track which ports are already defined to avoid duplicates
	definedPorts := make(map[string]bool)
	for _, p := range ports {
		key := fmt.Sprintf("%s-%d", p.Protocol, p.Port)
		definedPorts[key] = true
	}

	// Add user-exposed ports
	for _, exposed := range workspace.Spec.Expose {
		var protocol corev1.Protocol
		switch exposed.Protocol {
		case workspacev1.ExposeProtocolTCP, workspacev1.ExposeProtocolHTTP:
			protocol = corev1.ProtocolTCP
		case workspacev1.ExposeProtocolUDP:
			protocol = corev1.ProtocolUDP
		default:
			continue
		}

		key := fmt.Sprintf("%s-%d", protocol, exposed.Port)
		if definedPorts[key] {
			continue // Skip if already defined
		}
		definedPorts[key] = true

		ports = append(ports, corev1.ServicePort{
			Name:       fmt.Sprintf("exposed-%s-%d", exposed.Protocol, exposed.Port),
			Protocol:   protocol,
			Port:       exposed.Port,
			TargetPort: intstr.FromInt32(exposed.Port),
		})
	}

	return ports
}

// ensureWorkspaceService ensures a Service is created for the workspace
func (r *WorkspaceReconciler) ensureWorkspaceService(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	serviceName := workspace.Name

	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: targetNamespace}}

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, svc, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		svc.Labels = map[string]string{
			"app":       "workspace",
			"workspace": workspace.Name,
		}
		svc.Spec.Selector = map[string]string{
			"app":       "workspace",
			"workspace": workspace.Name,
		}
		svc.Spec.Ports = r.buildServicePorts(workspace)
		svc.Spec.Type = corev1.ServiceTypeClusterIP

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create/update Service: %w", err)
	}

	logger.Info("Service reconciled", zap.String("name", serviceName), zap.String("result", string(result)))
	return nil
}

func (r *WorkspaceReconciler) setupWorkspaceIngress(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	// Fetch the DomainRequest to get subdomain
	domainRequest := &domainrequestv1.DomainRequest{}
	if err := r.Get(ctx, fn.NN("", "installation-domain"), domainRequest); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("DomainRequest not found yet, skipping Ingress creation")
			return nil
		}
		return fmt.Errorf("failed to get DomainRequest: %w", err)
	}

	// Check if subdomain is available
	if domainRequest.Status.Subdomain == "" {
		logger.Info("Subdomain not available yet, skipping Ingress creation")
		return nil
	}

	// Define HTTP services to expose via Ingress
	// Map of service-prefix -> port
	httpServices := map[string]int32{
		"vscode":   8080,
		"tty":      7681,
		"claude":   7682,
		"opencode": 7683,
		"codex":    7684,
	}

	// Service name that ingress will route to
	serviceName := workspace.Name

	// Generate hash of owner-workspaceName for unique, DNS-friendly hostnames
	wsHash := generateHash(fmt.Sprintf("%s-%s", workspace.Spec.OwnedBy, workspace.Name))

	// Build Ingress rules
	var ingressRules []networkingv1.IngressRule

	for prefix, port := range httpServices {
		// Use pattern: {prefix}-{hash(owner-workspaceName)}.{subdomain}
		// Example: claude-a1b2c3d4.eastman.khost.dev
		host := fmt.Sprintf("%s-%s.%s", prefix, wsHash, domainRequest.Status.Subdomain)

		pathType := networkingv1.PathTypePrefix
		ingressRules = append(ingressRules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: serviceName,
									Port: networkingv1.ServiceBackendPort{
										Number: port,
									},
								},
							},
						},
					},
				},
			},
		})
	}

	// Add ingress rules for user-exposed HTTP ports
	// These use the workspace service and pattern: p{port}-{hash}.{subdomain}
	for _, exposed := range workspace.Spec.Expose {
		if exposed.Protocol != workspacev1.ExposeProtocolHTTP {
			continue
		}

		// Use pattern: p{port}-{hash(owner-workspaceName)}.{subdomain}
		// Example: p3000-a1b2c3d4.eastman.khost.dev
		host := fmt.Sprintf("p%d-%s.%s", exposed.Port, wsHash, domainRequest.Status.Subdomain)

		pathType := networkingv1.PathTypePrefix
		ingressRules = append(ingressRules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: serviceName,
									Port: networkingv1.ServiceBackendPort{
										Number: exposed.Port,
									},
								},
							},
						},
					},
				},
			},
		})
		logger.Info("Adding ingress rule for exposed HTTP port",
			zap.Int32("port", exposed.Port),
			zap.String("host", host))
	}

	ingress := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: workspace.Name, Namespace: targetNamespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		// Set owner reference for automatic cleanup via garbage collection
		if err := controllerutil.SetControllerReference(workspace, ingress, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on Ingress: %w", err)
		}

		// Note: TLS is handled by the wm-ingress-controller using a central wildcard cert
		// from kloudlite-ingress namespace, so we don't need to specify TLS in the Ingress
		ingress.Spec.TLS = nil
		ingress.Spec.Rules = ingressRules
		ingress.SetLabels(fn.MapMerge(ingress.GetLabels(), map[string]string{
			"app":                               "workspace",
			"workspace":                         workspace.Name,
			"workspaces.kloudlite.io/workspace": workspace.Name,
			"kloudlite.io/workmachine":          workspace.Spec.WorkmachineName,
		}))
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// ensureWorkspaceHeadlessService ensures a headless Service is created for the workspace
// This headless service is used by service intercepts to route traffic to workspace pods
func (r *WorkspaceReconciler) ensureWorkspaceHeadlessService(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	headlessServiceName := fmt.Sprintf("%s-headless", workspace.Name)

	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	// Check if headless Service exists
	headlessSvc := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Name: headlessServiceName, Namespace: targetNamespace}, headlessSvc)

	if apierrors.IsNotFound(err) {
		// Create new headless Service
		headlessSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      headlessServiceName,
				Namespace: targetNamespace,
				Labels: map[string]string{
					"app":                               "workspace",
					"workspace":                         workspace.Name,
					"workspaces.kloudlite.io/workspace": workspace.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Type:      corev1.ServiceTypeClusterIP,
				ClusterIP: "None", // Headless service - DNS returns pod IP directly
				Selector: map[string]string{
					"workspaces.kloudlite.io/workspace-name": workspace.Name,
				},
				// No ports needed - SOCAT connects directly to pod IP:port via DNS
				// The port is specified in the SOCAT command, not resolved through the service
				Ports: []corev1.ServicePort{},
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, headlessSvc, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		if err := r.Create(ctx, headlessSvc); err != nil {
			return fmt.Errorf("failed to create headless Service: %w", err)
		}
		logger.Info("Created headless Service for workspace", zap.String("name", headlessServiceName))
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get headless Service: %w", err)
	}

	// Headless service exists, no need to update
	logger.Info("Headless service already exists", zap.String("name", headlessServiceName))
	return nil
}
