package refer_to_mres

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
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
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	CopiedMresCredentials   string = "copied-mres-credentials"
	DeleteCopiedCredentials string = "delete-copied-mres-credentials"
)

var ApplyCheckList = []rApi.CheckMeta{
	{Name: CopiedMresCredentials, Title: "Copied Managed Resource Credentials"},
}

var DeleteCheckList = []rApi.CheckMeta{
	{Name: DeleteCopiedCredentials, Title: "Delete Copied Managed Resource Credentials"},
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ReferToManagedResource{})
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.copyManagedResourceCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ReferToManagedResource]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) copyManagedResourceCredentials(req *rApi.Request[*crdsv1.ReferToManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(CopiedMresCredentials, req)

	mr, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.ManagedResourceRef.Namespace, obj.Spec.ManagedResourceRef.Name), &crdsv1.ManagedResource{})
	if err != nil {
		return check.Failed(errors.NewEf(err, "failed to find managed resource %s/%s", obj.Spec.ManagedResourceRef.Namespace, obj.Spec.ManagedResourceRef.Name))
	}

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(mr.Namespace, mr.Output.CredentialsRef.Name), &corev1.Secret{})
	if err != nil {
		return check.Failed(errors.NewEf(err, "failed to find managedresource secret credentials %s/%s", obj.Spec.ManagedResourceRef.Namespace, mr.Output.CredentialsRef.Name))
	}

	resultScrt := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: scrt.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &resultScrt, func() error {
		resultScrt.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		resultScrt.Data = scrt.Data
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ReferToManagedResource{})
	builder.Owns(&corev1.Secret{})

	builder.Watches(&crdsv1.ManagedResource{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			var rtmr crdsv1.ReferToManagedResourceList
			if err := r.List(ctx, &rtmr, &client.ListOptions{
				LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{"kloudlite.io/mres.ref": fmt.Sprintf("%s %s", obj.GetNamespace(), obj.GetName())}),
			}); err != nil {
				return nil
			}

			rr := make([]reconcile.Request, 0, len(rtmr.Items))
			for i := range rtmr.Items {
				rr = append(rr, reconcile.Request{NamespacedName: fn.NN(rtmr.Items[i].Namespace, rtmr.Items[i].Name)})
			}

			return rr
		}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
