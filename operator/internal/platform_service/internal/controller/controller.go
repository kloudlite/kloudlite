package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codingconcepts/env"
	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/platform_service/internal/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`
}

// Reconciler reconciles a PlatformService object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	env        Env
	YAMLClient kubectl.YAMLClient

	dynamicWatch           func(apiVersion, kind string) error
	templatePluginResource []byte

	watchingTypes map[string]struct{}
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/platform-service-controller"
}

// +kubebuilder:rbac:groups=kloudlite.io,resources=platformservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=platformservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=platformservices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.PlatformService{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.PlatformService]{
		{
			Name:     "setup-platform-service",
			Title:    "Setup Platform Service",
			OnCreate: r.createPlatformService,
			OnDelete: r.cleanupPlatformService,
		},
	})
}

func (r *Reconciler) createPlatformService(check *reconciler.Check[*v1.PlatformService], obj *v1.PlatformService) reconciler.StepResult {
	b, err := templates.ParseBytes(r.templatePluginResource, templates.PluginResourceTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			Labels:          obj.GetLabels(),
			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Plugin: &obj.Spec.Plugin,
	})
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	if err := r.OwnDynamicResource(check, obj); err != nil {
		return check.Failed(err).NoRequeue()
	}

	objects, err := r.YAMLClient.ApplyYAML(check.Context(), b)
	if err != nil {
		return check.Failed(err)
	}

	if len(objects) == 0 {
		return check.Failed(fmt.Errorf("no objects returned from YAMLClient.ApplyYAML"))
	}

	m, err := json.Marshal(objects[0])
	if err != nil {
		return check.Failed(err)
	}

	var pluginResource struct {
		Status reconciler.Status `json:"status"`
	}
	if err := json.Unmarshal(m, &pluginResource); err != nil {
		return check.Failed(err).NoRequeue()
	}

	if !pluginResource.Status.IsReady {
		errorMsg := ""
		for _, v := range pluginResource.Status.CheckList {
			if pluginResource.Status.Checks[v.Name].State == reconciler.ErroredState && pluginResource.Status.Checks[v.Name].Message != "" {
				errorMsg = pluginResource.Status.Checks[v.Name].Message
				break
			}
		}

		if errorMsg == "" {
			return check.Abort("waiting for plugin resource to reconcile").NoRequeue()
		}

		return check.Failed(fmt.Errorf(errorMsg)).NoRequeue()
	}

	return check.Passed()
}

func (r *Reconciler) cleanupPlatformService(check *reconciler.Check[*v1.PlatformService], obj *v1.PlatformService) reconciler.StepResult {
	b, err := templates.ParseBytes(r.templatePluginResource, templates.PluginResourceTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			Labels:          obj.GetLabels(),
			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Plugin: &obj.Spec.Plugin,
	})
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	// if err := r.OwnDynamicResource(req, obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind); err != nil {
	// 	return check.Failed(err).NoRequeue()
	// }

	if err := r.YAMLClient.DeleteYAML(check.Context(), b); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) OwnDynamicResource(check *reconciler.Check[*v1.PlatformService], obj *v1.PlatformService) error {
	apiVersion, kind := obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind
	if _, ok := r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)]; ok {
		return nil
	}

	r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)] = struct{}{}

	if !fn.IsGVKInstalled(r.Client, apiVersion, kind) {
		check.Logger().Warn("plugin CRDs not installed", "APIVersion", apiVersion, "Kind", kind)
		return nil
	}

	if err := r.dynamicWatch(apiVersion, kind); err != nil {
		check.Logger().Error("failed to call Complete() on builder, got", "err", err)
		return err
	}

	check.Logger().Info("ADDED watch for owned-resources", "APIVersion", apiVersion, "Kind", kind)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if err := env.Set(&r.env); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	if r.YAMLClient == nil {
		return fmt.Errorf(".YAMLClient must be set prior to invocation")
	}

	var err error
	r.templatePluginResource, err = templates.Read(templates.PluginResourceTemplate)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.PlatformService{}).Named(r.GetName())

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	builder.Owns(&corev1.Pod{})
	builder.Owns(&corev1.Service{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))

	tc, err := builder.Build(r)
	if err != nil {
		return err
	}

	r.dynamicWatch = func(apiVersion, kind string) error {
		obj := fn.NewUnstructured(metav1.TypeMeta{APIVersion: apiVersion, Kind: kind})
		return tc.Watch(source.Kind(mgr.GetCache(), obj, handler.TypedEnqueueRequestForOwner[*unstructured.Unstructured](mgr.GetScheme(), mgr.GetRESTMapper(), &v1.PlatformService{}, handler.OnlyControllerOwner())))
	}

	return nil
}
