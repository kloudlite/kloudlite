package msvc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	builder       *builder.Builder
	watchingTypes map[string]struct{}

	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateCommonMsvc []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ManagedServiceApplied string = "managed-service-applied"
	ManagedServiceReady   string = "managed-service-ready"
	OwnManagedResources   string = "own-managed-resources"

	ManagedServiceDeleted string = "managed-service-deleted"
	DefaultsPatched       string = "defaults-patched"
)

var ApplyCheckList = []rApi.CheckMeta{
	{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
	{Name: OwnManagedResources, Title: "Own Managed Resources"},
	{Name: ManagedServiceApplied, Title: "Managed Service Applied"},
	{Name: ManagedServiceReady, Title: "Managed Service Ready"},
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ManagedService{})
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

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
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

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(DefaultsPatched, req)

	hasUpdate := false

	if obj.Output.CredentialsRef.Name == "" {
		hasUpdate = true
		obj.Output.CredentialsRef.Name = fmt.Sprintf("msvc-%s-creds", obj.Name)
	}

	if obj.Spec.Plugin != nil && obj.Spec.Plugin.Export.ViaSecret == "" {
		hasUpdate = true
		obj.Spec.Plugin.Export.ViaSecret = obj.Name + "-export"
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for resource to reconcile")).RequeueAfter(200 * time.Millisecond)
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	req.LogPreCheck("finalizing")
	defer req.LogPostCheck("finalizing")

	check := rApi.NewRunningCheck("finalizing", req)

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: "finalizing", Title: "Cleanup Owned Resources"},
	}); !step.ShouldProceed() {
		return step
	}

	if result := req.CleanupOwnedResourcesV2(check); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) ownManagedResources(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(OwnManagedResources, req)

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
		req.AddToOwnedResources(rApi.ParseResourceRef(mr))
	}

	return check.Completed()
}

func (r *Reconciler) ensureRealMsvcCreated(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ManagedServiceApplied, req)

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

			"output": obj.Output,
		})
		if err != nil {
			return check.Failed(err).NoRequeue()
		}

		if err := r.OwnDynamicResource(obj.Spec.ServiceTemplate.APIVersion, obj.Spec.ServiceTemplate.Kind); err != nil {
			return check.Failed(err).NoRequeue()
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.Failed(err)
		}
		req.AddToOwnedResources(rr...)
	}

	if obj.Spec.Plugin != nil {
		b, err := templates.ParseBytes(r.templateCommonMsvc, map[string]any{
			"api-version": obj.Spec.Plugin.APIVersion,
			"kind":        obj.Spec.Plugin.Kind,

			"name":       obj.Name,
			"namespace":  obj.Namespace,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"labels":      obj.GetLabels(),
			"annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

			"service-template-spec": obj.Spec.Plugin.Spec,

			"export": obj.Spec.Plugin.Export,

			"output": map[string]string{
				"secretName": obj.Output.CredentialsRef.Name,
			},
		})
		if err != nil {
			return check.Failed(err).NoRequeue()
		}

		if err := r.OwnDynamicResource(obj.Spec.Plugin.APIVersion, obj.Spec.Plugin.Kind); err != nil {
			return check.Failed(err).NoRequeue()
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.Failed(err)
		}
		req.AddToOwnedResources(rr...)
	}

	return check.Completed()
}

func (r *Reconciler) ensureRealMsvcReady(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ManagedServiceReady, req)

	var uobj *unstructured.Unstructured
	if obj.Spec.ServiceTemplate != nil {
		uobj = fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.ServiceTemplate.APIVersion, Kind: obj.Spec.ServiceTemplate.Kind})
	}

	if obj.Spec.Plugin != nil {
		uobj = fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.Plugin.APIVersion, Kind: obj.Spec.Plugin.Kind})
	}

	realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), uobj)
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	b, err := json.Marshal(realMsvc)
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	var realMsvcObj struct {
		Status rApi.Status `json:"status"`
	}
	if err := json.Unmarshal(b, &realMsvcObj); err != nil {
		return check.Failed(err).NoRequeue()
	}

	if !realMsvcObj.Status.IsReady {
		if realMsvcObj.Status.Message == nil {
			return check.Failed(fmt.Errorf("waiting for real managed service to reconcile")).NoRequeue()
		}
		b, err := realMsvcObj.Status.Message.MarshalJSON()
		if err != nil {
			return check.Failed(err).NoRequeue()
		}
		return check.Failed(fmt.Errorf("%s", b)).NoRequeue()
	}

	return check.Completed()
}

func (r *Reconciler) OwnDynamicResource(apiVersion, kind string) error {
	if _, ok := r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)]; ok {
		return nil
	}

	r.watchingTypes[fmt.Sprintf("%s.%s", apiVersion, kind)] = struct{}{}

	if !fn.IsGVKInstalled(r.Client, apiVersion, kind) {
		r.logger.Warnf("plugin CRD not installed, APIVersion: %s, Kind=%s", apiVersion, kind)
		return nil
	}

	// Dynamically add the watch
	r.builder.Owns(fn.NewUnstructured(metav1.TypeMeta{APIVersion: apiVersion, Kind: kind}))
	if err := r.builder.Complete(r); err != nil {
		r.logger.Errorf(err, "failed to call Complete() on builder")
		return err
	}
	r.logger.Infof("ADDED watch for owned-resources with GVK %s/%s", apiVersion, kind)
	return nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})
	r.watchingTypes = make(map[string]struct{})

	var err error
	r.templateCommonMsvc, err = templates.Read(templates.CommonMsvcTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedService{})
	builder.Owns(&crdsv1.ManagedResource{})

	r.builder = builder

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
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
