package project

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	v1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	NamespacedRBACsReady        string = "namespaced-rbacs-ready"
	NamespaceExists             string = "namespace-exists"
	WorkspaceRouteSwitcherReady string = "workspace-route-switcher-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &v1.Project{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

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

	if step := r.ensureWorkspaceRouteSwitcher(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
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
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(NamespaceExists, check, err.Error()).Err(nil)
		}
		ns = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 3)
		}

		ns.Labels[constants.AccountNameKey] = obj.Spec.AccountName
		ns.Labels[constants.ClusterNameKey] = obj.Spec.ClusterName
		ns.Labels[constants.ProjectNameKey] = obj.Name
		return nil
	}); err != nil {
		return req.CheckFailed(NamespaceExists, check, err.Error())
	}

	check.Status = true
	if check != obj.Status.Checks[NamespaceExists] {
		obj.Status.Checks[NamespaceExists] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureNamespacedRBACs(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespacedRBACsReady)
	defer req.LogPostCheck(NamespacedRBACsReady)

	var pullSecrets crdsv1.ImagePullSecretList
	if err := r.List(ctx, &pullSecrets, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error())
	}

	secretNames := make([]string, 0, len(pullSecrets.Items))
	for i := range pullSecrets.Items {
		secretNames = append(secretNames, pullSecrets.Items[i].Name)
	}

	b, err := templates.Parse(
		templates.ProjectRBAC, map[string]any{
			"namespace":          obj.Spec.TargetNamespace,
			"svc-account-name":   r.Env.SvcAccountName,
			"image-pull-secrets": secretNames,
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

func (r *Reconciler) ensureWorkspaceRouteSwitcher(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(WorkspaceRouteSwitcherReady)
	defer req.LogPostCheck(WorkspaceRouteSwitcherReady)

	d := &v1.App{ObjectMeta: metav1.ObjectMeta{Name: r.Env.WorkspaceRouteSwitcherName, Namespace: obj.Spec.TargetNamespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, d, func() error {
		d.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		d.Spec = v1.AppSpec{
			DisplayName: "workspace router switcher",
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
					Image:           r.Env.WorkspaceRouteSwitcherImage,
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
		return req.CheckFailed(WorkspaceRouteSwitcherReady, check, err.Error()).Err(nil)
	}

	a, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.TargetNamespace, "workspace-route-switcher"), &v1.App{})
	if err != nil {
		return req.CheckFailed(WorkspaceRouteSwitcherReady, check, err.Error())
	}

	if !a.Status.IsReady {
		return req.CheckFailed(WorkspaceRouteSwitcherReady, check, "waiting for workspace-route-switcher to be ready").Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[WorkspaceRouteSwitcherReady] {
		obj.Status.Checks[WorkspaceRouteSwitcherReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Project{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.Role{})
	builder.Owns(&rbacv1.RoleBinding{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&crdsv1.App{})

	builder.Watches(
		&corev1.Namespace{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			if v, ok := obj.GetLabels()[constants.ProjectNameKey]; ok {
				return []reconcile.Request{{NamespacedName: fn.NN("", v)}}
			}
			return nil
		}),
	)

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
