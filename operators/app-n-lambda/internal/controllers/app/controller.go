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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	hpaTemplate           []byte
	appInterceptTemplate  []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ImagesLabelled  string = "images-labelled"
	DeploymentReady string = "deployment-ready"
	HPAConfigured   string = "hpa-configured"

	CleanedOwnedResources string = "cleaned-owned-resources"
	DeploymentSvcCreated  string = "deployment-svc-created"

	AppInterceptCreated string = "app-intercept-created"

	DefaultsPatched string = "defaults-patched"
)

var DeleteChecklist = []rApi.CheckMeta{
	{Name: CleanedOwnedResources, Title: "Cleaning up resources"},
}

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

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: DeploymentSvcCreated, Title: "Deployment And Service Created"},
		{Name: DeploymentReady, Title: "Deployment Ready", Hide: req.Object.IsInterceptEnabled()},
		{Name: HPAConfigured, Title: "Horizontal pod autoscaling configured", Hide: req.Object.IsInterceptEnabled()},
		{Name: AppInterceptCreated, Title: "App Intercept Created", Hide: !req.Object.IsInterceptEnabled()},
	}); !step.ShouldProceed() {
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

	if step := r.ensureHPA(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkDeploymentReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkAppIntercept(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) reconLabellingImages(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ImagesLabelled, req)

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
			return check.StillRunning(err)
		}

		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	return check.Completed()
}

func (r *Reconciler) ensureDeploymentThings(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(DeploymentSvcCreated, req)

	volumes, vMounts := crdsv1.ParseVolumes(obj.Spec.Containers)

	b, err := templates.ParseBytes(
		r.appDeploymentTemplate, map[string]any{
			"object":             obj,
			"volumes":            volumes,
			"volume-mounts":      vMounts,
			"owner-refs":         []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"account-name":       obj.GetAnnotations()[constants.AccountNameKey],
			"pod-labels":         fn.MapFilter(obj.Labels, "kloudlite.io/"),
			"pod-annotations":    fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			"cluster-dns-suffix": r.Env.ClusterInternalDNS,
		},
	)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	resRefs, err := r.YamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(resRefs...)

	return check.Completed()
}

func (r *Reconciler) checkAppIntercept(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	podname := obj.Name + "-intercept"
	podns := obj.Namespace
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podns}}

	check := rApi.NewRunningCheck(AppInterceptCreated, req)

	if obj.Spec.Intercept == nil || !obj.Spec.Intercept.Enabled {
		if err := r.Delete(ctx, pod); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}
		return check.Completed()
	}

	if obj.Spec.Services == nil {
		return check.Failed(fmt.Errorf("no services configured on app, failed to intercept")).NoRequeue()
	}

	if obj.Spec.Intercept.PortMappings == nil {
		obj.Spec.Intercept.PortMappings = make([]crdsv1.AppInterceptPortMappings, len(obj.Spec.Services))
		for i := range obj.Spec.Services {
			obj.Spec.Intercept.PortMappings[i] = crdsv1.AppInterceptPortMappings{
				AppPort:    obj.Spec.Services[i].Port,
				DevicePort: obj.Spec.Services[i].Port,
			}
		}
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.Completed()
	}

	appGenerationLabel := "kloudlite.io/app-generation"

	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		if apiErrors.IsNotFound(err) {
			portMappings := make(map[uint16]uint16, len(obj.Spec.Intercept.PortMappings))
			for _, pm := range obj.Spec.Intercept.PortMappings {
				portMappings[pm.AppPort] = pm.DevicePort
			}

			deviceHostSuffix := "device.local"
			if obj.Spec.Intercept.DeviceHostSuffix != nil {
				deviceHostSuffix = *obj.Spec.Intercept.DeviceHostSuffix
			}

			b, err := templates.ParseBytes(r.appInterceptTemplate, map[string]any{
				"name":             podname,
				"namespace":        podns,
				"labels":           fn.MapMerge(fn.MapFilter(obj.Labels, "kloudlite.io/"), map[string]string{appGenerationLabel: fmt.Sprintf("%d", obj.Generation)}),
				"owner-references": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"device-host":      fmt.Sprintf("%s.%s", obj.Spec.Intercept.ToDevice, deviceHostSuffix),
				"port-mappings":    portMappings,
			})
			if err != nil {
				return check.Failed(err).NoRequeue()
			}

			if _, err := r.YamlClient.ApplyYAML(ctx, b); err != nil {
				return check.Failed(err).NoRequeue()
			}

			return check.StillRunning(fmt.Errorf("waiting for intercept pod to start"))
		}
	}

	if pod.Labels[appGenerationLabel] != fmt.Sprintf("%d", obj.Generation) {
		if err := r.Delete(ctx, pod); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for previous generation pod to be deleted"))
	}

	if pod.Status.Phase != corev1.PodRunning {
		return check.StillRunning(fmt.Errorf("waiting for pod to start running"))
	}

	return check.Completed()
}

func (r *Reconciler) ensureHPA(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.NewRunningCheck(HPAConfigured, req)

	if obj.IsInterceptEnabled() {
		return check.Completed()
	}

	hpaVars := templates.HPATemplateVars{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			Labels:          obj.GetLabels(),
			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		HPA: obj.Spec.Hpa,
	}

	if obj.Spec.Hpa == nil || !obj.Spec.Hpa.Enabled {
		hpa, err := rApi.Get(ctx, r.Client, fn.NN(hpaVars.Metadata.Namespace, hpaVars.Metadata.Name), &autoscalingv2.HorizontalPodAutoscaler{})
		if err != nil {
			if apiErrors.IsNotFound(err) {
				return check.Completed()
			}
			return check.Failed(err)
		}

		if err := r.Delete(ctx, hpa); err != nil {
			return check.Failed(err)
		}
		return check.Completed()
	}

	b, err := templates.ParseBytes(r.hpaTemplate, hpaVars)
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.YamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.StillRunning(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) checkDeploymentReady(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(DeploymentReady, req)

	if obj.IsInterceptEnabled() {
		return check.Completed()
	}

	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		return check.Failed(err)
	}

	cds, err := conditions.FromObject(deployment)
	if err != nil {
		return check.StillRunning(err)
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
			return check.StillRunning(err)
		}

		pMessages := rApi.GetMessagesFromPods(podList.Items...)
		bMsg, err := json.Marshal(pMessages)
		if err != nil {
			return check.StillRunning(err)
		}
		return check.StillRunning(fmt.Errorf(string(bMsg)))
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return check.StillRunning(fmt.Errorf("ready-replicas (%d) != total replicas (%d)", deployment.Status.ReadyReplicas, deployment.Status.Replicas))
	}

	return check.Completed()
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

	r.hpaTemplate, err = templates.Read(templates.HPATemplate)
	if err != nil {
		return err
	}

	r.appInterceptTemplate, err = templates.Read(templates.AppIntercept)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.App{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Pod{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&autoscalingv2.HorizontalPodAutoscaler{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
