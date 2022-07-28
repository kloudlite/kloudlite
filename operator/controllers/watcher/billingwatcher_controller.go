package watcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	mysqlStandalone "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
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
	Env    *env.Env
	*Notifier
	logger logging.Logger
}

func (r *BillingWatcherReconciler) GetName() string {
	return "billing-watcher"
}

func (r *BillingWatcherReconciler) SendBillingEvent(ctx context.Context, obj client.Object, billing ResourceBilling) (ctrl.Result, error) {
	klMetadata := ExtractMetadata(obj)
	if obj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(obj, constants.BillingFinalizer) {
			if err := r.notifyBilling(ctx, getMsgKey(obj), klMetadata, &billing, Stages.Deleted); err != nil {
				return ctrl.Result{}, err
			}
		}
		return r.RemoveBillingFinalizer(ctx, obj)
	}

	if !controllerutil.ContainsFinalizer(obj, constants.BillingFinalizer) {
		return r.AddBillingFinalizer(ctx, obj)
	}

	if err := r.notifyBilling(ctx, getMsgKey(obj), klMetadata, &billing, Stages.Exists); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
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

	gvk, err := parseGroup(wName.Group)
	if err != nil {
		r.logger.Errorf(err, "badly formatted group-version-kind (%s) received, aborting ...", wName.Group)
		return ctrl.Result{}, nil
	}

	if gvk == nil {
		return ctrl.Result{}, nil
	}

	switch *gvk {
	case crdsv1.GroupVersion.WithKind("App"):
		{
			app, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.App{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			replicaCount, ok := app.Status.DisplayVars.GetInt("readyReplicas")
			if !ok {
				return ctrl.Result{}, errors.Newf("no readyReplicas key found in .DisplayVars")
			}
			billing := ResourceBilling{
				Name:  fmt.Sprintf("%s/%s", app.Namespace, app.Name),
				Items: []k8sItem{newK8sItem(app, Pod, replicaCount)},
			}
			return r.SendBillingEvent(ctx, app, billing)
		}

	case crdsv1.GroupVersion.WithKind("ManagedService"):
		{
			msvc, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.ManagedService{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			realMsvcType := metav1.TypeMeta{APIVersion: msvc.Spec.MsvcKind.APIVersion, Kind: msvc.Spec.MsvcKind.Kind}

			switch realMsvcType.GetObjectKind().GroupVersionKind() {
			case mongodbStandalone.GroupVersion.WithKind("Service"):
				{
					realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), &mongodbStandalone.Service{})
					if err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}

					billing := ResourceBilling{
						Name: fmt.Sprintf("%s/%s", msvc.Namespace, msvc.Name),
						Items: []k8sItem{
							newK8sItem(msvc, Pod, realMsvc.Spec.ReplicaCount),
							newK8sItem(msvc, Pvc, realMsvc.Spec.Storage.ToInt()),
						},
					}
					return r.SendBillingEvent(ctx, msvc, billing)
				}
			case redisStandalone.GroupVersion.WithKind("Service"):
				{
					realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), &mongodbStandalone.Service{})
					if err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}

					billing := ResourceBilling{
						Name: fmt.Sprintf("%s/%s", msvc.Namespace, msvc.Name),
						Items: []k8sItem{
							newK8sItem(msvc, Pod, realMsvc.Spec.ReplicaCount),
							newK8sItem(msvc, Pvc, realMsvc.Spec.Storage.ToInt()),
						},
					}
					return r.SendBillingEvent(ctx, msvc, billing)
				}

			case mysqlStandalone.GroupVersion.WithKind("Service"):
				{
					realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), &mongodbStandalone.Service{})
					if err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}

					billing := ResourceBilling{
						Name: fmt.Sprintf("%s/%s", msvc.Namespace, msvc.Name),
						Items: []k8sItem{
							newK8sItem(msvc, Pod, realMsvc.Spec.ReplicaCount),
							newK8sItem(msvc, Pvc, realMsvc.Spec.Storage.ToInt()),
						},
					}
					return r.SendBillingEvent(ctx, msvc, billing)
				}
			}
		}

	case serverlessv1.GroupVersion.WithKind("Lambda"):
		{
			lambda, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &serverlessv1.Lambda{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: labels.SelectorFromValidatedSet(
						map[string]string{"kloudlite.io/lambda.name": lambda.Name},
					),
					Namespace: lambda.Namespace,
				},
			); err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			billing := ResourceBilling{
				Name: fmt.Sprintf("%s/%s", lambda.Namespace, lambda.Name),
				Items: []k8sItem{
					newK8sItem(lambda, Pod, len(podsList.Items)),
				},
			}
			return r.SendBillingEvent(ctx, lambda, billing)
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
	r.logger = logging.NewOrDie(
		&logging.Options{
			Name: "billing-watcher",
			Dev:  true,
		},
	)

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&crdsv1.App{})

	watchList := []client.Object{
		// &crdsv1.Project{},
		&crdsv1.App{},
		&serverlessv1.Lambda{},
		&crdsv1.ManagedService{},
		// &crdsv1.ManagedResource{},
		// &crdsv1.Router{},
	}

	for _, object := range watchList {
		builder.Watches(
			&source.Kind{Type: object},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					b64Group := base64.StdEncoding.EncodeToString(
						[]byte(obj.GetAnnotations()[constants.AnnotationKeys.GroupVersionKind]),
					)

					if len(b64Group) == 0 {
						return nil
					}

					wName, err := WrappedName{Name: obj.GetName(), Group: b64Group}.String()
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
