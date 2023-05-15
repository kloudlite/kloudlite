package project

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	BlueprintCreated      string = "blueprint-created"
	DefaultEnvCreated     string = "default-env-created"
	HarborAccessAvailable string = "harbor-creds-available"
	NamespacedRBACsReady  string = "namespaced-rbacs-ready"
	NamespaceExists       string = "namespace-exists"
	EnvRouteSwitcherReady string = "env-route-switcher-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &v1.Project{})
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

	if step := req.EnsureFinalizers(constants.CommonFinalizer, constants.ForegroundFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespacedRBACs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureEnvRouteSwitcher(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.Project]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) ensureNamespace(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespaceExists)
	defer req.LogPostCheck(NamespaceExists)

	ns, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.TargetNamespace), &corev1.Namespace{})
	if err != nil {
		req.Logger.Errorf(err, fmt.Sprintf("[check] %s", NamespaceExists))
		return req.CheckFailed(NamespaceExists, check, err.Error()).Err(nil)
	}

	if ns.Labels == nil {
		ns.Labels = make(map[string]string, 2)
	}

	ns.Labels[constants.AccountNameKey] = obj.Spec.AccountName
	ns.Labels[constants.ClusterNameKey] = obj.Spec.ClusterName

	if err := r.Update(ctx, ns); err != nil {
		return req.CheckFailed(NamespaceExists, check, err.Error())
	}

	check.Status = true
	if check != obj.Status.Checks[NamespaceExists] {
		obj.Status.Checks[NamespaceExists] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureNamespacedRBACs(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespacedRBACsReady)
	defer req.LogPreCheck(NamespacedRBACsReady)

	// copy docker creds
	ds, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.OperatorsNamespace, r.Env.DockerSecretName), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	nds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: r.Env.DockerSecretName, Namespace: obj.Spec.TargetNamespace}, Type: ds.Type}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, nds, func() error {
		nds.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		nds.Data = ds.Data
		return nil
	}); err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	b, err := templates.Parse(
		templates.ProjectRBAC, map[string]any{
			"namespace":          obj.Spec.TargetNamespace,
			"role-name":          r.Env.AdminRoleName,
			"role-binding-name":  r.Env.AdminRoleName + "-rb",
			"svc-account-name":   r.Env.SvcAccountName,
			"docker-secret-name": r.Env.DockerSecretName,
			"owner-refs":         []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
	)
	if err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	refs, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(refs...)
	check.Status = true
	if check != checks[NamespacedRBACsReady] {
		checks[NamespacedRBACsReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureEnvRouteSwitcher(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(EnvRouteSwitcherReady)
	defer req.LogPostCheck(EnvRouteSwitcherReady)

	d := &v1.App{ObjectMeta: metav1.ObjectMeta{Name: "env-route-switcher", Namespace: obj.Spec.TargetNamespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, d, func() error {
		d.Spec = v1.AppSpec{
			DisplayName: "env router switcher",
			Replicas:    0,
			Services: []v1.AppSvc{
				{
					Port:       80,
					TargetPort: 80,
					Type:       "tcp",
					Name:       "http",
				},
			},
			Containers: []v1.AppContainer{
				{
					Name:            "main",
					Image:           "registry.kloudlite.io/public/env-route-switcher:v1.0.5",
					ImagePullPolicy: "Always",
					LivenessProbe: &v1.Probe{
						Type: "httpGet",
						HttpGet: &v1.HttpGetProbe{
							Path: "/.kl/healthz",
							Port: 80,
						},
						FailureThreshold: 3,
						InitialDelay:     3,
						Interval:         10,
					},
					ReadinessProbe: &v1.Probe{
						Type: "httpGet",
						HttpGet: &v1.HttpGetProbe{
							Path: "/.kl/healthz",
							Port: 80,
						},
						FailureThreshold: 3,
						InitialDelay:     3,
						Interval:         10,
					},
				},
			},
		}
		return nil
	}); err != nil {
		return req.CheckFailed(EnvRouteSwitcherReady, check, err.Error()).Err(nil)
	}

	a, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.TargetNamespace, "env-route-switcher"), &v1.App{})
	if err != nil {
		return req.CheckFailed(EnvRouteSwitcherReady, check, err.Error())
	}

	if !a.Status.IsReady {
		return req.CheckFailed(EnvRouteSwitcherReady, check, "waiting for env-route-switcher to be ready").Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[EnvRouteSwitcherReady] {
		obj.Status.Checks[EnvRouteSwitcherReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Project{})
	builder.Owns(&corev1.Namespace{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.Role{})
	builder.Owns(&rbacv1.RoleBinding{})

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
