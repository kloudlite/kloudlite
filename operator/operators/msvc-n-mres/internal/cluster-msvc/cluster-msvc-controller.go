package cluster_msvc

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateCommonMsvc []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	NSCreated string = "namespace-created"
	MsvcReady string = "msvc-ready"

	NamespaceCreatedByLabel string = "kloudlite.io/created-by-cluster-msvc-controller"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ClusterManagedService{})
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

	if step := req.EnsureChecks(NSCreated, MsvcReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMsvcCreatedNReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ClusterManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.Check{Generation: obj.Generation}

	checkName := "finalizing"
	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	ns, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.Namespace), &corev1.Namespace{})
	if err != nil {
		return req.CheckFailed(checkName, check, err.Error())
	}

	if v, ok := ns.GetAnnotations()[NamespaceCreatedByLabel]; ok && v == "true" {
		if err := r.Delete(ctx, ns); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
		}
	}

	return req.Finalize()
}

func (r *Reconciler) ensureNamespace(req *rApi.Request[*crdsv1.ClusterManagedService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(NSCreated, check, err.Error())
	}

	if obj.Spec.Namespace == "" {
		obj.Spec.Namespace = fmt.Sprintf("cmsvc-%s", obj.Name)

		if err := r.Update(ctx, obj); err != nil {
			return failed(err)
		}
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Generation == 0 {
			fn.MapSet(&ns.Annotations, NamespaceCreatedByLabel, "true")
			ns.SetAnnotations(ns.Annotations)
		}
		return nil
	}); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[NSCreated] {
		checks[NSCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureMsvcCreatedNReady(req *rApi.Request[*crdsv1.ClusterManagedService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(MsvcReady, check, err.Error())
	}

	msvc := &crdsv1.ManagedService{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Spec.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, msvc, func() error {
		msvc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		fn.MapSet(&msvc.Labels, constants.CMsvcNameKey, obj.Name)

		msvc.Spec = obj.Spec.MSVCSepec

		return nil
	}); err != nil {
		return failed(err)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(msvc))

	msChecks := msvc.Status.Checks
	updated := false
	for k, c := range msChecks {
		if checks[k] != msChecks[k] {
			checks[k] = c
			updated = true
		}
	}

	if updated {
		obj.Status.Message = msvc.Status.Message
		if step := req.UpdateStatus(); !step.ShouldProceed() {
			_, err := step.ReconcilerResponse()
			return failed(err)
		}
		if err := r.Update(ctx, obj); err != nil {
			return failed(err)
		}
	}

	if !msvc.Status.IsReady {
		return failed(fmt.Errorf("managed service %q is not ready", msvc.Name))
	}

	// if err := func() error {
	// 	if m, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.Namespace, obj.Name), &crdsv1.ManagedService{}); err != nil {
	// 		if !apiErrors.IsNotFound(err) {
	// 			return err
	// 		}
	// 		resource := &crdsv1.ManagedService{
	// 			ObjectMeta: metav1.ObjectMeta{
	// 				Name:            obj.Name,
	// 				Namespace:       obj.Spec.Namespace,
	// 				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
	// 				Labels: map[string]string{
	// 					constants.CMsvcNameKey: obj.Name,
	// 				},
	// 			},
	// 			TypeMeta: metav1.TypeMeta{
	// 				Kind:       "ManagedService",
	// 				APIVersion: "crds.kloudlite.io/v1",
	// 			},
	// 			Spec: obj.Spec.MSVCSepec,
	// 		}
	//
	// 		if err := r.Create(ctx, resource); err != nil {
	// 			return err
	// 		}
	//
	// 		req.AddToOwnedResources(rApi.ParseResourceRef(resource))
	// 	} else {
	// 		msChecks := m.Status.Checks
	// 		updated := false
	// 		for k, c := range msChecks {
	// 			if checks[k] != msChecks[k] {
	// 				checks[k] = c
	// 				updated = true
	// 			}
	// 		}
	//
	// 		if updated {
	// 			obj.Status.Message = m.Status.Message
	// 			if step := req.UpdateStatus(); !step.ShouldProceed() {
	// 				_, err := step.ReconcilerResponse()
	// 				return err
	// 			}
	// 			if err := r.Update(ctx, obj); err != nil {
	// 				return err
	// 			}
	// 		}
	//
	// 		if !m.Status.IsReady {
	// 			return fmt.Errorf("managed service %q is not ready", m.Name)
	// 		}
	// 	}
	//
	// 	return nil
	// }(); err != nil {
	// 	return failed(err)
	// }

	check.Status = true
	if check != checks[MsvcReady] {
		checks[MsvcReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateCommonMsvc, err = templates.Read(templates.CommonMsvcTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ClusterManagedService{})
	msvcs := []client.Object{
		&crdsv1.ManagedService{},
	}

	for _, obj := range msvcs {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					if v, ok := obj.GetLabels()[constants.CMsvcNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN("", v)}}
					}
					return nil
				}))
		builder.Owns(obj)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
