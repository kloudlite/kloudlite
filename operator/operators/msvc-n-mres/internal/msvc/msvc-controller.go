package msvc

import (
	"context"
	"encoding/json"
	"github.com/kloudlite/operator/operator"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	elasticsearchmsvcv1 "github.com/kloudlite/operator/apis/elasticsearch.msvc/v1"
	influxdbmsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	neo4jmsvcv1 "github.com/kloudlite/operator/apis/neo4j.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	redpandamsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	zookeeperMsvcv1 "github.com/kloudlite/operator/apis/zookeeper.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ManagedServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *ManagedServiceReconciler) GetName() string {
	return r.Name
}

const (
	RealMsvcCreated string = "real-msvc-created"
	RealMsvcReady   string = "real-msvc-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.ManagedService{})
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

	if strings.HasSuffix(request.Namespace, "-blueprint") {
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(RealMsvcReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if req.Object.Enabled != nil && !*req.Object.Enabled {
		anchor := &crdsv1.Anchor{ObjectMeta: metav1.ObjectMeta{Name: req.GetAnchorName(), Namespace: req.Object.GetNamespace()}}
		return ctrl.Result{}, client.IgnoreNotFound(r.Delete(ctx, anchor))
	}

	if step := operator.EnsureAnchor(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMsvcCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMsvcReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *ManagedServiceReconciler) finalize(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	return req.Finalize()
}

func (r *ManagedServiceReconciler) ensureRealMsvcCreated(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RealMsvcCreated)
	defer req.LogPostCheck(RealMsvcCreated)

	anchor, _ := rApi.GetLocal[*crdsv1.Anchor](req, "anchor")

	b, err := templates.Parse(
		templates.CommonMsvc, map[string]any{
			"obj":        obj,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(anchor, true)},
		},
	)
	if err != nil {
		return req.CheckFailed(RealMsvcCreated, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(RealMsvcCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[RealMsvcCreated] {
		checks[RealMsvcCreated] = check
		req.UpdateStatus()
		//return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ManagedServiceReconciler) ensureRealMsvcReady(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RealMsvcReady)
	defer req.LogPostCheck(RealMsvcReady)

	realMsvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(
			metav1.TypeMeta{Kind: obj.Spec.MsvcKind.Kind, APIVersion: obj.Spec.MsvcKind.APIVersion},
		),
	)
	if err != nil {
		return req.CheckFailed(RealMsvcReady, check, err.Error()).Err(nil)
	}

	b, err := json.Marshal(realMsvc)
	if err != nil {
		return req.CheckFailed(RealMsvcReady, check, err.Error()).Err(nil)
	}

	var realMsvcObj struct {
		Status rApi.Status `json:"status"`
	}
	if err := json.Unmarshal(b, &realMsvcObj); err != nil {
		return req.CheckFailed(RealMsvcReady, check, err.Error()).Err(nil)
	}

	if !realMsvcObj.Status.IsReady {
		if realMsvcObj.Status.Message.RawMessage == nil {
			return req.CheckFailed(RealMsvcReady, check, "waiting for real managed resource to reconcile ...").Err(nil)
		}
		b, err := realMsvcObj.Status.Message.MarshalJSON()
		if err != nil {
			return req.CheckFailed(RealMsvcReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(RealMsvcReady, check, string(b)).Err(nil)
	}

	check.Status = true
	if check != checks[RealMsvcReady] {
		checks[RealMsvcReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedService{})
	msvcs := []client.Object{
		&mongodbMsvcv1.StandaloneService{},
		&mongodbMsvcv1.ClusterService{},
		&mysqlMsvcv1.StandaloneService{},
		&mysqlMsvcv1.ClusterService{},
		&redisMsvcv1.StandaloneService{},
		&redisMsvcv1.ClusterService{},
		&elasticsearchmsvcv1.Service{},
		&zookeeperMsvcv1.Service{},
		&influxdbmsvcv1.Service{},
		&redpandamsvcv1.Service{},
		&neo4jmsvcv1.StandaloneService{},
	}

	for i := range msvcs {
		builder.Watches(&source.Kind{Type: msvcs[i]},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if v, ok := obj.GetLabels()[constants.MsvcNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
					}
					return nil
				}))
		builder.Owns(msvcs[i])
	}

	return builder.Complete(r)
}
