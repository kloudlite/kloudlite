package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
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

type AppReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *AppReconciler) GetName() string {
	return r.Name
}

const (
	AppReady string = "app-ready"
	HPAReady string = "hpa-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *AppReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.App{})
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

	// TODO: initialize all checks here
	if step := req.EnsureChecks(AppReady, HPAReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.cleanup(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconApp(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *AppReconciler) finalize(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	return req.Finalize()
}

func (r *AppReconciler) reconApp(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		req.Logger.Infof("deployment %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name))
	}

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		req.Logger.Infof("service %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name))
	}

	// svcInternal, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name+"-internal"), &corev1.Service{})
	// if err != nil {
	// 	req.Logger.Infof("service %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name+"-internal"))
	// }

	if deployment == nil || svc == nil || check.Generation > checks[AppReady].Generation {

		volumes, vMounts := crdsv1.ParseVolumes(obj.Spec.Containers)

		b, err := templates.Parse(
			templates.CrdsV1.App, map[string]any{
				"object":        obj,
				"volumes":       volumes,
				"volume-mounts": vMounts,
				"freeze": obj.GetLabels()[constants.LabelKeys.Freeze] == "true" || obj.GetLabels()[constants.LabelKeys.
					IsIntercepted] == "true",
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

				// for intercepting
				"is-intercepted": obj.GetLabels()[constants.LabelKeys.IsIntercepted] == "true",
				"device-ref":     obj.GetLabels()[constants.LabelKeys.DeviceRef],
				"account-ref":    obj.GetAnnotations()[constants.AnnotationKeys.AccountRef],
			},
		)

		if err != nil {
			return req.CheckFailed(AppReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AppReady, check, err.Error()).Err(nil)
		}

		return req.UpdateStatus()
	}

	cds, err := conditions.FromObject(deployment)
	if err != nil {
		return req.CheckFailed(AppReady, check, err.Error()).Err(nil)
	}

	deploymentIsReady := meta.IsStatusConditionTrue(cds, "Available")
	check.Status = deploymentIsReady

	if !deploymentIsReady {
		var podList corev1.PodList
		if err := r.List(
			ctx, &podList, &client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(
					map[string]string{
						"app": obj.Name,
					},
				),
				Namespace: obj.Namespace,
			},
		); err != nil {
			return req.CheckFailed(AppReady, check, err.Error())
		}

		pMessages := rApi.GetMessagesFromPods(podList.Items...)
		bMsg, err := json.Marshal(pMessages)
		if err != nil {
			check.Message = err.Error()
			return req.CheckFailed(AppReady, check, err.Error())
		}
		check.Message = string(bMsg)
		return req.CheckFailed(AppReady, check, err.Error())
	}

	if err := obj.Status.DisplayVars.Set("readyRplicas", deployment.Status.ReadyReplicas); err != nil {
		return req.CheckFailed(AppReady, check, err.Error())
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return req.CheckFailed(
			AppReady,
			check,
			fmt.Sprintf("ready-replicas (%d) != total replicas (%d)", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
		)
	}

	check.Status = true

	if check != checks[AppReady] {
		checks[AppReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.App{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	return builder.Complete(r)
}
