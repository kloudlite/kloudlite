package crds

import (
	"context"

	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
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
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
}

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
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &crdsv1.Project{})
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

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) finalize(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	return req.Finalize()
}

func (r *ProjectReconciler) reconcileStatus(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
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

func (r *ProjectReconciler) reconcileOperations(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx := req.Context()
	project := req.Object

	if !fn.ContainsFinalizers(project, constants.CommonFinalizer, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(project, constants.CommonFinalizer)
		controllerutil.AddFinalizer(project, constants.ForegroundFinalizer)
		if err := r.Update(ctx, project); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	var dockerConfigJson []byte
	accountRef, ok := project.Annotations[constants.AnnotationKeys.Account]
	if !ok {
		req.Logger.Infof("project=%s does not have any account id annotation", project.Name)
	}

	b, err := templates.Parse(
		templates.Project, map[string]any{
			"name": project.Name,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
			"account-ref":        accountRef,
			"docker-config-json": string(dockerConfigJson),
			"docker-secret-name": r.env.DockerSecretName,
			"role-name":          r.env.NamespaceAdminRoleName,
			"svc-account-name":   r.env.NamespaceSvcAccountName,
		},
	)

	if err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		return req.FailWithOpError(err).Err(nil)
	}
	return req.Next()
}

func (r *ProjectReconciler) GetName() string {
	return "project"
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName("project")
	r.env = envVars

	harborCli, err := harbor.NewClient(
		harbor.Args{
			HarborAdminUsername: r.env.HarborAdminUsername,
			HarborAdminPassword: r.env.HarborAdminPassword,
			HarborRegistryHost:  r.env.HarborImageRegistryHost,
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
		Owns(&artifactsv1.HarborUserAccount{}).
		Complete(r)
}
