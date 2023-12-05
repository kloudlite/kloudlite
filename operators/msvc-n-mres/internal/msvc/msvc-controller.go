package msvc

import (
	"context"
	"encoding/json"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
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
	RealMsvcCreated string = "real-msvc-created"
	RealMsvcReady   string = "real-msvc-ready"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
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

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMsvcCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRealMsvcReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	req.LogPreCheck("finalizing")
	defer req.LogPostCheck("finalizing")

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) ensureRealMsvcCreated(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RealMsvcCreated)
	defer req.LogPostCheck(RealMsvcCreated)

	b, err := templates.ParseBytes(r.templateCommonMsvc, map[string]any{
		"api-version": obj.Spec.ServiceTemplate.APIVersion,
		"kind":        obj.Spec.ServiceTemplate.Kind,

		"name":       obj.Name,
		"namespace":  obj.Namespace,
		"labels":     obj.GetEnsuredLabels(),
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

		"service-template-spec": obj.Spec.ServiceTemplate.Spec,
	})
	if err != nil {
		return req.CheckFailed(RealMsvcCreated, check, err.Error()).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(RealMsvcCreated, check, err.Error()).Err(nil)
	}
	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != obj.Status.Checks[RealMsvcCreated] {
		obj.Status.Checks[RealMsvcCreated] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureRealMsvcReady(req *rApi.Request[*crdsv1.ManagedService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RealMsvcReady)
	defer req.LogPostCheck(RealMsvcReady)

	uobj := fn.NewUnstructured(metav1.TypeMeta{APIVersion: obj.Spec.ServiceTemplate.APIVersion, Kind: obj.Spec.ServiceTemplate.Kind})

	realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), uobj)
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
		if realMsvcObj.Status.Message == nil {
			return req.CheckFailed(RealMsvcReady, check, "waiting for real managed resource to reconcile ...").Err(nil)
		}
		b, err := realMsvcObj.Status.Message.MarshalJSON()
		if err != nil {
			return req.CheckFailed(RealMsvcReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(RealMsvcReady, check, string(b)).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[RealMsvcReady] {
		obj.Status.Checks[RealMsvcReady] = check
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
	r.templateCommonMsvc, err = templates.Read(templates.CommonMsvcTemplate)
	if err != nil {
		return err
	}

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
		builder.Watches(
			&source.Kind{Type: msvcs[i]},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if v, ok := obj.GetLabels()[constants.MsvcNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
					}
					return nil
				}))
		builder.Owns(msvcs[i])
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
