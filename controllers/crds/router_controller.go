package crds

import (
	"context"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/lib/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouterReconciler reconciles a Router object
type RouterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	types.MessageSender
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/finalizers,verbs=update

func (r *RouterReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &crdsv1.Router{})

	if req == nil {
		return ctrl.Result{}, nil
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabels(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *RouterReconciler) finalize(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
	return req.Finalize()
}

func (r *RouterReconciler) reconcileStatus(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
	ctx := req.Context()
	router := req.Object

	isReady := false
	var cs []metav1.Condition

	_, err := rApi.Get(ctx, r.Client, fn.NN(router.Namespace, router.Name), &networkingv1.Ingress{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithOpError(errors.NewEf(err, "failed to get ingress resource"))
		}
		isReady = false
	}

	for idx, domain := range router.Spec.Domains {
		httpReq, err := http.NewRequest(http.MethodHead, fmt.Sprintf("https://%s/", domain), nil)
		if err != nil {
			return req.FailWithStatusError(errors.NewEf(err, "could not create http request"))
		}
		httpResp, err := http.DefaultClient.Do(httpReq)
		if err != nil || httpReq == nil || httpResp.StatusCode < 200 || httpResp.StatusCode > 300 {
			isReady = false
			cs = append(
				cs,
				conditions.New(
					fmt.Sprintf("%d-HasValidSSL", idx),
					false,
					"SSLCheckFailed",
					errors.NewEf(err, "while making http request to (url=%s)", domain).Error(),
				),
			)
		}
	}

	newConditions, hasUpdated, err := conditions.Patch(router.Status.Conditions, cs)
	if err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not patch conditions"))
	}

	if !hasUpdated && isReady == router.Status.IsReady {
		return req.Next()
	}

	router.Status.IsReady = isReady
	router.Status.Conditions = newConditions
	router.Status.OpsConditions = []metav1.Condition{}

	if err := r.Status().Update(ctx, router); err != nil {
		return req.FailWithStatusError(err)
	}

	return req.Done()
}

func (r *RouterReconciler) reconcileOperations(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
	router := req.Object

	lambdaGroups := map[string]map[string][]crdsv1.Route{}
	appRoutes := map[string][]crdsv1.Route{}

	for routePath, routes := range router.Spec.Routes {
		for _, route := range routes {
			if s := route.Lambda; s != "" {
				if _, ok := lambdaGroups[route.Lambda]; !ok {
					lambdaGroups[route.Lambda] = map[string][]crdsv1.Route{}
				}

				lambdaGroups[route.Lambda][routePath] = append(lambdaGroups[route.Lambda][routePath], route)
			}

			if s := route.App; s != "" {
				appRoutes[routePath] = append(appRoutes[routePath], route)
			}
		}
	}

	var kubeYamls [][]byte

	for lName, lMapRoutes := range lambdaGroups {
		b, err := templates.Parse(
			templates.CoreV1.Ingress, map[string]any{
				"ingress-class":  constants.DefaultIngressClass,
				"cluster-issuer": constants.DefaultClusterIssuer,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(router, true),
				},
				"labels": map[string]string{
					"operators.kloudlite.io/router-name": router.Name,
				},

				"virtual-hostname":   fmt.Sprintf("%s.%s", lName, router.Namespace),
				"force-ssl-redirect": router.Spec.ForceSSLRedirect,
				"domains":            router.Spec.Domains,
				"routes":             lMapRoutes,

				"name":      fmt.Sprintf("r-%s-lambda-%s", router.Name, lName),
				"namespace": router.Namespace,
			},
		)

		if err != nil {
			return req.FailWithOpError(
				errors.NewEf(err, "could not parse (template=%s) into Object", templates.IngressLambda),
			)
		}

		kubeYamls = append(kubeYamls, b)
	}
	if len(appRoutes) > 0 {
		b, err := templates.Parse(
			templates.CoreV1.Ingress, map[string]any{
				"ingress-class":  constants.DefaultIngressClass,
				"cluster-issuer": constants.DefaultClusterIssuer,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(router, true),
				},

				"labels": map[string]string{
					"operators.kloudlite.io/router-name": router.Name,
				},

				"force-ssl-redirect": router.Spec.ForceSSLRedirect,
				"domains":            router.Spec.Domains,
				"routes":             appRoutes,

				"name":      router.Name,
				"namespace": router.Namespace,
			},
		)
		if err != nil {
			return req.FailWithOpError(errors.NewEf(err, "could not parse (template=%s) into Object", templates.Ingress))
		}
		kubeYamls = append(kubeYamls, b)
	}

	if _, err := fn.KubectlApplyExec(kubeYamls...); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not apply ingress ingressObj"))
	}

	return req.Done()
}

func (r *RouterReconciler) notify(req *rApi.Request[*crdsv1.Router]) rApi.StepResult {
	router := req.Object

	err := r.SendMessage(
		req.Context(), router.NameRef(), types.MessageReply{
			Key:        router.NameRef(),
			Conditions: router.Status.Conditions,
			IsReady:    router.Status.IsReady,
		},
	)
	if err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not send message into kafka"))
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
