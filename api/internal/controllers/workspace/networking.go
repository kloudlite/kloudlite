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

// ensureWorkspaceService ensures a Service is created for the workspace
func (r *WorkspaceReconciler) ensureWorkspaceService(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	serviceName := workspace.Name

	// Get target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	// Check if Service exists
	svc := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: targetNamespace}, svc)

	if apierrors.IsNotFound(err) {
		// Create new Service
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: targetNamespace,
				Labels: map[string]string{
					"app":       "workspace",
					"workspace": workspace.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app":       "workspace",
					"workspace": workspace.Name,
				},
				Ports: []corev1.ServicePort{
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
				},
				Type: corev1.ServiceTypeClusterIP,
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, svc, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		if err := r.Create(ctx, svc); err != nil {
			return fmt.Errorf("failed to create Service: %w", err)
		}
		logger.Info("Created Service", zap.String("name", serviceName))
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get Service: %w", err)
	}

	// Service exists, no need to update
	logger.Info("Service already exists", zap.String("name", serviceName))
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
