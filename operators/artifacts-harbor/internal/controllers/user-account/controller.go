package user_account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/env"
	"operators.kloudlite.io/pkg/constants"
	"operators.kloudlite.io/pkg/errors"
	fn "operators.kloudlite.io/pkg/functions"
	"operators.kloudlite.io/pkg/harbor"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	HarborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched   string = "defaults-patched"
	RobotAccountReady string = "robot-account-ready"
	OutputReady       string = "output-ready"
)

// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harboruseraccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harboruseraccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harboruseraccounts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		context.WithValue(ctx, "logger", r.logger),
		r.Client,
		request.NamespacedName,
		&artifactsv1.HarborUserAccount{},
	)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(DefaultsPatched, OutputReady, RobotAccountReady); !step.ShouldProceed() {
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

	//if step := r.reconOutput(req); !step.ShouldProceed() {
	//	return step.ReconcilerResponse()
	//}

	if step := r.ensureRobotAccount(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	//if step := r.reconRobotAccount(req); !step.ShouldProceed() {
	//	return step.ReconcilerResponse()
	//}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	userDeleted := "user-deleted"
	if step := req.EnsureChecks(userDeleted); !step.ShouldProceed() {
		return step
	}

	check := rApi.Check{Generation: obj.Generation}

	req.Logger.Infof("deleting user")
	defer func() {
		req.Logger.Infof("deleting user (deleted: %s)", check.Status)
	}()

	if obj.Spec.OperatorProps.HarborUser != nil {
		if err := r.HarborCli.DeleteUserAccount(ctx, int64(obj.Spec.OperatorProps.HarborUser.Id)); err != nil {
			return req.CheckFailed(userDeleted, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.DockerConfigName == "" {
		hasUpdated = true
		obj.Spec.DockerConfigName = r.Env.DockerSecretName
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	checks := req.Object.Status.Checks
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func getDockerConfig(imageRegistry, username, password string) ([]byte, error) {
	encAuthPass := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", username, password)),
	)

	return json.Marshal(
		map[string]any{
			"auths": map[string]any{
				imageRegistry: map[string]any{
					"auth": encAuthPass,
				},
			},
		},
	)
}

func patchServiceAccount(ctx context.Context, client client.Client, namespace, svcAccName, secretName string) error {
	var svcAccount corev1.ServiceAccount
	if err := client.Get(ctx, fn.NN(namespace, svcAccName), &svcAccount); err != nil {
		return err
	}

	for _, secret := range svcAccount.ImagePullSecrets {
		if secret.Name == secretName {
			return nil
		}
	}

	svcAccount.ImagePullSecrets = append(
		svcAccount.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secretName,
		},
	)

	return client.Update(ctx, &svcAccount)
}

func (r *Reconciler) ensureRobotAccount(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RobotAccountReady)
	defer req.LogPostCheck(RobotAccountReady)

	robotUsername := fmt.Sprintf("%s-%s", obj.Namespace, obj.Name)
	robotUser, err := r.HarborCli.FindUserAccountByName(ctx, obj.Spec.ProjectRef, robotUsername)
	if err != nil {
		httpErr, ok := err.(*errors.HttpError)
		if !ok || httpErr.Code != http.StatusNotFound {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("robot account (%s) does not exist, will be creating now...", robotUsername)
	}

	harborAccessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.DockerConfigName), &crdsv1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
		harborAccessSecret = nil
	}

	if harborAccessSecret != nil {
		if !fn.IsOwner(obj, fn.AsOwner(harborAccessSecret)) {
			obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(harborAccessSecret)))
			if err := r.Update(ctx, obj); err != nil {
				return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
			}
			return req.Done().RequeueAfter(1 * time.Second)
		}
	}

	if robotUser != nil && harborAccessSecret == nil {
		if err := r.HarborCli.DeleteUserAccount(ctx, int64(robotUser.Id)); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
		robotUser = nil
	}

	if robotUser == nil {
		user, err := r.HarborCli.CreateUserAccount(ctx, obj.Spec.ProjectRef, robotUsername)
		if err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}

		dockerCfg, err := getDockerConfig(r.Env.HarborImageRegistryHost, user.Name, user.Password)
		if err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}

		secret := &crdsv1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.DockerConfigName, Namespace: obj.Namespace}, Type: corev1.SecretTypeDockerConfigJson}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
			if !fn.IsOwner(secret, fn.AsOwner(obj)) {
				secret.SetOwnerReferences(append(secret.GetOwnerReferences(), fn.AsOwner(obj, true)))
			}
			if secret.Data == nil {
				secret.Data = make(map[string][]byte, 1)
				secret.StringData = make(map[string]string, 5)
			}
			secret.Data[".dockerconfigjson"] = dockerCfg
			secret.StringData["username"] = user.Name
			secret.StringData["password"] = user.Password
			secret.StringData["registry"] = r.Env.HarborImageRegistryHost
			secret.StringData["project"] = obj.Spec.ProjectRef
			secret.StringData["harborUserId"] = fmt.Sprintf("%d", user.Id)
			return nil
		}); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if check.Generation > checks[RobotAccountReady].Generation {
		if err := r.HarborCli.UpdateUserAccount(ctx, int64(robotUser.Id), obj.Spec.Enabled); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[RobotAccountReady] {
		checks[RobotAccountReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&artifactsv1.HarborUserAccount{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&crdsv1.Secret{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
