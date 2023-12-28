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

	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/types"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

const ClusterScopeNamespace = "kloudlite-cluster-scope"

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

func (r *Reconciler) dispatchEvent(ctx context.Context, obj *unstructured.Unstructured) (ctrl.Result, error) {
	mctx, cf := func() (context.Context, context.CancelFunc) {
		if r.Env.IsDev {
			return context.WithCancel(context.TODO())
		}
		return context.WithTimeout(ctx, 2*time.Second)
	}()
	defer cf()

	belongsTo := func(group string) bool {
		return strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, group)
	}

	switch {
	case belongsTo("infra.kloudlite.io"):
		{
			if err := r.MsgSender.DispatchInfraResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	case belongsTo("clusters.kloudlite.io"):
		{
			if err := r.MsgSender.DispatchInfraResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	case belongsTo("wireguard.kloudlite.io"):
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

					if err := r.MsgSender.DispatchInfraResourceUpdates(mctx, t.ResourceUpdate{
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
	case belongsTo("distribution.kloudlite.io"):
		{
			if err := r.MsgSender.DispatchContainerRegistryResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	case belongsTo("kloudlite.io"):
		{
			if err := r.MsgSender.DispatchConsoleResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	case belongsTo(""):
		if err := r.MsgSender.DispatchConsoleResourceUpdates(mctx, t.ResourceUpdate{
			ClusterName: r.Env.ClusterName,
			AccountName: r.Env.AccountName,
			Object:      obj.Object,
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SendResourceEvents(ctx context.Context, obj *unstructured.Unstructured, logger logging.Logger) (ctrl.Result, error) {
	obj.SetManagedFields(nil)

	if obj.GetDeletionTimestamp() != nil {
		// resource is about to be deleted
		if t.HasOtherKloudliteFinalizers(obj) {
			// 1. send deleting event
			obj.Object[t.ResourceStatusKey] = t.ResourceStatusDeleting
			return r.dispatchEvent(ctx, obj)
		}

		// 2. send deleted event
		obj.Object[t.ResourceStatusKey] = t.ResourceStatusDeleted
		if controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
			if rr, err := r.RemoveWatcherFinalizer(ctx, obj); err != nil {
				return rr, err
			}
		}
		return r.dispatchEvent(ctx, obj)
	}

	if !controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
		if rr, err := r.AddWatcherFinalizer(ctx, obj); err != nil {
			return rr, err
		}
	}

	obj.Object[t.ResourceStatusKey] = t.ResourceStatusUpdated
	return r.dispatchEvent(ctx, obj)
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
	fmt.Println(gvk.Group)
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

type WatchResource struct {
	metav1.TypeMeta `json:",inline"`
}

func NewWatchResource(apiVersion string, kind string) WatchResource {
	return WatchResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		},
	}
}

// SetupWithManager sets up the controllers with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name).WithKV("accountName", r.Env.AccountName).WithKV("clusterName", r.Env.ClusterName)

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&corev1.Node{})

	watchList := []WatchResource{
		NewWatchResource("crds.kloudlite.io/v1", "Project"),
		NewWatchResource("crds.kloudlite.io/v1", "App"),
		NewWatchResource("crds.kloudlite.io/v1", "ManagedService"),
		NewWatchResource("crds.kloudlite.io/v1", "ManagedResource"),
		NewWatchResource("crds.kloudlite.io/v1", "Workspace"),
		NewWatchResource("crds.kloudlite.io/v1", "Router"),
		NewWatchResource("clusters.kloudlite.io/v1", "NodePool"),
		NewWatchResource("wireguard.kloudlite.io/v1", "Device"),
		NewWatchResource("distribution.kloudlite.io/v1", "BuildRun"),

		// native resources
		NewWatchResource("v1", "PersistentVolumeClaim"),
		NewWatchResource("v1", "PersistentVolume"),
		NewWatchResource("networking.k8s.io/v1", "Ingress"),
	}

	for i := range watchList {
		func(obj WatchResource) {
			fmt.Println(obj.APIVersion, obj.Kind)
			builder.Watches(
				fn.NewUnstructured(obj.TypeMeta),
				handler.EnqueueRequestsFromMapFunc(
					func(_ context.Context, obj client.Object) []reconcile.Request {
						gvk := obj.GetObjectKind().GroupVersionKind().String()

						b64Group := base64.StdEncoding.EncodeToString([]byte(gvk))
						if len(b64Group) == 0 {
							return nil
						}

						wName, err := types.WrappedName{Name: obj.GetName(), Group: b64Group}.String()
						if err != nil {
							return nil
						}
						if obj.GetNamespace() == "" {
							return []reconcile.Request{
								{NamespacedName: fn.NN(ClusterScopeNamespace, wName)},
							}
						} else {
							return []reconcile.Request{
								{NamespacedName: fn.NN(obj.GetNamespace(), wName)},
							}
						}
					},
				),
			)

		}(watchList[i])
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
