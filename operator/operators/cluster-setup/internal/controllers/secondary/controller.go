package secondary

import (
	"context"
	v1 "github.com/kloudlite/operator/apis/cluster-setup/v1"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

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
	CheckReady           string = "check-ready"
	DefaultsPatched      string = "defaults-patched"
	PriorityClassesReady string = "priority-classes-ready"
)

// +kubebuilder:rbac:groups=cluster-setup.kloudlite.io,resources=secondaryclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-setup.kloudlite.io,resources=secondaryclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster-setup.kloudlite.io,resources=secondaryclusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &v1.SecondaryCluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(CheckReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensurePriorityClasses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.SecondaryCluster]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*v1.SecondaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	sharedC := v1.SecondarySharedConstants{
		StatefulPriorityClass: "stateful",
	}

	var hasUpdated bool

	if obj.Spec.SharedConstants != sharedC {
		hasUpdated = true
		obj.Spec.SharedConstants = sharedC
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}
	return req.Next()

}

func (r *Reconciler) ensurePriorityClasses(req *rApi.Request[*v1.SecondaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(PriorityClassesReady)
	defer req.LogPostCheck(PriorityClassesReady)

	pc := &schedulingv1.PriorityClass{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.SharedConstants.StatefulPriorityClass}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, pc, func() error {
		pc.PreemptionPolicy = fn.New(corev1.PreemptLowerPriority)
		pc.Value = 1000000
		return nil
	}); err != nil {
		return req.CheckFailed(PriorityClassesReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[PriorityClassesReady] {
		checks[PriorityClassesReady] = check
		return req.UpdateStatus()
	}
	return req.Next()

}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.SecondaryCluster{})
	builder.Owns(&appsv1.Deployment{})
	return builder.Complete(r)
}
