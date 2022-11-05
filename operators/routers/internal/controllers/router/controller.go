package router

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/kubectl"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/routers/internal/controllers"
	"operators.kloudlite.io/operators/routers/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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
	IngressReady string = "ingress-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.Router{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := req.EnsureChecks(IngressReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconBasicAuth(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconIngresses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	return req.Finalize()
}

type Config struct {
	WildcardDomains []string `json:"wildcard-domains"`
}

func (r *Reconciler) parseAccountDomains(ctx context.Context, accountRef string) []string {
	klAcc := fn.NewUnstructured(constants.KloudliteAccountType)
	if err := r.Get(ctx, fn.NN("", accountRef), klAcc); err != nil {
		return nil
	}
	b, err := json.Marshal(klAcc.Object)
	if err != nil {
		return nil
	}

	var j struct {
		Spec struct {
			OwnedDomains []string `json:"ownedDomains,omitempty"`
		}
	}
	if err := json.Unmarshal(b, &j); err != nil {
		return nil
	}
	return j.Spec.OwnedDomains
}

func (r *Reconciler) readFromProjectConfig(req *rApi.Request[*crdsv1.Router]) Config {
	ctx, obj := req.Context(), req.Object

	var rcfg Config

	projectCfg := &corev1.ConfigMap{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, "project-config"), projectCfg); err != nil {
		return rcfg
	}
	if err := yaml.Unmarshal([]byte(projectCfg.Data["router"]), &rcfg); err != nil {
		return rcfg
	}
	return rcfg
}

func isNonWildcardDomain(wildcardDomains []string, domain string) bool {
	sp := strings.SplitN(domain, ".", 2)
	if len(sp) != 2 {
		return true
	}

	for i := range wildcardDomains {
		if wildcardDomains[i] == sp[1] {
			return false
		}
	}

	return true
}

func (r *Reconciler) reconBasicAuth(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	// ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	// check := rApi.Check{Generation: obj.Generation}

	return req.Next()
}

func (r *Reconciler) reconIngresses(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, router, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: router.Generation}

	accRef := func() string {
		v, ok := router.GetLabels()[constants.AccountRef]
		if ok {
			return v
		}
		v, ok = router.GetAnnotations()[constants.AccountRef]
		if ok {
			return v
		}
		return ""
	}()

	if accRef == "" {
		return req.CheckFailed(IngressReady, check, "could not get account-ref from labels/annotations").Err(nil)
	}

	routerCfg := r.readFromProjectConfig(req)
	accDomains := r.parseAccountDomains(ctx, accRef)

	wcDomains := make([]string, 0, len(routerCfg.WildcardDomains)+len(accDomains))
	for _, v := range routerCfg.WildcardDomains {
		if strings.HasPrefix(v, "*.") {
			wcDomains = append(wcDomains, v[2:])
		}
	}
	for _, v := range accDomains {
		if strings.HasPrefix(v, "*.") && strings.HasSuffix(v, ".kloudlite.io") {
			wcDomains = append(wcDomains, v[2:])
		}
	}

	nonWildCardDomains := make([]string, 0, len(router.Spec.Domains))
	for i := range router.Spec.Domains {
		if isNonWildcardDomain(wcDomains, router.Spec.Domains[i]) {
			nonWildCardDomains = append(nonWildCardDomains, router.Spec.Domains[i])
		}
	}

	if len(router.Spec.Routes) == 0 {
		return req.CheckFailed(IngressReady, check, "no routes specified in ingress resource").Err(nil)
	}

	// var ingressList []string

	lambdaGroups := map[string][]crdsv1.Route{}
	var appRoutes []crdsv1.Route

	for _, route := range router.Spec.Routes {
		// if !strings.HasSuffix(route.Path, "/") {
		// 	route.Path = route.Path + "/"
		// }
		//
		if route.Lambda != "" {
			if _, ok := lambdaGroups[route.Lambda]; !ok {
				lambdaGroups[route.Lambda] = []crdsv1.Route{}
			}

			// ingressList = append(ingressList, fmt.Sprintf("r-%s-lambda-%s", router.Name, route.Lambda))
			lambdaGroups[route.Lambda] = append(lambdaGroups[route.Lambda], route)
		}

		if route.App != "" {
			// ingressList = append(ingressList, router.Name)
			appRoutes = append(appRoutes, route)
		}
	}

	var kubeYamls [][]byte

	for lName, lRoutes := range lambdaGroups {
		ingName := fmt.Sprintf("r-%s-lambda-%s", router.Name, lName)

		b, err := templates.Parse(
			templates.CoreV1.Ingress, map[string]any{
				"name":       ingName,
				"namespace":  router.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(router, true)},

				"domains":          nonWildCardDomains,
				"wildcard-domains": wcDomains,

				"router-ref":       router,
				"routes":           lRoutes,
				"virtual-hostname": fmt.Sprintf("%s.%s", lName, router.Namespace),

				"ingress-class": func() string {
					if router.Spec.Region == "" {
						return "ingress-nginx"
					}
					return controllers.GetIngressClassName(router.Spec.Region)
				}(),

				"cluster-issuer":     controllers.GetClusterIssuerName(router.Spec.Region),
				"cert-ingress-class": controllers.GetIngressClassName(router.Spec.Region),
			},
		)
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
				"name":             router.Name,
				"namespace":        router.Namespace,
				"owner-refs":       []metav1.OwnerReference{fn.AsOwner(router, true)},
				"domains":          nonWildCardDomains,
				"wildcard-domains": wcDomains,
				"labels": map[string]any{
					constants.RouterNameKey: router.Name,
				},
				"router-ref": router,
				"routes":     appRoutes,

				"ingress-class": func() string {
					if router.Spec.Region == "" {
						return "ingress-nginx"
					}
					return controllers.GetIngressClassName(router.Spec.Region)
				}(),
				"cluster-issuer":     controllers.GetClusterIssuerName(router.Spec.Region),
				"cert-ingress-class": controllers.GetIngressClassName(router.Spec.Region),
				// "virtual-hostname":   fmt.Sprintf("%s.%s", router.Name, router.Namespace),
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
	return builder.Complete(r)
}
