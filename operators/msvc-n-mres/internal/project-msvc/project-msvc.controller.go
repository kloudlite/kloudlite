package project_msvc

import (
	"context"
	"fmt"
	"strings"

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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	NSCreated   string = "namespace-created"
	MsvcReady   string = "msvc-ready"
	MsvcDeleted string = "msvc-deleted"

	NamespaceCreatedByLabel string = "kloudlite.io/created-by-project-msvc-controller"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ProjectManagedService{})
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

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ProjectManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.Check{Generation: obj.Generation}

	checkName := "finalizing"
	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(MsvcDeleted, check, err.Error())
	}

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	if msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Name, obj.Spec.TargetNamespace), &crdsv1.ManagedService{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
	} else {
		if msvc != nil && msvc.DeletionTimestamp == nil {
			if err := r.Delete(ctx, msvc); err != nil {
				return failed(err)
			}

			return failed(fmt.Errorf("managed service %q is scheduled for deletion", msvc.Name))
		}

		return failed(fmt.Errorf("managed service %q is being deleted", msvc.Name))
	}

	ns, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.TargetNamespace), &corev1.Namespace{})
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

func (r *Reconciler) ensureNamespace(req *rApi.Request[*crdsv1.ProjectManagedService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(NSCreated, check, err.Error())
	}

	if obj.Spec.TargetNamespace == "" {
		obj.Spec.TargetNamespace = fmt.Sprintf("pmsvc-%s", obj.Name)

		if err := r.Update(ctx, obj); err != nil {
			return failed(err)
		}
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
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
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureMsvcCreatedNReady(req *rApi.Request[*crdsv1.ProjectManagedService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(MsvcReady, check, err.Error())
	}

	msvc := &crdsv1.ManagedService{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, msvc, func() error {
		fn.MapSet(&msvc.Labels, constants.ProjectManagedServiceRefKey, fmt.Sprintf("%s/%s", obj.Namespace, obj.Name))
		msvc.Spec = obj.Spec.MSVCSpec

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

	check.Status = true
	if check != checks[MsvcReady] {
		checks[MsvcReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateCommonMsvc, err = templates.Read(templates.CommonMsvcTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ProjectManagedService{})
	msvcs := []client.Object{
		&crdsv1.ManagedService{},
	}

	for _, obj := range msvcs {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					if v, ok := obj.GetLabels()[constants.ProjectManagedServiceRefKey]; ok {
						sp := strings.SplitN(v, "/", 2)
						if len(sp) != 2 {
							return nil
						}
						return []reconcile.Request{{NamespacedName: fn.NN(sp[0], sp[1])}}
					}
					return nil
				}),
		)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
