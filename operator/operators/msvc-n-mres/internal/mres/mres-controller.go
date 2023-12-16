package mres

import (
	"context"
	"encoding/json"
	"time"

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
	RealMresCreated string = "real-mres-created"
	RealMresReady   string = "real-mres-ready"
	MsvcIsOwner     string = "msvc-is-owner"
	DefaultsPatched string = "defaults-patched"
	OwnedByMsvc     string = "owned-by-msvc"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ManagedResource{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

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

	if req.Object.Enabled != nil && !*req.Object.Enabled {
		// TODO (nxtcoder17): need to finalize this resource
		return req.Done().ReconcilerResponse()
	}

	if step := r.ensureOwnedByMsvc(req); !step.ShouldProceed() {
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

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	checkName := "finalizing"
	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) ensureOwnedByMsvc(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(OwnedByMsvc)
	defer req.LogPostCheck(OwnedByMsvc)

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ResourceTemplate.MsvcRef.Name), &crdsv1.ManagedService{},
	)
	if err != nil {
		return req.CheckFailed(OwnedByMsvc, check, err.Error())
	}

	hasUpdated := false
	if !fn.MapContains(obj.Labels, msvc.Labels) {
		hasUpdated = true
		for k, v := range msvc.Labels {
			obj.Labels[k] = v
		}
	}

	if !fn.IsOwner(obj, fn.AsOwner(msvc)) {
		hasUpdated = true
		obj.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(msvc, true)})
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(OwnedByMsvc, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[OwnedByMsvc] {
		obj.Status.Checks[OwnedByMsvc] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureRealMresCreated(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RealMresCreated)
	defer req.LogPostCheck(RealMresCreated)

	b, err := templates.ParseBytes(r.templateCommonMres, map[string]any{
		"api-version": obj.Spec.ResourceTemplate.APIVersion,
		"kind":        obj.Spec.ResourceTemplate.Kind,

		"name":       obj.Name,
		"namespace":  obj.Namespace,
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"labels":     obj.GetEnsuredLabels(),

		"msvc-ref":               obj.Spec.ResourceTemplate.MsvcRef,
		"resource-template-spec": obj.Spec.ResourceTemplate.Spec,
	})
	if err != nil {
		return req.CheckFailed(RealMresCreated, check, err.Error()).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(RealMresCreated, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != obj.Status.Checks[RealMresCreated] {
		obj.Status.Checks[RealMresCreated] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureRealMresReady(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RealMresReady)
	defer req.LogPreCheck(RealMresReady)

	uobj := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": obj.Spec.ResourceTemplate.APIVersion,
			"kind":       obj.Spec.ResourceTemplate.Kind,
			"metadata": map[string]any{
				"name":      obj.Name,
				"namespace": obj.Namespace,
			},
		},
	}

	if err := r.Get(ctx, client.ObjectKeyFromObject(&uobj), &uobj); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return req.CheckFailed(RealMresReady, check, err.Error())
		}
	}

	b, err := json.Marshal(uobj.Object)
	if err != nil {
		return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
	}

	var realMresObj struct {
		Status rApi.Status `json:"status"`
	}

	if err := json.Unmarshal(b, &realMresObj); err != nil {
		return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
	}

	if !realMresObj.Status.IsReady {
		if realMresObj.Status.Message == nil {
			return req.CheckFailed(RealMresReady, check, "waiting for real managed resource to reconcile ...")
		}
		b, err := json.Marshal(realMresObj.Status.Message)
		if err != nil {
			return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(RealMresReady, check, string(b)).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[RealMresReady] {
		obj.Status.Checks[RealMresReady] = check
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
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

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
