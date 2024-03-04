package prefix

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/operator/pkg/kubectl"

	redisMsvcV1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
}

func (r *Reconciler) GetName() string {
	return r.Name
}

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=prefixes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=prefixes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=prefixes/finalizers,verbs=update
func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &redisMsvcV1.Prefix{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureAccessCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*redisMsvcV1.Prefix]) stepResult.Result {
	checkName := "finalizing"
	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*redisMsvcV1.Prefix]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "patch-defaults"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hasPatched := false

	if obj.Spec.PrefixKey == "" {
		hasPatched = true
		obj.Spec.PrefixKey = fmt.Sprintf("%s:", obj.Name)
	}

	if obj.Spec.Output.Credentials.Name == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Name = fmt.Sprintf("mres-%s-creds", obj.Name)
	}

	if obj.Spec.Output.Credentials.Namespace == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Namespace = obj.Namespace
	}

	if hasPatched {
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) getMsvcCredentials(req *rApi.Request[*redisMsvcV1.Prefix]) (*types.MsvcOutput, error) {
	ctx, obj := req.Context(), req.Object

	switch obj.Spec.MsvcRef.Kind {
	case "StandaloneService":
		{
			msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &redisMsvcV1.StandaloneService{})
			if err != nil {
				return nil, err
			}

			s, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Spec.Output.Credentials.Namespace, msvc.Spec.Output.Credentials.Name), &corev1.Secret{})
			if err != nil {
				return nil, err
			}

			mo, err := fn.ParseFromSecret[types.MsvcOutput](s)
			if err != nil {
				return nil, err
			}

			return mo, nil
		}
	case "ClusterService":
		{
			return nil, fmt.Errorf("not implemented")
		}
	default:
		{
			return nil, fmt.Errorf("unknown msvc kind (%s), must of one of [StandaloneService, ClusterService]", obj.Spec.MsvcRef.Kind)
		}
	}
}

func (r *Reconciler) ensureAccessCredentials(req *rApi.Request[*redisMsvcV1.Prefix]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "ensure-access-credentials"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	msvcCreds, err := r.getMsvcCredentials(req)
	if err != nil {
		return fail(err)
	}

	creds, err := fn.JsonConvert[map[string]string](types.PrefixCredentialsData{
		Hosts:    msvcCreds.Hosts,
		Password: msvcCreds.RootPassword,
		Username: msvcCreds.RootUsername,
		Prefix:   obj.Spec.PrefixKey,
		Uri:      msvcCreds.Uri,
	})
	if err != nil {
		return fail(err)
	}

	mresCredsSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.Output.Credentials.Name, Namespace: obj.Spec.Output.Credentials.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, mresCredsSecret, func() error {
		mresCredsSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		mresCredsSecret.StringData = creds
		return nil
	}); err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(mresCredsSecret))

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
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

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcV1.Prefix{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&redisMsvcV1.StandaloneService{},
		&redisMsvcV1.ClusterService{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj, handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList redisMsvcV1.PrefixList
					if err := r.List(
						context.TODO(), &dbList, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{constants.MsvcNameKey: msvcName},
							),
							Namespace: obj.GetNamespace(),
						},
					); err != nil {
						return nil
					}

					reqs := make([]reconcile.Request, 0, len(dbList.Items))
					for j := range dbList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(dbList.Items[j].GetNamespace(), dbList.Items[j].GetName())})
					}

					return reqs
				},
			),
		)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
