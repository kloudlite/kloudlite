package msvc

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	templates2 "github.com/kloudlite/operator/operators/msvc-n-mres/internal/msvc/templates"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
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
	Env        *env.Env
	YAMLClient kubectl.YAMLClient

	dynamicWatch func(apiVersion, kind string) error

	watchingTypes map[string]struct{}

	templateCommonMsvc []byte
}

//go:embed templates/common-msvc.yml.tpl
var templateCommonManagedService []byte

func (r *Reconciler) GetName() string {
	return "managed-services"
}

const (
	ManagedServiceApplied string = "managed-service-applied"
	ManagedServiceReady   string = "managed-service-ready"
	OwnManagedResources   string = "own-managed-resources"

	ManagedServiceDeleted string = "managed-service-deleted"
	DefaultsPatched       string = "defaults-patched"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.ManagedService{})
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
		{Name: OwnManagedResources, Title: "Own Managed Resources"},
		{Name: ManagedServiceApplied, Title: "Managed Service Applied"},
		{Name: ManagedServiceReady, Title: "Managed Service Ready"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMsvcCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ownManagedResources(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMsvcReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) patchDefaults(req *reconciler.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(DefaultsPatched, req)

	if hasPatched := obj.PatchWithDefaults(); hasPatched {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("waiting for reconcilation"))
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *reconciler.Request[*crdsv1.ManagedService]) stepResult.Result {
	req.LogPreCheck("finalizing")
	defer req.LogPostCheck("finalizing")

	check := reconciler.NewRunningCheck("finalizing", req)

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: "finalizing", Title: "Cleanup Owned Resources"},
	}); !step.ShouldProceed() {
		return step
	}

	if result := req.CleanupOwnedResources(check); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) ownManagedResources(req *reconciler.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(OwnManagedResources, req)

	result, err := kubectl.PaginatedList[*crdsv1.ManagedResource](ctx, r.Client, &crdsv1.ManagedResourceList{}, &client.ListOptions{
		Namespace: obj.Namespace,
		Limit:     10,
	})
	if err != nil {
		return check.Failed(err)
	}

	for mr := range result {
		if mr.GetDeletionTimestamp() != nil {
			continue
		}
		if !fn.IsOwner(mr, fn.AsOwner(obj, true)) {
			mr.SetOwnerReferences(append(mr.GetOwnerReferences(), fn.AsOwner(obj, true)))
			if err := r.Update(ctx, mr); err != nil {
				return check.Failed(err)
			}
		}
		req.AddToOwnedResources(reconciler.ParseResourceRef(mr))
	}

	return check.Completed()
}

func (r *Reconciler) ensureRealMsvcCreated(req *reconciler.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(ManagedServiceApplied, req)

	if obj.Spec.ServiceTemplate != nil {
		if !fn.IsGVKInstalled(r.Client, obj.Spec.ServiceTemplate.APIVersion, obj.Spec.ServiceTemplate.Kind) {
			return check.Failed(fmt.Errorf("CRD not installed for (apiVersion: %s, kind: %s)", obj.Spec.ServiceTemplate.APIVersion, obj.Spec.ServiceTemplate.Kind))
		}

		b, err := templates.ParseBytes(r.templateCommonMsvc, map[string]any{
			"api-version": obj.Spec.ServiceTemplate.APIVersion,
			"kind":        obj.Spec.ServiceTemplate.Kind,

			"name":       obj.Name,
			"namespace":  obj.Namespace,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"labels":      obj.GetLabels(),
			"annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

			// "node-selector":         obj.Spec.NodeSelector,
			// "tolerations":           obj.Spec.Tolerations,
			"service-template-spec": obj.Spec.ServiceTemplate.Spec,
		})
		if err != nil {
			return check.Failed(err).NoRequeue()
		}

		if err := r.OwnDynamicResource(req, obj.Spec.ServiceTemplate.APIVersion, obj.Spec.ServiceTemplate.Kind); err != nil {
			return check.Failed(err).NoRequeue()
		}

		rr, err := r.YAMLClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.Failed(err)
		}
		req.AddToOwnedResources(rr...)
	}

	if obj.Spec.Plugin != nil {
		b, err := templates.ParseBytes(templateCommonManagedService, templates2.CommonManagedServiceParams{
			Metadata: metav1.ObjectMeta{
				Name:            obj.Name,
				Namespace:       obj.Namespace,
				Labels:          obj.GetLabels(),
				Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
			PluginTemplate: *obj.Spec.Plugin,
		})
		// b, err := templates.ParseBytes(r.templateCommonMsvc, map[string]any{
		// 	"api-version": obj.Spec.Plugin.APIVersion,
		// 	"kind":        obj.Spec.Plugin.Kind,
		//
		// 	"name":       obj.Name,
		// 	"namespace":  obj.Namespace,
		// 	"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
		//
		// 	"labels":      obj.GetLabels(),
		// 	"annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
		//
		// 	"service-template-spec": obj.Spec.Plugin.Spec,
		//
		// 	"export": obj.Spec.Plugin.Export,
		//
		// 	// "output": map[string]string{
		// 	// 	"secretName": obj.Output.CredentialsRef.Name,
		// 	// },
		// })
		if err != nil {
			return check.Failed(err).NoRequeue()
		}

		os.WriteFile("/tmp/sample.yml", b, 0o666)

		if err := r.OwnDynamicResource(req, obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind); err != nil {
			return check.Failed(err).NoRequeue()
		}

		rr, err := r.YAMLClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.Failed(err)
		}
		req.AddToOwnedResources(rr...)
	}

	return check.Completed()
}

func (r *Reconciler) ensureRealMsvcReady(req *reconciler.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(ManagedServiceReady, req)

	var uobj *unstructured.Unstructured
	if obj.Spec.ServiceTemplate != nil {
		uobj = fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.ServiceTemplate.APIVersion, Kind: obj.Spec.ServiceTemplate.Kind})
	}

	if obj.Spec.Plugin != nil {
		uobj = fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.Plugin.APIVersion, Kind: obj.Spec.Plugin.Kind})
	}

	realMsvc, err := reconciler.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), uobj)
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	b, err := json.Marshal(realMsvc)
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	var realMsvcObj struct {
		Status reconciler.Status `json:"status"`
	}
	if err := json.Unmarshal(b, &realMsvcObj); err != nil {
		return check.Failed(err).NoRequeue()
	}

	if !realMsvcObj.Status.IsReady {
		errorMsg := ""
		for _, v := range realMsvcObj.Status.CheckList {
			if realMsvcObj.Status.Checks[v.Name].State == reconciler.ErroredState && realMsvcObj.Status.Checks[v.Name].Message != "" {
				errorMsg = realMsvcObj.Status.Checks[v.Name].Message
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

func (r *Reconciler) OwnDynamicResource(req *reconciler.Request[*crdsv1.ManagedService], apiVersion, kind string) error {
	if _, ok := r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)]; ok {
		return nil
	}

	r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)] = struct{}{}

	if !fn.IsGVKInstalled(r.Client, apiVersion, kind) {
		req.Logger.Warn("plugin CRD not installed", "APIVersion", apiVersion, "kind", kind)
		return nil
	}

	if err := r.dynamicWatch(apiVersion, kind); err != nil {
		req.Logger.Error("failed to call Complete() on builder, got", "err", err)
		return err
	}

	// r.controller.Watch(source.TypedKind(cache cache.Cache, obj object, handler handler.TypedEventHandler[object, request], predicates ...predicate.TypedPredicate[object]))

	// r.builder.Watches(source.Kind(cache, &Type{}, handler.EnqueueRequestForOwner(&source.Kind{Type: obj},, &OwnerType{}, OnlyControllerOwner()))).

	// Dynamically add the watch
	// r.builder.Owns(fn.NewUnstructured(metav1.TypeMeta{APIVersion: apiVersion, Kind: kind}))
	// if err := r.builder.Complete(r); err != nil {
	// 	req.Logger.Error("failed to call Complete() on builder, got", "err", err)
	// 	return err
	// }
	req.Logger.Info(fmt.Sprintf("ADDED watch for owned-resources GVK %s/%s", apiVersion, kind))
	return nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YAMLClient == nil {
		return fmt.Errorf("yaml client must be set")
	}

	r.watchingTypes = make(map[string]struct{})

	var err error
	r.templateCommonMsvc, err = templates.Read(templates.CommonMsvcTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedService{}).Named(r.GetName())
	builder.Owns(&crdsv1.ManagedResource{})

	// watchlist := []client.Object{
	// 	&crdsv1.ManagedResource{},
	// }
	//
	// for _, obj := range watchlist {
	// 	builder.Watches(obj, handler.EnqueueRequestsFromMapFunc(
	// 		func(_ context.Context, obj client.Object) []reconcile.Request {
	// 			if v, ok := obj.GetLabels()[constants.MsvcNameKey]; ok {
	// 				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
	// 			}
	// 			return nil
	// 		}))
	// 	builder.Owns(obj)
	// }

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))

	// r.dynamicWatch = func(apiVersion, kind string) error {
	// 	return nil
	// }

	tc, err := builder.Build(r)
	if err != nil {
		return err
	}

	r.dynamicWatch = func(apiVersion, kind string) error {
		obj := fn.NewUnstructured(metav1.TypeMeta{APIVersion: apiVersion, Kind: kind})
		return tc.Watch(source.Kind(mgr.GetCache(), obj, handler.TypedEnqueueRequestForOwner[*unstructured.Unstructured](mgr.GetScheme(), mgr.GetRESTMapper(), &crdsv1.ManagedService{}, handler.OnlyControllerOwner())))
	}

	// return builder.Complete(r)
	return nil
}
