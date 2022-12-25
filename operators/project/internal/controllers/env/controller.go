package env

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operators/project/internal/env"
	"operators.kloudlite.io/pkg/constants"
	fn "operators.kloudlite.io/pkg/functions"
	jsonPatch "operators.kloudlite.io/pkg/json-patch"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	"operators.kloudlite.io/pkg/templates"
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
	RBACReady            string = "rbac-ready"
	HarborAccessReady    string = "harbor-access-ready"
	HarborCredsAvailable string = "harbor-creds-available"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=envs/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.Env{})
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

	//if step := r.copyHarborCreds(req); !step.ShouldProceed() {
	//	return step.ReconcilerResponse()
	//}

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
		nCfg := &crdsv1.Config{ObjectMeta: metav1.ObjectMeta{Name: cfg.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, nCfg, func() error {
			if !fn.IsOwner(nCfg, fn.AsOwner(obj)) {
				nCfg.SetOwnerReferences(append(nCfg.GetOwnerReferences(), fn.AsOwner(obj, true)))
			}
			if nCfg.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(cfg.Data, nCfg.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &nCfg.Data)
			}
			nCfg.Data = cfg.Data
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
		nScrt := &crdsv1.Secret{ObjectMeta: metav1.ObjectMeta{Name: scrt.Name, Namespace: obj.Name}, Type: scrt.Type}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, nScrt, func() error {
			if !fn.IsOwner(nScrt, fn.AsOwner(obj)) {
				nScrt.SetOwnerReferences(append(nScrt.GetOwnerReferences(), fn.AsOwner(obj, true)))
			}

			if nScrt.Overrides != nil {
				b1, err := jsonPatch.ApplyPatch(scrt.Data, scrt.Overrides.Patches)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(b1, &nScrt.Data); err != nil {
					return err
				}

				b2, err := jsonPatch.ApplyPatch(scrt.StringData, scrt.Overrides.Patches)
				if err != nil {
					return err
				}

				return json.Unmarshal(b2, &nScrt.StringData)
			}

			nScrt.Data = scrt.Data
			nScrt.StringData = scrt.StringData
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
		return req.CheckFailed(RBACReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(RBACReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[RBACReady] {
		checks[RBACReady] = check
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
		nMsvc := &crdsv1.ManagedService{ObjectMeta: metav1.ObjectMeta{Name: msvc.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, nMsvc, func() error {
			if !fn.IsOwner(nMsvc, fn.AsOwner(obj)) {
				nMsvc.SetOwnerReferences(append(nMsvc.GetOwnerReferences(), fn.AsOwner(obj, true)))
			}
			nMsvc.Labels = msvc.Labels
			if nMsvc.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(msvc.Spec, nMsvc.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &nMsvc.Spec)
			}
			nMsvc.Spec = msvc.Spec
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
		nMres := &crdsv1.ManagedResource{ObjectMeta: metav1.ObjectMeta{Name: mres.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, nMres, func() error {
			if nMres.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(mres.Spec, nMres.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &nMres.Spec)
			}
			nMres.Spec = mres.Spec
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
		nApp := &crdsv1.App{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: obj.Name}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, nApp, func() error {
			if !fn.IsOwner(nApp, fn.AsOwner(obj)) {
				nApp.SetOwnerReferences(append(nApp.GetOwnerReferences(), fn.AsOwner(obj, true)))
			}
			if nApp.Overrides != nil {
				patchedBytes, err := jsonPatch.ApplyPatch(app.Spec, nApp.Overrides.Patches)
				if err != nil {
					return err
				}
				return json.Unmarshal(patchedBytes, &nApp.Spec)
			}
			nApp.Spec = app.Spec
			return nil
		}); err != nil {
			return req.CheckFailed(AppsCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[AppsCreated] {
		checks[AppsCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

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

	envMap := map[string]bool{}

	for i := range watchList {
		builder.Watches(&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
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
