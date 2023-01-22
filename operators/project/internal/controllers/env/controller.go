package env

import (
	"context"
	"encoding/json"
	"fmt"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	jsonPatch "github.com/kloudlite/operator/pkg/json-patch"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	IsDev      bool
	recorder   record.EventRecorder
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	NamespaceReady       string = "namespace-ready"
	AppsCreated          string = "apps-created"
	CfgNSecretsCreated   string = "config-n-secrets-created"
	MsvcCreated          string = "msvc-created"
	MresCreated          string = "mres-created"
	NamespacedRBACsReady string = "namespaced-rbac-ready"
	RoutersCreated       string = "routers-created"
)

func ensureOwnership(childRes client.Object, ownerRes client.Object) {
	if !fn.IsOwner(childRes, fn.AsOwner(ownerRes)) {
		childRes.SetOwnerReferences(append(childRes.GetOwnerReferences(), fn.AsOwner(ownerRes, true)))
	}
}

func copyMap(into map[string]string, from map[string]string) {
	if into == nil {
		into = make(map[string]string, 1)
	}

	for k, v := range from {
		into[k] = v
	}

	if _, ok := into[constants.ShouldReconcile]; !ok {
		into[constants.ShouldReconcile] = "false"
	}
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.Env{})
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

	if step := req.EnsureChecks(NamespaceReady, AppsCreated, CfgNSecretsCreated, MsvcCreated, MresCreated, NamespacedRBACsReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespaces(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespacedRBACs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCfgAndSecrets(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureApps(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMsvc(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMres(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRouters(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) ensureNamespaces(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespaceReady)
	defer req.LogPostCheck(NamespaceReady)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Name}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if !fn.IsOwner(ns, fn.AsOwner(obj)) {
			ns.SetOwnerReferences(append(ns.GetOwnerReferences(), fn.AsOwner(obj)))
		}
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 1)
		}
		ns.Labels[constants.EnvNameKey] = obj.Name
		return nil
	}); err != nil {
		return req.CheckFailed(NamespaceReady, check, err.Error())
	}

	check.Status = true
	if check != checks[NamespaceReady] {
		checks[NamespaceReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCfgAndSecrets(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CfgNSecretsCreated)
	defer req.LogPostCheck(CfgNSecretsCreated)

	var cfgList crdsv1.ConfigList
	if err := r.List(ctx, &cfgList, &client.ListOptions{
		Namespace: obj.Spec.BlueprintName,
	}); err != nil {
		return req.CheckFailed(CfgNSecretsCreated, check, err.Error()).Err(nil)
	}

	for i := range cfgList.Items {
		cfg := cfgList.Items[i]
		lCfg := &crdsv1.Config{ObjectMeta: metav1.ObjectMeta{Name: cfg.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lCfg, func() error {
			ensureOwnership(lCfg, obj)
			copyMap(lCfg.Labels, cfg.Labels)
			copyMap(lCfg.Annotations, cfg.Annotations)
			fn.MapSet(lCfg.Annotations, constants.EnvironmentRef, obj.Annotations[constants.ResourceRef])
			if lCfg.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(cfg.Data, lCfg.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &lCfg.Data)
			}
			lCfg.Data = cfg.Data
			return nil
		}); err != nil {
			return req.CheckFailed(CfgNSecretsCreated, check, err.Error())
		}
	}

	var scrtList crdsv1.SecretList
	if err := r.List(ctx, &scrtList, &client.ListOptions{
		Namespace: obj.Spec.BlueprintName,
	}); err != nil {
		return req.CheckFailed(CfgNSecretsCreated, check, err.Error()).Err(nil)
	}

	for i := range scrtList.Items {
		scrt := scrtList.Items[i]
		lScrt := &crdsv1.Secret{ObjectMeta: metav1.ObjectMeta{Name: scrt.Name, Namespace: obj.Name}, Type: scrt.Type}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lScrt, func() error {
			ensureOwnership(lScrt, obj)
			copyMap(lScrt.Labels, scrt.Labels)
			copyMap(lScrt.Annotations, scrt.Annotations)
			fn.MapSet(lScrt.Annotations, constants.EnvironmentRef, obj.Annotations[constants.ResourceRef])

			if lScrt.Overrides != nil {
				if scrt.Data != nil {
					b1, err := jsonPatch.ApplyPatch(scrt.Data, scrt.Overrides.Patches)
					if err != nil {
						return err
					}
					if err := json.Unmarshal(b1, &lScrt.Data); err != nil {
						return err
					}
				}

				if scrt.StringData != nil {
					b2, err := jsonPatch.ApplyPatch(scrt.StringData, scrt.Overrides.Patches)
					if err != nil {
						return err
					}
					if err := json.Unmarshal(b2, &lScrt.StringData); err != nil {
						return err
					}
				}
				return nil
			}
			lScrt.Data = scrt.Data
			lScrt.StringData = scrt.StringData
			return nil
		}); err != nil {
			return req.CheckFailed(CfgNSecretsCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[CfgNSecretsCreated] {
		checks[CfgNSecretsCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureNamespacedRBACs(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	namespace := obj.Name
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespacedRBACsReady)
	defer req.LogPreCheck(NamespacedRBACsReady)

	b, err := templates.Parse(
		templates.ProjectRBAC, map[string]any{
			"namespace":          namespace,
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[NamespacedRBACsReady] {
		checks[NamespacedRBACsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureMsvc(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(MsvcCreated)
	defer req.LogPostCheck(MsvcCreated)

	var msvcList crdsv1.ManagedServiceList
	if err := r.List(ctx, &msvcList, &client.ListOptions{
		Namespace: obj.Spec.BlueprintName,
	}); err != nil {
		return req.CheckFailed(MsvcCreated, check, err.Error()).Err(nil)
	}

	for i := range msvcList.Items {
		msvc := msvcList.Items[i]
		lMsvc := &crdsv1.ManagedService{ObjectMeta: metav1.ObjectMeta{Name: msvc.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lMsvc, func() error {
			ensureOwnership(lMsvc, obj)
			copyMap(lMsvc.Labels, msvc.Labels)
			copyMap(lMsvc.Annotations, msvc.Annotations)
			fn.MapSet(lMsvc.Annotations, constants.EnvironmentRef, obj.Annotations[constants.ResourceRef])

			if lMsvc.Enabled == nil {
				lMsvc.Enabled = fn.New(false)
			}

			if lMsvc.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(msvc.Spec, lMsvc.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &lMsvc.Spec)
			}
			lMsvc.Spec = msvc.Spec
			return nil
		}); err != nil {
			return req.CheckFailed(MsvcCreated, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[MsvcCreated] {
		checks[MsvcCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureMres(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(MresCreated)
	defer req.LogPostCheck(MresCreated)

	var mresList crdsv1.ManagedResourceList
	if err := r.List(ctx, &mresList, &client.ListOptions{
		Namespace: obj.Spec.BlueprintName,
	}); err != nil {
		return req.CheckFailed(MresCreated, check, err.Error()).Err(nil)
	}

	for i := range mresList.Items {
		mres := mresList.Items[i]
		lMres := &crdsv1.ManagedResource{ObjectMeta: metav1.ObjectMeta{Name: mres.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lMres, func() error {
			ensureOwnership(lMres, obj)
			copyMap(lMres.Labels, mres.Labels)
			copyMap(lMres.Annotations, mres.Annotations)
			fn.MapSet(lMres.Annotations, constants.EnvironmentRef, obj.Annotations[constants.ResourceRef])

			if lMres.Enabled == nil {
				lMres.Enabled = fn.New(false)
			}

			if lMres.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(mres.Spec, lMres.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &lMres.Spec)
			}
			lMres.Spec = mres.Spec
			return nil
		}); err != nil {
			return req.CheckFailed(MresCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[MresCreated] {
		checks[MresCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureApps(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AppsCreated)
	defer req.LogPostCheck(AppsCreated)

	var appsList crdsv1.AppList
	if err := r.List(ctx, &appsList, &client.ListOptions{
		Namespace: obj.Spec.BlueprintName,
	}); err != nil {
		return req.CheckFailed(AppsCreated, check, err.Error()).Err(nil)
	}

	for i := range appsList.Items {
		app := appsList.Items[i]
		lApp := &crdsv1.App{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lApp, func() error {
			ensureOwnership(lApp, obj)
			copyMap(lApp.Labels, app.Labels)
			copyMap(lApp.Annotations, app.Annotations)
			fn.MapSet(lApp.Annotations, constants.EnvironmentRef, obj.Annotations[constants.ResourceRef])

			if lApp.Enabled == nil {
				lApp.Enabled = fn.New(false)
			}

			if lApp.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(app.Spec, lApp.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &lApp.Spec)
			}
			lApp.Spec = app.Spec
			return nil
		}); err != nil {
			return req.CheckFailed(AppsCreated, check, err.Error()).Err(nil)
		}
		r.recorder.Event(lApp, "Normal", "EnvEnsureApp", "hi")
	}

	check.Status = true
	if check != checks[AppsCreated] {
		checks[AppsCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureRouters(req *rApi.Request[*crdsv1.Env]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	var routers crdsv1.RouterList
	if err := r.List(ctx, &routers, &client.ListOptions{
		Namespace: obj.Spec.BlueprintName,
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
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Env{})
	watchList := []client.Object{
		&crdsv1.App{},
		&corev1.ServiceAccount{},
		&crdsv1.ManagedService{},
		&crdsv1.ManagedResource{},
		&crdsv1.Config{},
		&crdsv1.Secret{},
	}

	for i := range watchList {
		builder.Owns(watchList[i])
	}

	for i := range watchList {
		builder.Watches(&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
				envMap := map[string]bool{}
				if !strings.HasSuffix(obj.GetNamespace(), "-blueprint") {
					return nil
				}
				sp := strings.Split(obj.GetNamespace(), "-blueprint")

				if len(sp) > 0 {
					ns := sp[0]
					var envList crdsv1.EnvList
					if err := r.List(context.TODO(), &envList, &client.ListOptions{
						LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.ProjectNameKey: ns}),
					}); err != nil {
						return nil
					}

					var reqs []reconcile.Request
					for j := range envList.Items {
						envRes := envList.Items[j]
						if !envMap[envRes.Name] {
							reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(envRes.Name, "")})
							envMap[envRes.Name] = true
						}
					}
					return reqs
				}
				return nil
			}))
	}

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
