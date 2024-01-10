package app

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/client-go/tools/record"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/templates"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Logger     logging.Logger
	Name       string
	Env        *env.Env
	YamlClient kubectl.YAMLClient
	recorder   record.EventRecorder

	appDeploymentTemplate []byte
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
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.Logger), r.Client, request.NamespacedName, &crdsv1.App{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
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
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}
	return req.Finalize()
}

func (r *Reconciler) reconLabellingImages(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
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
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[ImagesLabelled] {
		obj.Status.Checks[ImagesLabelled] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureDeploymentThings(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DeploymentSvcAndHpaCreated)
	defer req.LogPostCheck(DeploymentSvcAndHpaCreated)

	volumes, vMounts := crdsv1.ParseVolumes(obj.Spec.Containers)

	b, err := templates.ParseBytes(
		r.appDeploymentTemplate, map[string]any{
			"object": obj,

			"volumes":       volumes,
			"volume-mounts": vMounts,
			"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"account-name":  obj.GetAnnotations()[constants.AccountNameKey],

			"pod-annotations": fn.FilterObservabilityAnnotations(obj),

			"cluster-dns-suffix": r.Env.ClusterInternalDNS,
		},
	)
	if err != nil {
		return req.CheckFailed(DeploymentSvcAndHpaCreated, check, err.Error()).Err(nil)
	}

	resRefs, err := r.YamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(DeploymentSvcAndHpaCreated, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(resRefs...)

	check.Status = true
	if check != obj.Status.Checks[DeploymentSvcAndHpaCreated] {
		fn.MapSet(&obj.Status.Checks, DeploymentSvcAndHpaCreated, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) checkDeploymentReady(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
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

	if !isReady {
		var podList corev1.PodList
		if err := r.List(
			ctx, &podList, &client.ListOptions{
				LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{"app": obj.Name}),
				Namespace:     obj.Namespace,
			},
		); err != nil {
			return req.CheckFailed(DeploymentReady, check, err.Error())
		}

		pMessages := rApi.GetMessagesFromPods(podList.Items...)
		bMsg, err := json.Marshal(pMessages)
		if err != nil {
			return req.CheckFailed(DeploymentReady, check, err.Error())
		}
		return req.CheckFailed(DeploymentReady, check, string(bMsg))
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return req.CheckFailed(
			DeploymentReady,
			check,
			fmt.Sprintf("ready-replicas (%d) != total replicas (%d)", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
		)
	}

	check.Status = true
	if check != obj.Status.Checks[DeploymentReady] {
		obj.Status.Checks[DeploymentReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Logger = logger.WithName(r.Name)
	r.YamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.Logger})
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	var err error
	r.appDeploymentTemplate, err = templates.Read(templates.AppDeployment)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.App{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&autoscalingv2.HorizontalPodAutoscaler{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
