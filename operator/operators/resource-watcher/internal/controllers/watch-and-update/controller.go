package watch_and_update

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	wireguardv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/types"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

// Reconciler reconciles a StatusWatcher object
type Reconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	logger      logging.Logger
	Name        string
	Env         *env.Env
	accessToken string
	MsgSender   MessageSender
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func (r *Reconciler) SendResourceEvents(ctx context.Context, obj *unstructured.Unstructured, logger logging.Logger) (ctrl.Result, error) {
	obj.SetManagedFields(nil)

	// mctx, cf := context.WithTimeout(ctx, 100*time.Second)
	_ = time.Second
	mctx, cf := context.WithCancel(ctx)
	defer cf()

	switch {
	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "infra.kloudlite.io"):
		{
			if err := r.MsgSender.DispatchInfraUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "clusters.kloudlite.io"):
		{
			switch obj.GetObjectKind().GroupVersionKind().Kind {
			case "Cluster":
				{
					if err := r.MsgSender.DispatchInfraUpdates(mctx, t.ResourceUpdate{
						ClusterName: obj.GetName(),
						AccountName: obj.Object["spec"].(map[string]any)["accountName"].(string),
						Object:      obj.Object,
					}); err != nil {
						return ctrl.Result{}, err
					}
				}
			default:
				{
					if err := r.MsgSender.DispatchInfraUpdates(mctx, t.ResourceUpdate{
						ClusterName: r.Env.ClusterName,
						AccountName: r.Env.AccountName,
						Object:      obj.Object,
					}); err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		}

	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "wireguard.kloudlite.io"):
		{
			switch obj.GetObjectKind().GroupVersionKind().Kind {
			case "Device":
				{
					deviceConfig := &corev1.Secret{}
					if err := r.Get(ctx, fn.NN(obj.GetNamespace(), fmt.Sprintf("wg-configs-%s", obj.GetName())), deviceConfig); err != nil {
						r.logger.Infof("wireguard secret for device (%s), not found", obj.GetName())
						deviceConfig = nil
					}

					if deviceConfig != nil {
						obj.Object["resource-watcher-wireguard-config"] = map[string]any{
							"value":    base64.StdEncoding.EncodeToString(deviceConfig.Data["config"]),
							"encoding": "base64",
						}
					}

					if err := r.MsgSender.DispatchInfraUpdates(mctx, t.ResourceUpdate{
						ClusterName: r.Env.ClusterName,
						AccountName: r.Env.AccountName,
						Object:      obj.Object,
					}); err != nil {
						return ctrl.Result{}, err
					}
				}
			default:
				{
					return ctrl.Result{}, nil
				}
			}
		}

	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "kloudlite.io"):
		{
			if err := r.MsgSender.DispatchResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	default:
		{
			logger.Infof("ignoring resource status update, as it does not belong to group kloudlite.io")
			return ctrl.Result{}, nil
		}
	}

	if obj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
			return r.RemoveWatcherFinalizer(mctx, obj)
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
		return r.AddWatcherFinalizer(mctx, obj)
	}

	return ctrl.Result{}, nil
}

// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/finalizers,verbs=update
func (r *Reconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	if r.MsgSender == nil {
		r.logger.Infof("message-sender is nil")
		return ctrl.Result{
			// RequeueAfter: 2 *time.Second,
		}, errors.New("waiting for message sender to be initialized")
	}

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

	logger := r.logger.WithKV("NN", fn.NN(oReq.Namespace, wName.Name).String()).WithKV("gvk", gvk.String())
	logger.Infof("request received")
	defer func() {
		logger.Infof("request processed")
	}()

	tm := metav1.TypeMeta{Kind: gvk.Kind, APIVersion: fmt.Sprintf("%s/%s", gvk.Group, gvk.Version)}
	obj, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), fn.NewUnstructured(tm))
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.SendResourceEvents(ctx, obj, logger)
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
	r.logger = logger.WithName(r.Name).WithKV("accountName", r.Env.AccountName).WithKV("clusterName", r.Env.ClusterName)

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&crdsv1.Project{})

	watchList := []client.Object{
		&crdsv1.Project{},
		&crdsv1.App{},
		&serverlessv1.Lambda{},
		&crdsv1.ManagedService{},
		&crdsv1.ManagedResource{},
		&crdsv1.Router{},
		&crdsv1.Workspace{},
		&crdsv1.Config{},
		&crdsv1.Secret{},

		&clustersv1.Cluster{},
		&clustersv1.NodePool{},
		// &clustersv1.Node{},

		&corev1.PersistentVolumeClaim{},
		&wireguardv1.Device{},
	}

	for _, object := range watchList {
		builder.Watches(
			object,
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
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
						{NamespacedName: fn.NN(obj.GetNamespace(), wName)},
					}
				},
			),
		)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
