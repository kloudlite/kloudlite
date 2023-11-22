package mres

import (
	"context"
	"encoding/json"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	influxdbMsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
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
	return req.Finalize()
}

func (r *Reconciler) ensureOwnedByMsvc(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(OwnedByMsvc)
	defer req.LogPostCheck(OwnedByMsvc)

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.MsvcRef.Name), &crdsv1.ManagedService{},
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

	mresBytes, err := templates.Parse(
		templates.CommonMres, map[string]any{
			"object":     obj,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
	)
	if err != nil {
		return req.CheckFailed(RealMresCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, mresBytes); err != nil {
		return req.CheckFailed(RealMresCreated, check, err.Error()).Err(nil)
	}

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

	realMres, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.MsvcRef.APIVersion, Kind: obj.Spec.MresKind.Kind}),
	)
	if err != nil {
		req.Logger.Infof("real managed resource (%s) does not exist, creating it now...", fn.NN(obj.Namespace, obj.Name).String())
	}

	b, err := json.Marshal(realMres)
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

	for i := range children {
		builder.Watches(
			&source.Kind{Type: children[i]},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
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
