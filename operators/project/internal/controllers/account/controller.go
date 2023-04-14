package account

import (
	"context"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccountReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
}

func (r *AccountReconciler) GetName() string {
	return r.Name
}

const (
	HarborProjectExists     string = "harbor-project-exists"
	HarborUserAccountExists string = "harbor-user-account-exists"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts/finalizers,verbs=update

func (r *AccountReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
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

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

// func (r *AccountReconciler) ensureHarborProjectExists(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.Check{Generation: obj.Generation}
//
// 	req.LogPreCheck(HarborProjectExists)
// 	defer req.LogPostCheck(HarborProjectExists)
//
// 	hp := &artifactsv1.HarborProject{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.HarborProjectName}}
// 	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, hp, func() error {
// 		return nil
// 	}); err != nil {
// 		return nil
// 	}
//
// 	check.Status = true
// 	if check != obj.Status.Checks[HarborProjectExists] {
// 		obj.Status.Checks[HarborProjectExists] = check
// 		req.UpdateStatus()
// 	}
//
// 	return req.Next()
// }
//
// func (r *AccountReconciler) ensureHarborUserAccountExists(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.Check{Generation: obj.Generation}
//
// 	req.LogPreCheck(HarborUserAccountExists)
// 	defer req.LogPostCheck(HarborUserAccountExists)
//
// 	hua := &artifactsv1.HarborUserAccount{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-admin", obj.Name)}}
// 	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, hua, func() error {
// 		hua.Spec = artifactsv1.HarborUserAccountSpec{
// 			Enabled:           true,
// 			HarborProjectName: obj.Spec.HarborProjectName,
// 			TargetSecret:      "",
// 			DockerConfigName:  "",
// 			Permissions: []harbor.Permission{
// 				harbor.PullRepository,
// 			},
// 		}
// 		hua.Spec.Enabled = true
// 		hua.Spec.HarborProjectName = obj.Spec.HarborProjectName
// 		return nil
// 	}); err != nil {
// 		return nil
// 	}
//
// 	check.Status = true
// 	if check != obj.Status.Checks[HarborUserAccountExists] {
// 		obj.Status.Checks[HarborUserAccountExists] = check
// 		req.UpdateStatus()
// 	}
//
// 	return req.Next()
// }

func (r *AccountReconciler) finalize(req *rApi.Request[*crdsv1.Account]) stepResult.Result {
	return req.Finalize()
}

func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Account{})
	return builder.Complete(r)
}
