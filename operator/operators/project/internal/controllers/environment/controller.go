package environment

import (
	"context"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/operators/project/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	reconcinler "github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	Scheme     *runtime.Scheme
	Env        *env.Env
	YAMLClient kubectl.YAMLClient

	templateNamespaceRBAC    []byte
	templateHelmIngressNginx []byte
}

func (r *Reconciler) GetName() string {
	return "environment"
}

const (
	patchDefaults        string = "patch-defaults"
	ensureNamespace      string = "ensure-namespace"
	ensureNamespaceRBACs string = "ensure-namespace-rbac"
	setupEnvIngress      string = "setup-env-ingress"
	updateRouterIngress  string = "update-router-ingress"
	suspendEnvironment   string = "suspend-environment"

	envFinalizing string = "env-finalizing"

	envServiceAccount string = "kloudlite-env-sa"
)

var DestroyChecklist = []reconcinler.CheckMeta{
	{Name: envFinalizing, Title: "Finalizing Environment"},
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconcinler.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.Environment{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	if step := req.EnsureChecks(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

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

	if step := req.EnsureCheckList([]reconcinler.CheckMeta{
		{Name: patchDefaults, Title: "Patching Defaults"},
		{Name: ensureNamespace, Title: "Ensuring Namespace"},
		{Name: ensureNamespaceRBACs, Title: "Ensuring Namespace RBACs"},
		{Name: suspendEnvironment, Title: "Suspending Environment", Hide: !req.Object.Spec.Suspend},
	}); !step.ShouldProceed() {
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

	// if step := r.setupEnvIngressController(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }
	//
	// if step := r.updateRouterIngressClasses(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := r.suspendEnvironment(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconcinler.NewRunningCheck(envFinalizing, req)

	if step := req.EnsureCheckList(DestroyChecklist); !step.ShouldProceed() {
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

	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, mres...); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	// routers
	var routersList crdsv1.RouterList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &routersList, obj.Spec.TargetNamespace); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	routers := make([]client.Object, len(routersList.Items))
	for i := range routersList.Items {
		routers[i] = &routersList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, routers...); err != nil {
		return check.StillRunning(err).NoRequeue()
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

	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, apps...); err != nil {
		return check.StillRunning(err).NoRequeue()
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

	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, configs...); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	// secrets
	var secretsList corev1.SecretList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &secretsList, obj.Spec.TargetNamespace); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	secrets := make([]client.Object, len(secretsList.Items))
	for i := range secretsList.Items {
		secrets[i] = &secretsList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, secrets...); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	var helmList crdsv1.HelmChartList
	if err := findResourceBelongingToEnvironment(ctx, r.Client, &helmList, obj.Spec.TargetNamespace); err != nil {
		return check.Failed(err)
	}

	helmCharts := make([]client.Object, len(helmList.Items))
	for i := range helmList.Items {
		helmCharts[i] = &helmList.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, helmCharts...); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	if step := req.CleanupOwnedResources(check); !step.ShouldProceed() {
		return step
	}

	// deleting namespace
	if err := fn.DeleteAndWait(ctx, req.Logger, r.Client, &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace},
	}); err != nil {
		return check.StillRunning(err).NoRequeue()
	}

	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconcinler.NewRunningCheck(patchDefaults, req)

	hasUpdated := false

	if obj.Spec.Routing == nil {
		hasUpdated = true
		obj.Spec.Routing = &crdsv1.EnvironmentRouting{}
	}

	if obj.Spec.Routing.Mode == "" {
		hasUpdated = true
		obj.Spec.Routing.Mode = crdsv1.EnvironmentRoutingModePrivate
	}

	// if obj.Spec.Routing.PublicIngressClass == "" {
	// 	hasUpdated = true
	// 	obj.Spec.Routing.PublicIngressClass = r.Env.DefaultIngressClass
	// }
	//
	// if obj.Spec.Routing.PrivateIngressClass == "" {
	// 	hasUpdated = true
	// 	// obj.Spec.Routing.PrivateIngressClass = fmt.Sprintf("%s-env-%s", obj.Spec.TargetNamespace, obj.Name)
	// 	obj.Spec.Routing.PrivateIngressClass = fmt.Sprintf("k-%s", fn.Md5([]byte(fmt.Sprintf("%s-env-%s", obj.Spec.TargetNamespace, obj.Name))))
	// }

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return req.Done()
	}

	return check.Completed()
}

func (r *Reconciler) ensureNamespace(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconcinler.NewRunningCheck(ensureNamespace, req)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 4)
		}

		ns.Labels[constants.EnvironmentNameKey] = obj.Name
		ns.Labels[constants.KloudliteGatewayEnabledLabel] = "true"
		ns.Labels[constants.KloudliteNamespaceForEnvironment] = obj.Name

		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string, 1)
		}

		ns.Annotations[constants.DescriptionKey] = fmt.Sprintf("this namespace is now being managed by kloudlite environment (%s)", obj.Name)

		// ns.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		return nil
	}); err != nil {
		return check.StillRunning(err)
	}

	return check.Completed()
}

func (r *Reconciler) ensureNamespaceRBACs(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconcinler.NewRunningCheck(ensureNamespaceRBACs, req)

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
		"svc-account-name":   envServiceAccount,
		"image-pull-secrets": secretNames,
	},
	)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	rr, err := r.YAMLClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

// DEPRECATED: no use as of now
// func (r *Reconciler) setupEnvIngressController(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := reconcinler.NewRunningCheck(setupEnvIngress, req)
//
// 	// releaseName := fmt.Sprintf("%s-env-ingress-%s", obj.Spec.TargetNamespace, obj.Name)
// 	releaseName := obj.Spec.Routing.PrivateIngressClass
// 	releaseNamespace := obj.Spec.TargetNamespace
//
// 	b, err := templates.ParseBytes(r.templateHelmIngressNginx, map[string]any{
// 		"release-name":      releaseName,
// 		"release-namespace": releaseNamespace,
//
// 		"labels": map[string]string{
// 			constants.EnvironmentNameKey: obj.Name,
// 		},
//
// 		"ingress-class-name": obj.Spec.Routing.PrivateIngressClass,
// 	})
// 	if err != nil {
// 		return check.Failed(err).Err(nil)
// 	}
//
// 	rr, err := r.YAMLClient.ApplyYAML(ctx, b)
// 	if err != nil {
// 		return check.StillRunning(err)
// 	}
//
// 	req.AddToOwnedResources(rr...)
//
// 	// wait for helm chart to be ready
// 	hc, err := reconcinler.Get(ctx, r.Client, fn.NN(releaseNamespace, releaseName), &crdsv1.HelmChart{})
// 	if err != nil {
// 		return check.Failed(err)
// 	}
//
// 	if !hc.Status.IsReady {
// 		if hc.Status.Message != nil {
// 			check.Message = hc.Status.Message.ToString()
// 		}
// 		return check.StillRunning(fmt.Errorf("waiting for helm chart to be ready")).NoRequeue()
// 	}
//
// 	return check.Completed()
// }

// DEPRECATED: no use as of now
func (r *Reconciler) updateRouterIngressClasses(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconcinler.NewRunningCheck(updateRouterIngress, req)

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

func (r *Reconciler) suspendEnvironment(req *reconcinler.Request[*crdsv1.Environment]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconcinler.NewRunningCheck(suspendEnvironment, req)

	rquota := &corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-quota", obj.Name), Namespace: obj.Spec.TargetNamespace}}

	if !obj.Spec.Suspend {
		if err := r.Delete(ctx, rquota); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}
		return check.Completed()
	}

	// creating resource quota for the environment namespace
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rquota, func() error {
		rquota.Spec.Hard = corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("0Mi"),
			corev1.ResourceCPU:    resource.MustParse("0m"),
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// evicting all pods from the environment namespace, to make them honour created resource quota
	var podsList corev1.PodList
	if err := r.List(ctx, &podsList, client.InNamespace(obj.Spec.TargetNamespace)); err != nil {
		return check.Failed(err)
	}

	for i := range podsList.Items {
		if err := r.Delete(ctx, &podsList.Items[i]); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YAMLClient == nil {
		return fmt.Errorf("yamlclient must be set")
	}

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
	builder.WithEventFilter(reconcinler.ReconcileFilter())
	return builder.Complete(r)
}
