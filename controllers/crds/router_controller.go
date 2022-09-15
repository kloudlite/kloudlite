package crds

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// RouterReconciler reconciles a Router object
type RouterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	env    *env.Env
	logger logging.Logger
	Name   string
}

func (r *RouterReconciler) GetName() string {
	return r.Name
}

const (
	KeyIngressResourcesList string = "ingress-resources"
)

const (
	IngressExistsCondition string = "ingress.exists/%v"
)

type RouterConfig struct {
	WildcardDomains []string `json:"wildcard-domains"`
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/finalizers,verbs=update

func (r *RouterReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &crdsv1.Router{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *RouterReconciler) finalize(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	return req.Finalize()
}

func getIngressResources(router *crdsv1.Router) []string {
	var items []string
	if err := router.Status.DisplayVars.Get(KeyIngressResourcesList, &items); err != nil {
		return []string{}
	}
	return items
}

func (r *RouterReconciler) reconcileStatus(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx := req.Context()
	router := req.Object

	isReady := false
	var cs []metav1.Condition

	for _, ingressRes := range getIngressResources(router) {
		_, err := rApi.Get(ctx, r.Client, fn.NN(router.Namespace, ingressRes), &networkingv1.Ingress{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return req.FailWithOpError(errors.NewEf(err, "failed to get ingress resource"))
			}
			isReady = false
			cs = append(
				cs,
				conditions.New(fmt.Sprintf(IngressExistsCondition, ingressRes), false, conditions.NotFound, err.Error()),
			)
		} else {
			cs = append(
				cs, conditions.New(
					fmt.Sprintf(IngressExistsCondition, ingressRes),
					true,
					conditions.Found,
					fmt.Sprintf("Ingress: %v exists", ingressRes),
				),
			)
		}
	}

	nConditions, hasUpdated, err := conditions.Patch(router.Status.Conditions, cs)
	if err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not patch conditions"))
	}

	if !hasUpdated && isReady == router.Status.IsReady {
		return req.Next()
	}

	router.Status.IsReady = isReady
	router.Status.Conditions = nConditions

	if err := r.Status().Update(ctx, router); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *RouterReconciler) readFromProjectConfig(req *rApi.Request[*crdsv1.Router]) RouterConfig {
	ctx, obj := req.Context(), req.Object

	var rcfg RouterConfig

	projectCfg := &corev1.ConfigMap{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, "project-config"), projectCfg); err != nil {
		return rcfg
	}
	if err := yaml.Unmarshal([]byte(projectCfg.Data["router"]), &rcfg); err != nil {
		return rcfg
	}
	return rcfg
}

func (r *RouterReconciler) reconcileOperations(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	router := req.Object

	accRef := router.GetLabels()[constants.AccountRef]
	if accRef == "" {
		return req.FailWithOpError(fmt.Errorf("label %s must be present in resource", constants.AccountRef)).Err(nil)
	}

	routerCfg := r.readFromProjectConfig(req)

	wcDomains := make(map[string]bool, len(routerCfg.WildcardDomains))
	for i := range routerCfg.WildcardDomains {
		wcDomains[routerCfg.WildcardDomains[i]] = true
	}

	nonWildCardDomains := make([]string, 0, len(router.Spec.Domains))
	for i := range router.Spec.Domains {
		if v, ok := wcDomains[router.Spec.Domains[i]]; ok && v {
			nonWildCardDomains = append(nonWildCardDomains, router.Spec.Domains[i])
		}
	}

	lambdaGroups := map[string][]crdsv1.Route{}
	var appRoutes []crdsv1.Route

	var ingressList []string

	for _, route := range router.Spec.Routes {
		if !strings.HasSuffix(route.Path, "/") {
			route.Path = route.Path + "/"
		}
		if s := route.Lambda; s != "" {
			if _, ok := lambdaGroups[route.Lambda]; !ok {
				lambdaGroups[route.Lambda] = []crdsv1.Route{}
			}

			lambdaGroups[route.Lambda] = append(lambdaGroups[route.Lambda], route)
		}

		if s := route.App; s != "" {
			appRoutes = append(appRoutes, route)
		}
	}

	var kubeYamls [][]byte

	for lName, lMapRoutes := range lambdaGroups {
		ingName := fmt.Sprintf("r-%s-lambda-%s", router.Name, lName)
		args := map[string]any{
			"name":       ingName,
			"namespace":  router.Namespace,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(router, true)},

			"domains":          nonWildCardDomains,
			"wildcard-domains": routerCfg.WildcardDomains,

			"router-ref":       router,
			"routes":           lMapRoutes,
			"virtual-hostname": fmt.Sprintf("%s.%s", lName, router.Namespace),

			"ingress-class":  "ingress-nginx-" + accRef,
			"cluster-issuer": r.env.ClusterCertIssuer,
		}

		ingressList = append(ingressList, ingName)

		b, err := templates.Parse(templates.CoreV1.Ingress, args)
		if err != nil {
			return req.FailWithOpError(err).Err(nil)
		}
		if err != nil {
			return req.FailWithOpError(
				errors.NewEf(err, "could not parse (template=%s)", templates.Ingress),
			).Err(nil)
		}
		kubeYamls = append(kubeYamls, b)
	}

	if len(appRoutes) > 0 {
		args := map[string]any{
			"name":             router.Name,
			"namespace":        router.Namespace,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(router, true)},
			"domains":          nonWildCardDomains,
			"wildcard-domains": routerCfg.WildcardDomains,
			"labels": map[string]any{
				constants.RouterNameKey: router.Name,
			},
			"router-ref": router,
			"routes":     appRoutes,

			"ingress-class":  "ingress-nginx-" + accRef,
			"cluster-issuer": r.env.ClusterCertIssuer,
		}
		ingressList = append(ingressList, router.Name)
		b, err := templates.Parse(templates.CoreV1.Ingress, args)
		if err != nil {
			return req.FailWithOpError(err).Err(nil)
		}
		if err != nil {
			return req.FailWithOpError(errors.NewEf(err, "could not parse (template=%s)", templates.Ingress)).Err(nil)
		}

		kubeYamls = append(kubeYamls, b)
	}

	if err := fn.KubectlApplyExec(req.Context(), kubeYamls...); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not apply ingress ingressObj")).Err(nil)
	}

	if !reflect.DeepEqual(getIngressResources(router), ingressList) {
		if err := router.Status.DisplayVars.Set(KeyIngressResourcesList, ingressList); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(req.Context(), router); err != nil {
			return req.FailWithOpError(err)
		}
	}

	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *RouterReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.env = envVars
	r.logger = logger.WithName(r.Name)
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Router{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
