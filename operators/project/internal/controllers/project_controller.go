package controllers

import (
	"context"
	"encoding/json"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *ProjectReconciler) GetName() string {
	return r.Name
}

const (
	NamespaceReady     string = "namespace-ready"
	ProjectCfgReady    string = "project-config-ready"
	RBACReady          string = "rbac-ready"
	HarborAccessReady  string = "harbor-access-ready"
	AccountRouterReady string = "account-router-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status-watcher,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *ProjectReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.Project{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(NamespaceReady, ProjectCfgReady, RBACReady, HarborAccessReady, AccountRouterReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconProjectCfg(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconProjectRBAC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHarborAccess(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.reconAccountRouter(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *ProjectReconciler) finalize(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	return req.Finalize()
}

func (r *ProjectReconciler) reconNamespace(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.AccountRef == "" {
		accRef, ok := obj.GetAnnotations()[constants.AccountRef]
		if !ok {
			return req.CheckFailed(NamespaceReady, check, "no account-ref found in annotations").Err(nil)
		}
		if ok {
			obj.Spec.AccountRef = accRef
			if err := r.Update(ctx, obj); err != nil {
				return req.FailWithOpError(err)
			}
			return req.Done()
		}
	}

	ns := &corev1.Namespace{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, obj.Name), ns); err != nil {
		req.Logger.Infof("namespace (%s) does not exist, will be creating one now", obj.Name)
	}

	if ns == nil || check.Generation > checks[NamespaceReady].Generation {
		b, err := templates.Parse(
			templates.CoreV1.Namespace, map[string]any{
				"name":       obj.Name,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"labels": map[string]string{
					constants.ProjectNameKey: obj.Name,
					constants.AccountRef:     obj.Spec.AccountRef,
				},
			},
		)

		if err != nil {
			return req.CheckFailed(NamespaceReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(NamespaceReady, check, err.Error()).Err(nil)
		}

		checks[NamespaceReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[NamespaceReady] {
		checks[NamespaceReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ProjectReconciler) reconProjectCfg(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	projectCfg, err := rApi.Get[*corev1.ConfigMap](ctx, r.Client, fn.NN(obj.Name, r.env.ProjectCfgName), &corev1.ConfigMap{})
	if err != nil {
		projectCfg = nil
		req.Logger.Infof("obj configmap does not exist, will be creating it")
	}

	check := rApi.Check{Generation: obj.Generation}
	if projectCfg == nil || check.Generation > checks[ProjectCfgReady].Generation {
		b, err := templates.Parse(
			templates.CoreV1.ConfigMap, map[string]any{
				"name":       r.env.ProjectCfgName,
				"namespace":  obj.Name,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"data": map[string]string{
					"app":    "",
					"router": "",
				},
			},
		)
		if err != nil {
			return req.CheckFailed(ProjectCfgReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(ProjectCfgReady, check, err.Error()).Err(nil)
		}

		checks[ProjectCfgReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[ProjectCfgReady] {
		checks[ProjectCfgReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ProjectReconciler) reconProjectRBAC(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	namespace := obj.Name

	check := rApi.Check{Generation: obj.Generation}

	svcAccount, err := rApi.Get(ctx, r.Client, fn.NN(namespace, r.env.NamespaceSvcAccountName), &corev1.ServiceAccount{})
	if err != nil {
		req.Logger.Infof("service account %s does not exist, creating now...", r.env.NamespaceSvcAccountName)
	}

	role, err := rApi.Get(ctx, r.Client, fn.NN(namespace, r.env.NamespaceAdminRoleName), &rbacv1.Role{})
	if err != nil {
		req.Logger.Infof("service account %s does not exist, creating now...", r.env.NamespaceSvcAccountName)
	}

	roleBinding, err := rApi.Get(ctx, r.Client, fn.NN(namespace, r.env.NamespaceAdminRoleName+"-rb"), &rbacv1.RoleBinding{})
	if err != nil {
		req.Logger.Infof("service account %s does not exist, creating now...", r.env.NamespaceSvcAccountName)
	}

	if svcAccount == nil || role == nil || roleBinding == nil || check.Generation > checks[RBACReady].Generation {
		b, err := templates.Parse(
			templates.ProjectRBAC, map[string]any{
				"namespace":          namespace,
				"role-name":          r.env.NamespaceAdminRoleName,
				"role-binding-name":  r.env.NamespaceAdminRoleName + "-rb",
				"svc-account-name":   r.env.NamespaceSvcAccountName,
				"docker-secret-name": r.env.DockerSecretName,
				"owner-refs":         []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
		)
		if err != nil {
			return req.CheckFailed(RBACReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(RBACReady, check, err.Error()).Err(nil)
		}

		checks[RBACReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[RBACReady] {
		checks[RBACReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ProjectReconciler) reconHarborAccess(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	namespace := obj.Name

	check := rApi.Check{Generation: obj.Generation}

	harborProject, harborUserAcc := &artifactsv1.HarborProject{}, &artifactsv1.HarborUserAccount{}

	if err := r.Get(ctx, fn.NN(namespace, obj.Spec.AccountRef), harborProject); err != nil {
		req.Logger.Infof("harbor project (%s) does not exist, creating now ...", obj.Spec.AccountRef)
	}

	if err := r.Get(ctx, fn.NN(namespace, r.env.DockerSecretName), harborUserAcc); err != nil {
		req.Logger.Infof("harbor user account (%s) does not exist, creating now ...", obj.Spec.AccountRef)
	}

	if harborProject == nil || harborUserAcc == nil || check.Generation > checks[HarborAccessReady].Generation {
		b, err := templates.Parse(
			templates.ProjectHarbor, map[string]any{
				"acc-ref":            obj.Spec.AccountRef,
				"docker-secret-name": r.env.DockerSecretName,
				"namespace":          namespace,
				"owner-refs":         []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
		)
		if err != nil {
			return req.CheckFailed(HarborAccessReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(HarborAccessReady, check, err.Error()).Err(nil)
		}

		checks[HarborAccessReady] = check
		return req.UpdateStatus()
	}

	if !harborProject.Status.IsReady {
		bMessage, err := json.Marshal(harborProject.Status.Message)
		if err != nil {
			return req.CheckFailed(HarborAccessReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(HarborAccessReady, check, string(bMessage)).Err(nil)
	}

	if !harborUserAcc.Status.IsReady {
		bMessage, err := json.Marshal(harborUserAcc.Status.Message)
		if err != nil {
			return req.CheckFailed(HarborAccessReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(HarborAccessReady, check, string(bMessage)).Err(nil)
	}

	check.Status = true
	if check != checks[HarborAccessReady] {
		checks[HarborAccessReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ProjectReconciler) reconAccountRouter(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	accNamespace := "wg-" + obj.Spec.AccountRef

	accRouter, err := rApi.Get[*crdsv1.AccountRouter](ctx, r.Client, fn.NN(accNamespace, r.env.AccountRouterName), &crdsv1.AccountRouter{})
	if err != nil {
		req.Logger.Infof("account router (%s) does not exist, would be creating it now...", r.env.AccountRouterName)
	}

	if accRouter == nil {
		b, err := templates.Parse(
			templates.CrdsV1.AccountRouter, map[string]any{
				"name":      r.env.AccountRouterName,
				"namespace": accNamespace,
				"acc-ref":   obj.Spec.AccountRef,
			},
		)
		if err != nil {
			return req.CheckFailed(AccountRouterReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AccountRouterReady, check, err.Error()).Err(nil)
		}

		checks[AccountRouterReady] = check

		return req.UpdateStatus()
	}

	if !accRouter.Status.IsReady {
		bMsg, err := json.Marshal(accRouter.Status.Message)
		if err != nil {
			return req.CheckFailed(AccountRouterReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(AccountRouterReady, check, string(bMsg)).Err(nil)
	}

	check.Status = true
	if check != checks[AccountRouterReady] {
		checks[AccountRouterReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Project{})
	builder.Owns(&corev1.Namespace{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.Role{})
	builder.Owns(&rbacv1.RoleBinding{})
	builder.Owns(&artifactsv1.HarborUserAccount{})

	return builder.Complete(r)
}
