package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codingconcepts/env"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	v1 "github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/app/internal/templates"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`
}

// Reconciler reconciles a App object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	YAMLClient kubectl.YAMLClient

	env Env

	templateHPASpec             []byte
	templateAppInterceptPodSpec []byte
	templateDeployment          []byte
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/app-controller"
}

// +kubebuilder:rbac:groups=kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=apps/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.App{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.App]{
		{
			Name:     "setup-deployment",
			Title:    "Setup Deployment",
			OnCreate: r.setupDeployment,
			OnDelete: r.cleanupDeployment,
		},
		{
			Name:     "setup-service",
			Title:    "Setup Service",
			OnCreate: r.setupService,
			OnDelete: r.cleanupService,
		},
		{
			Name:     "setup-hpa",
			Title:    "Setup Horizontal Pod Autoscaler",
			OnCreate: r.setupHPA,
			OnDelete: r.cleanupHPA,
		},
		{
			Name:     "setup-app-intercept",
			Title:    "Setup App Intercept",
			OnCreate: r.setupAppIntercept,
			OnDelete: r.cleanupAppIntercept,
		},
		{
			Name:     "setup-app-router",
			Title:    "Setup App Router",
			OnCreate: r.setupAppRouter,
			OnDelete: r.cleanupAppRouter,
		},
	})
}

// func getMatchSelector(obj *v1.App) map[string]string {
// 	return recommendedLabels(obj)
// 	// return fn.MapMerge(
// 	// 	fn.MapFilterWithPrefix(obj.GetLabels(), reconciler.ProjectDomain),
// 	// 	map[string]string{
// 	// 		"app.kubernetes.io/name": obj.GetName(),
// 	// 	},
// 	// )
// }

func (r *Reconciler) setupDeployment(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	labels := fn.MapFilterWithPrefix(obj.GetLabels(), v1.ProjectDomain)
	annotations := fn.MapFilterWithPrefix(obj.GetAnnotations(), v1.ProjectDomain)

	b, err := templates.ParseBytes(r.templateDeployment, templates.DeploymentParams{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Paused:         obj.Spec.Paused,
		Replicas:       obj.Spec.Replicas,
		PodLabels:      labels,
		PodAnnotations: annotations,
		PodSpec:        obj.Spec.PodSpec,
	})
	if err != nil {
		return check.Failed(err)
	}

	results, err := r.YAMLClient.ApplyYAML(check.Context(), b)
	if err != nil {
		return check.Failed(err)
	}

	if len(results) != 1 {
		return check.Failed(fmt.Errorf("wanted 1 result from apply YAML client, go %d", len(results)))
	}

	deployment := &appsv1.Deployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(results[0].Object, deployment); err != nil {
		fmt.Printf("Error converting unstructured to pod: %v\n", err)
		return check.Failed(err)
	}

	for _, c := range deployment.Status.Conditions {
		switch c.Type {
		case appsv1.DeploymentAvailable:
			if c.Status != corev1.ConditionTrue {
				var podList corev1.PodList
				if err := r.List(
					check.Context(), &podList, &client.ListOptions{
						LabelSelector: apiLabels.SelectorFromValidatedSet(deployment.Spec.Template.Labels),
						Namespace:     obj.Namespace,
					},
				); err != nil {
					return check.Errored(fmt.Errorf("failed to list pods: %w", err))
				}

				if len(podList.Items) > 0 {
					pMessages := fn.GetMessagesFromPods(podList.Items...)
					bMsg, err := json.Marshal(pMessages)
					if err != nil {
						return check.Errored(fmt.Errorf("failed to marshal pod message: %w", err))
					}
					return check.Errored(fmt.Errorf("deployment is not ready: %s", bMsg))
				}
			}
		case appsv1.DeploymentReplicaFailure:
			if c.Status == corev1.ConditionTrue {
				return check.Failed(fmt.Errorf(c.Message))
			}
		}
	}

	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return check.Errored(fmt.Errorf("ready-replicas (%d) != total replicas (%d)", deployment.Status.ReadyReplicas, deployment.Status.Replicas))
	}

	return check.Passed()
}

func (r *Reconciler) cleanupDeployment(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}

	if err := fn.DeleteAndWait(check.Context(), r.Client, deployment); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) setupService(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}

	if len(obj.Spec.Services) == 0 {
		return r.cleanupService(check, obj)
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svc, func() error {
		svc.SetLabels(fn.MapMerge(
			svc.GetLabels(),
			fn.MapFilterWithPrefix(obj.GetLabels(), reconciler.ProjectDomain),
		))
		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svc.Spec.Selector = obj.GetEnsuredLabels()
		svc.Spec.Ports = make([]corev1.ServicePort, 0, len(obj.Spec.Services))
		for i := range obj.Spec.Services {
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
				Name:       fmt.Sprintf("p-%d", obj.Spec.Services[i].Port),
				Protocol:   obj.Spec.Services[i].Protocol,
				Port:       obj.Spec.Services[i].Port,
				TargetPort: intstr.FromInt32(obj.Spec.Services[i].Port),
			})
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupService(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}

	if err := fn.DeleteAndWait(check.Context(), r.Client, svc); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) setupHPA(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}

	if obj.IsInterceptEnabled() || !obj.IsHPAEnabled() {
		return r.cleanupHPA(check, obj)
	}

	b, err := templates.ParseBytes(r.templateHPASpec, templates.HPASpecParams{
		DeploymentName: obj.Name,
		HPA:            obj.Spec.HPA,
	})
	if err != nil {
		return check.Failed(err)
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, hpa, func() error {
		hpa.SetLabels(fn.MapMerge(hpa.GetLabels(), fn.MapFilterWithPrefix(obj.GetLabels(), reconciler.ProjectDomain)))
		hpa.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		if err := yaml.Unmarshal(b, &hpa.Spec); err != nil {
			return fmt.Errorf("failed to unmarshal into hpa spec: %w", err)
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupHPA(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}); err != nil {
		return check.Errored(err)
	}
	return check.Passed()
}

func (r *Reconciler) setupAppIntercept(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	podname := obj.Name + "-intercept"
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: obj.Namespace}}

	if !obj.IsInterceptEnabled() {
		return r.cleanupAppIntercept(check, obj)
	}

	if obj.Spec.Services == nil {
		return check.Failed(fmt.Errorf("no services configured on app, failed to intercept")).NoRequeue()
	}

	if obj.Spec.Intercept.PortMappings == nil {
		obj.Spec.Intercept.PortMappings = make([]v1.AppInterceptPortMappings, len(obj.Spec.Services))
		for i := range obj.Spec.Services {
			obj.Spec.Intercept.PortMappings[i] = v1.AppInterceptPortMappings{
				Protocol:   obj.Spec.Services[i].Protocol,
				AppPort:    obj.Spec.Services[i].Port,
				DevicePort: obj.Spec.Services[i].Port,
			}
		}
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}

		return check.Abort("updated .spec.intercept.portMappings, waiting for reconcilation")
	}

	appGenerationLabel := "kloudlite.io/app.gen"

	if err := r.Get(check.Context(), client.ObjectKeyFromObject(pod), pod); err != nil {
		if apiErrors.IsNotFound(err) {
			tcpPortMappings := make(map[int32]int32, len(obj.Spec.Intercept.PortMappings))
			udpPortMappings := make(map[int32]int32, len(obj.Spec.Intercept.PortMappings))

			for _, pm := range obj.Spec.Intercept.PortMappings {
				switch pm.Protocol {
				case corev1.ProtocolTCP:
					tcpPortMappings[pm.AppPort] = pm.DevicePort
				case corev1.ProtocolUDP:
					udpPortMappings[pm.AppPort] = pm.DevicePort
				}
			}

			b, err := templates.ParseBytes(r.templateAppInterceptPodSpec, templates.AppInterceptPodSpecParams{
				DeviceHost:      obj.Spec.Intercept.ToHost,
				TCPPortMappings: tcpPortMappings,
				UDPPortMappings: udpPortMappings,
			})
			if err != nil {
				return check.Failed(err).NoRequeue()
			}

			if err := yaml.Unmarshal(b, &pod.Spec); err != nil {
				return check.Failed(err)
			}

			if err := r.Create(check.Context(), pod); err != nil {
				return check.Failed(err)
			}

			return check.Abort("waiting for intercept pod to start")
		}
	}

	if pod.Labels[appGenerationLabel] != fmt.Sprintf("%d", obj.Generation) {
		if result := r.cleanupAppIntercept(check, obj); !result.ShouldProceed() {
			return result
		}

		return check.Abort("waiting for previous generation intercept pod to be deleted")
	}

	if pod.Status.Phase != corev1.PodRunning {
		return check.Errored(fmt.Errorf("waiting for pod to start running"))
	}

	return check.Passed()
}

func (r *Reconciler) cleanupAppIntercept(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: obj.Name + "-intercept", Namespace: obj.Namespace}}); err != nil {
		return check.Failed(err)
	}
	return check.Passed()
}

func (r *Reconciler) setupAppRouter(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	if !obj.IsRouterEnabled() {
		return r.cleanupAppRouter(check, obj)
	}

	if len(obj.Spec.Router.Routes) == 0 {
		return check.Failed(fmt.Errorf("must specify at least 1 route"))
	}

	for i := range obj.Spec.Router.Routes {
		// INFO: app router will only route to current app, for any such usecases Router kind must be used
		obj.Spec.Router.Routes[i].Service = obj.Name
	}

	router := &v1.Router{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name + "-app-router",
			Namespace: obj.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, router, func() error {
		router.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		router.Spec = obj.Spec.Router.RouterSpec
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupAppRouter(check *reconciler.Check[*v1.App], obj *v1.App) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &v1.Router{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name + "-app-router",
			Namespace: obj.Namespace,
		},
	}); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if err := env.Set(&r.env); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.App{}).Named(r.GetName())
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&autoscalingv2.HorizontalPodAutoscaler{})

	var err error
	r.templateHPASpec, err = templates.Read(templates.HPASpec)
	if err != nil {
		return err
	}

	r.templateAppInterceptPodSpec, err = templates.Read(templates.AppInterceptPodSpec)
	if err != nil {
		return err
	}

	r.templateDeployment, err = templates.Read(templates.Deployment)
	if err != nil {
		return err
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))
	return builder.Complete(r)
}
