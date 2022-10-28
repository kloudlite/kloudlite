package billing_watcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/operators/status-n-billing/internal/env"
	t "operators.kloudlite.io/operators/status-n-billing/internal/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler reconciles a BillingWatcher object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	*t.Notifier
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func (r *Reconciler) SendBillingEvent(ctx context.Context, obj client.Object, billing *t.ResourceBilling) (ctrl.Result,
	error) {
	klMetadata := t.ExtractMetadata(obj)
	if obj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(obj, constants.BillingFinalizer) {
			if err := r.NotifyBilling(ctx, t.GetMsgKey(obj), klMetadata, billing, t.Stages.Deleted); err != nil {
				return ctrl.Result{}, err
			}
		}
		return r.RemoveBillingFinalizer(ctx, obj)
	}

	if !controllerutil.ContainsFinalizer(obj, constants.BillingFinalizer) {
		return r.AddBillingFinalizer(ctx, obj)
	}

	if err := r.NotifyBilling(ctx, t.GetMsgKey(obj), klMetadata, billing, t.Stages.Exists); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=billingwatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=billingwatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watcher.kloudlite.io,resources=billingwatchers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	var wName t.WrappedName
	if err := json.Unmarshal([]byte(oReq.Name), &wName); err != nil {
		return ctrl.Result{}, nil
	}

	gvk, err := wName.ParseGroup()
	if err != nil {
		r.logger.Errorf(
			err, "badly formatted group-version-kind (%s) received, aborting ...", wName.Group,
		)
		return ctrl.Result{}, nil
	}

	if gvk == nil {
		return ctrl.Result{}, nil
	}

	logger := r.logger.WithName(fn.NN(oReq.Namespace, wName.Name).String()).WithKV("RefKind", gvk.String())
	logger.Infof("request received ...")

	switch *gvk {
	case crdsv1.GroupVersion.WithKind("App"):
		{
			app, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.App{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			var readyReplicas int
			if err := app.Status.DisplayVars.Get("readyReplicas", &readyReplicas); err != nil {
				return ctrl.Result{}, err
			}

			s, ok := app.GetAnnotations()[constants.AnnotationKeys.BillableQuantity]
			if !ok {
				logger.Infof("missing annotation key billable quantity: %v", constants.AnnotationKeys.BillableQuantity)
				return ctrl.Result{}, nil
			}
			billableQ, err := strconv.ParseFloat(s, 32)
			if err != nil {
				logger.Errorf(err, "could not convert annotation %s value into float64", constants.AnnotationKeys.BillableQuantity)
				return ctrl.Result{}, nil
			}

			billing := t.ResourceBilling{
				Name: fmt.Sprintf("%s/%s", app.Namespace, app.Name),
				Items: []t.K8sItem{
					t.NewK8sItem(app, t.Compute, float32(billableQ), readyReplicas),
				},
			}

			return r.SendBillingEvent(ctx, app, &billing)
		}

	case crdsv1.GroupVersion.WithKind("ManagedService"):
		{
			msvc, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &crdsv1.ManagedService{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			if msvc.GetDeletionTimestamp() != nil {
				return r.SendBillingEvent(ctx, msvc, nil)
			}

			realMsvcType := metav1.TypeMeta{APIVersion: msvc.Spec.MsvcKind.APIVersion, Kind: msvc.Spec.MsvcKind.Kind}

			switch realMsvcType.GetObjectKind().GroupVersionKind() {
			case mongodbMsvcv1.GroupVersion.WithKind("Service"):
				{
					realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), &mongodbMsvcv1.StandaloneService{})
					if err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}

					s, ok := msvc.GetAnnotations()[constants.AnnotationKeys.BillableQuantity]
					if !ok {
						logger.Infof(
							"missing annotation key billable quantity: %v",
							constants.AnnotationKeys.BillableQuantity,
						)
						return ctrl.Result{}, nil
					}
					billableQ, err := strconv.ParseFloat(s, 32)
					if err != nil {
						logger.Errorf(err, "could not convert annotation %s value into float64", constants.AnnotationKeys.BillableQuantity)
						return ctrl.Result{}, nil
					}

					billing := t.ResourceBilling{
						Name: fmt.Sprintf("%s/%s", msvc.Namespace, msvc.Name),
						Items: []t.K8sItem{
							t.NewK8sItem(msvc, t.Compute, float32(billableQ), realMsvc.Spec.ReplicaCount),
							t.NewK8sItem(msvc, t.BlockStorage, float32(realMsvc.Spec.Storage.ToInt()), realMsvc.Spec.ReplicaCount),
						},
					}
					return r.SendBillingEvent(ctx, msvc, &billing)
				}
			case redisStandalone.GroupVersion.WithKind("Service"):
				{
					realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), &mongodbStandalone.Service{})
					if err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}

					s, ok := realMsvc.GetAnnotations()[constants.AnnotationKeys.BillableQuantity]
					if !ok {
						logger.Infof("missing annotation key billable quantity: %v", constants.AnnotationKeys.BillableQuantity)
						return ctrl.Result{}, nil
					}
					billableQ, err := strconv.ParseFloat(s, 32)
					if err != nil {
						logger.Errorf(err, "could not convert annotation %s value into float64", constants.AnnotationKeys.BillableQuantity)
						return ctrl.Result{}, nil
					}

					billing := t.ResourceBilling{
						Name: fmt.Sprintf("%s/%s", msvc.Namespace, msvc.Name),
						Items: []t.K8sItem{
							t.NewK8sItem(msvc, t.Compute, float32(billableQ), realMsvc.Spec.ReplicaCount),
							t.NewK8sItem(msvc, t.BlockStorage, float32(realMsvc.Spec.Storage.ToInt()), realMsvc.Spec.ReplicaCount),
						},
					}
					return r.SendBillingEvent(ctx, msvc, &billing)
				}

			case mysqlStandalone.GroupVersion.WithKind("Service"):
				{
					realMsvc, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), &mongodbStandalone.Service{})
					if err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}
					s, ok := realMsvc.GetAnnotations()[constants.AnnotationKeys.BillableQuantity]
					if !ok {
						r.logger.Infof("missing annotation key billable quantity: %v", constants.AnnotationKeys.BillableQuantity)
						return ctrl.Result{}, nil
					}
					billableQ, err := strconv.ParseFloat(s, 32)
					if err != nil {
						logger.Errorf(err, "could not convert annotation %s value into float64", constants.AnnotationKeys.BillableQuantity)
						return ctrl.Result{}, nil
					}

					billing := t.ResourceBilling{
						Name: fmt.Sprintf("%s/%s", msvc.Namespace, msvc.Name),
						Items: []t.K8sItem{
							t.NewK8sItem(msvc, t.Compute, float32(billableQ), realMsvc.Spec.ReplicaCount),
							t.NewK8sItem(msvc, t.BlockStorage, float32(realMsvc.Spec.Storage.ToInt()), realMsvc.Spec.ReplicaCount),
						},
					}
					return r.SendBillingEvent(ctx, msvc, &billing)
				}
			}
		}

	case serverlessv1.GroupVersion.WithKind("Lambda"):
		{
			lambda, err := rApi.Get(ctx, r.Client, fn.NN(oReq.Namespace, wName.Name), &serverlessv1.Lambda{})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			s, ok := lambda.GetAnnotations()[constants.AnnotationKeys.BillableQuantity]
			if !ok {
				logger.Infof("missing annotation key billable quantity: %v", constants.AnnotationKeys.BillableQuantity)
				return ctrl.Result{}, nil
			}

			billableQ, err := strconv.ParseFloat(s, 32)
			if err != nil {
				logger.Errorf(err, "could not convert annotation %s value into float64", constants.AnnotationKeys.BillableQuantity)
				return ctrl.Result{}, nil
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

			billing := t.ResourceBilling{
				Name: fmt.Sprintf("%s/%s", lambda.Namespace, lambda.Name),
				Items: []t.K8sItem{
					t.NewK8sItem(lambda, t.Lambda, float32(billableQ), len(podsList.Items)),
				},
			}
			return r.SendBillingEvent(ctx, lambda, &billing)
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) AddBillingFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.AddFinalizer(obj, constants.BillingFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

func (r *Reconciler) RemoveBillingFinalizer(ctx context.Context, obj client.Object) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(obj, constants.BillingFinalizer)
	return ctrl.Result{}, r.Update(ctx, obj)
}

// SetupWithManager sets up the controllers with the Manager.

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&crdsv1.Project{})

	watchList := []client.Object{
		// &crdsv1.ProjectRef{},
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

					wName, err := t.WrappedName{Name: obj.GetName(), Group: b64Group}.String()
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
