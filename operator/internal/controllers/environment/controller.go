package environment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codingconcepts/env"
	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	"github.com/kloudlite/operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type envVars struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`
}

// EnvironmentReconciler reconciles a Environment object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	env envVars
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/environment-controller"
}

// +kubebuilder:rbac:groups=kloudlite.io,resources=environments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=environments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=environments/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.Environment{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.Environment]{
		{
			Name:     "setup environment namespace",
			Title:    "Setup Environment Namespace",
			OnCreate: r.createNamespace,
			OnDelete: r.cleanupNamespace,
		},
		{
			Name:     "setup default service account",
			Title:    "Setup Default Service Account",
			OnCreate: r.createServiceAccount,
			OnDelete: r.cleanupServiceAccount,
		},
		{
			Name:     "pause/unpause environment",
			Title:    "Pause/UnPause Environment",
			OnCreate: r.pauseEnvironment,
			OnDelete: r.unpauseEnvironment,
		},
	})
}

func (r *Reconciler) createNamespace(check *reconciler.Check[*v1.Environment], obj *v1.Environment) reconciler.StepResult {
	if obj.Spec.TargetNamespace == "" {
		obj.Spec.TargetNamespace = obj.Name
		if !strings.HasPrefix(obj.Name, "env-") {
			obj.Spec.TargetNamespace = "env-" + obj.Name
		}

		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}

		return check.Abort("waiting for .spec.targetNamespace to be set")
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, ns, func() error {
		ns.SetLabels(fn.MapMerge(ns.GetLabels(), map[string]string{
			v1.EnvironmentNameKey:     obj.Name,
			v1.GatewayEnabledLabelKey: v1.GatewayEnabledLabelValue,
		}))

		ns.SetAnnotations(fn.MapMerge(ns.GetAnnotations(), map[string]string{
			reconciler.AnnotationDescriptionKey: fmt.Sprintf("this namespace is now being managed by kloudlite environment (%s)", obj.Name),
		}))

		// ns.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) deleteEnvRouters(ctx context.Context, namespace string) error {
	var routers v1.RouterList
	if err := r.Client.List(ctx, &routers, client.InNamespace(namespace)); err != nil {
		return err
	}

	if err := fn.DeleteAndWait(ctx, r.Client, fn.TransformSlice(routers.Items, func(v v1.Router) client.Object { return &v })...); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) deleteEnvApps(ctx context.Context, namespace string) error {
	var apps v1.AppList
	if err := r.Client.List(ctx, &apps, client.InNamespace(namespace)); err != nil {
		return err
	}

	if err := fn.DeleteAndWait(ctx, r.Client, fn.TransformSlice(apps.Items, func(v v1.App) client.Object { return &v })...); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) deleteEnvServiceIntercepts(ctx context.Context, namespace string) error {
	var apps v1.ServiceInterceptList
	if err := r.Client.List(ctx, &apps, client.InNamespace(namespace)); err != nil {
		return err
	}

	if err := fn.DeleteAndWait(ctx, r.Client, fn.TransformSlice(apps.Items, func(v v1.ServiceIntercept) client.Object { return &v })...); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) cleanupNamespace(check *reconciler.Check[*v1.Environment], obj *v1.Environment) reconciler.StepResult {
	if err := r.deleteEnvRouters(check.Context(), obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err).RequeueAfter(500 * time.Millisecond)
	}

	if err := r.deleteEnvApps(check.Context(), obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err).RequeueAfter(500 * time.Millisecond)
	}

	if err := r.deleteEnvServiceIntercepts(check.Context(), obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err).RequeueAfter(500 * time.Millisecond)
	}

	return check.Passed()
}

func (r *Reconciler) pauseEnvironment(check *reconciler.Check[*v1.Environment], obj *v1.Environment) reconciler.StepResult {
	rquota := &corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-quota", obj.Name),
		Namespace: obj.Spec.TargetNamespace,
	}}

	if !obj.Spec.Paused {
		return r.unpauseEnvironment(check, obj)
	}

	// creating resource quota for the environment namespace
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, rquota, func() error {
		rquota.Spec.Hard = corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("0Mi"),
			corev1.ResourceCPU:    resource.MustParse("0m"),
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// INFO: need to evict all pods from the environment's namespace, to let them honour the created quota
	var podsList corev1.PodList
	if err := r.List(check.Context(), &podsList, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return check.Failed(err)
	}

	for i := range podsList.Items {
		if err := r.Delete(check.Context(), &podsList.Items[i]); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}
	}

	return check.Passed()
}

func (r *Reconciler) unpauseEnvironment(check *reconciler.Check[*v1.Environment], obj *v1.Environment) reconciler.StepResult {
	rquota := &corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-quota", obj.Name),
		Namespace: obj.Spec.TargetNamespace,
	}}

	if err := r.Delete(check.Context(), rquota); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	}
	return check.Passed()
}

func (r *Reconciler) createServiceAccount(check *reconciler.Check[*v1.Environment], obj *v1.Environment) reconciler.StepResult {
	if obj.Spec.ServiceAccount == "" {
		obj.Spec.ServiceAccount = "kloudlite-env-sa"
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.Abort("waiting for reconcilation")
	}

	var pullSecrets corev1.SecretList
	if err := r.List(check.Context(), &pullSecrets, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return check.Failed(err)
	}

	secretNames := make([]corev1.LocalObjectReference, 0, len(pullSecrets.Items))
	for i := range pullSecrets.Items {
		if pullSecrets.Items[i].Type == corev1.SecretTypeDockerConfigJson {
			secretNames = append(secretNames, corev1.LocalObjectReference{Name: pullSecrets.Items[i].Name})
		}
	}

	svca := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Spec.ServiceAccount,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svca, func() error {
		svca.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svca.ImagePullSecrets = secretNames

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupServiceAccount(check *reconciler.Check[*v1.Environment], obj *v1.Environment) reconciler.StepResult {
	if obj.Spec.ServiceAccount == "" {
		return check.Passed()
	}

	svca := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Spec.ServiceAccount,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if err := fn.DeleteAndWait(check.Context(), r.Client, svca); err != nil {
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

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Environment{}).Named(r.GetName())
	builder.Owns(&corev1.Namespace{})
	builder.Owns(&corev1.ResourceQuota{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))
	return builder.Complete(r)
}
