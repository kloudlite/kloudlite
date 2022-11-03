package user_account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/kubectl"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/env"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	req.Logger.Infof("NEW RECONCILATION")
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

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

	if step := r.reconOutput(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconRobotAccount(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
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

	if obj.Spec.HarborUser.Id > 0 {
		if err := r.HarborCli.DeleteUserAccount(ctx, obj.Spec.HarborUser); err != nil {
			return req.CheckFailed(userDeleted, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.DockerConfigName == "" {
		obj.Spec.DockerConfigName = r.Env.DockerSecretName
	}

	if obj.Spec.HarborUser == nil {
		obj.Spec.HarborUser = &harbor.User{
			Name: fmt.Sprintf("robot-%s+%s-%s", obj.Spec.ProjectRef, obj.Namespace, obj.Name),
		}
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

func (r *Reconciler) patchDockerSecret(req *rApi.Request[*artifactsv1.HarborUserAccount], username, password string) error {
	ctx, obj := req.Context(), req.Object
	dockerScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.DockerConfigName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("docker secret, does not exist will be creating now...")
	}

	harborDockerConfig, err := getDockerConfig(r.Env.HarborImageRegistryHost, username, password)
	if err != nil {
		return err
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, dockerScrt, func() error {
		b, err := json.Marshal(types.UserAccountOutput{
			DockerConfigJson: string(harborDockerConfig),
			Username:         username,
			Password:         password,
			Registry:         r.Env.HarborImageRegistryHost,
			Project:          obj.Spec.ProjectRef,
		})
		if err != nil {
			return err
		}

		json.Unmarshal(b, &dockerScrt.StringData)
		return nil
	}); err != nil {
		return err
	}
	return nil
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

	svcAccount.ImagePullSecrets = append(svcAccount.ImagePullSecrets, corev1.LocalObjectReference{
		Name: secretName,
	})

	return client.Update(ctx, &svcAccount)
}

func (r *Reconciler) reconOutput(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	dockerScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.DockerConfigName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("docker secret, does not exist will be creating now...")
	}

	if dockerScrt == nil {
		b, err := templates.Parse(
			templates.CoreV1.Secret, map[string]any{
				"name":       obj.Spec.DockerConfigName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.UserAccountOutput{
					DockerConfigJson: string("{}"),
				},
				"secret-type": "kubernetes.io/dockerconfigjson",
				"immutable":   false,
			},
		)
		if err != nil {
			return req.CheckFailed(OutputReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(OutputReady, check, err.Error()).Err(nil)
		}

		checks[OutputReady] = check
		return req.UpdateStatus().RequeueAfter(2 * time.Second)
	}

	if fn.IsOwner(obj, fn.AsOwner(dockerScrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(dockerScrt)))
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(OutputReady, check, err.Error())
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	if err := patchServiceAccount(ctx, r.Client, obj.Namespace, r.Env.ServiceAccountName, obj.Spec.DockerConfigName); err != nil {
		return req.CheckFailed(OutputReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[OutputReady] {
		checks[OutputReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconRobotAccount(req *rApi.Request[*artifactsv1.HarborUserAccount]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	exists, err := r.HarborCli.CheckIfUserAccountExists(ctx, obj.Spec.HarborUser)
	if err != nil {
		return req.CheckFailed(RobotAccountReady, check, err.Error())
	}

	if !exists {
		user, err := r.HarborCli.CreateUserAccount(ctx, obj.Spec.ProjectRef, fn.Md5([]byte(fmt.Sprintf("%s-%s", obj.Namespace, obj.Name))))
		if err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error())
		}

		if err := r.patchDockerSecret(req, user.Name, user.Password); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error())
		}

		obj.Spec.HarborUser = user
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error())
		}

		checks[RobotAccountReady] = check
		return req.UpdateStatus().RequeueAfter(2 * time.Second)
	}

	if check.Generation > checks[RobotAccountReady].Generation {
		if err := r.HarborCli.UpdateUserAccount(ctx, obj.Spec.HarborUser, obj.Spec.Enabled); err != nil {
			return req.CheckFailed(RobotAccountReady, check, err.Error())
		}

		checks[RobotAccountReady] = check
		return req.UpdateStatus()
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
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
