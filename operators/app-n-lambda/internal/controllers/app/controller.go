package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/operator/operator"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"

	"k8s.io/client-go/tools/record"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
	recorder   record.EventRecorder
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DeploymentSvcAndHpaCreated string = "deployment-svc-and-hpa-created"
	ImagesLabelled             string = "images-labelled"
	DeploymentReady            string = "deployment-ready"
	AnchorReady                string = "anchor-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.App{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if crdsv1.IsBlueprintNamespace(ctx, r.Client, request.Namespace) {
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DeploymentSvcAndHpaCreated, ImagesLabelled); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if req.Object.Enabled != nil && !*req.Object.Enabled {
		anchor := &crdsv1.Anchor{ObjectMeta: metav1.ObjectMeta{Name: req.GetAnchorName(), Namespace: req.Object.Namespace}}
		return ctrl.Result{}, client.IgnoreNotFound(r.Delete(ctx, anchor))
	}

	if step := operator.EnsureAnchor(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconLabellingImages(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDeploymentThings(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkDeploymentReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	req.Object.Status.DisplayVars.Set("intercepted", func() string {
		if req.Object.GetLabels()[constants.LabelKeys.IsIntercepted] == "true" {
			return "true/" + req.Object.GetLabels()[constants.LabelKeys.DeviceRef]
		}
		return "false"
	}())
	req.Object.Status.DisplayVars.Set("frozen", req.Object.GetLabels()[constants.LabelKeys.Freeze] == "true")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconLabellingImages(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ImagesLabelled)
	defer req.LogPostCheck(ImagesLabelled)

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
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	if check != checks[ImagesLabelled] {
		checks[ImagesLabelled] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureDeploymentThings(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DeploymentSvcAndHpaCreated)
	defer req.LogPostCheck(DeploymentSvcAndHpaCreated)

	volumes, vMounts := crdsv1.ParseVolumes(obj.Spec.Containers)

	isIntercepted := obj.GetLabels()[constants.LabelKeys.IsIntercepted] == "true"
	isFrozen := obj.GetLabels()[constants.LabelKeys.Freeze] == "true"

	anchor, _ := rApi.GetLocal[*crdsv1.Anchor](req, "anchor")

	b, err := templates.Parse(
		templates.CrdsV1.App, map[string]any{
			"object":        obj,
			"volumes":       volumes,
			"volume-mounts": vMounts,
			"freeze":        isFrozen || isIntercepted,
			"owner-refs":    []metav1.OwnerReference{fn.AsOwner(anchor, true)},

			// for intercepting
			"is-intercepted": obj.GetLabels()[constants.LabelKeys.IsIntercepted] == "true",
			"device-ref":     obj.GetLabels()[constants.LabelKeys.DeviceRef],
			"account-ref":    obj.GetAnnotations()[constants.AnnotationKeys.AccountRef],
		},
	)

	if err != nil {
		return req.CheckFailed(DeploymentSvcAndHpaCreated, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(DeploymentSvcAndHpaCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[DeploymentSvcAndHpaCreated] {
		checks[DeploymentSvcAndHpaCreated] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) checkDeploymentReady(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DeploymentReady)
	defer req.LogPostCheck(DeploymentReady)

	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		return req.CheckFailed(DeploymentReady, check, err.Error()).Err(nil)
	}

	cds, err := conditions.FromObject(deployment)
	if err != nil {
		return req.CheckFailed(DeploymentReady, check, err.Error()).Err(nil)
	}

	isReady := meta.IsStatusConditionTrue(cds, "Available")
	check.Status = isReady

	if !isReady {
		var podList corev1.PodList
		if err := r.List(
			ctx, &podList, &client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(
					map[string]string{"app": obj.Name},
				),
				Namespace: obj.Namespace,
			},
		); err != nil {
			return req.CheckFailed(DeploymentReady, check, err.Error())
		}

		pMessages := rApi.GetMessagesFromPods(podList.Items...)
		bMsg, err := json.Marshal(pMessages)
		if err != nil {
			check.Message = err.Error()
			return req.CheckFailed(DeploymentReady, check, err.Error())
		}
		check.Message = string(bMsg)
		return req.CheckFailed(DeploymentReady, check, "deployment is not ready")
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return req.CheckFailed(
			DeploymentReady,
			check,
			fmt.Sprintf("ready-replicas (%d) != total replicas (%d)", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
		)
	}

	check.Status = true
	if check != checks[DeploymentReady] {
		checks[DeploymentReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.App{})
	builder.Owns(&crdsv1.Anchor{})

	watchList := []client.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&autoscalingv2.HorizontalPodAutoscaler{},
	}

	for i := range watchList {
		builder.Watches(&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				if v, ok := obj.GetLabels()[constants.AppNameKey]; ok {
					return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
				}
				return nil
			}))
	}
	builder.WithOptions(controller.Options{
		MaxConcurrentReconciles: 1,
		//MaxConcurrentReconciles: 10,
	})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
