package aclaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	libRedis "operators.kloudlite.io/lib/redis"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	AccessCredsReady string = "access-creds-ready"
	ACLUserReady     string = "acl-user-ready"
	IsOwnedByMsvc    string = "is-owned-by-msvc"
)

const (
	KeyMsvcOutput string = "msvc-output"
	KeyOutput     string = "output"
)

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &redisMsvcv1.ACLAccount{})

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

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(AccessCredsReady, ACLUserReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
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
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconOwnership(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.MsvcRef.Name), fn.NewUnstructured(
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
			return req.FailWithOpError(err)
		}
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[IsOwnedByMsvc] {
		checks[IsOwnedByMsvc] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func parseMsvcOutput(scrt *corev1.Secret) (*types.MsvcOutput, error) {
	b, err := json.Marshal(scrt)
	if err != nil {
		return nil, err
	}
	var output types.MsvcOutput
	if err := json.Unmarshal(b, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func parseMresOutput(scrt *corev1.Secret) (*types.MresOutput, error) {
	b, err := json.Marshal(scrt)
	if err != nil {
		return nil, err
	}
	var output types.MresOutput
	if err := json.Unmarshal(b, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	accessSecretName := "mres-" + obj.Name

	accessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, accessSecretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, accessSecretName).String())
	}

	// msvc output ref
	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output does not exist").Error()).Err(nil)
	}

	msvcOutput, err := parseMsvcOutput(msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	if accessSecret == nil {
		passwd := fn.CleanerNanoid(40)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       accessSecretName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.MresOutput{
					Hosts:    msvcOutput.Hosts,
					Password: passwd,
					Username: obj.Name,
					Prefix:   obj.Name,
					Uri:      fmt.Sprintf("redis://%s:%s@%s?allowUsernameInURI=true", obj.Name, passwd, msvcOutput.Hosts),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	mresOutput, err := parseMresOutput(accessSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	rApi.SetLocal(req, KeyMsvcOutput, msvcOutput)
	rApi.SetLocal(req, KeyOutput, mresOutput)
	return req.Next()
}

func (r *Reconciler) reconACLUser(req *rApi.Request[*redisMsvcv1.ACLAccount]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(ACLUserReady, check, fmt.Sprintf("key %q does not exist in req-locals", KeyMsvcOutput))
	}

	output, ok := rApi.GetLocal[types.MresOutput](req, KeyOutput)
	if !ok {
		return req.CheckFailed(ACLUserReady, check, fmt.Sprintf("key %q does not exist in req-locals", KeyOutput))
	}

	redisCli, err := libRedis.NewClient(msvcOutput.Hosts, "", msvcOutput.RootPassword)
	if err != nil {
		return req.CheckFailed(ACLUserReady, check, err.Error())
	}
	defer redisCli.Close()

	tCtx, _ := context.WithTimeout(ctx, 3*time.Second)
	exists, err := redisCli.UserExists(tCtx, obj.Name)
	if err != nil {
		return req.CheckFailed(ACLUserReady, check, err.Error())
	}

	if !exists {
		tCtx, _ := context.WithTimeout(ctx, 3*time.Second)
		if err := redisCli.UpsertUser(tCtx, output.Prefix, output.Username, output.Password); err != nil {
			return req.CheckFailed(ACLUserReady, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[ACLUserReady] {
		checks[ACLUserReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.ACLAccount{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
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

	return builder.Complete(r)
}
