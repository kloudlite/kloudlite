package env

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
)

type EnvReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
}

func (r *EnvReconciler) GetName() string {
	return r.Name
}

const (
	NamespaceReady       string = "namespace-ready"
	DefaultsPatched      string = "defaults-patched"
	NamespacedRBACsReady string = "namespaced-rbacs-ready"
	RoutersCreated string = "routers-created"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/finalizers,verbs=update

func (r *EnvReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		rApi.NewReconcilerCtx(ctx, r.logger),
		r.Client,
		request.NamespacedName,
		&crdsv1.Env{},
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

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
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

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *EnvReconciler) finalize(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	return req.Finalize()
}

func (r *EnvReconciler) reconDefaults(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false
	if obj.Spec.TargetNamespace == "" {
		obj.Spec.TargetNamespace = obj.Spec.ProjectName + "-" + obj.Name
		hasUpdated = true
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
	}

	check.Status = true
	if check != obj.Status.Checks[DefaultsPatched] {
		obj.Status.Checks[DefaultsPatched] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *EnvReconciler) ensureNamespace(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespaceReady)
	defer req.LogPostCheck(NamespaceReady)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}

	var project crdsv1.Project
	if err := r.Get(ctx, fn.NN("", obj.Spec.ProjectName), &project); err != nil {
		return req.CheckFailed(NamespaceReady, check, err.Error())
	}

	controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if !fn.IsOwner(ns, fn.AsOwner(obj, true)) {
			ns.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		labels := obj.GetLabels()
		if labels == nil {
			labels = make(map[string]string, 2)
		}

		labels[constants.AccountNameKey] = project.Spec.AccountName
		labels[constants.ClusterNameKey] = project.Spec.ClusterName
		ns.SetLabels(labels)

		return nil
	})

	check.Status = true
	if check != obj.Status.Checks[NamespaceReady] {
		obj.Status.Checks[NamespaceReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *EnvReconciler) ensureNamespaceRBACs(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespacedRBACsReady)
	defer req.LogPreCheck(NamespacedRBACsReady)

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

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != checks[NamespacedRBACsReady] {
		checks[NamespacedRBACsReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *EnvReconciler) ensureRoutingFromProject(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	var routers crdsv1.RouterList
	if err := r.List(ctx, &routers, &client.ListOptions{
		Namespace: obj.Namespace,
	}); err != nil {
		return req.CheckFailed(RoutersCreated, check, err.Error()).Err(nil)
	}

	for i := range routers.Items {
		router := routers.Items[i]

		localRouter := &crdsv1.Router{ObjectMeta: metav1.ObjectMeta{Name: router.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, localRouter, func() error {
			ensureOwnership(localRouter, obj)
			copyMap(localRouter.Labels, router.Labels)
			copyMap(localRouter.Annotations, router.Annotations)

			localRouter.Spec = router.Spec
			for j := range router.Spec.Domains {
				localRouter.Spec.Domains[j] = fmt.Sprintf("env.%s.%s", obj.Name, router.Spec.Domains[j])
			}

			//	if localRouter.Overrides != nil {
			//		patchedBytes, err := jsonPatch.ApplyPatch(router.Spec, localRouter.Overrides.Patches)
			//		if err != nil {
			//			return err
			//		}
			//		return json.Unmarshal(patchedBytes, &localRouter.Spec)
			//	}
			//	localRouter.Spec = router.Spec
			return nil
		}); err != nil {
			return req.CheckFailed(RoutersCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[RoutersCreated] {
		checks[RoutersCreated] = check
		req.UpdateStatus()
	}
	return req.Next()

}

func (r *EnvReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Env{})
	return builder.Complete(r)
}
