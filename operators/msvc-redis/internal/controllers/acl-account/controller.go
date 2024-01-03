package aclaccount

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/operator/pkg/kubectl"

	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	libRedis "github.com/kloudlite/operator/pkg/redis"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

const (
	AccessCredsReady string = "access-creds-ready"
	ACLUserReady     string = "acl-user-ready"
	IsOwnedByMsvc    string = "is-owned-by-msvc"

	ACLUserDeleted string = "acl-user-deleted"
)

const (
	KeyMsvcOutput string = "msvc-output"
	KeyMresOutput string = "mres-output"
)

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &redisMsvcv1.ACLAccount{})
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

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconOwnership(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconAccessCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconACLUser(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	// msvc output ref
	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("msvc output does not exist, i.e. msvc does not exist, so no need keeping mres, so finalizing it")
		return req.Finalize()
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	redisCli, err := libRedis.NewClient(msvcOutput.Hosts, "", msvcOutput.RootPassword)
	if err != nil {
		return req.CheckFailed(ACLUserReady, check, err.Error())
	}
	defer redisCli.Close()

	tctx, cf := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cf()
	if err := redisCli.DeleteUser(tctx, obj.Name); err != nil {
		return req.CheckFailed(ACLUserDeleted, check, err.Error())
	}

	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	if obj.Spec.KeyPrefix == "" {
		obj.Spec.KeyPrefix = obj.Name
		if err := r.Update(ctx, obj); err != nil {
			return req.Done().Err(err)
		}
	}

	return req.Next()
}

func (r *Reconciler) reconOwnership(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IsOwnedByMsvc)
	defer req.LogPostCheck(IsOwnedByMsvc)

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), fn.NewUnstructured(
			metav1.TypeMeta{
				Kind:       obj.Spec.MsvcRef.Kind,
				APIVersion: obj.Spec.MsvcRef.APIVersion,
			},
		),
	)
	if err != nil {
		return req.CheckFailed(IsOwnedByMsvc, check, err.Error())
	}

	if !fn.IsOwner(obj, fn.AsOwner(msvc)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(msvc)))
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(IsOwnedByMsvc, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Second)
	}

	check.Status = true
	if check != obj.Status.Checks[IsOwnedByMsvc] {
		obj.Status.Checks[IsOwnedByMsvc] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	secretName := "mres-" + obj.Name

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, secretName).String())
	}

	// msvc output ref
	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output does not exist").Error()).Err(nil)
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	if scrt == nil {
		passwd := fn.CleanerNanoid(40)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"owner-refs": obj.GetOwnerReferences(),
				"labels": map[string]string{
					constants.MsvcNamespaceKey: obj.Spec.MsvcRef.Namespace,
					constants.MsvcNameKey:      obj.Spec.MsvcRef.Name,
					constants.IsMresOutput:     "true",
				},
				"string-data": types.MresOutput{
					Hosts:    msvcOutput.Hosts,
					Password: passwd,
					Username: obj.Name,
					Prefix:   obj.Spec.KeyPrefix,
					Uri:      fmt.Sprintf("redis://%s:%s@%s?allowUsernameInURI=true", obj.Name, passwd, msvcOutput.Hosts),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[AccessCredsReady] {
		obj.Status.Checks[AccessCredsReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	mresOutput, err := fn.ParseFromSecret[types.MresOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)
	rApi.SetLocal(req, KeyMresOutput, *mresOutput)
	return req.Next()
}

func (r *Reconciler) reconACLUser(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ACLUserReady)
	defer req.LogPostCheck(ACLUserReady)

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(ACLUserReady, check, errors.NotInLocals(KeyMsvcOutput).Error())
	}

	output, ok := rApi.GetLocal[types.MresOutput](req, KeyMresOutput)
	if !ok {
		return req.CheckFailed(ACLUserReady, check, errors.NotInLocals(KeyMresOutput).Error())
	}

	redisCli, err := libRedis.NewClient(msvcOutput.Hosts, "", msvcOutput.RootPassword)
	if err != nil {
		return req.CheckFailed(ACLUserReady, check, err.Error())
	}
	defer redisCli.Close()

	exists, err := redisCli.UserExists(context.TODO(), output.Username)
	if err != nil {
		return req.CheckFailed(ACLUserReady, check, err.Error())
	}

	if !exists {
		tCtx, cancelFn := context.WithTimeout(ctx, 3*time.Second)
		defer cancelFn()
		if err := redisCli.UpsertUser(tCtx, output.Prefix, output.Username, output.Password); err != nil {
			return req.CheckFailed(ACLUserReady, check, err.Error())
		}
	}

	check.Status = true
	if check != obj.Status.Checks[ACLUserReady] {
		obj.Status.Checks[ACLUserReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.ACLAccount{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&redisMsvcv1.StandaloneService{},
		&redisMsvcv1.ClusterService{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj, handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList redisMsvcv1.ACLAccountList
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
