package mysqlstandalonemsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

	mysqlStandalone "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type Output struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	DbHosts      string `json:"HOSTS"`
}

type ServiceReq struct {
	client.Client
	ctrl.Request
	mysqlSvc *mysqlStandalone.Service
	logger   *zap.SugaredLogger
}

func (req *ServiceReq) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	req.mysqlSvc.Status.Conditions.Build("", metav1.Condition{
		Type:    constants.ConditionReady.Type,
		Status:  metav1.ConditionFalse,
		Reason:  constants.ConditionReady.ErrorReason,
		Message: err.Error(),
	})

	return req.notify(ctx)
}

func (req *ServiceReq) notify(ctx context.Context) (ctrl.Result, error) {
	if err := req.Status().Update(ctx, req.mysqlSvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", req.mysqlSvc.NameRef()))
	}
	return reconcileResult.OK()
}

func (req *ServiceReq) reconcileStatus(ctx context.Context) (*ctrl.Result, error) {
	prevStatus := req.mysqlSvc.Status

	if err := req.mysqlSvc.Status.Conditions.BuildFromHelmMsvc(ctx, req.Client, constants.HelmMySqlDBKind, types.NamespacedName{Namespace: req.mysqlSvc.Namespace, Name: req.mysqlSvc.Name}); err != nil {
		return &ctrl.Result{}, errors.NewEf(err, "while building conditions from Helm Resource")
	}

	if err := req.mysqlSvc.Status.Conditions.BuildFromStatefulset(ctx, req.Client, types.NamespacedName{Namespace: req.mysqlSvc.Namespace, Name: req.mysqlSvc.Name}); err != nil {
		return &ctrl.Result{}, errors.NewEf(err, "while building conditions from statefulset resource")
	}

	if !cmp.Equal(prevStatus, req.mysqlSvc.Status, cmpopts.IgnoreUnexported(t.Conditions{})) {
		return &ctrl.Result{}, nil
	}

	return nil, nil
}

func (req *ServiceReq) updateStatus(ctx context.Context, ctrlResult ctrl.Result) (ctrl.Result, error) {
	if err := req.Status().Update(ctx, req.mysqlSvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", req.mysqlSvc.NameRef()))
	}
	return ctrlResult, nil
}

func (req *ServiceReq) reconcileOps(ctx context.Context) (*ctrl.Result, error) {
	var helmSecret corev1.Secret
	if err := req.Get(ctx, (types.NamespacedName{Namespace: req.mysqlSvc.Namespace, Name: req.mysqlSvc.Name}), &helmSecret); err != nil {
		req.logger.Error(err)
		req.logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", types.NamespacedName{Namespace: req.mysqlSvc.Namespace, Name: req.mysqlSvc.Name}.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(req.mysqlSvc.Spec.Inputs, &m); err != nil {
		return &ctrl.Result{}, errors.NewEf(err, "while unmarshalling inputs")
	}
	x, ok := helmSecret.Data["mysql-root-password"]
	m["root_password"] = fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40))
	y, ok := helmSecret.Data["mysql-password"]
	m["password"] = fn.IfThenElse(ok, string(y), fn.CleanerNanoid(40))
	marshal, err := json.Marshal(m)
	if err != nil {
		return &ctrl.Result{}, err
	}
	req.mysqlSvc.Spec.Inputs = marshal

	b, err := templates.Parse(templates.MySqlStandalone, req.mysqlSvc)
	if err != nil {
		return &ctrl.Result{}, err
	}
	if _, err := fn.KubectlApply(b); err != nil {
		return &ctrl.Result{}, err
	}

	return &ctrl.Result{}, nil
}

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &ServiceReq{
		Client:   r.Client,
		Request:  orgReq,
		logger:   crds.GetLogger(orgReq.NamespacedName),
		mysqlSvc: new(mysqlStandalone.Service),
	}

	if err := r.Get(ctx, req.NamespacedName, req.mysqlSvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if !req.mysqlSvc.HasLabels() {
		req.mysqlSvc.EnsureLabels()
		if err := req.Update(ctx, req.mysqlSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if req.mysqlSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req.mysqlSvc)
	}

	if ctrlResult, err := req.reconcileStatus(ctx); ctrlResult != nil {
		if err != nil {
			req.mysqlSvc.Status.Conditions.MarkNotReady(err)
		}
		if err := req.Status().Update(ctx, req.mysqlSvc); err != nil {
			return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", req.mysqlSvc.NameRef()))
		}
		return *ctrlResult, nil
	}

	if ctrlResult, err := req.reconcileOps(ctx); ctrlResult != nil {
		if err != nil {
			req.mysqlSvc.Status.Conditions.MarkNotReady(err)
		}
		if err := req.Status().Update(ctx, req.mysqlSvc); err != nil {
			return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", req.mysqlSvc.NameRef()))
		}
		return *ctrlResult, nil
	}

	if err := r.buildOutput(ctx, req); err != nil {
		return ctrl.Result{}, err
	}

	return reconcileResult.OK()
}

func (r *ServiceReconciler) buildOutput(ctx context.Context, req *ServiceReq) error {
	m, err := fn.Json.FromRawMessage(req.mysqlSvc.Spec.Inputs)
	if err != nil {
		return err
	}
	out := Output{
		RootPassword: m["root_password"].(string),
		DbHosts:      fmt.Sprintf("%s.%s.svc.cluster.local", req.mysqlSvc.Name, req.mysqlSvc.Namespace),
	}

	var outMap map[string]string
	if err := fn.Json.FromTo(out, &outMap); err != nil {
		return err
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", req.mysqlSvc.Name),
			Namespace: req.mysqlSvc.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, req.Client, scrt, func() error {
		scrt.StringData = outMap
		return controllerutil.SetControllerReference(req.mysqlSvc, scrt, r.Scheme)
	})
	if err != nil {
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
		Watches(&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Complete(r)
}
