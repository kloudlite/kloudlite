package watcher

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// StatusWatcherReconciler reconciles a StatusWatcher object
type StatusWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env
	*Notifier
	logger logging.Logger
}

func (r *StatusWatcherReconciler) GetName() string {
	return "status-watcher"
}

// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=statuswatchers/finalizers,verbs=update

func (r *StatusWatcherReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	r.logger.Infof("request received ...")
	var wName WrappedName
	if err := json.Unmarshal([]byte(oReq.Name), &wName); err != nil {
		return ctrl.Result{}, nil
	}

	switch wName.Group {
	case crdsv1.GroupVersion.WithKind("Project").String():
		{
			project, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.Project{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			klMetadata := ExtractMetadata(project)
			if project.GetDeletionTimestamp() != nil {
				if controllerutil.ContainsFinalizer(project, constants.StatusWatcherFinalizer) {
					if err := r.notify(ctx, getMsgKey(project), klMetadata, project.Status, Stages.Deleted); err != nil {
						return ctrl.Result{}, err
					}
					return r.RemoveWatcherFinalizer(ctx, project)
				}
				return ctrl.Result{}, nil
			}
			if !controllerutil.ContainsFinalizer(project, constants.StatusWatcherFinalizer) {
				return r.AddWatcherFinalizer(ctx, project)
			}
			if err := r.notify(ctx, getMsgKey(project), klMetadata, project.Status, Stages.Exists); err != nil {
				return ctrl.Result{}, err
			}
		}

	case crdsv1.GroupVersion.WithKind("App").String():
		{
			app, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.App{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			klMetadata := ExtractMetadata(app)
			if app.GetDeletionTimestamp() != nil {
				if controllerutil.ContainsFinalizer(app, constants.StatusWatcherFinalizer) {
					if err := r.notify(ctx, getMsgKey(app), klMetadata, app.Status, Stages.Deleted); err != nil {
						return ctrl.Result{}, err
					}
				}
				return r.RemoveWatcherFinalizer(ctx, app)
			}
			if !controllerutil.ContainsFinalizer(app, constants.StatusWatcherFinalizer) {
				return r.AddWatcherFinalizer(ctx, app)
			}
			if err := r.notify(ctx, getMsgKey(app), klMetadata, app.Status, Stages.Exists); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *StatusWatcherReconciler) AddWatcherFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.AddFinalizer(obj, constants.StatusWatcherFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

func (r *StatusWatcherReconciler) RemoveWatcherFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(obj, constants.StatusWatcherFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StatusWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logger, err := logging.New(
		&logging.Options{
			Name: "status-watcher",
			Dev:  true,
		},
	)

	if err != nil {
		panic(err)
	}

	r.logger = logger

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&crdsv1.Project{})

	watchList := []client.Object{
		&crdsv1.Project{},
		&crdsv1.App{},
		&crdsv1.ManagedService{},
		&crdsv1.ManagedResource{},
		&crdsv1.Router{},
	}

	for _, object := range watchList {
		builder.Watches(
			&source.Kind{Type: object},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {

					wName, err := WrappedName{Name: obj.GetName(), Group: obj.GetAnnotations()[constants.AnnotationKeys.GroupVersionKind]}.
						String()
					if err != nil {
						return nil
					}
					return []reconcile.Request{
						{
							NamespacedName: types2.NamespacedName{Namespace: obj.GetNamespace(), Name: wName},
						},
					}
				},
			),
		)
	}
	return builder.Complete(r)
}
