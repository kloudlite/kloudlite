package workspace

import (
	"context"
	"fmt"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ensurePackageRequest creates or updates PackageRequest for the workspace
func (r *WorkspaceReconciler) ensurePackageRequest(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	pkgReqName := fmt.Sprintf("%s-packages", workspace.Name)

	// Check if PackageRequest exists
	pkgReq := &workspacev1.PackageRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: pkgReqName, Namespace: workspace.Namespace}, pkgReq)

	if apierrors.IsNotFound(err) {
		// PackageRequest doesn't exist
		// Only create if packages are defined
		if len(workspace.Spec.Packages) == 0 {
			logger.Info("No packages defined for workspace and no PackageRequest exists, skipping creation")
			return nil
		}

		// Create new PackageRequest
		pkgReq = &workspacev1.PackageRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pkgReqName,
				Namespace: workspace.Namespace,
			},
			Spec: workspacev1.PackageRequestSpec{
				WorkspaceRef: workspace.Name,
				Packages:     workspace.Spec.Packages,
				ProfileName:  fmt.Sprintf("workspace-%s-packages", workspace.Name),
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(workspace, pkgReq, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

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

	// Get the PackageRequest
	pkgReqName := fmt.Sprintf("%s-packages", workspace.Name)
	pkgReq := &workspacev1.PackageRequest{}
	err := r.Get(ctx, client.ObjectKey{Name: pkgReqName, Namespace: workspace.Namespace}, pkgReq)
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

	// Check if Service exists
	svc := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: workspace.Namespace}, svc)

	if apierrors.IsNotFound(err) {
		// Create new Service
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: workspace.Namespace,
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
						Name:       "vscode-tunnel",
						Protocol:   corev1.ProtocolTCP,
						Port:       8000,
						TargetPort: intstr.FromInt(8000),
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

// ensureWorkspaceHeadlessService ensures a headless Service is created for the workspace
// This headless service is used by service intercepts to route traffic to workspace pods
func (r *WorkspaceReconciler) ensureWorkspaceHeadlessService(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	headlessServiceName := fmt.Sprintf("workspace-%s-headless", workspace.Name)

	// Check if headless Service exists
	headlessSvc := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: headlessServiceName, Namespace: workspace.Namespace}, headlessSvc)

	if apierrors.IsNotFound(err) {
		// Create new headless Service
		headlessSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      headlessServiceName,
				Namespace: workspace.Namespace,
				Labels: map[string]string{
					"app":                           "workspace",
					"workspace":                     workspace.Name,
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