package mres

import (
	"context"
	"fmt"
	"slices"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	influxdbMsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
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

var ApplyCheckList = []rApi.CheckMeta{
	{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
	{Name: UnderlyingManagedResourceCreated, Title: "Underlying Managed Resource Created"},
	{Name: UnderlyingManagedResourceReady, Title: "Underlying Managed Resource Ready"},
}

var DeleteCheckList = []rApi.CheckMeta{
	{Name: Cleanup, Title: "Cleanup"},
}

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

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
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
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := DefaultsPatched

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hasUpdated := false

	if obj.Output.CredentialsRef.Name == "" {
		hasUpdated = true
		obj.Output.CredentialsRef.Name = fmt.Sprintf("mres-%s-creds", getRealResourceName(obj))
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}
		return req.Done().RequeueAfter(500 * time.Second)
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

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	obj := req.Object

	if !slices.Equal(obj.Status.CheckList, DeleteCheckList) {
		if step := req.EnsureCheckList(DeleteCheckList); !step.ShouldProceed() {
			return step
		}
	}

	checkName := "finalizing"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

// func (r *Reconciler) ensureOwnedByMsvc(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}
//
// 	checkName := ResourceOwnedByManagedSvc
//
// 	req.LogPreCheck(checkName)
// 	defer req.LogPostCheck(checkName)
//
// 	fail := func(err error) stepResult.Result {
// 		return req.CheckFailed(checkName, check, err.Error())
// 	}
//
// 	msvc, err := rApi.Get(
// 		ctx, r.Client, fn.NN(obj.Spec.ResourceTemplate.MsvcRef.Namespace, obj.Spec.ResourceTemplate.MsvcRef.Name), &crdsv1.ManagedService{},
// 	)
// 	if err != nil {
// 		return fail(err)
// 	}
//
// 	hasUpdated := false
//
// 	msvcLabels := fn.MapFilter(msvc.Labels, constants.KloudliteLabelPrefix)
// 	if !fn.MapContains(obj.Labels, msvcLabels) {
// 		hasUpdated = true
// 		fn.MapJoin(&obj.Labels, msvcLabels)
// 	}
//
// 	if !fn.IsOwner(obj, fn.AsOwner(msvc)) {
// 		hasUpdated = true
// 		obj.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(msvc, true)})
// 	}
//
// 	if hasUpdated {
// 		if err := r.Update(ctx, obj); err != nil {
// 			return fail(err)
// 		}
// 		return req.Done().RequeueAfter(100 * time.Millisecond)
// 	}
//
// 	check.Status = true
// 	if check != obj.Status.Checks[checkName] {
// 		fn.MapSet(&obj.Status.Checks, checkName, check)
// 		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
// 			return sr
// 		}
// 	}
//
// 	return req.Next()
// }

func (r *Reconciler) ensureRealMresCreated(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := UnderlyingManagedResourceCreated

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

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

func (r *Reconciler) ensureRealMresReady(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := UnderlyingManagedResourceReady

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	uobj := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": obj.Spec.ResourceTemplate.APIVersion,
			"kind":       obj.Spec.ResourceTemplate.Kind,
		},
	}

	if err := r.Get(ctx, fn.NN(obj.GetNamespace(), getRealResourceName(obj)), &uobj); err != nil {
		return fail(err)
	}

	realMresObj, err := fn.JsonConvert[struct {
		Status rApi.Status `json:"status"`
		Output struct {
			Credentials ct.SecretRef `json:"credentials"`
		} `json:"output"`
	}](uobj.Object)
	if err != nil {
		return fail(err).Err(nil)
	}

	if !realMresObj.Status.IsReady {
		if realMresObj.Status.Message == nil {
			return req.CheckFailed(checkName, check, "waiting for real managed service to reconcile ...").Err(nil)
		}
		b, err := realMresObj.Status.Message.MarshalJSON()
		if err != nil {
			return fail(err).Err(nil)
		}
		return fail(fmt.Errorf("%s", b)).Err(nil)
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

	children := []client.Object{
		&redisMsvcv1.StandaloneService{},
		&redisMsvcv1.ClusterService{},
		&redisMsvcv1.ACLAccount{},
		&redisMsvcv1.Prefix{},

		&mongodbMsvcv1.StandaloneService{},
		&mongodbMsvcv1.ClusterService{},
		&mongodbMsvcv1.Database{},

		&mysqlMsvcv1.ClusterService{},
		&mysqlMsvcv1.StandaloneService{},
		&mysqlMsvcv1.Database{},

		&influxdbMsvcv1.Bucket{},
		&influxdbMsvcv1.Service{},
	}

	for _, obj := range children {
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
