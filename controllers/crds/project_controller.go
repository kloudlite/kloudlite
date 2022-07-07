package crds

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/lib/types"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	types.MessageSender
	Env       env.Env
	harborCli *harbor.Client
}

const (
	KeyRobotAccId        string = "robotAccId"
	KeyRobotUserName     string = "robotUserName"
	KeyRobotUserPassword string = "robotUserPassword"
)

const (
	HarborProjectExists        conditions.Type = "HarborProjectExists"
	HarborProjectAccountExists conditions.Type = "HarborProjectAccountExists"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *ProjectReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &crdsv1.Project{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabels(); !x.ShouldProceed() {
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

func (r *ProjectReconciler) finalize(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	// TODO: delete correspoding harbor account, and user account
	return req.Finalize()
}

func (r *ProjectReconciler) reconcileStatus(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	ctx := req.Context()
	project := req.Object

	isReady := true
	var cs []metav1.Condition

	ns, err := rApi.Get(ctx, r.Client, fn.NN(project.Namespace, project.Name), &corev1.Namespace{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("NamespaceExists", false, "NotFound", err.Error()))
		ns = nil
	}

	if ns != nil {
		cs = append(cs, conditions.New("NamespaceExists", true, "Found"))
	}

	ok, err := r.harborCli.CheckIfProjectExists(ctx, project.Name)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !ok {
		isReady = false
		cs = append(cs, conditions.New(HarborProjectExists, false, conditions.NotFound))
	} else {
		cs = append(cs, conditions.New(HarborProjectExists, true, conditions.Found))
	}

	if project.Status.GeneratedVars.Exists(KeyRobotAccId) {
		robotAccId, ok := project.Status.GeneratedVars.GetInt(KeyRobotAccId)
		if ok {
			ok2, err := r.harborCli.CheckIfAccountExists(ctx, robotAccId)
			if err != nil {
				return req.FailWithStatusError(err)
			}
			if !ok2 {
				cs = append(cs, conditions.New("HarborProjectAccountExists", false, "NotFound"))
			} else {
				cs = append(cs, conditions.New("HarborProjectAccountExists", true, "Found"))
			}
		}
	} else {
		isReady = false
		cs = append(cs, conditions.New("HarborProjectAccountExists", false, "NotFound"))
	}

	newConditions, hasUpdated, err := conditions.Patch(project.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}
	if !hasUpdated && isReady == project.Status.IsReady {
		return req.Next()
	}

	project.Status.IsReady = isReady
	project.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, project); err != nil {
		return req.FailWithOpError(err)
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

func (r *ProjectReconciler) reconcileOperations(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	ctx := req.Context()
	project := req.Object

	if !controllerutil.ContainsFinalizer(project, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(project, constants.CommonFinalizer)
		controllerutil.AddFinalizer(project, constants.ForegroundFinalizer)
		return req.FailWithOpError(r.Update(ctx, project))
	}

	accountRef, ok := project.Annotations[constants.AccountAnnotation]
	if !ok {
		return req.FailWithOpError(errors.Newf("could not read kloudlite acocunt annotation from project resource"))
	}

	if err2 := func() error {
		if meta.IsStatusConditionFalse(project.Status.Conditions, HarborProjectExists.String()) {
			if err := r.harborCli.CreateProject(ctx, accountRef); err != nil {
				return errors.NewEf(err, "creating harbor project")
			}
		}
		if meta.IsStatusConditionFalse(project.Status.Conditions, HarborProjectAccountExists.String()) {
			userAcc, err := r.harborCli.CreateUserAccount(ctx, accountRef)
			if err != nil {
				return errors.NewEf(err, "creating harbor project user-account")
			}
			if err := project.Status.GeneratedVars.Set(KeyRobotAccId, userAcc.Id); err != nil {
				return errors.NewEf(err, "could not set robotAccId")
			}
			if err := project.Status.GeneratedVars.Set(KeyRobotUserName, userAcc.Name); err != nil {
				return errors.NewEf(err, "could not set robotUserName")
			}
			if err := project.Status.GeneratedVars.Set(KeyRobotUserPassword, userAcc.Secret); err != nil {
				return errors.NewEf(err, "could not set robotUserPassword")
			}
			return r.Status().Update(ctx, project)
		}
		return nil
	}(); err2 != nil {
		return req.FailWithOpError(err2)
	}

	robotUserName, ok := project.Status.GeneratedVars.GetString(KeyRobotUserName)
	if !ok {
		return req.FailWithOpError(errors.Newf("key: %s not found in .GeneratedVars", KeyRobotUserName))
	}
	robotUserPassword, ok := project.Status.GeneratedVars.GetString(KeyRobotUserPassword)
	if !ok {
		return req.FailWithOpError(errors.Newf("key: %s not found in .GeneratedVars", KeyRobotUserPassword))
	}

	hDockerConfig, err := getDockerConfig(r.Env.HarborImageRegistryHost, robotUserName, robotUserPassword)
	if err != nil {
		return req.FailWithOpError(err)
	}

	b, err := templates.Parse(
		templates.Project, map[string]any{
			"name": project.Name,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
			"docker-config-json": string(hDockerConfig),
			"docker-secret-name": "kloudlite-docker-registry",
			"role-name":          "kloudlite-ns-admin",
			"svc-account-name":   "kloudlite-svc-account",
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}
	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&crdsv1.Project{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
