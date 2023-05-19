package watch_and_update

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/types"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

// Reconciler reconciles a StatusWatcher object
type Reconciler struct {
	client.Client
	Scheme                    *runtime.Scheme
	logger                    logging.Logger
	Name                      string
	Env                       *env.Env
	GetGrpcConnection         func() (*grpc.ClientConn, error)
	dispatchResourceUpdates   func(ctx context.Context, stu t.ResourceUpdate) error
	dispatchInfraUpdates      func(ctx context.Context, stu t.ResourceUpdate) error
	dispatchBYOCClientUpdates func(ctx context.Context, stu t.ResourceUpdate) error
	accessToken               string
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func (r *Reconciler) SendResourceEvents(ctx context.Context, obj client.Object, logger logging.Logger) (ctrl.Result, error) {
	obj.SetManagedFields(nil)

	b, err := json.Marshal(obj)
	if err != nil {
		return ctrl.Result{}, nil
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return ctrl.Result{}, err
	}

	switch {
	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "infra.kloudlite.io"):
		{
			if err := r.dispatchInfraUpdates(ctx, t.ResourceUpdate{
				// ClusterName: obj.GetLabels()[constants.ClusterNameKey],
				// AccountName: obj.GetLabels()[constants.AccountNameKey],
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      m,
			}); err != nil {
				return ctrl.Result{}, err
			}
		}

	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "clusters.kloudlite.io"):
		{
			if obj.GetObjectKind().GroupVersionKind().Kind == "BYOC" {
				var byoc clustersv1.BYOC
				if err := json.Unmarshal(b, &byoc); err != nil {
					return ctrl.Result{}, err
				}

				if err := r.dispatchBYOCClientUpdates(ctx, t.ResourceUpdate{
					ClusterName: byoc.Name,
					AccountName: byoc.Spec.AccountName,
					Object:      m,
				}); err != nil {
					return ctrl.Result{}, err
				}
			}
		}

	case strings.HasSuffix(obj.GetObjectKind().GroupVersionKind().Group, "kloudlite.io"):
		{
			if err := r.dispatchResourceUpdates(ctx, t.ResourceUpdate{
				// ClusterName: obj.GetLabels()[constants.ClusterNameKey],
				// AccountName: obj.GetLabels()[constants.AccountNameKey],
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      m,
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

	logger.WithKV("timestamp", time.Now()).Infof("dispatched update to message office api")

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
	if r.accessToken == "" {
		r.logger.Infof("trying to read accessToken")
		var clusterIdentity corev1.Secret
		if err := r.Get(ctx, fn.NN(r.Env.ClusterIdentitySecretNamespace, r.Env.ClusterIdentitySecretName), &clusterIdentity); err != nil {
			r.logger.Infof("waiting to read accessToken, retrying every 2s till then")
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		r.accessToken = string(clusterIdentity.Data["ACCESS_TOKEN"])
		r.logger.Infof("successfully retrieved accessToken")
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

	logger := r.logger.WithName(fn.NN(oReq.Namespace, wName.Name).String()).WithKV("gvk", gvk.String())
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
	r.logger = logger.WithName(r.Name)

	r.dispatchResourceUpdates = func(ctx context.Context, ru t.ResourceUpdate) error {
		return fmt.Errorf("grpc connection not established yet")
	}

	r.dispatchInfraUpdates = func(ctx context.Context, ru t.ResourceUpdate) error {
		return fmt.Errorf("grpc connection not established yet")
	}

	r.dispatchBYOCClientUpdates = func(ctx context.Context, ru t.ResourceUpdate) error {
		return fmt.Errorf("grpc connection not established yet")
	}

	go func() {
		handlerCh := make(chan error, 1)
		for {
			logger.Infof("Waiting for grpc connection to setup")
			cc, err := r.GetGrpcConnection()
			if err != nil {
				log.Fatalf("Failed to connect after retries: %v", err)
			}

			logger.Infof("GRPC connection successful")

			msgDispatchCli := messages.NewMessageDispatchServiceClient(cc)

			mds, err := msgDispatchCli.ReceiveResourceUpdates(context.Background())
			if err != nil {
				logger.Errorf(err, "ReceiveStatusMessages")
			}

			r.dispatchResourceUpdates = func(_ context.Context, ru t.ResourceUpdate) error {
				b, err := json.Marshal(ru)
				if err != nil {
					return err
				}
				if err = mds.Send(&messages.ResourceUpdate{
					AccessToken: r.accessToken,
					ClusterName: r.Env.ClusterName,
					AccountName: r.Env.AccountName,
					Message:     b,
				}); err != nil {
					handlerCh <- err
					return err
				}
				return nil
			}

			infraMessagesCli, err := msgDispatchCli.ReceiveInfraUpdates(context.Background())
			if err != nil {
				log.Fatalf(err.Error())
			}

			r.dispatchInfraUpdates = func(_ context.Context, ru t.ResourceUpdate) error {
				b, err := json.Marshal(ru)
				if err != nil {
					return err
				}

				if err = infraMessagesCli.Send(&messages.InfraUpdate{
					AccessToken: r.accessToken,
					ClusterName: r.Env.ClusterName,
					AccountName: r.Env.AccountName,
					Message:     b,
				}); err != nil {
					handlerCh <- err
					return err
				}
				return nil
			}

			byocClientUpdatesCli, err := msgDispatchCli.ReceiveBYOCClientUpdates(context.Background())
			if err != nil {
				logger.Errorf(err, "ReceiveBYOCClientUpdates")
			}

			r.dispatchBYOCClientUpdates = func(_ context.Context, ru t.ResourceUpdate) error {
				b, err := json.Marshal(ru)
				if err != nil {
					return err
				}

				if err = byocClientUpdatesCli.Send(&messages.BYOCClientUpdate{
					AccessToken: r.accessToken,
					ClusterName: r.Env.ClusterName,
					AccountName: r.Env.AccountName,
					Message:     b,
				}); err != nil {
					handlerCh <- err
					return err
				}
				return nil
			}

			connState := cc.GetState()
			go func(cs connectivity.State) {
				for cs != connectivity.Ready && connState != connectivity.Shutdown {
					handlerCh <- fmt.Errorf("connection lost")
				}
			}(connState)
			<-handlerCh
			cc.Close()
		}
	}()

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

		&clustersv1.BYOC{},

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
