package admin

import (
	"context"
	"fmt"

	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/redpanda"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	if step := r.createAdminCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createRedpandaAdmin(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	adminDeleted := "admin-deleted"

	req.LogPreCheck(adminDeleted)
	defer req.LogPostCheck(adminDeleted)

	if obj.Spec.AuthFlags == nil || !obj.Spec.AuthFlags.Enabled {
		return req.Finalize()
	}

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.AuthFlags.TargetSecret.Namespace, obj.Spec.AuthFlags.TargetSecret.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(adminDeleted, check, err.Error()).Err(nil)
	}
	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](scrt)
	if err != nil {
		return req.CheckFailed(adminDeleted, check, err.Error()).Err(nil)
	}

	adminCli := redpanda.NewAdminClient(adminCreds.AdminEndpoint, "", nil)

	exists, err := adminCli.UserExists(adminCreds.Username)
	if err != nil {
		return req.CheckFailed(adminDeleted, check, err.Error()).Err(nil)
	}

	if exists {
		if err := adminCli.DeleteUser(adminCreds.Username); err != nil {
			return req.CheckFailed(adminDeleted, check, err.Error()).Err(nil)
		}
	}
	check.Status = true
	return req.Finalize()
}

func (r *Reconciler) createAdminCreds(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	if obj.Spec.AuthFlags == nil || !obj.Spec.AuthFlags.Enabled {
		return req.Next()
	}

	sc := fn.NN(obj.Spec.AuthFlags.TargetSecret.Namespace, obj.Spec.AuthFlags.TargetSecret.Name)
	scrt, err := rApi.Get(ctx, r.Client, sc, &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s, does not exist, will be creating now...", sc.String())
	}

	if scrt == nil {
		password := fn.CleanerNanoid(40)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       obj.Spec.AuthFlags.TargetSecret.Name,
				"namespace":  obj.Spec.AuthFlags.TargetSecret.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.AdminUserCreds{
					AdminEndpoint: obj.Spec.AdminEndpoint,
					KafkaBrokers:  obj.Spec.KafkaBrokers,
					Username:      obj.Name,
					Password:      password,

					RpkAdminFlags: fmt.Sprintf("--user %s --password %s --api-urls %s", obj.Name, password, obj.Spec.AdminEndpoint),
					RpkSASLFlags: fmt.Sprintf(
						"--user %s --password %s --brokers %s --sasl-mechanism %s", obj.Name, password,
						obj.Spec.KafkaBrokers, redpanda.ScramSHA256,
					),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}
		req.AddToOwnedResources(rr...)
	}

	if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
		err := r.Update(ctx, obj)
		return req.Done().Err(err)
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	adminSecret, err := fn.ParseFromSecret[types.AdminUserCreds](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}
	rApi.SetLocal(req, KeyAdminCreds, adminSecret)

	return req.Next()
}

func (r *Reconciler) createRedpandaAdmin(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	obj, checks := req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RedpandaAdminReady)
	defer req.LogPostCheck(RedpandaAdminReady)

	if obj.Spec.AuthFlags == nil || !obj.Spec.AuthFlags.Enabled {
		return req.Next()
	}

	adminCreds, ok := rApi.GetLocal[*types.AdminUserCreds](req, KeyAdminCreds)
	if !ok {
		return req.CheckFailed(RedpandaAdminReady, check, errors.NotInLocals(KeyAdminCreds).Error()).Err(nil)
	}

	adminExists := true

	err, _, _ := fn.Exec(
		fmt.Sprintf(
			"rpk acl user list --user %s --password '%s' --api-urls %s | grep -i %s", adminCreds.Username, adminCreds.Password,
			adminCreds.AdminEndpoint, adminCreds.Username,
		),
	)

	if err != nil {
		adminExists = false
		req.Logger.Infof("admin user does not exist, would be creating now...")
	}

	if !adminExists {
		err, _, stderr := fn.Exec(
			fmt.Sprintf(
				"rpk acl user create %s -p '%s' --api-urls %s", adminCreds.Username, adminCreds.Password, adminCreds.AdminEndpoint,
			),
		)
		if err != nil {
			return req.CheckFailed(RedpandaAdminReady, check, errors.NewEf(err, stderr.String()).Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[RedpandaAdminReady] {
		checks[RedpandaAdminReady] = check
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
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&redpandaMsvcv1.Admin{})
	builder.Owns(&corev1.Secret{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
