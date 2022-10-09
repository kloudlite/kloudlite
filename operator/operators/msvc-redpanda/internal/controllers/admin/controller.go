package admin

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	redpandaMsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/msvc-redpanda/internal/env"
	"operators.kloudlite.io/operators/msvc-redpanda/internal/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	RedpandaAdminReady string = "redpanda-admin-ready"
	AccessCredsReady   string = "access-creds-ready"
)

const (
	KeyAdminCreds string = "admin-creds"
)

// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=admins,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=admins/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=admins/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &redpandaMsvcv1.Admin{})
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
	if step := req.EnsureChecks(RedpandaAdminReady, AccessCredsReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createAdminCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createRedpandaAdmin(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) createAdminCreds(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	scrtName := "msvc-redpanda-" + obj.Name + "-creds"

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, scrtName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s, does not exist, will be creating now...", fn.NN(obj.Namespace, scrtName).String())
	}

	if scrt == nil {
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       scrtName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.AdminUserCreds{
					AdminEndpoint: obj.Spec.AdminEndpoint,
					KafkaBrokers:  obj.Spec.KafkaBrokers,
					Username:      obj.Name,
					Password:      fn.CleanerNanoid(40),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
		err := r.Update(ctx, obj)
		return req.Done().Err(err)
	}

	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		obj.Status.DisplayVars.Set("output-secret-name", scrtName)
		return req.UpdateStatus()
	}

	adminSecret, err := fn.ParseFromSecret[types.AdminUserCreds](scrt)
	rApi.SetLocal(req, KeyAdminCreds, adminSecret)

	return req.Next()
}

func (r *Reconciler) createRedpandaAdmin(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	obj, checks := req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	adminCreds, ok := rApi.GetLocal[*types.AdminUserCreds](req, KeyAdminCreds)
	if !ok {
		return req.CheckFailed(RedpandaAdminReady, check, errors.NotInLocals(KeyAdminCreds).Error()).Err(nil)
	}

	adminExists := true

	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl user list --user %s --password '%s' --api-urls %s | grep -i admin", adminCreds.Username, adminCreds.Password,
			adminCreds.AdminEndpoint,
		),
	)
	fmt.Println(stderr.String())

	if err != nil {
		adminExists = false
		req.Logger.Infof("admin user does not exist, would be creating now...")
	}

	fmt.Println(adminExists)

	if !adminExists {
		err, stdout, stderr := fn.Exec(
			fmt.Sprintf(
				"rpk acl user create %s -p '%s' --api-urls %s", adminCreds.Username, adminCreds.Password, adminCreds.AdminEndpoint,
			),
		)
		fmt.Println(stderr.String(), stdout.String())
		if err != nil {
			return req.CheckFailed(RedpandaAdminReady, check, errors.NewEf(err, stderr.String()).Error()).Err(nil)
		}
	}

	if check != checks[RedpandaAdminReady] {
		checks[RedpandaAdminReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&redpandaMsvcv1.Admin{})
	builder.Owns(&corev1.Secret{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
