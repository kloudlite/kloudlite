package account

import (
	"context"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/account/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ensureAccountNamespace string = "ensure-account-namespace"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.Account{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureAccountNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) ensureAccountNamespace(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ensureAccountNamespace)
	defer req.LogPostCheck(ensureAccountNamespace)

	if obj.Spec.TargetNamespace == nil {
		return req.CheckFailed(ensureAccountNamespace, check, ".spec.targetNamespace is nil, it must be non-nil")
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: *obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 1)
		}
		ns.Labels[constants.AccountNameKey] = obj.Name

		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string, 1)
		}
		ns.Annotations[constants.DescriptionKey] = "namespace to hold resources, only for account: " + obj.Name
		return nil
	}); err != nil {
		req.CheckFailed(ensureAccountNamespace, check, err.Error())
	}

	check.Status = true
	if check != obj.Status.Checks[ensureAccountNamespace] {
		obj.Status.Checks[ensureAccountNamespace] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}
	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Account{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
