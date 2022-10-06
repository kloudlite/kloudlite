package account_router

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goombaio/namegenerator"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/routers/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type AccountRouterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *AccountRouterReconciler) GetName() string {
	return r.Name
}

const (
	IngressControllersReady string = "ingress-controllers-ready"
	IngressRoutesReady      string = "ingress-routes-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *AccountRouterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.AccountRouter{})
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

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(IngressControllersReady, IngressRoutesReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSpec(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconIngressController(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconIngressRoutes(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *AccountRouterReconciler) finalize(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
	return req.Finalize()
}

func (r *AccountRouterReconciler) reconSpec(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	if obj.Spec.ControllerName == "" {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		obj.Spec.ControllerName = "ingress-" + nameGenerator.Generate()
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
	}
	return req.Next()
}

func (r *AccountRouterReconciler) reconIngressController(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	ingressC, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ControllerName), fn.NewUnstructured(constants.HelmIngressNginx))
	if err != nil {
		req.Logger.Infof("ingress controller (%s) does not exist, will be creating it", fn.NN(obj.Namespace, obj.Spec.ControllerName).String())
	}

	if ingressC == nil || check.Generation > checks[IngressControllersReady].Generation {
		b, err := templates.Parse(
			templates.HelmIngressNginx, map[string]any{
				"obj":              obj,
				"controllers-name": obj.Spec.ControllerName,
				"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"labels": map[string]string{
					constants.AccountRouterNameKey: obj.Name,
				},
			},
		)
		if err != nil {
			return req.CheckFailed(IngressControllersReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(IngressControllersReady, check, err.Error())
		}

		checks[IngressControllersReady] = check
		return req.UpdateStatus()
	}

	cds, err := conditions.FromObject(ingressC)
	if err != nil {
		return req.CheckFailed(IngressControllersReady, check, err.Error())
	}

	deployedC := meta.FindStatusCondition(cds, "Deployed")
	if deployedC == nil {
		return req.Done()
	}

	if deployedC.Status == metav1.ConditionFalse {
		return req.CheckFailed(IngressControllersReady, check, check.Message)
	}

	if deployedC.Status == metav1.ConditionTrue {
		check.Status = true
	}

	if check != checks[IngressControllersReady] {
		checks[IngressControllersReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *AccountRouterReconciler) parseAccountDomains(ctx context.Context, accountRef string) []string {
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

func (r *AccountRouterReconciler) reconIngressRoutes(req *rApi.Request[*crdsv1.AccountRouter]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	accDomains := r.parseAccountDomains(ctx, obj.Spec.AccountRef)
	domains := append(accDomains, obj.Spec.WildcardDomains...)

	if len(domains) == 0 {
		check.Status = true
		if check != checks[IngressRoutesReady] {
			checks[IngressRoutesReady] = check
			return req.UpdateStatus()
		}
		return req.Next()
	}

	b, err := templates.Parse(
		templates.AccountIngressBridge, map[string]any{
			"obj":                  obj,
			"controllers-name":     obj.Spec.ControllerName,
			"global-ingress-class": r.Env.GlobalIngressClass,
			"domains":              domains,
			"labels":               obj.GetEnsuredLabels(),
			"cluster-issuer":       r.Env.ClusterCertIssuer,
			"owner-refs":           []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"ingress-svc-name":     obj.Spec.ControllerName + "-controller",
		},
	)

	if err != nil {
		return req.CheckFailed(IngressRoutesReady, check, err.Error())
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		return req.CheckFailed(IngressRoutesReady, check, err.Error())
	}

	check.Status = true
	if check != checks[IngressRoutesReady] {
		checks[IngressRoutesReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *AccountRouterReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.AccountRouter{})
	builder.Owns(fn.NewUnstructured(constants.HelmIngressNginx))
	builder.Owns(&networkingv1.Ingress{})

	builder.Watches(
		&source.Kind{Type: fn.NewUnstructured(constants.KloudliteAccountType)}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				accId := obj.GetLabels()[constants.AccountRef]
				if accId == "" {
					return nil
				}
				var accRoutersList crdsv1.AccountRouterList
				if err := r.List(
					context.TODO(), &accRoutersList, &client.ListOptions{
						LabelSelector: labels.SelectorFromValidatedSet(
							map[string]string{constants.AccountRef: accId},
						),
					},
				); err != nil {
					return nil
				}

				rr := make([]reconcile.Request, 0, len(accRoutersList.Items))
				for i := range accRoutersList.Items {
					rr = append(
						rr,
						reconcile.Request{NamespacedName: fn.NN(accRoutersList.Items[i].GetNamespace(), accRoutersList.Items[i].GetName())},
					)
				}
				return rr
			},
		),
	)

	return builder.Complete(r)
}
