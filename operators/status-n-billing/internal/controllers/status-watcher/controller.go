package status_watcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	"github.com/kloudlite/operator/operators/status-n-billing/internal/env"
	"github.com/kloudlite/operator/operators/status-n-billing/internal/types"
	statusType "github.com/kloudlite/operator/operators/status-n-billing/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"github.com/kloudlite/operator/pkg/redpanda"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler reconciles a StatusWatcher object
type Reconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	Producer     redpanda.Producer
	logger       logging.Logger
	Name         string
	Env          *env.Env
	getTopicName func(group string) string
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func (r *Reconciler) SendStatusEvents(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return ctrl.Result{}, nil
	}

	var j struct {
		Spec struct {
			AccountName string `json:"accountName"`
		} `json:"spec"`
		Status rApi.Status `json:"status"`
	}

	if err := json.Unmarshal(b, &j); err != nil {
		return ctrl.Result{}, nil
	}

	obj.SetManagedFields(nil)

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return ctrl.Result{}, err
	}

	b, err = json.Marshal(statusType.StatusUpdate{
		ClusterName: r.Env.ClusterName,
		AccountName: j.Spec.AccountName,
		Object:      m,
		Status:      j.Status,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	pm, err := r.Producer.Produce(ctx, r.getTopicName(obj.GetObjectKind().GroupVersionKind().Group), r.Env.ClusterName, b)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.logger.Infof("dispatched update to %s @ %s", pm.Topic, pm.Timestamp.String())

	if obj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
			return r.RemoveWatcherFinalizer(ctx, obj)
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
		return r.AddWatcherFinalizer(ctx, obj)
	}

	return ctrl.Result{}, nil
}

// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	var wName types.WrappedName
	if err := json.Unmarshal([]byte(oReq.Name), &wName); err != nil {
		return ctrl.Result{}, nil
	}

	gvk, err := wName.ParseGroup()
	if err != nil {
		r.logger.Errorf(err, "badly formatted group-version-kind (%s) received, aborting ...", wName.Group)
		return ctrl.Result{}, nil
	}

	if gvk == nil {
		return ctrl.Result{}, nil
	}

	logger := r.logger.WithName(fn.NN(oReq.Namespace, wName.Name).String()).WithKV("RefKind", gvk.String())
	logger.Infof("request received ...")

	tm := metav1.TypeMeta{Kind: gvk.Kind, APIVersion: fmt.Sprintf("%s/%s", gvk.Group, gvk.Version)}
	obj, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), fn.NewUnstructured(tm))
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.SendStatusEvents(ctx, obj)
}

func (r *Reconciler) AddWatcherFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.AddFinalizer(obj, constants.StatusWatcherFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

func (r *Reconciler) RemoveWatcherFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(obj, constants.StatusWatcherFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

// SetupWithManager sets up the controllers with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	r.getTopicName = func(group string) string {
		if group == "infra.kloudlite.io" || group == "cmgr.kloudlite.io" {
			return r.Env.KafkaTopicInfraUpdates
		}
		return r.Env.KafkaTopicStatusUpdates
	}

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&crdsv1.Project{})

	watchList := []client.Object{
		&crdsv1.Project{},
		&crdsv1.App{},
		&serverlessv1.Lambda{},
		&crdsv1.ManagedService{},
		&crdsv1.ManagedResource{},
		&crdsv1.Router{},
		&crdsv1.Env{},
		&crdsv1.Config{},
		&crdsv1.Secret{},
		fn.NewUnstructured(constants.EdgeInfraType),
		fn.NewUnstructured(constants.CloudProviderType),
		fn.NewUnstructured(constants.WorkerNodeType),
		fn.NewUnstructured(constants.NodePoolType),

		fn.NewUnstructured(constants.ClusterType),
		fn.NewUnstructured(constants.MasterNodeType),
		// fn.NewUnstructured(constants.DeviceType),
	}

	for _, object := range watchList {
		builder.Watches(
			&source.Kind{Type: object},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					v, ok := obj.GetAnnotations()[constants.GVKKey]
					if !ok {
						return nil
					}
					b64Group := base64.StdEncoding.EncodeToString([]byte(v))
					if len(b64Group) == 0 {
						return nil
					}

					wName, err := types.WrappedName{Name: obj.GetName(), Group: b64Group}.String()
					if err != nil {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: types2.NamespacedName{Namespace: obj.GetNamespace(), Name: wName}},
					}
				},
			),
		)
	}
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
