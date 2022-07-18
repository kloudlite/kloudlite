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
	Env       *env.Env
	harborCli *harbor.Client
}

const (
	KeyRobotAccId           string = "robotAccId"
	KeyRobotUserName        string = "robotUserName"
	KeyRobotUserPassword    string = "robotUserPassword"
	KeyHarborProjectStorage string = "KeyHarborProjectStorage"
)

const (
	HarborProjectExists           conditions.Type = "HarborProjectExists"
	HarborProjectAccountExists    conditions.Type = "HarborProjectAccountExists"
	HarborProjectStorageAllocated conditions.Type = "HarborProjectStorageAllocated"
	NamespaceExists               conditions.Type = "NamespaceExists"
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

func (r *ProjectReconciler) finalize(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	ctx := req.Context()
	project := req.Object

	// accountRef, ok := project.Annotations[constants.AnnotationKeys.Account]
	// if !ok {
	// 	return req.FailWithOpError(errors.Newf("could not read kloudlite acocunt annotation from project resource"))
	// }
	//

	if err := func() error {
		robotAccId, ok := project.Status.GeneratedVars.GetInt(KeyRobotAccId)
		if !ok {
			return errors.Newf("key: %s is not found in .Status.GeneratedVars", KeyRobotAccId)
		}
		if err := r.harborCli.DeleteUserAccount(ctx, robotAccId); err != nil {
			return err
		}
		// if err := r.harborCli.DeleteProject(ctx, accountRef); err != nil {
		// 	return err
		// }
		return nil
	}(); err != nil {
		return req.FailWithOpError(err)
	}

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
		cs = append(cs, conditions.New(NamespaceExists, false, conditions.NotFound, err.Error()))
		ns = nil
	}

	if ns != nil {
		cs = append(cs, conditions.New(NamespaceExists, true, conditions.Found))
	}

	accountRef, ok := project.Annotations[constants.AnnotationKeys.Account]
	if !ok {
		// TODO: we need to have account labels on all the projects
		return rApi.NewStepResult(&ctrl.Result{}, nil)
		// return req.FailWithStatusError(
		// 	errors.Newf(
		// 		"Account Annotation (%s) not found in resource",
		// 		constants.AnnotationAccount,
		// 	),
		// )
	}

	ok, err = r.harborCli.CheckIfProjectExists(ctx, accountRef)
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
			ok2, err := r.harborCli.CheckIfUserAccountExists(ctx, robotAccId)
			if err != nil {
				return req.FailWithStatusError(err)
			}
			if !ok2 {
				cs = append(cs, conditions.New(HarborProjectAccountExists, false, conditions.NotFound))
			} else {
				cs = append(cs, conditions.New(HarborProjectAccountExists, true, conditions.Found))
			}
		}
	} else {
		isReady = false
		cs = append(cs, conditions.New(HarborProjectAccountExists, false, conditions.NotFound))
	}

	// if project.Spec.ArtifactRegistry.Enabled {
	// 	hStorage := project.Spec.ArtifactRegistry.Size
	// 	allocatedStorage, ok := project.Status.DisplayVars.GetInt(KeyHarborProjectStorage)
	// 	if !ok || hStorage != allocatedStorage {
	// 		isReady = false
	// 		cs = append(cs, conditions.New(HarborProjectStorageAllocated, false, conditions.NotReconciledYet))
	// 	} else {
	// 		cs = append(cs, conditions.New(HarborProjectStorageAllocated, true, conditions.Found))
	// 	}
	// }

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

	var dockerConfigJson []byte

	if accountRef, ok := project.Annotations[constants.AnnotationKeys.Account]; ok {
		// TODO: harbor project creation should be moved out to Account Creation
		// if meta.IsStatusConditionFalse(project.Status.Conditions, HarborProjectExists.String()) {
		// 	if err2 := func() error {
		// 		storageSize := 1000 * r.Env.HarborProjectStorageSize
		// 		// if project.Spec.ArtifactRegistry.Enabled && project.Spec.ArtifactRegistry.Size > 0 {
		// 		// 	storageSize = project.Spec.ArtifactRegistry.Size
		// 		// }
		// 		if err := r.harborCli.CreateProject(ctx, accountRef, storageSize); err != nil {
		// 			return errors.NewEf(err, "creating harbor project")
		// 		}
		// 		return project.Status.DisplayVars.Set(KeyHarborProjectStorage, storageSize)
		// 	}(); err2 != nil {
		// 		return req.FailWithOpError(err2)
		// 	}
		// 	if err := r.Status().Update(ctx, project); err != nil {
		// 		return req.FailWithOpError(err)
		// 	}
		// 	return req.Done(&ctrl.Result{RequeueAfter: 0})
		// }

		if meta.IsStatusConditionFalse(project.Status.Conditions, HarborProjectAccountExists.String()) {
			if err3 := func() error {
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
				return nil
			}(); err3 != nil {
				return req.FailWithOpError(err3)
			}
			return req.FailWithOpError(r.Status().Update(ctx, project))
		}

		// TODO: harbor project storage size allocation should be moved out to Account Creation
		// if meta.IsStatusConditionFalse(project.Status.Conditions, HarborProjectStorageAllocated.String()) {
		// 	if err := r.harborCli.SetProjectQuota(ctx, project.Name, project.Spec.ArtifactRegistry.Size); err != nil {
		// 		return req.FailWithOpError(err)
		// 	}
		// 	if err := project.Status.DisplayVars.Set(
		// 		KeyHarborProjectStorage,
		// 		project.Spec.ArtifactRegistry.Size,
		// 	); err != nil {
		// 		return nil
		// 	}
		// 	return req.FailWithOpError(r.Status().Update(ctx, project))
		// }

		robotUserName, ok := project.Status.GeneratedVars.GetString(KeyRobotUserName)
		if !ok {
			return req.FailWithOpError(errors.Newf("key: %s not found in .GeneratedVars", KeyRobotUserName))
		}

		robotUserPassword, ok := project.Status.GeneratedVars.GetString(KeyRobotUserPassword)
		if !ok {
			return req.FailWithOpError(errors.Newf("key: %s not found in .GeneratedVars", KeyRobotUserPassword))
		}

		// hDockerConfig, err := getDockerConfig(r.Env.HarborImageRegistryHost, robotUserName, robotUserPassword)
		harborDockerConfig, err := getDockerConfig(r.Env.HarborImageRegistryHost, robotUserName, robotUserPassword)
		if err != nil {
			return req.FailWithOpError(err)
		}
		dockerConfigJson = harborDockerConfig
	}

	b, err := templates.Parse(
		templates.Project, map[string]any{
			"name": project.Name,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
			"docker-config-json": string(dockerConfigJson),
			"docker-secret-name": r.Env.DockerSecretName,
			"role-name":          r.Env.NamespaceAdminRoleName,
			"svc-account-name":   r.Env.NamespaceSvcAccountName,
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
