package admin

import (
	"context"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
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
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	RedpandaAdminReady string = "redpanda-admin-ready"
	AccessCredsReady   string = "access-creds-ready"
	DefaultsPatched    string = "defaults-patched"
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
	if step := req.EnsureChecks(RedpandaAdminReady, AccessCredsReady, DefaultsPatched); !step.ShouldProceed() {
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

	if step := r.createAdminCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createRedpandaAdmin(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	adminDeleted := "admin-deleted"
	req.Logger.Infof("deleting admin ...")
	defer func() {
		if checks[adminDeleted].Status {
			req.Logger.Infof("redpanda admin deleted ...")
		}
		req.Logger.Infof("still ... deleting admin")
	}()

	check := rApi.Check{Generation: obj.Generation}

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.Output.SecretRef.Namespace, obj.Spec.Output.SecretRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(adminDeleted, check, err.Error()).Err(nil)
	}
	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](scrt)

	adminCli := redpanda.NewAdminClient(adminCreds.Username, adminCreds.Password, adminCreds.KafkaBrokers, adminCreds.AdminEndpoint)

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

func (r *Reconciler) reconDefaults(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.Output == nil {
		obj.Spec.Output = &ct.Output{
			SecretRef: &ct.SecretRef{
				Name:      "msvc-redpanda-" + obj.Name + "-creds",
				Namespace: obj.Namespace,
			},
		}

		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
		return req.Done().RequeueAfter(5 * time.Second)
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) createAdminCreds(req *rApi.Request[*redpandaMsvcv1.Admin]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	scrtNN := fn.NN(obj.Spec.Output.SecretRef.Namespace, obj.Spec.Output.SecretRef.Name)
	scrt, err := rApi.Get(ctx, r.Client, scrtNN, &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s, does not exist, will be creating now...", scrtNN.String())
	}

	if scrt == nil {
		password := fn.CleanerNanoid(40)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       obj.Spec.Output.SecretRef.Name,
				"namespace":  obj.Spec.Output.SecretRef.Namespace,
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

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
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

	err, _, _ := fn.Exec(
		fmt.Sprintf(
			"rpk acl user list --user %s --password '%s' --api-urls %s | grep -i admin", adminCreds.Username, adminCreds.Password,
			adminCreds.AdminEndpoint,
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
		return req.UpdateStatus()
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
