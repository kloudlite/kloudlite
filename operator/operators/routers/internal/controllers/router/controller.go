package router

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operators/routers/internal/controllers"
	"operators.kloudlite.io/operators/routers/internal/env"
	"operators.kloudlite.io/pkg/constants"
	"operators.kloudlite.io/pkg/errors"
	fn "operators.kloudlite.io/pkg/functions"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	"operators.kloudlite.io/pkg/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"time"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	IngressReady    string = "ingress-ready"
	BasicAuthReady  string = "basic-auth-ready"
	DefaultsPatched string = "patch-defaults"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.Router{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !req.ShouldReconcile() {
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(IngressReady); !step.ShouldProceed() {
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

	if step := r.reconBasicAuth(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIngresses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.BasicAuth.Enabled && obj.Spec.BasicAuth.SecretName == "" {
		hasUpdated = true
		obj.Spec.BasicAuth.SecretName = obj.Name + "-basic-auth"
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}
	return req.Next()

}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	return req.Finalize()
}

type Config struct {
	WildcardDomains []string `json:"wildcard-domains"`
}

func isBlueprint(obj client.Object) bool {
	return strings.HasSuffix(obj.GetNamespace(), "-blueprint")
}

func (r *Reconciler) reconBasicAuth(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(BasicAuthReady)
	defer req.LogPostCheck(BasicAuthReady)

	if obj.Spec.BasicAuth.Enabled {
		basicAuthScrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.BasicAuth.SecretName, Namespace: obj.Namespace}, Type: "Opaque"}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, basicAuthScrt, func() error {
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
			return req.CheckFailed(BasicAuthReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[BasicAuthReady] {
		checks[BasicAuthReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureIngresses(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IngressReady)
	defer req.LogPostCheck(IngressReady)

	if len(obj.Spec.Routes) == 0 {
		return req.CheckFailed(IngressReady, check, "no routes specified in ingress resource").Err(nil)
	}

	issuerName := controllers.GetClusterIssuerName(obj.Spec.Region)
	clusterIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", issuerName), &certmanagerv1.ClusterIssuer{})
	if err != nil {
		return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
	}

	wcDomainsMap := make(map[string]bool, 2)
	wcDomains := make([]string, 0, 2)

	for _, solver := range clusterIssuer.Spec.ACME.Solvers {
		if solver.DNS01 != nil {
			for _, dnsName := range solver.Selector.DNSNames {
				if strings.HasPrefix(dnsName, "*.") {
					wcDomainsMap[dnsName[2:]] = true
				}
				wcDomainsMap[dnsName] = true
				wcDomains = append(wcDomains, dnsName)
			}
		}
	}

	nonWildCardDomains := make([]string, 0, len(obj.Spec.Domains))
	for _, domain := range obj.Spec.Domains {
		if _, ok := wcDomainsMap[domain]; ok {
			continue
		}
		sp := strings.SplitN(domain, ".", 2)
		if len(sp) < 2 {
			continue
		}
		if _, ok := wcDomainsMap[sp[1]]; ok {
			continue
		}
		nonWildCardDomains = append(nonWildCardDomains, domain)
	}

	lambdaGroups := map[string][]crdsv1.Route{}
	var appRoutes []crdsv1.Route

	for _, route := range obj.Spec.Routes {
		if route.Lambda != "" {
			if _, ok := lambdaGroups[route.Lambda]; !ok {
				lambdaGroups[route.Lambda] = []crdsv1.Route{}
			}
			lambdaGroups[route.Lambda] = append(lambdaGroups[route.Lambda], route)
		}

		if route.App != "" {
			if isBlueprint(obj) {
				route.App = "env-route-switcher"
				appRoutes = append(appRoutes, route)
				continue
			}
			appRoutes = append(appRoutes, route)
		}

	}

	var kubeYamls [][]byte

	for lName, lRoutes := range lambdaGroups {
		ingName := fmt.Sprintf("r-%s-lambda-%s", obj.Name, lName)

		vals := map[string]any{
			"name":       ingName,
			"namespace":  obj.Namespace,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"labels": map[string]string{
				constants.RouterNameKey: obj.Name,
			},

			"domains":          nonWildCardDomains,
			"wildcard-domains": wcDomains,

			"router-ref":       obj,
			"routes":           lRoutes,
			"virtual-hostname": fmt.Sprintf("%s.%s", lName, obj.Namespace),

			"ingress-class":      controllers.GetIngressClassName(obj.Spec.Region),
			"cluster-issuer":     controllers.GetClusterIssuerName(obj.Spec.Region),
			"cert-ingress-class": controllers.GetIngressClassName(obj.Spec.Region),
			"is-blueprint":       isBlueprint(obj),
		}

		b, err := templates.Parse(templates.CoreV1.Ingress, vals)
		if err != nil {
			return req.FailWithOpError(
				errors.NewEf(err, "could not parse (template=%s)", templates.Ingress),
			).Err(nil)
		}
		kubeYamls = append(kubeYamls, b)
	}

	if len(appRoutes) > 0 {
		b, err := templates.Parse(
			templates.CoreV1.Ingress, map[string]any{
				"name":             obj.Name,
				"namespace":        obj.Namespace,
				"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"domains":          nonWildCardDomains,
				"wildcard-domains": wcDomains,
				"labels": map[string]any{
					constants.RouterNameKey: obj.Name,
				},
				"router-ref": obj,
				"routes":     appRoutes,

				"ingress-class": func() string {
					if obj.Spec.Region == "" {
						return "ingress-nginx"
					}
					return controllers.GetIngressClassName(obj.Spec.Region)
				}(),
				"is-blueprint":       isBlueprint(obj),
				"cluster-issuer":     controllers.GetClusterIssuerName(obj.Spec.Region),
				"cert-ingress-class": controllers.GetIngressClassName(obj.Spec.Region),
			},
		)
		if err != nil {
			return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
		}

		kubeYamls = append(kubeYamls, b)
	}

	if err := r.yamlClient.ApplyYAML(req.Context(), kubeYamls...); err != nil {
		return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[IngressReady] {
		checks[IngressReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Router{})
	builder.Owns(&networkingv1.Ingress{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
