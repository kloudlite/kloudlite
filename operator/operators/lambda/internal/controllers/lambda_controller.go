package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LambdaReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *LambdaReconciler) GetName() string {
	return r.Name
}

const (
	KnativeServingReady string = "knative-serving-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lambdas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lambdas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lambdas/finalizers,verbs=update

func (r *LambdaReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
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
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *LambdaReconciler) finalize(req *rApi.Request[*serverlessv1.Lambda]) stepResult.Result {
	return req.Finalize()
}

func (r *LambdaReconciler) reconLambda(req *rApi.Request[*serverlessv1.Lambda]) stepResult.Result {
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

func (r *LambdaReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&serverlessv1.Lambda{})
	builder.Owns(fn.NewUnstructured(constants.KnativeServiceType))
	return builder.Complete(r)
}
