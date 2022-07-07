package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
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

// BillingWatcherReconciler reconciles a BillingWatcher object
type BillingWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    env.Env
	*Notifier
	logger logging.Logger
}

// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=billingwatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=billingwatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=billingwatchers/finalizers,verbs=update

func (r *BillingWatcherReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	r.logger.Infof("request received ...")
	var wName WrappedName
	if err := json.Unmarshal([]byte(oReq.Name), &wName); err != nil {
		return ctrl.Result{}, nil
	}
	switch wName.Group {
	case fn.New(crdsv1.App{}).GroupVersionKind():
		{
			app, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.App{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			klMetadata := ExtractMetadata(app)

			replicaCount, ok := app.Status.DisplayVars.GetInt("readyReplicas")
			if !ok {
				return ctrl.Result{}, errors.Newf("no readyReplicas key found in .DisplayVars")
			}
			billing := ResourceBilling{
				Name: fmt.Sprintf("%s/%s", app.Namespace, app.Name),
				Items: []k8sItem{
					{Type: Pod, Count: replicaCount, Plan: Plan(klMetadata.Plan)},
				},
			}
			if app.GetDeletionTimestamp() != nil {
				if controllerutil.ContainsFinalizer(app, constants.BillingFinalizer) {
					billing.ToBeDeleted = true
					if err := r.notifyBilling(ctx, getMsgKey(app), klMetadata, &billing); err != nil {
						return ctrl.Result{}, err
					}
				}
				return r.RemoveBillingFinalizer(ctx, app)
			}

			if !controllerutil.ContainsFinalizer(app, constants.BillingFinalizer) {
				return r.AddBillingFinalizer(ctx, app)
			}

			if err := r.notifyBilling(ctx, getMsgKey(app), klMetadata, &billing); err != nil {
				return ctrl.Result{}, err
			}
		}
	case fn.New(crdsv1.ManagedService{}).GroupVersionKind():
		{
		}
	case fn.New(serverlessv1.Lambda{}).GroupVersionKind():
		{
		}
	}
	return ctrl.Result{}, nil
}

func (r *BillingWatcherReconciler) AddBillingFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.AddFinalizer(obj, constants.BillingFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

func (r *BillingWatcherReconciler) RemoveBillingFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(obj, constants.BillingFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

// SetupWithManager sets up the controller with the Manager.

func (r *BillingWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
