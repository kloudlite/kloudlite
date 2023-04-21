package account

import (
	"context"
	"time"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/account/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
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
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	HarborProjectExists     string = "harbor-project-exists"
	HarborUserAccountExists string = "harbor-user-account-exists"
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

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureHarborProjectExists(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureHarborUserAccountExists(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) ensureHarborProjectExists(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HarborProjectExists)
	defer req.LogPostCheck(HarborProjectExists)

	hp := &artifactsv1.HarborProject{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.HarborProjectName}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, hp, func() error {
		if !fn.IsOwner(hp, fn.AsOwner(hp)) {
			hp.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		req.AddToOwnedResources(rApi.ResourceRef{
			TypeMeta:  metav1.TypeMeta{Kind: hp.GetObjectKind().GroupVersionKind().Kind, APIVersion: hp.GetObjectKind().GroupVersionKind().GroupVersion().String()},
			Namespace: hp.GetNamespace(),
			Name:      hp.GetName(),
		})
		return nil
	}); err != nil {
		return nil
	}

	check.Status = true
	if check != obj.Status.Checks[HarborProjectExists] {
		obj.Status.Checks[HarborProjectExists] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureHarborUserAccountExists(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HarborUserAccountExists)
	defer req.LogPostCheck(HarborUserAccountExists)

	hua := &artifactsv1.HarborUserAccount{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.HarborUsername, Namespace: obj.Spec.HarborSecretsNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, hua, func() error {
		if !fn.IsOwner(hua, fn.AsOwner(obj)) {
			hua.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		hua.Spec = artifactsv1.HarborUserAccountSpec{
			Enabled:           true,
			HarborProjectName: obj.Spec.HarborProjectName,
			Permissions: []harbor.Permission{
				harbor.PullRepository,
			},
		}
		hua.Spec.Enabled = true
		hua.Spec.HarborProjectName = obj.Spec.HarborProjectName

		req.AddToOwnedResources(rApi.ResourceRef{
			TypeMeta:  metav1.TypeMeta{Kind: hua.GetObjectKind().GroupVersionKind().Kind, APIVersion: hua.GetObjectKind().GroupVersionKind().GroupVersion().String()},
			Namespace: hua.GetNamespace(),
			Name:      hua.GetName(),
		})
		return nil
	}); err != nil {
		return nil
	}

	check.Status = true
	if check != obj.Status.Checks[HarborUserAccountExists] {
		obj.Status.Checks[HarborUserAccountExists] = check
		req.UpdateStatus()
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
