package workspace

import (
	"context"
	"fmt"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
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

// ensurePackageRequest creates or updates PackageRequest for the workspace
func (r *WorkspaceReconciler) ensurePackageRequest(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	pkgReqName := fmt.Sprintf("%s-packages", workspace.Name)

	// Check if PackageRequest exists (cluster-scoped)
	pkgReq := &packagesv1.PackageRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: pkgReqName}, pkgReq)

	if apierrors.IsNotFound(err) {
		// PackageRequest doesn't exist
		// Only create if packages are defined
		if len(workspace.Spec.Packages) == 0 {
			logger.Info("No packages defined for workspace and no PackageRequest exists, skipping creation")
			return nil
		}

		// Create new PackageRequest (cluster-scoped, no namespace)
		pkgReq = &packagesv1.PackageRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: pkgReqName,
				// Do NOT set Namespace field for cluster-scoped resources
			},
			Spec: packagesv1.PackageRequestSpec{
				WorkspaceRef: workspace.Name,
				Packages:     workspace.Spec.Packages,
				ProfileName:  fmt.Sprintf("workspace-%s-packages", workspace.Name),
			},
		}

		// Set owner reference - use manual owner reference since both are cluster-scoped
		// Cannot use controllerutil.SetControllerReference for cluster-scoped resources
		blockOwnerDeletion := false
		controller := true
		ownerRef := metav1.OwnerReference{
			APIVersion:         workspace.APIVersion,
			Kind:               workspace.Kind,
			Name:               workspace.Name,
			UID:                workspace.UID,
			Controller:         &controller,
			BlockOwnerDeletion: &blockOwnerDeletion,
		}
		pkgReq.OwnerReferences = []metav1.OwnerReference{ownerRef}

		if err := r.Create(ctx, pkgReq); err != nil {
			return fmt.Errorf("failed to create PackageRequest: %w", err)
		}
		logger.Info("Created PackageRequest", zap.String("name", pkgReqName))
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get PackageRequest: %w", err)
	}

	// PackageRequest exists
	// If workspace has no packages, delete the PackageRequest
	// (CRD validation requires MinItems=1, so we can't set empty list)
	if len(workspace.Spec.Packages) == 0 {
		logger.Info("No packages defined for workspace, deleting PackageRequest", zap.String("name", pkgReqName))
		if err := r.Delete(ctx, pkgReq); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete PackageRequest: %w", err)
		}
		logger.Info("Deleted PackageRequest", zap.String("name", pkgReqName))
		return nil
	}

	// Check if packages changed
	packagesChanged := false
	if len(pkgReq.Spec.Packages) != len(workspace.Spec.Packages) {
		packagesChanged = true
	} else {
		for i, pkg := range workspace.Spec.Packages {
			if pkgReq.Spec.Packages[i].Name != pkg.Name ||
				pkgReq.Spec.Packages[i].Channel != pkg.Channel ||
				pkgReq.Spec.Packages[i].NixpkgsCommit != pkg.NixpkgsCommit {
				packagesChanged = true
				break
			}
		}
	}

	if packagesChanged {
		pkgReq.Spec.Packages = workspace.Spec.Packages
		if err := r.Update(ctx, pkgReq); err != nil {
			return fmt.Errorf("failed to update PackageRequest: %w", err)
		}

		// PackageManagerReconciler will be triggered by the spec change
		// and will handle the status update (don't update status here to avoid race conditions)
		logger.Info("Updated PackageRequest with new packages", zap.String("name", pkgReqName))
	}

	return nil
}

// syncPackageStatus syncs package installation status from PackageRequest to Workspace
func (r *WorkspaceReconciler) syncPackageStatus(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// Only sync if packages are defined
	if len(workspace.Spec.Packages) == 0 {
		return nil
	}

	// Get the PackageRequest (cluster-scoped)
	pkgReqName := fmt.Sprintf("%s-packages", workspace.Name)
	pkgReq := &packagesv1.PackageRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: pkgReqName}, pkgReq)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// PackageRequest not yet created
			return nil
		}
		return fmt.Errorf("failed to get PackageRequest: %w", err)
	}

	// Copy package status from PackageRequest to Workspace
	workspace.Status.InstalledPackages = pkgReq.Status.InstalledPackages
	workspace.Status.FailedPackages = pkgReq.Status.FailedPackages
	workspace.Status.PackageInstallationMessage = pkgReq.Status.Message

	return nil
}

// ensureWorkspaceService ensures a Service is created for the workspace
func (r *WorkspaceReconciler) ensureWorkspaceService(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	serviceName := fmt.Sprintf("workspace-%s", workspace.Name)

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
	// Map of service-prefix -> (service-port-name, service-port)
	httpServices := map[string]struct {
		portName string
		port     int32
	}{
		"vscode":   {"code-server", 8080},
		"tty":      {"ttyd", 7681},
		"claude":   {"claude-ttyd", 7682},
		"opencode": {"opencode-ttyd", 7683},
		"codex":    {"codex-ttyd", 7684},
	}

	// Service name that ingress will route to
	serviceName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Build Ingress rules
	var ingressRules []networkingv1.IngressRule
	var tlsHosts []string

	for prefix, svc := range httpServices {
		// Use simplified naming without nested periods:
		// {prefix}-{workspace}-{workmachine}.{subdomain}.khost.dev
		// Example: vscode-myworkspace-myworkmachine.subdomain.khost.dev
		host := fmt.Sprintf("%s-%s-%s.%s.khost.dev", prefix, workspace.Name, workspace.Spec.WorkmachineName, domainRequest.Status.Subdomain)
		tlsHosts = append(tlsHosts, host)

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
										Number: svc.port,
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
		ingress.Spec.IngressClassName = fn.Ptr("traefik")
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts: tlsHosts,
				// Use the global wildcard certificate created at server startup
				SecretName: "kloudlite-wildcard-cert-tls",
			},
		}
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
	headlessServiceName := fmt.Sprintf("workspace-%s-headless", workspace.Name)

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
