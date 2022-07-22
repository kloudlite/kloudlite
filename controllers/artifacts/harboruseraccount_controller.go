package artifacts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// HarborUserAccountReconciler reconciles a HarborUserAccount object
type HarborUserAccountReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Env       *env.Env
	harborCli *harbor.Client
}

func (r *HarborUserAccountReconciler) GetName() string {
	return "artifact-harbor-user-account"
}

const (
	KeyRobotAccId        string = "robotAccId"
	KeyRobotUserName     string = "robotUserName"
	KeyRobotUserPassword string = "robotUserPassword"
)

const (
	HarborUserAccountExists conditions.Type = "HarborUserAccountExists"
	HarborProjectReady      conditions.Type = "HarborProjectReady"
)

func getUsername(hAcc *artifactsv1.HarborUserAccount) string {
	return fmt.Sprintf("%s-%s", hAcc.Namespace, hAcc.Name)
	// return fmt.Sprintf("%s-%s", strings.ToLower(fn.CleanerNanoid(60)), hAcc.Name)
}

// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harboruseraccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harboruseraccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harboruseraccounts/finalizers,verbs=update

func (r *HarborUserAccountReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &artifactsv1.HarborUserAccount{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// STEP: cleaning up last run, clearing opsConditions
	if len(req.Object.Status.OpsConditions) > 0 {
		req.Object.Status.OpsConditions = []metav1.Condition{}
		return ctrl.Result{RequeueAfter: 0}, r.Status().Update(ctx, req.Object)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("----------------[Type: artifactsv1.HarborUserAccount] NEW RECONCILATION ----------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *HarborUserAccountReconciler) finalize(req *rApi.Request[*artifactsv1.HarborUserAccount]) rApi.StepResult {
	robotAccId, ok := req.Object.Status.GeneratedVars.GetInt(KeyRobotAccId)
	if !ok {
		return req.FailWithOpError(errors.Newf("Key=%s not found in GeneratedVars", KeyRobotAccId))
	}
	if err := r.harborCli.DeleteUserAccount(req.Context(), robotAccId); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "deleting harbor user account (id=%d)", robotAccId))
	}
	return req.Finalize()
}

func (r *HarborUserAccountReconciler) reconcileStatus(req *rApi.Request[*artifactsv1.HarborUserAccount]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	hProj, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ProjectRef), &artifactsv1.HarborProject{})
	if err != nil {
		cs = append(cs, conditions.New(HarborProjectExists, false, conditions.NotFound, err.Error()))
		nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
		if err != nil {
			return req.FailWithStatusError(err)
		}
		if !hasUpdated {
			return rApi.NewStepResult(&reconcile.Result{}, nil)
		}
		obj.Status.Conditions = nConditions
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
	}

	if !hProj.Status.IsReady {
		cs = append(cs, conditions.New(HarborProjectReady, false, conditions.NotReady))
		nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
		if err != nil {
			return req.FailWithStatusError(err)
		}
		if !hasUpdated {
			return rApi.NewStepResult(&reconcile.Result{}, nil)
		}
		obj.Status.Conditions = nConditions
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
	}

	// check if user account exists
	if obj.Status.GeneratedVars.Exists(KeyRobotAccId) {
		robotAccId, _ := obj.Status.GeneratedVars.GetInt(KeyRobotAccId)
		userAccExists, err := r.harborCli.CheckIfUserAccountExists(ctx, robotAccId)
		if err != nil {
			isReady = false
			cs = append(cs, conditions.New(HarborUserAccountExists, false, conditions.NotFound, err.Error()))
		}
		if !userAccExists {
			isReady = false
			cs = append(cs, conditions.New(HarborUserAccountExists, false, conditions.NotFound))
		}
		cs = append(cs, conditions.New(HarborUserAccountExists, true, conditions.Found))
	} else {
		isReady = false
		cs = append(cs, conditions.New(HarborUserAccountExists, false, conditions.NotFound))
	}

	nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		req.FailWithStatusError(err)
	}
	if !hasUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.Conditions = nConditions
	obj.Status.IsReady = isReady
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}

	return req.Done()
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

func (r *HarborUserAccountReconciler) reconcileOperations(req *rApi.Request[*artifactsv1.HarborUserAccount]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, HarborUserAccountExists.String()) {

		if err := func() error {
			if !obj.Status.GeneratedVars.Exists(KeyRobotAccId) {
				userAcc, err := r.harborCli.CreateUserAccount(ctx, obj.Spec.ProjectRef, getUsername(obj))
				if err != nil {
					return errors.NewEf(err, "creating harbor project user-account")
				}
				if userAcc == nil {
					return nil
				}
				if err := obj.Status.GeneratedVars.Set(KeyRobotAccId, userAcc.Id); err != nil {
					return errors.NewEf(err, "could not set robotAccId")
				}
				if err := obj.Status.GeneratedVars.Set(KeyRobotUserName, userAcc.Name); err != nil {
					return errors.NewEf(err, "could not set robotUserName")
				}
				if err := obj.Status.GeneratedVars.Set(KeyRobotUserPassword, userAcc.Secret); err != nil {
					return errors.NewEf(err, "could not set robotUserPassword")
				}
				return nil
			}
			robotAccId, _ := obj.Status.GeneratedVars.GetInt(KeyRobotAccId)
			return r.harborCli.UpdateUserAccount(ctx, robotAccId, obj.Spec.Disable)
		}(); err != nil {
			return req.FailWithOpError(err)
		}

		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{RequeueAfter: 0}, nil)
	}

	robotAccId, ok := obj.Status.GeneratedVars.GetInt(KeyRobotAccId)
	if !ok {
		return req.FailWithOpError(errors.Newf("Key=%s not found in GeneratedVars", KeyRobotAccId))
	}
	robotUserName, ok := obj.Status.GeneratedVars.GetString(KeyRobotUserName)
	if !ok {
		return req.FailWithOpError(errors.Newf("Key=%s not found in GeneratedVars", KeyRobotUserName))
	}
	robotUserPassword, ok := obj.Status.GeneratedVars.GetString(KeyRobotUserPassword)
	if !ok {
		return req.FailWithOpError(errors.Newf("Key=%s not found in GeneratedVars", KeyRobotUserPassword))
	}

	harborDockerConfig, err := getDockerConfig(r.Env.HarborImageRegistryHost, robotUserName, robotUserPassword)
	if err != nil {
		return req.FailWithOpError(err)
	}

	b, err := templates.Parse(
		templates.CoreV1.DockerConfigSecret, map[string]any{
			"name":      obj.Name,
			"namespace": obj.Namespace,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(obj, true),
			},
			"docker-config-json": string(harborDockerConfig),
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}

	if err := r.harborCli.UpdateUserAccount(ctx, robotAccId, obj.Spec.Disable); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *HarborUserAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	harborCli, err := harbor.NewClient(
		harbor.Args{
			HarborAdminUsername: r.Env.HarborAdminUsername,
			HarborAdminPassword: r.Env.HarborAdminPassword,
			HarborRegistryHost:  r.Env.HarborImageRegistryHost,
		},
	)
	if err != nil {
		return nil
	}
	r.harborCli = harborCli

	return ctrl.NewControllerManagedBy(mgr).
		For(&artifactsv1.HarborUserAccount{}).
		Watches(
			&source.Kind{Type: &artifactsv1.HarborProject{}}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					var userAccList artifactsv1.HarborUserAccountList
					if err := r.List(
						context.TODO(), &userAccList, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{
									constants.LabelKeys.HarborProjectRef: obj.GetName(),
								},
							),
						},
					); err != nil {
						return nil
					}

					var reqs []reconcile.Request
					for _, item := range userAccList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(item.Namespace, item.Name)})
					}
					return reqs
				},
			),
		).
		Complete(r)
}
