package mres

import (
	"context"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	// mongov1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	// mysqlv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	// postgresv1 "github.com/kloudlite/operator/apis/postgres.msvc/v1"
	// // influxdbMsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	// redis "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	harborCli          *harbor.Client
	logger             logging.Logger
	Name               string
	Env                *env.Env
	yamlClient         kubectl.YAMLClient
	templateCommonMres []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	Cleanup                          string = "cleanup"
	UnderlyingManagedResourceCreated string = "underlying-managed-resource-created"
	UnderlyingManagedResourceReady   string = "underlying-managed-resource-ready"
	DefaultsPatched                  string = "defaults-patched"
)

var DeleteCheckList = []rApi.CheckMeta{}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ManagedResource{})
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

	if step := req.EnsureCheckList([]rApi.CheckMeta{
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

func getRealResourceName(obj *crdsv1.ManagedResource) string {
	if obj.Spec.ResourceNamePrefix != nil {
		return fmt.Sprintf("%s-%s", *obj.Spec.ResourceNamePrefix, obj.Name)
	}

	return obj.Name
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(DefaultsPatched, req)

	hasUpdate := false

	if obj.Output.CredentialsRef.Name == "" {
		hasUpdate = true
		obj.Output.CredentialsRef.Name = fmt.Sprintf("mres-%s-creds", getRealResourceName(obj))
	}

	ms, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.ResourceTemplate.MsvcRef.Namespace, obj.Spec.ResourceTemplate.MsvcRef.Name), &crdsv1.ManagedService{})
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
		return req.Done().RequeueAfter(500 * time.Second)
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: "cleanup", Title: "Cleanup Owned Resources"},
	}); !step.ShouldProceed() {
		return step
	}

	check := rApi.NewRunningCheck("cleanup", req)

	if result := req.CleanupOwnedResourcesV2(check); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) ensureRealMresCreated(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(UnderlyingManagedResourceCreated, req)

	b, err := templates.ParseBytes(r.templateCommonMres, map[string]any{
		"api-version": obj.Spec.ResourceTemplate.APIVersion,
		"kind":        obj.Spec.ResourceTemplate.Kind,

		"name":       getRealResourceName(obj),
		"namespace":  obj.Namespace,
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"labels":     obj.GetEnsuredLabels(),

		"msvc-ref":               obj.Spec.ResourceTemplate.MsvcRef,
		"resource-template-spec": obj.Spec.ResourceTemplate.Spec,

		"output": obj.Output,
	})
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)
	return check.Completed()
}

func (r *Reconciler) ensureRealMresReady(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(UnderlyingManagedResourceReady, req)

	uobj := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": obj.Spec.ResourceTemplate.APIVersion,
			"kind":       obj.Spec.ResourceTemplate.Kind,
		},
	}

	if err := r.Get(ctx, fn.NN(obj.GetNamespace(), getRealResourceName(obj)), &uobj); err != nil {
		return check.Failed(err)
	}

	realMresObj, err := fn.JsonConvert[struct {
		Status rApi.Status `json:"status"`
		Output struct {
			Credentials ct.SecretRef `json:"credentials"`
		} `json:"output"`
	}](uobj.Object)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	if !realMresObj.Status.IsReady {
		if realMresObj.Status.Message == nil {
			return check.Failed(fmt.Errorf("waiting for real managed service to reconcile ...")).Err(nil)
		}
		b, err := realMresObj.Status.Message.MarshalJSON()
		if err != nil {
			return check.Failed(err).Err(nil)
		}
		return check.Failed(fmt.Errorf("%s", b)).Err(nil)
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateCommonMres, err = templates.Read(templates.CommonMresTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedResource{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&corev1.Secret{})

	// owns := []client.Object{
	// 	&mongov1.StandaloneDatabase{},
	// 	&mysqlv1.StandaloneDatabase{},
	// 	&postgresv1.StandaloneDatabase{},
	// }

	// for i := range owns {
	// 	builder.Owns(owns[i])
	// }

	watchlist := []client.Object{
		// &mongov1.StandaloneService{},
		// &mongov1.ClusterService{},
		// &redis.StandaloneService{},
		// &postgresv1.Standalone{},
		// &mysqlv1.StandaloneService{},
	}

	for _, obj := range watchlist {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				if v, ok := obj.GetLabels()[constants.MresNameKey]; ok {
					return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
				}
				return nil
			}))
	}
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
