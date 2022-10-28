package app

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/kubectl"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/app-n-lambda/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	AppReady       string = "app-ready"
	ImagesLabelled string = "images-labelled"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
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
	if step := req.EnsureChecks(AppReady, ImagesLabelled); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconlabellingImages(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconApp(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconlabellingImages(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	newLabels := make(map[string]string, len(obj.GetLabels()))
	for s, v := range obj.GetLabels() {
		newLabels[s] = v
	}

	for s := range newLabels {
		if strings.HasPrefix(s, "kloudlite.io/image-") {
			delete(newLabels, s)
		}
	}

	for i := range obj.Spec.Containers {
		newLabels[fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(obj.Spec.Containers[i].Image)))] = "true"
	}

	if !reflect.DeepEqual(newLabels, obj.GetLabels()) {
		obj.SetLabels(newLabels)
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(ImagesLabelled, check, err.Error())
		}
		checks[ImagesLabelled] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[ImagesLabelled] {
		checks[ImagesLabelled] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconApp(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		req.Logger.Infof("app (deployment) %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name))
	}

	// svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	// if err != nil {
	// 	req.Logger.Infof("service %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name))
	// }
	//
	// svcInternal, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name+"-internal"), &corev1.Service{})
	// if err != nil {
	// 	req.Logger.Infof("service %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name+"-internal"))
	// }
	//
	// shouldRecon := deployment == nil || svc == nil || svcInternal == nil
	// if obj.Spec.Hpa.Enabled {
	// 	hpa, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &autoscalingv2.HorizontalPodAutoscaler{})
	// 	if err != nil {
	// 		req.Logger.Infof("horizontal pod autoscalar %s does not exist, will be creating it", fn.NN(obj.Namespace, obj.Name))
	// 	}
	// 	shouldRecon = shouldRecon || hpa == nil
	// }

	volumes, vMounts := crdsv1.ParseVolumes(obj.Spec.Containers)

	b, err := templates.Parse(
		templates.CrdsV1.App, map[string]any{
			"object":        obj,
			"volumes":       volumes,
			"volume-mounts": vMounts,
			"freeze":        obj.GetLabels()[constants.LabelKeys.Freeze] == "true" || obj.GetLabels()[constants.LabelKeys.IsIntercepted] == "true",
			"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},

			// for intercepting
			"is-intercepted": obj.GetLabels()[constants.LabelKeys.IsIntercepted] == "true",
			"device-ref":     obj.GetLabels()[constants.LabelKeys.DeviceRef],
			"account-ref":    obj.GetAnnotations()[constants.AnnotationKeys.AccountRef],
		},
	)

	if err != nil {
		return req.CheckFailed(AppReady, check, err.Error()).Err(nil)
	}

	// if err := fn.KubectlApplyExec(ctx, b); err != nil {
	// 	return req.CheckFailed(AppReady, check, err.Error()).Err(nil)
	// }

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(AppReady, check, err.Error()).Err(nil)
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
					map[string]string{"app": obj.Name},
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
		return req.CheckFailed(AppReady, check, "deployment is not ready")
	}

	if err := obj.Status.DisplayVars.Set("readyReplicas", deployment.Status.ReadyReplicas); err != nil {
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

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.App{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&autoscalingv2.HorizontalPodAutoscaler{})
	return builder.Complete(r)
}
