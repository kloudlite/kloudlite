package router_controller

import (
	"context"
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	CreateIngressResource       string = "create-ingress-resource"

	CleaningUpResources string = "cleaning-up-resourcess"

	certCreatedByRouter string = "kloudlite.io/cert-created-by-router"
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

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: EnsuringHttpsCertsIfEnabled, Title: "Ensuring HTTPS Cert if enabled"},
		{Name: SettingUpBasicAuthIfEnabled, Title: "Setting Up Basic Auth if enabled"},
		{Name: CreateIngressResource, Title: "Creates kubernetes ingress resource"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.EnsuringHttpsCerts(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := r.reconBasicAuth(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIngresses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	check := reconciler.NewRunningCheck("finalizing", req)
	if step := req.CleanupOwnedResources(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) findClusterIssuer(req *reconciler.Request[*crdsv1.Router]) (*certmanagerv1.ClusterIssuer, error) {
	ctx, obj := req.Context(), req.Object
	https := obj.Spec.Https

	if https != nil && https.ClusterIssuer != "" {
		var issuer certmanagerv1.ClusterIssuer
		if err := r.Get(ctx, fn.NN("", https.ClusterIssuer), &issuer, &client.GetOptions{}); err != nil {
			return nil, err
		}

		return &issuer, nil
	}

	var issuerList certmanagerv1.ClusterIssuerList
	if err := r.List(ctx, &issuerList, &client.ListOptions{Limit: 1}); err != nil {
		return nil, nil
	}

	if len(issuerList.Items) != 1 {
		return nil, fmt.Errorf("no cluster issuer found")
	}

	return &issuerList.Items[0], nil
}

func (r *Reconciler) findIngressClass(req *reconciler.Request[*crdsv1.Router]) (string, error) {
	ctx, obj := req.Context(), req.Object

	if obj.Spec.IngressClass != "" {
		return obj.Spec.IngressClass, nil
	}

	var ingressClassList networkingv1.IngressClassList
	if err := r.List(ctx, &ingressClassList, &client.ListOptions{Limit: 1}); err != nil {
		return "", err
	}

	if len(ingressClassList.Items) != 1 {
		return "", fmt.Errorf("no/multiple ingress classes found")
	}

	return ingressClassList.Items[0].Name, nil
}

func isHttpsEnabled(obj *crdsv1.Router) bool {
	return obj.Spec.Https != nil && obj.Spec.Https.Enabled
}

func (r *Reconciler) reconBasicAuth(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(SettingUpBasicAuthIfEnabled, req)

	if obj.Spec.BasicAuth == nil || !obj.Spec.BasicAuth.Enabled {
		return check.Completed()
	}

	if len(obj.Spec.BasicAuth.Username) == 0 {
		return check.Failed(fmt.Errorf(".spec.basicAuth.username must be defined when .spec.basicAuth.enabled is set to true")).Err(nil)
	}

	if obj.Spec.BasicAuth.SecretName == "" {
		obj.Spec.BasicAuth.SecretName = obj.Name + "-basic-auth"
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for router reconcilation"))
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
		basicAuthScrt.StringData = map[string]string{
			"auth":     fmt.Sprintf("%s:%s", obj.Spec.BasicAuth.Username, ePass),
			"username": obj.Spec.BasicAuth.Username,
			"password": password,
		}
		return nil
	}); err != nil {
		return check.StillRunning(err)
	}

	req.AddToOwnedResources(reconciler.ParseResourceRef(basicAuthScrt))

	return check.Completed()
}

func groupHostsByKind(issuer *certmanagerv1.ClusterIssuer, obj *crdsv1.Router) (wildcardHosts []string, nonWildcardHosts []string) {
	var dnsNames []string

	for _, solver := range issuer.Spec.ACME.Solvers {
		if solver.DNS01 != nil {
			if solver.Selector != nil {
				dnsNames = append(dnsNames, solver.Selector.DNSNames...)
			}
		}
	}

	wcFilter := map[string]struct{}{}
	for _, pattern := range dnsNames {
		if strings.HasPrefix(pattern, "*.") {
			wcFilter[pattern[len("*."):]] = struct{}{}
			continue
		}
		wcFilter[pattern] = struct{}{}
	}

	for _, route := range obj.Spec.Routes {
		if _, ok := wcFilter[route.Host]; ok {
			wildcardHosts = append(wildcardHosts, route.Host)
			continue
		}

		idx := strings.IndexByte(route.Host, '.')
		if idx == -1 {
			nonWildcardHosts = append(nonWildcardHosts, route.Host)
			continue
		}

		if _, ok := wcFilter[route.Host[idx+1:]]; ok {
			wildcardHosts = append(wildcardHosts, route.Host)
			continue
		}

		nonWildcardHosts = append(nonWildcardHosts, route.Host)
	}

	return wildcardHosts, nonWildcardHosts
}

func (r *Reconciler) ensureIngresses(req *reconciler.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(CreateIngressResource, req)

	if obj.Spec.IngressClass == "" {
		ingClass, err := r.findIngressClass(req)
		if err != nil {
			return check.Failed(err)
		}

		obj.Spec.IngressClass = ingClass
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("updating .spec.ingressClass"))
	}

	if obj.Spec.Https != nil && obj.Spec.Https.ClusterIssuer == "" {
		issuer, err := r.findClusterIssuer(req)
		if err != nil {
			return check.Failed(err)
		}

		obj.Spec.Https.ClusterIssuer = issuer.Name
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("updating .spec.https.clusterIssuer"))
	}

	if len(obj.Spec.Routes) == 0 {
		return check.Completed()
	}

	issuer, err := r.findClusterIssuer(req)
	if err != nil {
		return check.Failed(err)
	}

	wcHosts, nonWcHosts := groupHostsByKind(issuer, obj)

	nginxIngressAnnotations := GenNginxIngressAnnotations(obj)

	// b, err := templates.ParseBytes(r.templateIngress, templates.IngressTemplateArgs{
	// 	Metadata: metav1.ObjectMeta{
	// 		Name:        obj.Name,
	// 		Namespace:   obj.Namespace,
	// 		Labels:      obj.GetLabels(),
	// 		Annotations: nginxIngressAnnotations,
	// 	},
	// 	IngressClassName:   obj.Spec.IngressClass,
	// 	HttpsEnabled:       isHttpsEnabled(obj),
	// 	WildcardDomains:    wcDomains,
	// 	NonWildcardDomains: nonWcDomains,
	// 	Routes:             obj.Spec.Routes,
	// })

	b, err := templates.ParseBytes(
		r.templateIngress, map[string]any{
			"name":      obj.Name,
			"namespace": obj.Namespace,

			"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"labels":      obj.GetLabels(),
			"annotations": nginxIngressAnnotations,

			"non-wildcard-domains": nonWcHosts,
			"wildcard-domains":     wcHosts,
			"ingress-class":        obj.Spec.IngressClass,

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

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YAMLClient == nil {
		return fmt.Errorf(".YAMLClient must be set")
	}

	var err error
	r.templateIngress, err = templates.Read(templates.IngressTemplate)
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
