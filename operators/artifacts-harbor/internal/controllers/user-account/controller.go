package user_account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	"github.com/kloudlite/operator/operators/artifacts-harbor/internal/env"
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
	"k8s.io/apimachinery/pkg/runtime"
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
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &artifactsv1.HarborUserAccount{})
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

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRobotAccount(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, req.Object); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	userDeleted := "user-deleted"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(userDeleted)
	defer req.LogPreCheck(userDeleted)

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(userDeleted, check, err.Error())
	}

	var dockerSecret corev1.Secret
	if err := r.Get(ctx, fn.NN(obj.Namespace, obj.Spec.TargetSecret), &dockerSecret); err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
		return req.Finalize()
	}

	if userId, ok := dockerSecret.Data["harborUserId"]; ok {
		hUserId, err := strconv.ParseInt(string(userId), 10, 64)
		if err != nil {
			return failed(err).Err(nil)
		}
		if err := r.HarborCli.DeleteUserAccount(ctx, hUserId); err != nil {
			return failed(err)
		}
	}

	if dockerSecret.GetDeletionTimestamp() == nil {
		if err := r.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: obj.Namespace, Name: obj.Spec.TargetSecret}}); err != nil {
			return failed(err)
		}
	}

	controllerutil.RemoveFinalizer(&dockerSecret, constants.CommonFinalizer2)
	if err := r.Update(ctx, &dockerSecret); err != nil {
		return failed(err)
	}

	check.Status = true
	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.TargetSecret == "" {
		hasUpdated = true
		obj.Spec.TargetSecret = fmt.Sprintf("harbor-creds-%s", obj.Name)
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	if check != obj.Status.Checks[DefaultsPatched] {
		obj.Status.Checks[DefaultsPatched] = check
		req.UpdateStatus()
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
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RobotAccountReady)
	defer req.LogPostCheck(RobotAccountReady)

	robotUsername := fmt.Sprintf("%s-%s", obj.Namespace, obj.Name)
	robotUser, err := r.HarborCli.FindUserAccountByName(ctx, obj.Spec.HarborProjectName, robotUsername)
	if err != nil {
		httpErr, ok := err.(*errors.HttpError)
		if !ok || httpErr.Code != http.StatusNotFound {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("robot account (%s) does not exist, will be creating now...", robotUsername)
	}

	harborAccessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.TargetSecret), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(RobotAccountReady, check, err.Error())
		}
		harborAccessSecret = nil
	}

	// if harborAccessSecret != nil {
	// 	if !fn.IsOwner(obj, fn.AsOwner(harborAccessSecret)) {
	// 		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(harborAccessSecret)))
	// 		if err := r.Update(ctx, obj); err != nil {
	// 			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
	// 		}
	// 		return req.Done().RequeueAfter(1 * time.Second)
	// 	}
	// }

	if robotUser != nil && harborAccessSecret == nil {
		if err := r.HarborCli.DeleteUserAccount(ctx, int64(robotUser.Id)); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
		robotUser = nil
	}

	if robotUser != nil {
		if check.Generation > obj.Status.Checks[RobotAccountReady].Generation {
			if err := r.HarborCli.UpdateUserAccount(ctx, int64(robotUser.Id), obj.Spec.Enabled); err != nil {
				return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
			}
		}
	}

	if robotUser == nil {
		user, err := r.HarborCli.CreateUserAccount(ctx, obj.Spec.HarborProjectName, robotUsername, obj.Spec.Permissions)
		if err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}

		dockerCfg, err := getDockerConfig(r.Env.HarborImageRegistryHost, user.Name, user.Password)
		if err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}

		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetSecret, Namespace: obj.Namespace}, Type: corev1.SecretTypeDockerConfigJson}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
			if !fn.IsOwner(secret, fn.AsOwner(obj)) {
				secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
			}

			controllerutil.AddFinalizer(secret, constants.CommonFinalizer2)

			if secret.Data == nil {
				secret.Data = make(map[string][]byte, 1)
			}

			if secret.StringData == nil {
				secret.StringData = make(map[string]string, 5)
			}
			secret.Data[".dockerconfigjson"] = dockerCfg
			secret.StringData["username"] = user.Name
			secret.StringData["password"] = user.Password
			secret.StringData["registry"] = r.Env.HarborImageRegistryHost
			secret.StringData["project"] = obj.Spec.HarborProjectName
			secret.StringData["harborUserId"] = fmt.Sprintf("%d", user.Id)
			return nil
		}); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != obj.Status.Checks[RobotAccountReady] {
		obj.Status.Checks[RobotAccountReady] = check
		req.UpdateStatus()
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
