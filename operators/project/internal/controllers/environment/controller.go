package environment

import (
	"context"
	"fmt"
	"slices"
	"time"

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
	networkingv1 "k8s.io/api/networking/v1"
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

const (
	patchDefaults        string = "patch-defaults"
	ensureNamespace      string = "ensure-namespace"
	ensureNamespaceRBACs string = "ensure-namespace-rbac"
	setupEnvIngress      string = "setup-env-ingress"
	updateRouterIngress  string = "update-router-ingress"

	envFinalizing string = "env-finalizing"
)

var (
	ENV_CHECKLIST = []rApi.CheckMeta{
		{Name: patchDefaults, Title: "Patching Defaults"},
		{Name: ensureNamespace, Title: "Ensuring Namespace"},
		{Name: ensureNamespaceRBACs, Title: "Ensuring Namespace RBACs"},
		{Name: setupEnvIngress, Title: "Setting up Environment Ingress"},
		{Name: updateRouterIngress, Title: "Updating Router Ingress"},
	}

	ENV_DESTROY_CHECKLIST = []rApi.CheckMeta{
		{Name: envFinalizing, Title: "Finalizing Environment"},
	}
)

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

	if step := req.EnsureCheckList(ENV_CHECKLIST); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(); !step.ShouldProceed() {
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

	if step := r.setupEnvIngressController(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.updateRouterIngressClasses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(envFinalizing, req)

	if !slices.Equal(obj.Status.CheckList, ENV_DESTROY_CHECKLIST) {
		req.Object.Status.CheckList = ENV_DESTROY_CHECKLIST
		if step := req.UpdateStatus(); !step.ShouldProceed() {
			return step
		}
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	// ensure deletion of other kloudlite resources, that belong to this environment
	var mresList crdsv1.ManagedResourceList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &mresList, obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err)
	}

	mres := make([]client.Object, len(mresList.Items))
	for i := range mresList.Items {
		mres[i] = &mresList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, mres...); err != nil {
		return check.Failed(err)
	}

	// routers
	var routersList crdsv1.RouterList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &routersList, obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err)
	}

	routers := make([]client.Object, len(routersList.Items))
	for i := range routersList.Items {
		routers[i] = &routersList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, routers...); err != nil {
		return check.Failed(err)
	}

	// apps
	var appsList crdsv1.AppList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &appsList, obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err)
	}

	apps := make([]client.Object, len(appsList.Items))
	for i := range appsList.Items {
		apps[i] = &appsList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, apps...); err != nil {
		return check.Failed(err)
	}

	// configs
	var configsList corev1.ConfigMapList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &configsList, obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err)
	}

	configs := make([]client.Object, 0, len(configsList.Items))
	for i := range configsList.Items {
		if configsList.Items[i].Name == "kube-root-ca.crt" {
			continue
		}
		configs = append(configs, &configsList.Items[i])
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, configs...); err != nil {
		return check.Failed(err)
	}

	// secrets
	var secretsList corev1.SecretList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &secretsList, obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err)
	}

	secrets := make([]client.Object, len(secretsList.Items))
	for i := range secretsList.Items {
		secrets[i] = &secretsList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, secrets...); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdated := false

	if obj.Spec.Routing == nil {
		hasUpdated = true
		obj.Spec.Routing = &crdsv1.EnvironmentRouting{}
	}

	if obj.Spec.Routing.Mode == "" {
		hasUpdated = true
		obj.Spec.Routing.Mode = crdsv1.EnvironmentRoutingModePrivate
	}

	if obj.Spec.Routing.PublicIngressClass == "" {
		hasUpdated = true
		obj.Spec.Routing.PublicIngressClass = r.Env.DefaultIngressClass
	}

	if obj.Spec.Routing.PrivateIngressClass == "" {
		hasUpdated = true
		// obj.Spec.Routing.PrivateIngressClass = fmt.Sprintf("%s-env-%s", obj.Spec.TargetNamespace, obj.Name)
		obj.Spec.Routing.PrivateIngressClass = fmt.Sprintf("k-%s", fn.Md5([]byte(fmt.Sprintf("%s-env-%s", obj.Spec.TargetNamespace, obj.Name))))
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return req.Done()
	}

	return check.Completed()
}

func (r *Reconciler) ensureNamespace(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ensureNamespace, req)

	var project crdsv1.Project
	if err := r.Get(ctx, fn.NN("", obj.Spec.ProjectName), &project); err != nil {
		return check.Failed(err)
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 4)
		}

		ns.Labels[constants.EnvironmentNameKey] = obj.Name
		ns.Labels[constants.ProjectNameKey] = obj.Spec.ProjectName

		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string, 1)
		}

		ns.Annotations[constants.DescriptionKey] = fmt.Sprintf("this namespace is now being managed by kloudlite environment (%s)", obj.Name)

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) ensureNamespaceRBACs(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ensureNamespaceRBACs, req)

	var pullSecrets corev1.SecretList
	if err := r.List(ctx, &pullSecrets, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return check.Failed(err)
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
		return check.Failed(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) setupEnvIngressController(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(setupEnvIngress, req)

	// releaseName := fmt.Sprintf("%s-env-ingress-%s", obj.Spec.TargetNamespace, obj.Name)
	releaseName := obj.Spec.Routing.PrivateIngressClass
	releaseNamespace := obj.Spec.TargetNamespace

	b, err := templates.ParseBytes(r.templateHelmIngressNginx, map[string]any{
		"release-name":      releaseName,
		"release-namespace": releaseNamespace,

		"labels": map[string]string{
			constants.ProjectNameKey:     obj.Spec.ProjectName,
			constants.EnvironmentNameKey: obj.Name,
		},

		"ingress-class-name": obj.Spec.Routing.PrivateIngressClass,
	})
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	// wait for helm chart to be ready
	hc, err := rApi.Get(ctx, r.Client, fn.NN(releaseNamespace, releaseName), &crdsv1.HelmChart{})
	if err != nil {
		return check.Failed(err)
	}

	if !hc.Status.IsReady {
		if hc.Status.Message != nil {
			check.Message = hc.Status.Message.ToString()
		}
		return check.StillRunning(fmt.Errorf("waiting for helm chart to be ready"))
	}

	return check.Completed()
}

func (r *Reconciler) updateRouterIngressClasses(req *rApi.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(updateRouterIngress, req)

	var routers crdsv1.RouterList
	if err := r.List(ctx, &routers, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return check.Failed(err)
	}

	for i := range routers.Items {
		routers.Items[i].Spec.IngressClass = obj.GetIngressClassName()
		if err := r.Update(ctx, &routers.Items[i]); err != nil {
			return check.Failed(err)
		}
	}

	var ingressList networkingv1.IngressList
	if err := r.List(ctx, &ingressList, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return check.Failed(err)
	}

	for i := range ingressList.Items {
		ingressList.Items[i].Spec.IngressClassName = fn.New(obj.GetIngressClassName())
		if err := r.Update(ctx, &ingressList.Items[i]); err != nil {
			return stepResult.New().RequeueAfter(1 * time.Second)
		}
	}

	return check.Completed()
}

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

	watchList := []client.Object{
		&crdsv1.HelmChart{},
		&crdsv1.Router{},
		&networkingv1.Ingress{},
		&corev1.Secret{},
	}

	for i := range watchList {
		builder.Watches(watchList[i],
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					var envList crdsv1.EnvironmentList
					if err := r.List(ctx, &envList, &client.ListOptions{
						LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
							constants.TargetNamespaceKey: obj.GetNamespace(),
						}),
					}); err != nil {
						return nil
					}

					rr := make([]reconcile.Request, 0, len(envList.Items))
					for i := range envList.Items {
						rr = append(rr, reconcile.Request{NamespacedName: fn.NN(envList.Items[i].GetNamespace(), envList.Items[i].GetName())})
					}

					return rr
				}),
		)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
