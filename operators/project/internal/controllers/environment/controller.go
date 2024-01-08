package environment

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/operators/project/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
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
	Scheme                   *runtime.Scheme
	Env                      *env.Env
	logger                   logging.Logger
	Name                     string
	yamlClient               kubectl.YAMLClient
	templateNamespaceRBAC    []byte
	templateHelmIngressNginx []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.Environment{})
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespaceRBACs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureEnvIngressController(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "patch-defaults"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hasUpdated := false

	if obj.Spec.IngressClassName == "" {
		hasUpdated = true
		obj.Spec.IngressClassName = fmt.Sprintf("env-%s", obj.Name)
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}
		return req.Done()
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureNamespace(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}
	checkName := "namespace"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	var project crdsv1.Project
	if err := r.Get(ctx, fn.NN("", obj.Spec.ProjectName), &project); err != nil {
		return fail(err)
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 4)
		}

		ns.Labels[constants.AccountNameKey] = project.Spec.AccountName
		ns.Labels[constants.ClusterNameKey] = project.Spec.ClusterName
		ns.Labels[constants.EnvironmentNameKey] = obj.Name
		ns.Labels[constants.ProjectNameKey] = obj.Spec.ProjectName

		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string, 1)
		}

		ns.Annotations[constants.DescriptionKey] = fmt.Sprintf("this namespace is now being managed by kloudlite environment (%s)", obj.Name)

		return nil
	}); err != nil {
		return fail(err)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureNamespaceRBACs(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "namespace-rbac"

	req.LogPreCheck(checkName)
	defer req.LogPreCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	var pullSecrets corev1.SecretList
	if err := r.List(ctx, &pullSecrets, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return fail(err)
	}

	secretNames := make([]string, 0, len(pullSecrets.Items))
	for i := range pullSecrets.Items {
		if pullSecrets.Items[i].Type == corev1.SecretTypeDockerConfigJson {
			secretNames = append(secretNames, pullSecrets.Items[i].Name)
		}
	}

	b, err := templates.ParseBytes(r.templateNamespaceRBAC, map[string]any{
		"namespace":          obj.Spec.TargetNamespace,
		"svc-account-name":   r.Env.SvcAccountName,
		"image-pull-secrets": secretNames,
	},
	)
	if err != nil {
		return fail(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return fail(err).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureEnvIngressController(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "env-ingress-controller"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	b, err := templates.ParseBytes(r.templateHelmIngressNginx, map[string]any{
		"release-name":      fmt.Sprintf("env-ingress-%s", obj.Name),
		"release-namespace": obj.Spec.TargetNamespace,

		"labels": map[string]string{
			constants.ProjectNameKey:     obj.Spec.ProjectName,
			constants.EnvironmentNameKey: obj.Name,
		},

		"ingress-class-name": obj.Spec.IngressClassName,
	})
	if err != nil {
		return fail(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

// func (r *Reconciler) ensureRoutingFromProject(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
// 	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
// 	check := rApi.Check{Generation: obj.Generation}
//
// 	var routers crdsv1.RouterList
// 	if err := r.List(ctx, &routers, &client.ListOptions{
// 		Namespace: obj.Namespace,
// 	}); err != nil {
// 		return req.CheckFailed(RoutersCreated, check, err.Error()).Err(nil)
// 	}
//
// 	for i := range routers.Items {
// 		router := routers.Items[i]
//
// 		localRouter := crdsv1.Router{ObjectMeta: metav1.ObjectMeta{Name: router.Name, Namespace: obj.Spec.TargetNamespace}}
//
// 		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &localRouter, func() error {
// 			// localRouter.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
//
// 			if localRouter.Labels == nil {
// 				localRouter.Labels = make(map[string]string, len(router.Labels))
// 			}
// 			for k, v := range router.Labels {
// 				localRouter.Labels[k] = v
// 			}
//
// 			if localRouter.Annotations == nil {
// 				localRouter.Annotations = make(map[string]string, len(router.Annotations))
// 			}
//
// 			for k, v := range router.Annotations {
// 				localRouter.Annotations[k] = v
// 			}
//
// 			localRouter.Spec = router.Spec
// 			localRouter.Spec.Https.Enabled = false
// 			for j := range router.Spec.Domains {
// 				localRouter.Spec.Domains[j] = fmt.Sprintf("env.%s.%s", obj.Name, router.Spec.Domains[j])
// 			}
// 			return nil
// 		}); err != nil {
// 			return req.CheckFailed(RoutersCreated, check, err.Error()).Err(nil)
// 		}
//
// 		req.AddToOwnedResources(rApi.ResourceRef{
// 			TypeMeta:  router.TypeMeta,
// 			Namespace: localRouter.Namespace,
// 			Name:      localRouter.Name,
// 		})
// 	}
//
// 	check.Status = true
// 	if check != checks[RoutersCreated] {
// 		checks[RoutersCreated] = check
// 		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
// 			return sr
// 		}
// 	}
// 	return req.Next()
// }

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateNamespaceRBAC, err = templates.Read(templates.NamespaceRBAC)
	if err != nil {
		return err
	}

	r.templateHelmIngressNginx, err = templates.Read(templates.HelmIngressNginx)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Environment{})
	builder.Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, obj client.Object) []reconcile.Request {
		if v, ok := obj.GetLabels()[constants.EnvironmentNameKey]; ok {
			return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
		}
		return nil
	}))

	builder.Watches(&crdsv1.HelmChart{}, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, obj client.Object) []reconcile.Request {
		if v, ok := obj.GetLabels()[constants.ProjectNameKey]; ok {
			var envList crdsv1.EnvironmentList
			if err := r.List(context.TODO(), &envList, &client.ListOptions{
				LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
					constants.EnvironmentNameKey: v,
				}),
			}); err != nil {
				return nil
			}

			reqs := make([]reconcile.Request, len(envList.Items))
			for i := range envList.Items {
				reqs[i] = reconcile.Request{NamespacedName: fn.NN(envList.Items[i].GetNamespace(), envList.Items[i].GetName())}
			}

			return reqs
		}
		return nil
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
