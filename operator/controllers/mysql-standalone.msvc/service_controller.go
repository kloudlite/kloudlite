package mysqlstandalonemsvc

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mysqlStandalone "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	logger   *zap.SugaredLogger
	mysqlSvc *mysqlStandalone.Service
}

func (r *ServiceReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	r.mysqlSvc.Status.Conditions.Build("", metav1.Condition{
		Type:    constants.ConditionReady.Type,
		Status:  metav1.ConditionFalse,
		Reason:  constants.ConditionReady.ErrorReason,
		Message: err.Error(),
	})

	return r.notify(ctx)
}

func (r *ServiceReconciler) notify(ctx context.Context) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, r.mysqlSvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.mysqlSvc.NameRef()))
	}
	return reconcileResult.OK()
}

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = crds.GetLogger(req.NamespacedName)
	r.logger.Infof("reconciling common service %s", req.NamespacedName)

	r.mysqlSvc = &mysqlStandalone.Service{}
	if err := r.Get(ctx, req.NamespacedName, r.mysqlSvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	r.logger.Infof("name: %v", r.mysqlSvc.Name)
	r.mysqlSvc.Status.Conditions.Reset()
	if !r.mysqlSvc.HasLabels() {
		r.mysqlSvc.EnsureLabels()
		if err := r.Update(ctx, r.mysqlSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if r.mysqlSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, r.mysqlSvc)
	}

	b, err := templates.Parse(templates.MySqlStandalone, r.mysqlSvc)
	r.logger.Infof("parsed template: %s", b)
	if err != nil {
		return reconcileResult.FailedE(err)
	}
	if err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(err)
	}

	if err := r.walk(ctx); err != nil {
		return r.notifyAndDie(ctx, err)
	}
	return r.notify(ctx)
}

func (r *ServiceReconciler) walk(ctx context.Context) error {
	if err := r.mysqlSvc.Status.Conditions.FromHelmMsvc(ctx, r.Client, constants.HelmMySqlDBKind, types.NamespacedName{Namespace: r.mysqlSvc.Namespace, Name: r.mysqlSvc.Name}); err != nil {
		return err
	}

	if err := r.mysqlSvc.Status.Conditions.FromStatefulset(ctx, r.Client, types.NamespacedName{Namespace: r.mysqlSvc.Namespace, Name: r.mysqlSvc.Name}); err != nil {
		return err
	}

	return nil
}

func (r *ServiceReconciler) finalize(ctx context.Context, m *mysqlStandalone.Service) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(m, finalizers.MsvcCommonService.String()) {
		if err := r.Delete(ctx, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmMySqlDBKind,
				"metadata": map[string]interface{}{
					"name":      m.Name,
					"namespace": m.Namespace,
				},
			},
		}); err != nil {
			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(m, finalizers.MsvcCommonService.String())
		if err := r.Update(ctx, m); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlStandalone.Service{}).
		Watches(&source.Kind{Type: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmMySqlDBKind,
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			var svcList mysqlStandalone.ServiceList
			key, value := mysqlStandalone.Service{}.LabelRef()
			if err := r.List(context.TODO(), &svcList, &client.ListOptions{
				LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{key: value}),
			}); err != nil {
				return nil
			}
			var reqs []reconcile.Request
			for _, item := range svcList.Items {
				nn := types.NamespacedName{
					Name:      item.Name,
					Namespace: item.Namespace,
				}

				for _, req := range reqs {
					if req.NamespacedName.String() == nn.String() {
						return nil
					}
				}
				reqs = append(reqs, reconcile.Request{NamespacedName: nn})
			}
			return reqs
		})).
		Watches(&source.Kind{
			Type: &appsv1.StatefulSet{},
		}, handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			labels := o.GetLabels()

			if s := labels["app.kubernetes.io/component"]; s != "primary" {
				return nil
			}
			if s := labels["app.kubernetes.io/name"]; s != "mysql" {
				return nil
			}
			resourceName := labels["app.kubernetes.io/instance"]
			nn := types.NamespacedName{Namespace: o.GetNamespace(), Name: resourceName}

			return []reconcile.Request{
				{NamespacedName: nn},
			}
		})).
		Complete(r)
}
