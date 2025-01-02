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
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

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
	ServicesReady   string = "services-ready"
	HPAConfigured   string = "hpa-configured"

	CleanedOwnedResources string = "cleaned-owned-resources"
	DeploymentSvcCreated  string = "deployment-svc-created"

	AppInterceptCreated string = "app-intercept-created"

	AppRouterReady string = "app-router-ready"
)

var DeleteChecklist = []reconciler.CheckMeta{
	{Name: CleanedOwnedResources, Title: "Cleaning up resources"},
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.App{})
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

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: DeploymentSvcCreated, Title: func() string {
			if len(req.Object.Spec.Services) > 0 {
				return "Deployment And Service Created"
			}
			return "Deployment Created"
		}(), Hide: req.Object.IsInterceptEnabled()},
		{Name: DeploymentReady, Title: "Deployment Ready", Hide: req.Object.IsInterceptEnabled()},
		{Name: HPAConfigured, Title: "Horizontal pod autoscaling configured", Hide: req.Object.IsInterceptEnabled() || !req.Object.IsHPAEnabled()},
		{Name: AppInterceptCreated, Title: "App Intercept Created", Hide: !req.Object.IsInterceptEnabled()},
		{Name: AppRouterReady, Title: "App Router Ready", Hide: req.Object.Spec.Router == nil},
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

	if step := r.checkAppRouter(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	check := reconciler.NewRunningCheck("finalizing", req)

	if step := req.CleanupOwnedResources(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) reconLabellingImages(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(ImagesLabelled, req)

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

func getServiceAccountName(obj *crdsv1.App) string {
	if obj.Spec.ServiceAccount != "" {
		return obj.Spec.ServiceAccount
	}
	if _, ok := obj.GetLabels()[constants.EnvNameKey]; ok {
		return "kloudlite-env-sa"
	}
	return ""
}

func (r *Reconciler) ensureDeploymentThings(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(DeploymentSvcCreated, req)

	volumes, vMounts := crdsv1.ParseVolumes(obj.Spec.Containers)

	b, err := templates.ParseBytes(
		r.appDeploymentTemplate, map[string]any{
			"object":               obj,
			"volumes":              volumes,
			"volume-mounts":        vMounts,
			"service-account-name": getServiceAccountName(obj),
			"owner-refs":           []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"account-name":         obj.GetAnnotations()[constants.AccountNameKey],
			"pod-labels":           fn.MapFilterWithPrefix(obj.GetLabels(), "kloudlite.io/"),
			"pod-annotations": fn.MapFilter(obj.GetAnnotations(),
				func(k string, _ string) bool {
					if k == "kloudlite.io/last-applied" {
						return false
					}
					return strings.HasPrefix(k, "kloudlite.io/") && !strings.HasPrefix(k, "kloudlite.io/operator.")
				},
			),
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

func (r *Reconciler) checkAppIntercept(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := reconciler.NewRunningCheck(AppInterceptCreated, req)

	podname := obj.Name + "-intercept"
	podns := obj.Namespace
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podns}}

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
				Protocol: func() crdsv1.ServiceProtocol {
					if obj.Spec.Services[i].Protocol != nil {
						return *obj.Spec.Services[i].Protocol
					}
					return crdsv1.ServiceProtocolTCP
				}(),
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
			tcpPortMappings := make(map[uint16]uint16, len(obj.Spec.Intercept.PortMappings))
			udpPortMappings := make(map[uint16]uint16, len(obj.Spec.Intercept.PortMappings))

			for _, pm := range obj.Spec.Intercept.PortMappings {
				switch pm.Protocol {
				case crdsv1.ServiceProtocolTCP:
					{
						tcpPortMappings[pm.AppPort] = pm.DevicePort
					}
				case crdsv1.ServiceProtocolUDP:
					{
						udpPortMappings[pm.AppPort] = pm.DevicePort
					}
				}
			}

			deviceHostSuffix := "device.local"
			if obj.Spec.Intercept.DeviceHostSuffix != nil {
				deviceHostSuffix = *obj.Spec.Intercept.DeviceHostSuffix
			}

			deviceHost := fmt.Sprintf("%s.%s", obj.Spec.Intercept.ToDevice, deviceHostSuffix)
			if obj.Spec.Intercept.ToIPAddr != "" {
				deviceHost = obj.Spec.Intercept.ToIPAddr
			}

			b, err := templates.ParseBytes(r.appInterceptTemplate, map[string]any{
				"name":             podname,
				"namespace":        podns,
				"labels":           fn.MapMerge(fn.MapFilterWithPrefix(obj.Labels, "kloudlite.io/"), map[string]string{appGenerationLabel: fmt.Sprintf("%d", obj.Generation)}),
				"owner-references": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				// "device-host":      fmt.Sprintf("%s.%s", obj.Spec.Intercept.ToDevice, deviceHostSuffix),
				"device-host":       deviceHost,
				"tcp-port-mappings": tcpPortMappings,
				"udp-port-mappings": udpPortMappings,
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

func (r *Reconciler) ensureHPA(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := reconciler.NewRunningCheck(HPAConfigured, req)

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

	if obj.IsInterceptEnabled() || !obj.IsHPAEnabled() {
		hpa, err := reconciler.Get(ctx, r.Client, fn.NN(hpaVars.Metadata.Namespace, hpaVars.Metadata.Name), &autoscalingv2.HorizontalPodAutoscaler{})
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

func (r *Reconciler) checkDeploymentReady(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(DeploymentReady, req)

	if obj.IsInterceptEnabled() {
		return check.Completed()
	}

	deployment, err := reconciler.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		return check.Failed(err)
	}

	if deployment.Spec.Template.Spec.ServiceAccountName != getServiceAccountName(obj) {
		return check.StillRunning(r.Delete(ctx, deployment))
	}

	for _, c := range deployment.Status.Conditions {
		switch c.Type {
		case appsv1.DeploymentAvailable:
			if c.Status != corev1.ConditionTrue {
				var podList corev1.PodList
				if err := r.List(
					ctx, &podList, &client.ListOptions{
						LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{"app": obj.Name}),
						Namespace:     obj.Namespace,
					},
				); err != nil {
					return check.StillRunning(err)
				}

				if len(podList.Items) > 0 {
					pMessages := fn.GetMessagesFromPods(podList.Items...)
					bMsg, err := json.Marshal(pMessages)
					if err != nil {
						return check.StillRunning(err)
					}
					return check.StillRunning(fmt.Errorf(string(bMsg)))
				}
			}
		case appsv1.DeploymentReplicaFailure:
			if c.Status == corev1.ConditionTrue {
				return check.Failed(fmt.Errorf(c.Message))
			}
		}
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return check.StillRunning(fmt.Errorf("ready-replicas (%d) != total replicas (%d)", deployment.Status.ReadyReplicas, deployment.Status.Replicas))
	}

	return check.Completed()
}

func (r *Reconciler) checkAppRouter(req *reconciler.Request[*crdsv1.App]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(AppRouterReady, req)

	if obj.Spec.Router == nil {
		return check.Completed()
	}

	if len(obj.Spec.Router.Routes) == 0 {
		if len(obj.Spec.Services) != 0 {
			return check.Failed(fmt.Errorf("app has multiple exposed services, cannot deduce router routes automatically from services, router routes must be explicity set via .spec.router.routes"))
		}

		obj.Spec.Router.Routes = append(obj.Spec.Router.Routes, crdsv1.Route{
			App:  obj.Name,
			Path: "/",
			Port: obj.Spec.Services[0].Port,
		})

		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("updating app router default routes"))
	}

	router := &crdsv1.Router{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-app-router", obj.Name), Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, router, func() error {
		router.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		router.Spec = obj.Spec.Router.RouterSpec
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YamlClient == nil {
		return fmt.Errorf("yamlClient must be set")
	}

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
	builder.WithEventFilter(reconciler.ReconcileFilter())
	return builder.Complete(r)
}
