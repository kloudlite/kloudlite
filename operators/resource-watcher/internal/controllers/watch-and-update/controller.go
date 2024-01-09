package watch_and_update

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	if r.MsgSender == nil {
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}

	r.logger.Infof("r.MsgSender is pointed to %p", r.MsgSender)

	gvk := newGVK(obj.GetAPIVersion(), obj.GetKind())

	switch gvk.String() {
	case ProjectGVK.String(), AppGVK.String(), EnvironmentGVK.String(), RouterGVK.String(), SecretGVK.String(), ConfigmapGVK.String():
		{
			// dispatch to console
			err := r.MsgSender.DispatchConsoleResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	case ManagedResourceGVK.String():
		{
			// dispatch to console
			mresConfig := &corev1.Secret{}
			if err := r.Get(ctx, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprintf("mres-%s-creds", obj.GetName())), mresConfig); err != nil {
				r.logger.Infof("mres secret for resource (%s), not found", obj.GetName())
				mresConfig = nil
			}

			if mresConfig != nil {
				obj.Object["resource-watcher-mres-config"] = map[string]any{
					"value":    base64.StdEncoding.EncodeToString(mresConfig.Data["config"]),
					"encoding": "base64",
				}
			}

			err := r.MsgSender.DispatchConsoleResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	case ProjectManageServiceGVK.String():
		{
			// dispatch to console
			pmsvcConfig := &corev1.Secret{}
			if err := r.Get(ctx, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprintf("msvc-%s-creds", obj.GetName())), pmsvcConfig); err != nil {
				r.logger.Infof("pmsvc secret for service (%s), not found", obj.GetName())
				pmsvcConfig = nil
			}

			if pmsvcConfig != nil {
				obj.Object["resource-watcher-pmsvc-config"] = map[string]any{
					"value":    base64.StdEncoding.EncodeToString(pmsvcConfig.Data["config"]),
					"encoding": "base64",
				}
			}

			err := r.MsgSender.DispatchConsoleResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	case ClusterManagedServiceGVK.String():
		{
			// dispatch to console
			cmsvcConfig := &corev1.Secret{}
			if err := r.Get(ctx, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprintf("msvc-%s-creds", obj.GetName())), cmsvcConfig); err != nil {
				r.logger.Infof("cmsvc secret for service (%s), not found", obj.GetName())
				cmsvcConfig = nil
			}

			if cmsvcConfig != nil {
				obj.Object["resource-watcher-cmsvc-config"] = map[string]any{
					"value":    base64.StdEncoding.EncodeToString(cmsvcConfig.Data["config"]),
					"encoding": "base64",
				}
			}

			err := r.MsgSender.DispatchConsoleResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	case BuildRunGVK.String():
		{
			// dispatch to container registry
			err := r.MsgSender.DispatchContainerRegistryResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	case DeviceGVK.String():
		{
			// dispatch to infra
			deviceConfig := &corev1.Secret{}
			if err := r.Get(ctx, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprintf("wg-configs-%s", obj.GetName())), deviceConfig); err != nil {
				r.logger.Infof("wireguard secret for device (%s), not found", obj.GetName())
				deviceConfig = nil
			}

			if deviceConfig != nil {
				obj.Object["resource-watcher-wireguard-config"] = map[string]any{
					"value":    base64.StdEncoding.EncodeToString(deviceConfig.Data["config"]),
					"encoding": "base64",
				}
			}

			err := r.MsgSender.DispatchInfraResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	case NodePoolGVK.String(), PersistentVolumeClaimGVK.String(), PersistentVolumeGVK.String(), VolumeAttachmentGVK.String(), IngressGVK.String(), HelmChartGVK.String():
		{
			// dispatch to infra
			err := r.MsgSender.DispatchInfraResourceUpdates(mctx, t.ResourceUpdate{
				ClusterName: r.Env.ClusterName,
				AccountName: r.Env.AccountName,
				Object:      obj.Object,
			})
			return ctrl.Result{}, err
		}

	default:
		{
			r.logger.Infof("message sender is not configured for resource (%s) of gvk(%s). Ignoring resource update", fmt.Sprintf(obj.GetNamespace(), obj.GetName()), obj.GetObjectKind().GroupVersionKind().String())
			return ctrl.Result{}, nil
		}
	}
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
		if controllerutil.ContainsFinalizer(obj, constants.StatusWatcherFinalizer) {
			if rr, err := r.RemoveWatcherFinalizer(ctx, obj); err != nil {
				return rr, err
			}
		}
		obj.Object[t.ResourceStatusKey] = t.ResourceStatusDeleted
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
	logger.Infof("resource update received")
	defer func() {
		logger.Infof("resource update processed")
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

type GVK struct {
	schema.GroupVersionKind
}

func newGVK(apiVersion, kind string) GVK {
	return GVK{
		GroupVersionKind: fn.ParseGVK(apiVersion, kind),
	}
}

var (
	ProjectGVK = newGVK("crds.kloudlite.io/v1", "Project")
	AppGVK     = newGVK("crds.kloudlite.io/v1", "App")

	// ManagedServiceGVK = newGVK("crds.kloudlite.io/v1", "ManagedService")
	ManagedResourceGVK       = newGVK("crds.kloudlite.io/v1", "ManagedResource")
	EnvironmentGVK           = newGVK("crds.kloudlite.io/v1", "Environment")
	RouterGVK                = newGVK("crds.kloudlite.io/v1", "Router")
	NodePoolGVK              = newGVK("clusters.kloudlite.io/v1", "NodePool")
	DeviceGVK                = newGVK("wireguard.kloudlite.io/v1", "Device")
	BuildRunGVK              = newGVK("distribution.kloudlite.io/v1", "BuildRun")
	ClusterManagedServiceGVK = newGVK("crds.kloudlite.io/v1", "ClusterManagedService")
	HelmChartGVK             = newGVK("crds.kloudlite.io/v1", "HelmChart")
	ProjectManageServiceGVK  = newGVK("crds.kloudlite.io/v1", "ProjectManagedService")

	// native resources
	PersistentVolumeClaimGVK = newGVK("v1", "PersistentVolumeClaim")
	PersistentVolumeGVK      = newGVK("v1", "PersistentVolume")
	VolumeAttachmentGVK      = newGVK("storage.k8s.io/v1", "VolumeAttachment")
	IngressGVK               = newGVK("networking.k8s.io/v1", "Ingress")
	SecretGVK                = newGVK("v1", "Secret")
	ConfigmapGVK             = newGVK("v1", "ConfigMap")
)

// SetupWithManager sets up the controllers with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name).WithKV("accountName", r.Env.AccountName).WithKV("clusterName", r.Env.ClusterName)

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&corev1.Node{})

	watchList := []GVK{
		ProjectGVK,
		AppGVK,
		ManagedResourceGVK,
		EnvironmentGVK,
		RouterGVK,

		BuildRunGVK,

		DeviceGVK,

		ClusterManagedServiceGVK,
		ProjectManageServiceGVK,
		NodePoolGVK,
		HelmChartGVK,

		// native resources
		PersistentVolumeClaimGVK,
		PersistentVolumeGVK,
		VolumeAttachmentGVK,
		IngressGVK,

		// filtered watch
		SecretGVK,
		ConfigmapGVK,
	}

	for i := range watchList {
		restype := fn.NewUnstructured(metav1.TypeMeta{Kind: watchList[i].Kind, APIVersion: watchList[i].GroupVersion().String()})
		builder.Watches(
			restype,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					gvk := obj.GetObjectKind().GroupVersionKind().String()

					if (gvk == SecretGVK.String()) && !fn.MapContains(obj.GetAnnotations(), t.SecretWatchingAnnotation) {
						return nil
					}

					if (gvk == ConfigmapGVK.String()) && !fn.MapContains(obj.GetAnnotations(), t.ConfigWatchingAnnotation) {
						return nil
					}

					b64Group := base64.StdEncoding.EncodeToString([]byte(gvk))
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
