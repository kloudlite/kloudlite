package crds

import (
	"context"
	"encoding/json"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouterReconciler reconciles a Router object
type RouterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env
}

func (r *RouterReconciler) GetName() string {
	return "router"
}

const (
	KeyIngressResourcesList string = "ingress-resources"
)

const (
	IngressExistsCondition string = "ingress.exists/%v"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/finalizers,verbs=update

func (r *RouterReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &crdsv1.Router{})

	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// STEP: cleaning up last run, clearing opsConditions
	if len(req.Object.Status.OpsConditions) > 0 {
		req.Object.Status.OpsConditions = []metav1.Condition{}
		return ctrl.Result{RequeueAfter: 0}, r.Status().Update(ctx, req.Object)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
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

func (r *RouterReconciler) finalize(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
	return req.Finalize()
}

func getIngressResources(router *crdsv1.Router) []string {
	items, ok := router.Status.DisplayVars.Get(KeyIngressResourcesList)
	if !ok {
		return []string{}
	}
	b, err := json.Marshal(items)
	if err != nil {
		return []string{}
	}
	var x []string
	if err := json.Unmarshal(b, &x); err != nil {
		return []string{}
	}
	return x
}

func (r *RouterReconciler) reconcileStatus(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
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

func (r *RouterReconciler) reconcileOperations(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
	router := req.Object

	lambdaGroups := map[string][]crdsv1.Route{}
	var appRoutes []crdsv1.Route

	var ingressList []string

	for _, route := range router.Spec.Routes {
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

			"router-ref":       router,
			"routes":           lMapRoutes,
			"virtual-hostname": fmt.Sprintf("%s.%s", lName, router.Namespace),

			"ingress-class":  r.Env.DefaultIngressClass,
			"cluster-issuer": r.Env.ClusterCertIssuer,

			"wildcard-domain-certificate": r.Env.WildcardDomainCertificate,
			"wildcard-domain-suffix":      r.Env.WildcardDomainSuffix,
		}

		ingressList = append(ingressList, ingName)

		b, err := templates.Parse(templates.CoreV1.Ingress, args)
		if err != nil {
			return req.FailWithOpError(
				errors.NewEf(err, "could not parse (template=%s)", templates.Ingress),
			)
		}
		kubeYamls = append(kubeYamls, b)
	}

	if len(appRoutes) > 0 {
		args := map[string]any{
			"name":      router.Name,
			"namespace": router.Namespace,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(router, true),
			},

			"router-ref": router,
			"routes":     appRoutes,

			"ingress-class":  r.Env.DefaultIngressClass,
			"cluster-issuer": r.Env.ClusterCertIssuer,

			"wildcard-domain-certificate": r.Env.WildcardDomainCertificate,
			"wildcard-domain-suffix":      r.Env.WildcardDomainSuffix,
		}
		ingressList = append(ingressList, router.Name)
		b, err := templates.Parse(templates.CoreV1.Ingress, args)
		if err != nil {
			return req.FailWithOpError(errors.NewEf(err, "could not parse (template=%s)", templates.Ingress))
		}

		kubeYamls = append(kubeYamls, b)
	}

	if _, err := fn.KubectlApplyExec(kubeYamls...); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not apply ingress ingressObj"))
	}

	if !reflect.DeepEqual(getIngressResources(router), ingressList) {
		if err := router.Status.DisplayVars.Set(KeyIngressResourcesList, ingressList); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(req.Context(), router); err != nil {
			return req.FailWithOpError(err)
		}
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Router{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
