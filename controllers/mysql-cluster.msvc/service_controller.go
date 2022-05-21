package mysqlclustermsvc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

	"encoding/json"

	mysqlCluster "operators.kloudlite.io/apis/mysql-cluster.msvc/v1"
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
	mysqlSvc *mysqlCluster.Service
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

type Output struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	DbHosts      string `json:"HOSTS"`
}

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = crds.GetLogger(req.NamespacedName)
	r.logger.Infof("reconciling common service %s", req.NamespacedName)

	r.mysqlSvc = &mysqlCluster.Service{}
	if err := r.Get(ctx, req.NamespacedName, r.mysqlSvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
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

	var helmSecret corev1.Secret
	if err := r.Get(ctx, (types.NamespacedName{Namespace: r.mysqlSvc.Namespace, Name: r.mysqlSvc.Name}), &helmSecret); err != nil {
		r.logger.Error(err)
		r.logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", types.NamespacedName{Namespace: r.mysqlSvc.Namespace, Name: r.mysqlSvc.Name}.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(r.mysqlSvc.Spec.Inputs, &m); err != nil {
		return reconcileResult.FailedE(err)
	}
	x, ok := helmSecret.Data["mysql-root-password"]
	m["root_password"] = fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40))
	y, ok := helmSecret.Data["mysql-password"]
	m["password"] = fn.IfThenElse(ok, string(y), fn.CleanerNanoid(40))
	marshal, err := json.Marshal(m)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}
	r.mysqlSvc.Spec.Inputs = marshal

	b, err := templates.Parse(templates.MySqlStandalone, r.mysqlSvc)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}
	if _, err := fn.KubectlApply(b); err != nil {
		return r.notifyAndDie(ctx, err)
	}

	if err := r.walk(ctx); err != nil {
		return r.notifyAndDie(ctx, err)
	}
	if err := r.buildOutput(ctx); err != nil {
		return r.notifyAndDie(ctx, err)
	}
	return r.notify(ctx)
}

func (r *ServiceReconciler) walk(ctx context.Context) error {
	if err := r.mysqlSvc.Status.Conditions.FromHelmMsvc(ctx, r.Client, constants.HelmMySqlDBKind, types.NamespacedName{Namespace: r.mysqlSvc.Namespace, Name: r.mysqlSvc.Name}); err != nil {
		return err
	}

	if err := r.mysqlSvc.Status.Conditions.FromStatefulset(ctx, r.Client, types.NamespacedName{Namespace: r.mysqlSvc.Namespace, Name: r.mysqlSvc.Name}); err != nil {
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *ServiceReconciler) buildOutput(ctx context.Context) error {
	m, err := fn.Json.FromRawMessage(r.mysqlSvc.Spec.Inputs)
	if err != nil {
		return err
	}
	out := Output{
		RootPassword: m["root_password"].(string),
		DbHosts:      fmt.Sprintf("%s.%s.svc.cluster.local", r.mysqlSvc.Name, r.mysqlSvc.Namespace),
	}

	var outMap map[string]string
	if err := fn.Json.FromTo(out, &outMap); err != nil {
		return err
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", r.mysqlSvc.Name),
			Namespace: r.mysqlSvc.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, scrt, func() error {
		scrt.StringData = outMap
		return controllerutil.SetControllerReference(r.mysqlSvc, scrt, r.Scheme)
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceReconciler) finalize(ctx context.Context, m *mysqlCluster.Service) (ctrl.Result, error) {
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

func (r *ServiceReconciler) kWatcherMap(o client.Object) []reconcile.Request {
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
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlCluster.Service{}).
		Watches(&source.Kind{Type: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmMySqlDBKind,
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			var svcList mysqlCluster.ServiceList
			key, value := mysqlCluster.Service{}.LabelRef()
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
		Watches(&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Complete(r)
}
