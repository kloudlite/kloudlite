package acluser

import (
	"context"

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
	AccessCredsReady  string = "access-creds-ready"
	RedpandaUserReady string = "redpanda-user-ready"
)

const (
	KeyAdminCreds  string = "admin-creds"
	KeyAccessCreds string = "access-creds"
)

// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=aclusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=aclusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=aclusers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &redpandaMsvcv1.ACLUser{})
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
	if step := req.EnsureChecks(AccessCredsReady, RedpandaUserReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
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
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*redpandaMsvcv1.ACLUser]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	aclUserDeleted := "acl-user-deleted"
	req.Logger.Infof("deleting user")
	defer func() {
		if checks[aclUserDeleted].Status {
			req.Logger.Infof("acl user deletion successfull")
		}
		req.Logger.Infof("still ... deleting user")
	}()

	if step := req.EnsureChecks(aclUserDeleted); !step.ShouldProceed() {
		return step
	}

	check := rApi.Check{Generation: obj.Generation}

	adminScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.AdminSecretRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
	}

	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](adminScrt)
	if err != nil {
		return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
	}

	adminCli := redpanda.NewAdminClient(adminCreds.AdminEndpoint, "", nil)

	exists, err := adminCli.UserExists(obj.Name)
	if err != nil {
		return req.CheckFailed(aclUserDeleted, check, err.Error())
	}
	if exists {
		if err := adminCli.DeleteUser(obj.Name); err != nil {
			return req.CheckFailed(aclUserDeleted, check, err.Error())
		}
	}

	check.Status = true

	return req.Finalize()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*redpandaMsvcv1.ACLUser]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	adminScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.AdminSecretRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
	}

	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](adminScrt)
	if err != nil {
		return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
	}

	scrtName := "mres-redpanda-acl-" + obj.Name

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, scrtName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret (%s) does not exist, will be creating it shortly...", fn.NN(obj.Namespace, obj.Name).String())
	}

	if scrt == nil {
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       scrtName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.ACLUserCreds{
					KafkaBrokers: adminCreds.KafkaBrokers,
					Username:     obj.Name,
					Password:     fn.CleanerNanoid(40),
				},
			},
		)

		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}

		checks[RedpandaUserReady] = check
		return req.UpdateStatus()
	}

	aclUserCreds, err := fn.ParseFromSecret[types.ACLUserCreds](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	rApi.SetLocal(req, KeyAdminCreds, adminCreds)
	rApi.SetLocal(req, KeyAccessCreds, aclUserCreds)

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) reconACLUser(req *rApi.Request[*redpandaMsvcv1.ACLUser]) stepResult.Result {
	obj, checks := req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	adminCreds, ok := rApi.GetLocal[*types.AdminUserCreds](req, KeyAdminCreds)
	if !ok {
		return req.CheckFailed(RedpandaUserReady, check, errors.NotInLocals(KeyAdminCreds).Error()).Err(nil)
	}

	aclUserCreds, ok := rApi.GetLocal[*types.ACLUserCreds](req, KeyAccessCreds)
	if !ok {
		return req.CheckFailed(RedpandaUserReady, check, errors.NotInLocals(KeyAccessCreds).Error()).Err(nil)
	}

	adminCli := redpanda.NewAdminClient(adminCreds.AdminEndpoint, "", nil)

	userExists, err := adminCli.UserExists(aclUserCreds.Username)
	if err != nil {
		return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
	}

	if !userExists {
		if err := adminCli.CreateUser(aclUserCreds.Username, aclUserCreds.Password); err != nil {
			req.Logger.Error(err)
			return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
		}
	}

	// if err := adminCli.AllowUserOnTopics(aclUserCreds.Username, r.Env.AclAllowedOperations, obj.Spec.Topics...); err != nil {
	if err := adminCli.AllowUserOnTopics(aclUserCreds.Username, "all", obj.Spec.Topics...); err != nil {
		req.Logger.Error(err)
		return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[RedpandaUserReady] {
		checks[RedpandaUserReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&redpandaMsvcv1.ACLUser{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
