package watcher

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/env"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// StatusWatcherReconciler reconciles a StatusWatcher object
type StatusWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    env.Env
	*Notifier
	logger logging.Logger
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
	case fn.New(crdsv1.Router{}).GroupVersionKind(),
		fn.New(crdsv1.Project{}).GroupVersionKind(),
		fn.New(crdsv1.ManagedResource{}).GroupVersionKind():
		{
			project, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.Project{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			klMetadata := ExtractMetadata(project)
			if project.GetDeletionTimestamp() != nil {
				if err := r.notify(ctx, getMsgKey(project), klMetadata, project.Status); err != nil {
					return ctrl.Result{}, err
				}
			}
			if err := r.notify(ctx, getMsgKey(project), klMetadata, project.Status); err != nil {
				return ctrl.Result{}, err
			}
		}

	case fn.New(crdsv1.App{}).GroupVersionKind():
		{
			app, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.App{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			klMetadata := ExtractMetadata(app)
			if app.GetDeletionTimestamp() != nil {
				if err := r.notify(ctx, getMsgKey(app), klMetadata, app.Status); err != nil {
					return ctrl.Result{}, err
				}
			}
			if err := r.notify(ctx, getMsgKey(app), klMetadata, app.Status); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
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
	builder.For(&corev1.Namespace{})

	watchList := []client.Object{
		&crdsv1.Project{},
		&crdsv1.App{},
		&serverlessv1.Lambda{},
		&crdsv1.ManagedService{},
		&crdsv1.ManagedResource{},
		&crdsv1.Router{},
	}

	for _, object := range watchList {
		builder.Watches(
			&source.Kind{Type: object},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					wName, err := WrappedName{Name: obj.GetName(), Group: obj.GetObjectKind().GroupVersionKind()}.String()
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
