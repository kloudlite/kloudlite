package domainrequest

import (
	"context"
	"fmt"
	"time"

	domainrequestsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	domainRequestFinalizer = "domains.kloudlite.io/finalizer"
	consoleAPIBaseURL      = "https://console.kloudlite.io"
)

// DomainRequestReconciler reconciles DomainRequest objects
type DomainRequestReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Logger             *zap.Logger
	InstallationKey    string
	InstallationSecret string
}

// Reconcile handles DomainRequest reconciliation
func (r *DomainRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("name", req.Name),
	)

	logger.Info("Reconciling DomainRequest")

	// Fetch the DomainRequest instance (cluster-scoped)
	domainRequest := &domainrequestsv1.DomainRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, domainRequest); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("DomainRequest resource not found, ignoring")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get DomainRequest", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Handle deletion
	if !domainRequest.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, domainRequest, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(domainRequest, domainRequestFinalizer) {
		controllerutil.AddFinalizer(domainRequest, domainRequestFinalizer)
		if err := r.Update(ctx, domainRequest); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Added finalizer to DomainRequest")
	}

	// Reconcile based on current state
	switch domainRequest.Status.State {
	case "", "Pending":
		// New flow: Download installation's origin certificate first
		return r.handleOriginCertificateDownload(ctx, domainRequest, logger)
	case "CertificateDownloading":
		return r.handleOriginCertificateDownload(ctx, domainRequest, logger)
	case "CertificateGenerated":
		// Create HAProxy with landing page (or configured backend)
		return r.handleHAProxyCreation(ctx, domainRequest, logger)
	case "HAProxyCreating":
		return r.handleHAProxyStatusCheck(ctx, domainRequest, logger)
	case "HAProxyReady":
		// HAProxy is ready, now configure DNS/IP
		return r.handleIPRegistration(ctx, domainRequest, logger)
	case "IPRegistering":
		return r.handleIPRegistration(ctx, domainRequest, logger)
	case "Ready":
		// Check if HAProxy pod still exists
		if domainRequest.Status.HAProxyPodName != "" {
			pod := &corev1.Pod{}
			err := r.Get(ctx, client.ObjectKey{
				Name:      domainRequest.Status.HAProxyPodName,
				Namespace: domainRequest.Spec.WorkloadNamespace,
			}, pod)

			if errors.IsNotFound(err) {
				// HAProxy pod was deleted, need to recreate it
				logger.Warn("HAProxy pod deleted while DomainRequest is Ready, recreating",
					zap.String("podName", domainRequest.Status.HAProxyPodName))

				// Reset HAProxyReady status and trigger recreation
				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
					domainRequest.Status.State = "HAProxyCreating"
					domainRequest.Status.Message = "HAProxy pod was deleted, recreating"
					domainRequest.Status.HAProxyReady = false
					return nil
				}, logger); err != nil {
					logger.Error("Failed to update status for HAProxy recreation", zap.Error(err))
					return reconcile.Result{}, err
				}

				// Requeue immediately to recreate the pod
				return reconcile.Result{Requeue: true}, nil
			} else if err != nil {
				logger.Error("Failed to check HAProxy pod existence", zap.Error(err))
				return reconcile.Result{}, err
			}
		}

		// Ensure image-registry ingress exists when subdomain is assigned
		if err := r.ensureImageRegistryIngress(ctx, domainRequest, logger); err != nil {
			logger.Error("Failed to ensure image-registry ingress", zap.Error(err))
			// Don't fail the reconciliation, just log the error
		}

		// Check if DomainRoutes have changed and need DNS reconciliation
		// With GenerationChangedPredicate, this only runs when spec actually changes
		currentRoutesHash, err := computeDomainRoutesHash(domainRequest.Spec.DomainRoutes)
		if err != nil {
			logger.Error("Failed to compute DomainRoutes hash", zap.Error(err))
			return reconcile.Result{}, err
		}

		if currentRoutesHash != domainRequest.Status.LastReconciledRoutesHash {
			logger.Info("DomainRoutes have changed, updating HAProxy config and DNS records",
				zap.String("previousHash", domainRequest.Status.LastReconciledRoutesHash),
				zap.String("currentHash", currentRoutesHash))

			// Update HAProxy ConfigMap with new routing rules
			if err := r.createHAProxyConfigMap(ctx, domainRequest, logger); err != nil {
				logger.Error("Failed to update HAProxy ConfigMap", zap.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("HAProxy ConfigMap updated successfully, triggering DNS reconciliation")

			// Transition to IPRegistering state to update DNS records
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
				domainRequest.Status.State = "IPRegistering"
				domainRequest.Status.Message = "DomainRoutes updated, reconciling DNS records and HAProxy config"
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update status for DNS reconciliation", zap.Error(err))
				return reconcile.Result{}, err
			}

			// Requeue immediately to trigger IP registration (which handles DNS)
			return reconcile.Result{Requeue: true}, nil
		}

		logger.Info("DomainRequest is ready, no action needed")
		return reconcile.Result{RequeueAfter: 24 * time.Hour}, nil
	case "Failed":
		// Retry the failed operation after 30 seconds
		logger.Info("DomainRequest failed, determining which step to retry")

		// Check if origin certificate secret exists
		secretName := fmt.Sprintf("%s-origin-cert", domainRequest.Name)
		secret := &corev1.Secret{}
		secretExists := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: domainRequest.Spec.WorkloadNamespace}, secret) == nil

		// Determine which step failed based on what's been completed
		if !secretExists {
			// Origin certificate not downloaded yet
			logger.Info("Retrying origin certificate download")
			return r.handleOriginCertificateDownload(ctx, domainRequest, logger)
		} else if secretExists && domainRequest.Status.HAProxyPodName == "" {
			// Origin cert exists but HAProxy not created
			logger.Info("Origin cert exists but HAProxy not created, retrying HAProxy creation")
			return r.handleHAProxyCreation(ctx, domainRequest, logger)
		} else if domainRequest.Status.HAProxyPodName != "" && !domainRequest.Status.HAProxyReady {
			// HAProxy exists but not ready
			logger.Info("HAProxy exists but not ready, checking status")
			return r.handleHAProxyStatusCheck(ctx, domainRequest, logger)
		} else if domainRequest.Status.HAProxyReady && domainRequest.Status.LastIPRegistrationTime == nil {
			// HAProxy ready but IP not registered
			logger.Info("HAProxy ready but IP not registered, retrying IP registration")
			return r.handleIPRegistration(ctx, domainRequest, logger)
		} else if domainRequest.Status.OriginCertificateSecretName == "" {
			// Origin certificate not yet downloaded
			logger.Info("Origin certificate not downloaded, downloading now")
			return r.handleOriginCertificateDownload(ctx, domainRequest, logger)
		} else {
			// Unknown state or IP registration failed
			logger.Info("Unknown failed state, retrying IP registration")
			return r.handleIPRegistration(ctx, domainRequest, logger)
		}
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *DomainRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainrequestsv1.DomainRequest{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Secret{}).    // Watch Secrets owned by DomainRequest
		Owns(&corev1.Pod{}).       // Watch HAProxy Pods owned by DomainRequest
		Owns(&corev1.ConfigMap{}). // Watch HAProxy ConfigMaps owned by DomainRequest
		Complete(r)
}

// ensureImageRegistryIngress creates or updates the Ingress for image-registry
// This exposes the image-registry at cr.{subdomain} with TLS termination
func (r *DomainRequestReconciler) ensureImageRegistryIngress(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	if domainRequest.Status.Subdomain == "" {
		logger.Debug("Subdomain not yet assigned, skipping image-registry ingress creation")
		return nil
	}

	// Build the host: cr.{subdomain}.khost.dev
	host := fmt.Sprintf("cr.%s.khost.dev", domainRequest.Status.Subdomain)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "image-registry",
			Namespace: "kloudlite",
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		ingress.Labels = map[string]string{
			"app":                     "image-registry",
			"kloudlite.io/managed-by": "domainrequest-controller",
		}
		ingress.Annotations = map[string]string{
			// Large body size for container image layers (0 = unlimited)
			"nginx.ingress.kubernetes.io/proxy-body-size": "0",
		}
		ingress.Spec = networkingv1.IngressSpec{
			IngressClassName: ptr.To("kloudlite"),
			TLS: []networkingv1.IngressTLS{{
				Hosts:      []string{host},
				SecretName: "kloudlite-wildcard-cert-tls",
			}},
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: ptr.To(networkingv1.PathTypePrefix),
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: "image-registry",
									Port: networkingv1.ServiceBackendPort{Number: 5000},
								},
							},
						}},
					},
				},
			}},
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to create/update image-registry ingress", zap.Error(err))
		return err
	}

	logger.Info("Image registry ingress ensured", zap.String("host", host))
	return nil
}
