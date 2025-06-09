package kibana

import (
	"context"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	elasticsearchMsvcv1 "github.com/kloudlite/operator/apis/elasticsearch.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-elasticsearch/internal/env"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	HelmReady   string = "helm-ready"
	RouterReady string = "router-ready"
)

// +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=kibanas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=kibanas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elasticsearch.msvc.kloudlite.io,resources=kibanas/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &elasticsearchMsvcv1.Kibana{})
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
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(HelmReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.reconRouter(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*elasticsearchMsvcv1.Kibana]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*elasticsearchMsvcv1.Kibana]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	kibanaHelm, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmKibanaType))
	if err != nil {
		req.Logger.Infof("helm kibana not found, will be creating now...")
	}

	if kibanaHelm == nil || check.Generation > checks[HelmReady].Generation {
		b, err := templates.Parse(
			templates.Kibana, map[string]any{
				"obj":         obj,
				"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"elastic-url": obj.Spec.ElasticUrl,
			},
		)
		if err != nil {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}

		checks[HelmReady] = check
		return req.UpdateStatus()
	}

	cds, err := conditions.FromObject(kibanaHelm)
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	deployedC := meta.FindStatusCondition(cds, "Deployed")
	if deployedC == nil {
		return req.Done()
	}

	if deployedC.Status == metav1.ConditionFalse {
		return req.CheckFailed(HelmReady, check, deployedC.Message).Err(nil)
	}

	if deployedC.Status == metav1.ConditionTrue {
		check.Status = true
	}

	if check != checks[HelmReady] {
		checks[HelmReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

// func (r *Reconciler) reconRouter(req *rApi.Request[*elasticsearchMsvcv1.Kibana]) stepResult.Result {
// 	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
// 	check := rApi.Check{Generation: obj.Generation}
//
// 	router, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &crdsv1.Router{})
// 	if err != nil {
// 		router = nil
// 		req.Logger.Infof("kibana router does not exist yet, will be creating now")
// 	}
//
// 	if router == nil || check.Generation > checks[RouterReady].Generation {
// 		if err := r.Create(
// 			ctx, &crdsv1.Router{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:            obj.Name,
// 					Namespace:       obj.Namespace,
// 					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
// 					Labels:          obj.GetLabels(),
// 				},
// 				Spec: crdsv1.RouterSpec{
// 					Region: obj.Spec.Region,
// 					Https: crdsv1.Https{
// 						Enabled:       true,
// 						ForceRedirect: true,
// 					},
// 					Domains: []string{obj.Spec.Expose.Domain},
// 					Routes: []crdsv1.Route{
// 						{
// 							App:  obj.Name + "-kibana",
// 							Path: "/",
// 							Port: 5601,
// 						},
// 					},
// 					BasicAuth: crdsv1.BasicAuth{
// 						Enabled:    true,
// 						SecretName: obj.Spec.Expose.BasicAuthSecret,
// 					},
// 				},
// 			},
// 		); err != nil {
// 			return req.CheckFailed(RouterReady, check, err.Error()).Err(nil)
// 		}
// 		checks[RouterReady] = check
// 		return req.UpdateStatus()
// 	}
//
// 	check.Status = true
// 	if check != checks[RouterReady] {
// 		checks[RouterReady] = check
// 		return req.UpdateStatus()
// 	}
// 	return req.Next()
// }

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&elasticsearchMsvcv1.Kibana{})
	builder.Owns(fn.NewUnstructured(constants.HelmKibanaType))
	builder.Owns(&crdsv1.Router{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
