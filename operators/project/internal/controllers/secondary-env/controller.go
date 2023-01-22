package secondary_env

import (
	"context"
	"fmt"
	fn "github.com/kloudlite/operator/pkg/functions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	v1 "github.com/kloudlite/operator/apis/crds/v1"

	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	// TODO: add checks
	CheckReady     string = "check-ready"
	NamespaceReady string = "namespace-ready"
)

const (
	MsvcReady      string = "msvc-ready"
	ServicesSynced string = "services-synced"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=secondaryEnvs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=secondaryEnvs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=secondaryEnvs/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &v1.SecondaryEnv{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
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

	if step := req.EnsureChecks(CheckReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespaces(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.syncServices(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.SecondaryEnv]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) ensureNamespaces(req *rApi.Request[*v1.SecondaryEnv]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespaceReady)
	defer req.LogPostCheck(NamespaceReady)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Name}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if !fn.IsOwner(ns, fn.AsOwner(obj)) {
			ns.SetOwnerReferences(append(ns.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}
		return nil
	}); err != nil {
		return req.CheckFailed(NamespaceReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[NamespaceReady] {
		checks[NamespaceReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) syncServices(req *rApi.Request[*v1.SecondaryEnv]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ServicesSynced)
	defer req.LogPostCheck(ServicesSynced)

	var svcList corev1.ServiceList
	if err := r.List(ctx, &svcList, &client.ListOptions{
		Namespace: obj.Spec.PrimaryEnvName,
	}); err != nil {
		return req.CheckFailed(ServicesSynced, check, err.Error()).Err(nil)
	}

	for i := range svcList.Items {
		svc := svcList.Items[i]
		lSvc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: svc.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lSvc, func() error {
			if !fn.IsOwner(lSvc, fn.AsOwner(obj)) {
				lSvc.SetOwnerReferences(append(lSvc.GetOwnerReferences(), fn.AsOwner(obj, true)))
			}
			lSvc.Spec.Type = corev1.ServiceTypeExternalName
			lSvc.Spec.ExternalName = fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace)
			return nil
		}); err != nil {
			return req.CheckFailed(ServicesSynced, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[ServicesSynced] {
		checks[ServicesSynced] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.SecondaryEnv{})
	return builder.Complete(r)
}
