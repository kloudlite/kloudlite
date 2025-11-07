package workmachine

import (
	"fmt"
	"time"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// syncDomainRequest creates or updates the DomainRequest with the latest WorkMachine IP
// This runs on every reconcile to keep the DomainRequest in sync
func (r *WorkMachineReconciler) syncDomainRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// First check if DomainRequest exists and has the correct IP
	domainRequest := &domainrequestv1.DomainRequest{}
	err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, domainRequest)

	if err != nil && !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Always update/create DomainRequest to sync workspace routes
	// CreateOrUpdate will handle both creation and updates efficiently
	return r.createDomainRequest(check, obj)
}

// createDomainRequest creates or updates the DomainRequest with workspace routes
func (r *WorkMachineReconciler) createDomainRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Fetch subdomain from installation DomainRequest instead of env var
	// DomainRequest is cluster-scoped
	installationDR := &domainrequestv1.DomainRequest{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: "installation-domain"}, installationDR); err != nil {
		return check.Errored(fmt.Errorf("failed to get installation DomainRequest: %w", err)).RequeueAfter(5 * time.Second)
	}

	subDomain := installationDR.Status.Subdomain
	if subDomain == "" {
		return check.Errored(fmt.Errorf("installation subdomain not yet configured")).RequeueAfter(5 * time.Second)
	}

	// List all cluster-scoped workspaces and filter by WorkmachineName
	var wsList workspacev1.WorkspaceList
	if err := r.List(check.Context(), &wsList); err != nil {
		return check.Failed(err)
	}

	// Filter workspaces owned by this WorkMachine
	var ownedWorkspaces []workspacev1.Workspace
	for _, ws := range wsList.Items {
		if ws.Spec.WorkmachineName == obj.Name {
			ownedWorkspaces = append(ownedWorkspaces, ws)
		}
	}

	var domainRoutes []domainrequestv1.DomainRoute
	for _, ws := range ownedWorkspaces {
		serviceName := "workspace-" + ws.Name + "-headless"
		domainRoutes = append(domainRoutes,
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("vscode-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      8080,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("tty-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7681,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("claude-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7682,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("opencode-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7683,
			},
			domainrequestv1.DomainRoute{
				Domain:           fmt.Sprintf("codex-%s.%s.%s", ws.Name, obj.Name, subDomain),
				ServiceName:      serviceName,
				ServiceNamespace: obj.Spec.TargetNamespace,
				ServicePort:      7684,
			},
		)
	}

	// DomainRequest is cluster-scoped, with workloads running in shared workloadNamespace
	domainRequest := &domainrequestv1.DomainRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: obj.Name,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, domainRequest, func() error {
		// Set WorkMachine as owner for cascading deletion
		// MUST be set inside the mutate function to ensure it's preserved on updates
		blockOwnerDeletion := false
		controller := true
		ownerRef := metav1.OwnerReference{
			APIVersion:         obj.APIVersion,
			Kind:               obj.Kind,
			Name:               obj.Name,
			UID:                obj.UID,
			Controller:         &controller,
			BlockOwnerDeletion: &blockOwnerDeletion,
		}
		domainRequest.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

		hostManagerName := fmt.Sprintf("hm-%s", obj.Name)
		domainRequest.Spec = domainrequestv1.DomainRequestSpec{
			NodeName:          obj.Name,
			WorkloadNamespace: "kloudlite-ingress", // Shared namespace for all DomainRequest workloads
			IPAddress:         obj.Status.PublicIP,
			CertificateScope:  "workmachine",
			OriginCertificateHostnames: []string{
				fmt.Sprintf("%s.%s", obj.Name, subDomain),
				fmt.Sprintf("*.%s.%s", obj.Name, subDomain),
			},
			SSHBackend: &domainrequestv1.IngressBackendConfig{
				ServiceName:      hostManagerName,
				ServiceNamespace: hostManagerNamespace,
				ServicePort:      22,
			},
			DomainRoutes: domainRoutes,
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Store the DNS host in WorkMachine status if available
	// Don't block on DomainRequest readiness - it will become ready asynchronously
	// This allows the WorkMachine to proceed with cloud machine setup
	if domainRequest.Status.Domain != "" {
		obj.Status.DNSHost = domainRequest.Status.Domain
	}

	return check.Passed()
}

// deleteDomainRequest deletes the DomainRequest associated with the WorkMachine
func (r *WorkMachineReconciler) deleteDomainRequest(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// DomainRequest is cluster-scoped
	domainRequest := &domainrequestv1.DomainRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: obj.Name,
		},
	}

	if err := r.Delete(check.Context(), domainRequest); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
		// Already deleted, that's fine
	}

	return check.Passed()
}
