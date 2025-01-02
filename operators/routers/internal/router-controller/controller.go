package router_controller

import (
	"context"
	"fmt"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/routers/internal/env"
	"github.com/kloudlite/operator/operators/routers/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	apiLabels "k8s.io/apimachinery/pkg/labels"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Name       string
	Env        *env.Env
	YAMLClient kubectl.YAMLClient

	templateIngress []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	IngressReady    string = "ingress-ready"
	BasicAuthReady  string = "basic-auth-ready"
	DefaultsPatched string = "patch-defaults"

	Finalizing         string = "finalizing"
	CheckHttpsCerteady string = "https-cert-ready"

	EnsuringHttpsCertsIfEnabled string = "ensuring-https-certs-if-enabled"
	SettingUpBasicAuthIfEnabled string = "setting-up-basic-auth-if-enabled"
	CreatingIngressResources    string = "creating-ingress-resources"

	CleaningUpResources string = "cleaning-up-resourcess"

	certCreatedByRouter string = "kloudlite.io/cert-created-by-router"
)

var (
	ApplyChecklist = []reconciler.CheckMeta{
		{Name: DefaultsPatched, Title: "Defaults Patched"},
		{Name: EnsuringHttpsCertsIfEnabled, Title: "Ensuring HTTPS Cert if enabled"},
		{Name: SettingUpBasicAuthIfEnabled, Title: "Setting Up Basic Auth if enabled"},
	}

	DeleteChecklist = []reconciler.CheckMeta{
		{Name: CleaningUpResources, Title: "Cleaning Up Resources"},
	}
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.Router{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !req.ShouldReconcile() {
		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(ApplyChecklist); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.EnsuringHttpsCerts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconBasicAuth(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIngresses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) patchDefaults(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(DefaultsPatched, req)

	hasUpdate := false

	if obj.Spec.BasicAuth != nil && obj.Spec.BasicAuth.Enabled && obj.Spec.BasicAuth.SecretName == "" {
		hasUpdate = true
		obj.Spec.BasicAuth.SecretName = obj.Name + "-basic-auth"
	}

	if obj.Spec.IngressClass == "" {
		hasUpdate = true
		obj.Spec.IngressClass = r.Env.DefaultIngressClass
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	check := reconciler.NewRunningCheck("finalizing", req)
	if step := req.CleanupOwnedResources(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func genTLSCertName(domain string) string {
	return fmt.Sprintf("%s-tls", domain)
}

func (r *Reconciler) getRouterClusterIssuer(obj *crdsv1.Router) string {
	https := obj.Spec.Https

	if https != nil && https.ClusterIssuer != "" {
		return https.ClusterIssuer
	}

	return r.Env.DefaultClusterIssuer
}

func isHttpsEnabled(obj *crdsv1.Router) bool {
	return obj.Spec.Https != nil && obj.Spec.Https.Enabled
}

func (r *Reconciler) EnsuringHttpsCerts(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(EnsuringHttpsCertsIfEnabled, req)

	if !isHttpsEnabled(obj) {
		return check.Completed()
	}

	_, nonWildcardDomains, err := r.parseAndExtractDomains(req)
	if err != nil {
		return check.Failed(err)
	}

	for _, domain := range nonWildcardDomains {

		tlsCertLabel := fmt.Sprintf("kloudlite.io/tls-cert.%s", fn.Md5([]byte(genTLSCertName(domain))))
		if v, ok := obj.Labels[tlsCertLabel]; !ok || v != "true" {
			fn.MapSet(&obj.Labels, tlsCertLabel, "true")
			if err := r.Update(ctx, obj); err != nil {
				return check.StillRunning(err)
			}
		}

		cert, err := reconciler.Get(ctx, r.Client, fn.NN(r.Env.CertificateNamespace, genTLSCertName(domain)), &certmanagerv1.Certificate{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.StillRunning(err)
			}
			cert = nil
		}

		if cert == nil {
			cert := &certmanagerv1.Certificate{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Certificate",
					APIVersion: certmanagerv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      genTLSCertName(domain),
					Namespace: r.Env.CertificateNamespace,
					Labels: map[string]string{
						certCreatedByRouter: obj.Name,
					},
				},
				Spec: certmanagerv1.CertificateSpec{
					DNSNames: []string{domain},
					IssuerRef: certmanagermetav1.ObjectReference{
						Name:  r.getRouterClusterIssuer(obj),
						Kind:  "ClusterIssuer",
						Group: certmanagerv1.SchemeGroupVersion.Group,
					},
					RenewBefore: &metav1.Duration{
						Duration: 15 * 24 * time.Hour, // 15 days prior
					},
					SecretName: genTLSCertName(domain),
					Usages: []certmanagerv1.KeyUsage{
						certmanagerv1.UsageDigitalSignature,
						certmanagerv1.UsageKeyEncipherment,
					},
				},
			}
			if err := r.Create(ctx, cert); err != nil {
				return check.StillRunning(err)
			}
		}

		if _, err := IsHttpsCertReady(cert); err != nil {
			return check.StillRunning(err).NoRequeue()
			// return check.StillRunning(err)
			// return check.StillRunning(err).RequeueAfter(1 * time.Second)
		}

		certSecret, err := reconciler.Get(ctx, r.Client, fn.NN(r.Env.CertificateNamespace, genTLSCertName(domain)), &corev1.Secret{})
		if err != nil {
			return check.StillRunning(err)
		}

		copyTLSSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: genTLSCertName(domain), Namespace: obj.Namespace}, Type: corev1.SecretTypeTLS}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, copyTLSSecret, func() error {
			if copyTLSSecret.Annotations == nil {
				copyTLSSecret.Annotations = make(map[string]string, 1)
			}
			copyTLSSecret.Annotations["kloudlite.io/secret.cloned-by"] = "router"

			copyTLSSecret.Data = certSecret.Data
			copyTLSSecret.StringData = certSecret.StringData
			return nil
		}); err != nil {
			return check.StillRunning(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) reconBasicAuth(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(SettingUpBasicAuthIfEnabled, req)

	if obj.Spec.BasicAuth != nil && obj.Spec.BasicAuth.Enabled {
		if len(obj.Spec.BasicAuth.Username) == 0 {
			return check.Failed(fmt.Errorf(".spec.basicAuth.username must be defined when .spec.basicAuth.enabled is set to true")).Err(nil)
		}

		basicAuthScrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.BasicAuth.SecretName, Namespace: obj.Namespace}, Type: "Opaque"}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, basicAuthScrt, func() error {
			basicAuthScrt.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
			if _, ok := basicAuthScrt.Data["password"]; ok {
				return nil
			}

			password := fn.CleanerNanoid(48)
			ePass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			basicAuthScrt.Data = map[string][]byte{
				"auth":     []byte(fmt.Sprintf("%s:%s", obj.Spec.BasicAuth.Username, ePass)),
				"username": []byte(obj.Spec.BasicAuth.Username),
				"password": []byte(password),
			}
			return nil
		}); err != nil {
			return check.StillRunning(err)
		}

		req.AddToOwnedResources(reconciler.ParseResourceRef(basicAuthScrt))
	}

	return check.Completed()
}

func (r *Reconciler) ensureIngresses(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(CreatingIngressResources, req)

	wcDomains, nonWcDomains, err := r.parseAndExtractDomains(req)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	nginxIngressAnnotations := GenNginxIngressAnnotations(obj)

	if len(obj.Spec.Routes) > 0 {
		b, err := templates.ParseBytes(
			r.templateIngress, map[string]any{
				"name":      obj.Name,
				"namespace": obj.Namespace,

				"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"labels":      obj.GetLabels(),
				"annotations": nginxIngressAnnotations,

				"non-wildcard-domains": nonWcDomains,
				"wildcard-domains":     wcDomains,
				"router-domains":       obj.Spec.Domains,

				"ingress-class": obj.Spec.IngressClass,
				"cluster-issuer": func() string {
					if obj.Spec.Https != nil && obj.Spec.Https.ClusterIssuer != "" {
						return obj.Spec.Https.ClusterIssuer
					}
					return r.Env.DefaultClusterIssuer
				}(),

				"routes": obj.Spec.Routes,

				"is-https-enabled": isHttpsEnabled(obj),
			},
		)
		if err != nil {
			return check.Failed(err).Err(nil)
		}

		rr, err := r.YAMLClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.StillRunning(err)
		}

		req.AddToOwnedResources(rr...)
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YAMLClient == nil {
		return fmt.Errorf("r.YAMLClient must be set")
	}

	var err error
	r.templateIngress, err = templates.ReadIngressTemplate()
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Router{})
	builder.Owns(&networkingv1.Ingress{})

	builder.Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		var routersList crdsv1.RouterList
		if err := r.List(ctx, &routersList, &client.ListOptions{
			LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
				fmt.Sprintf("kloudlite.io/tls-cert.%s", fn.Md5([]byte(obj.GetName()))): "true",
			}),
		}); err != nil {
			return nil
		}

		rr := make([]reconcile.Request, 0, len(routersList.Items))
		for i := range routersList.Items {
			rr = append(rr, reconcile.Request{NamespacedName: fn.NN(routersList.Items[i].GetNamespace(), routersList.Items[i].GetName())})
		}

		return rr
	}))
	// builder.Owns(&certmanagerv1.Certificate{})

	builder.WithEventFilter(reconciler.ReconcileFilter())
	return builder.Complete(r)
}
