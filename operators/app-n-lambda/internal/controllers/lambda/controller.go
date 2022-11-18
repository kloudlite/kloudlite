package lambda

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/pkg/conditions"
	"operators.kloudlite.io/pkg/constants"
	fn "operators.kloudlite.io/pkg/functions"
	"operators.kloudlite.io/pkg/harbor"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	"operators.kloudlite.io/pkg/templates"
	"operators.kloudlite.io/operators/app-n-lambda/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	KnativeServingReady string = "knative-serving-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lambdas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lambdas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lambdas/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &serverlessv1.Lambda{})
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

	if step := req.EnsureChecks(KnativeServingReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconLambda(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*serverlessv1.Lambda]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconLambda(req *rApi.Request[*serverlessv1.Lambda]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	knServing, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.KnativeServiceType))
	if err != nil {
		req.Logger.Infof("knative serving (%s) does not exist, will be creating now...", fn.NN(obj.Namespace, obj.Name).String())
	}

	if knServing == nil || check.Generation > checks[KnativeServingReady].Generation {
		b, err := templates.Parse(
			templates.ServerlessLambda, map[string]any{
				"object": obj,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(KnativeServingReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(KnativeServingReady, check, err.Error()).Err(nil)
		}

		checks[KnativeServingReady] = check
		return req.UpdateStatus()
	}

	cds, err := conditions.FromObject(knServing)
	if err != nil {
		return req.CheckFailed(KnativeServingReady, check, err.Error())
	}

	cfgReady := meta.FindStatusCondition(cds, "ConfigurationsReady")
	if cfgReady == nil {
		return req.Done()
	}

	if cfgReady.Status == metav1.ConditionFalse {
		return req.CheckFailed(KnativeServingReady, check, "knative serving is not ready yet").Err(nil)
	}

	check.Status = true
	if check != checks[KnativeServingReady] {
		checks[KnativeServingReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&serverlessv1.Lambda{})
	builder.Owns(fn.NewUnstructured(constants.KnativeServiceType))
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
