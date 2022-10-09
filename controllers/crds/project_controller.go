package crds

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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

// ProjectReconciler reconciles a ProjectRef object
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
	NamespaceExists    conditions.Type = "project.namespace/Exists"
	HarborProjectReady conditions.Type = "project.harbor-project/Ready"
	HarborUserReady    conditions.Type = "project.harbor-user/Ready"
)

const (
	ProjectConfigMapName string = "project-config"
)

const (
	keyProjectConfigReady string = "project-config-ready"
)

// type ProjectConfig struct {
// }

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

	if step := req.EnsureChecks(keyProjectConfigReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if step := r.reconcileProjectConfig(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, nil
}

func (r *ProjectReconciler) finalize(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	return req.Finalize()
}

func (r *ProjectReconciler) reconcileProjectConfig(req *rApi.Request[*crdsv1.Project]) stepResult.Result {
	ctx, project, checks := req.Context(), req.Object, req.Object.Status.Checks

	projectCfg := &corev1.ConfigMap{}
	if err := r.Get(ctx, fn.NN(project.Name, ProjectConfigMapName), projectCfg); err != nil {
		req.Logger.Infof("project configmap does not exist, will be creating it")
	}

	check := rApi.Check{Generation: project.Generation}
	if projectCfg == nil {
		if err := r.Create(
			ctx, &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ProjectConfigMapName,
					Namespace: project.Name,
				},
				Data: map[string]string{
					"app":    "",
					"router": "",
				},
			},
		); err != nil {
			return req.CheckFailed(keyProjectConfigReady, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[keyProjectConfigReady] {
		checks[keyProjectConfigReady] = check
		if err := r.Status().Update(ctx, project); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	return req.Next()
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

	if accRef, ok := project.Labels["kloudlite.io/account-ref"]; ok {
		harborProject, err := rApi.Get(ctx, r.Client, fn.NN(project.Name, accRef), &artifactsv1.HarborProject{})
		if err != nil {
			isReady = false
			cs = append(cs, conditions.New(HarborProjectReady, false, conditions.NotFound))
			if !apiErrors.IsNotFound(err) {
				return req.FailWithStatusError(err, cs...)
			}
		}
		if !harborProject.Status.IsReady {
			isReady = false
			cs = append(cs, conditions.New(HarborProjectReady, false, conditions.NotReady))
		} else {
			cs = append(cs, conditions.New(HarborProjectReady, true, conditions.Ready))
		}
	}

	// artifacts-harbor user account is ready
	harborUser, err := rApi.Get(ctx, r.Client, fn.NN(project.Name, r.env.DockerSecretName), &artifactsv1.HarborUserAccount{})
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(HarborUserReady, false, conditions.NotFound, err.Error()))
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err, cs...)
		}
	}
	if !harborUser.Status.IsReady {
		isReady = false
		cs = append(cs, conditions.New(HarborUserReady, false, conditions.NotReady))
	} else {
		cs = append(cs, conditions.New(HarborUserReady, true, conditions.Ready))
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
	accountRef, ok := project.Annotations[constants.AnnotationKeys.AccountRef]
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

	project.Status.OpsConditions = []metav1.Condition{}
	// project.Status.Generation = project.Generation
	if err := r.Status().Update(ctx, project); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
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
		Watches(
			&source.Kind{Type: &artifactsv1.HarborProject{}}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					var projects crdsv1.ProjectList
					if err := r.List(
						context.TODO(), &projects, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{
									"kloudlite.io/account-ref": obj.GetLabels()["kloudlite.io/account-ref"],
								},
							),
						},
					); err != nil {
						return nil
					}
					reqs := make([]reconcile.Request, 0, len(projects.Items))
					for _, item := range projects.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(item.Namespace, item.Name)})
					}
					return reqs
				},
			),
		).
		Complete(r)
}
