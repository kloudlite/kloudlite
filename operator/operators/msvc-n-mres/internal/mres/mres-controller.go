package mres

import (
	"context"
	_ "embed"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/mres/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *Env
	YAMLClient kubectl.YAMLClient
	// templateCommonMres []byte

	dynamicOwnership func(apiVersion, kind string) error

	watchingTypes map[string]struct{}
}

func (r *Reconciler) GetName() string {
	return "managed-resource"
}

//go:embed templates/common-mres.yml.tpl
var templateCommonManagedResource []byte

const (
	Cleanup                          string = "cleanup"
	UnderlyingManagedResourceCreated string = "underlying-managed-resource-created"
	UnderlyingManagedResourceReady   string = "underlying-managed-resource-ready"
	DefaultsPatched                  string = "defaults-patched"
)

var DeleteCheckList = []reconciler.CheckMeta{}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.ManagedResource{})
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

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
		{Name: UnderlyingManagedResourceCreated, Title: "Underlying Managed Resource Created"},
		{Name: UnderlyingManagedResourceReady, Title: "Underlying Managed Resource Ready"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMresCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMresReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

// func getRealResourceName(obj *crdsv1.ManagedResource) string {
// 	if obj.Spec.ResourceNamePrefix != nil {
// 		return fmt.Sprintf("%s-%s", *obj.Spec.ResourceNamePrefix, obj.Name)
// 	}
//
// 	return obj.Name
// }

func (r *Reconciler) patchDefaults(req *reconciler.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(DefaultsPatched, req)

	hasUpdate := false

	// if obj.Output.CredentialsRef.Name == "" {
	// 	hasUpdate = true
	// 	obj.Output.CredentialsRef.Name = fmt.Sprintf("mres-%s-creds", getRealResourceName(obj))
	// }

	ms, err := reconciler.Get(ctx, r.Client, fn.NN(obj.Spec.ManagedServiceRef.Namespace, obj.Spec.ManagedServiceRef.Name), &crdsv1.ManagedService{})
	if err != nil {
		return check.Failed(err)
	}

	if !fn.IsOwner(obj, fn.AsOwner(ms, true)) {
		hasUpdate = true
		obj.OwnerReferences = append(obj.OwnerReferences, fn.AsOwner(ms, true))
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for resource reconcilation"))
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *reconciler.Request[*crdsv1.ManagedResource]) stepResult.Result {
	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: "cleanup", Title: "Cleanup Owned Resources"},
	}); !step.ShouldProceed() {
		return step
	}

	check := reconciler.NewRunningCheck("cleanup", req)

	if result := req.CleanupOwnedResources(check); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) ensureRealMresCreated(req *reconciler.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(UnderlyingManagedResourceCreated, req)

	if !fn.IsGVKInstalled(r.Client, obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind) {
		return check.Failed(fmt.Errorf("CRD not installed for (apiVersion: %s, kind: %s)", obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind))
	}

	b, err := templates.ParseBytes(templateCommonManagedResource, templates.CommonManagedResourceParams{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			Labels:          obj.GetLabels(),
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		PluginTemplate:    obj.Spec.Plugin,
		ManagedServiceRef: obj.Spec.ManagedServiceRef,
	})
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	if err := r.OwnDynamicResource(req, obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind); err != nil {
		return check.Failed(err)
	}

	rr, err := r.YAMLClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)
	return check.Completed()
}

func (r *Reconciler) OwnDynamicResource(req *reconciler.Request[*crdsv1.ManagedResource], apiVersion, kind string) error {
	if _, ok := r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)]; ok {
		return nil
	}

	r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)] = struct{}{}

	if !fn.IsGVKInstalled(r.Client, apiVersion, kind) {
		req.Logger.Warn("plugin CRD not installed", "APIVersion", apiVersion, "kind", kind)
		return nil
	}

	if err := r.dynamicOwnership(apiVersion, kind); err != nil {
		req.Logger.Error("failed to call Complete() on builder, got", "err", err)
		return err
	}

	req.Logger.Info("ADDED watch for owned-resources with GVK %s/%s", apiVersion, kind)
	return nil
}

func (r *Reconciler) ensureRealMresReady(req *reconciler.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(UnderlyingManagedResourceReady, req)

	uobj := fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.Plugin.APIVersion, Kind: obj.Spec.Plugin.Kind})
	if err := r.Get(ctx, fn.NN(obj.GetNamespace(), obj.Name), uobj); err != nil {
		return check.Failed(err)
	}

	realMresObj, err := fn.JsonConvert[struct {
		Status reconciler.Status `json:"status"`
	}](uobj.Object)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	if !realMresObj.Status.IsReady {
		errorMsg := ""
		for _, v := range realMresObj.Status.CheckList {
			if realMresObj.Status.Checks[v.Name].State == reconciler.ErroredState && realMresObj.Status.Checks[v.Name].Message != "" {
				errorMsg = realMresObj.Status.Checks[v.Name].Message
				break
			}
		}

		if errorMsg == "" {
			return check.Failed(fmt.Errorf("waiting for real managed service to reconcile")).NoRequeue()
		}

		return check.Failed(fmt.Errorf(errorMsg)).NoRequeue()
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.Env == nil {
		return fmt.Errorf("env must be set")
	}

	if r.YAMLClient == nil {
		return fmt.Errorf("yaml client is required")
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedResource{}).Named(r.GetName())
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&corev1.Secret{})

	r.watchingTypes = make(map[string]struct{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))

	tc, err := builder.Build(r)
	if err != nil {
		return err
	}

	r.dynamicOwnership = func(apiVersion, kind string) error {
		obj := fn.NewUnstructured(metav1.TypeMeta{APIVersion: apiVersion, Kind: kind})
		return tc.Watch(source.Kind(mgr.GetCache(), obj, handler.TypedEnqueueRequestForOwner[*unstructured.Unstructured](mgr.GetScheme(), mgr.GetRESTMapper(), &crdsv1.ManagedResource{}, handler.OnlyControllerOwner())))
	}

	return nil
}
