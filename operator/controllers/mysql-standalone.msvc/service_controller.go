package mysqlstandalonemsvc

import (
	"context"
	"encoding/json"
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
	lt     metav1.Time
}

type Output struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	DbHosts      string `json:"HOSTS"`
}

const (
	MysqlRootPasswordKey = "mysql-root-password"
	MysqlPasswordKey     = "mysql-password"
)

type ServiceReconReq struct {
	t.ReconReq
	ctrl.Request
	logger      *zap.SugaredLogger
	condBuilder fn.StatusConditions
	mysqlSvc    *mysqlStandalone.Service
}

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &ServiceReconReq{
		Request:  orgReq,
		logger:   crds.GetLogger(orgReq.NamespacedName),
		mysqlSvc: new(mysqlStandalone.Service),
	}

	if err := r.Get(ctx, orgReq.NamespacedName, req.mysqlSvc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.condBuilder = fn.Conditions.From(req.mysqlSvc.Status.Conditions)

	if !req.mysqlSvc.HasLabels() {
		req.mysqlSvc.EnsureLabels()
		if err := r.Update(ctx, req.mysqlSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if req.mysqlSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req.mysqlSvc)
	}

	reconResult, err := r.reconcileStatus(ctx, req)
	if err != nil {
		return r.failWithErr(ctx, req, err)
	}
	if reconResult != nil {
		return *reconResult, nil
	}

	req.logger.Infof("status is in sync, so proceeding with ops")
	return r.reconcileOperations(ctx, req)
}

func (r *ServiceReconciler) finalize(ctx context.Context, m *mysqlStandalone.Service) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(m, finalizers.MsvcCommonService.String()) {
		if err := r.Delete(
			ctx, &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": constants.MsvcApiVersion,
					"kind":       constants.HelmMySqlDBKind,
					"metadata": map[string]interface{}{
						"name":      m.Name,
						"namespace": m.Namespace,
					},
				},
			},
		); err != nil {
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

func (r *ServiceReconciler) statusUpdate(ctx context.Context, req *ServiceReconReq) error {
	req.mysqlSvc.Status.Conditions = req.condBuilder.GetAll()
	return r.Status().Update(ctx, req.mysqlSvc)
}

func (r *ServiceReconciler) failWithErr(ctx context.Context, req *ServiceReconReq, err error) (ctrl.Result, error) {
	req.condBuilder.MarkNotReady(err)
	if err2 := r.statusUpdate(ctx, req); err2 != nil {
		return ctrl.Result{}, err2
	}
	return reconcileResult.FailedE(err)
}

func (r *ServiceReconciler) reconcileStatus(ctx context.Context, req *ServiceReconReq) (*ctrl.Result, error) {
	prevStatus := req.mysqlSvc.Status
	req.condBuilder.Reset()

	err := req.condBuilder.BuildFromHelmMsvc(
		ctx,
		r.Client,
		constants.HelmMySqlDBKind,
		types.NamespacedName{
			Namespace: req.mysqlSvc.GetNamespace(),
			Name:      req.mysqlSvc.GetName(),
		},
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	err = req.condBuilder.BuildFromStatefulset(
		ctx,
		r.Client,
		types.NamespacedName{
			Namespace: req.mysqlSvc.GetNamespace(),
			Name:      req.mysqlSvc.GetName(),
		},
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	var helmSecret corev1.Secret
	nn := types.NamespacedName{
		Namespace: req.mysqlSvc.GetNamespace(),
		Name:      req.mysqlSvc.GetName(),
	}
	if err := r.Get(ctx, nn, &helmSecret); err != nil {
		req.logger.Info(
			"helm release %s is not available yet, assuming resource not yet installed, so installing",
			nn.String(),
		)
	}
	x, ok := helmSecret.Data[MysqlRootPasswordKey]
	req.SetStateData(MysqlRootPasswordKey, fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40)))
	y, ok := helmSecret.Data[MysqlPasswordKey]
	req.SetStateData(MysqlPasswordKey, fn.IfThenElse(ok, string(y), fn.CleanerNanoid(40)))

	if req.condBuilder.Equal(prevStatus.Conditions) {
		req.logger.Infof("Status is already in sync, so moving forward with ops")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	if err := r.statusUpdate(ctx, req); err != nil {
		return nil, err
	}

	return reconcileResult.OKP()
}

func (r *ServiceReconciler) reconcileOperations(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(req.mysqlSvc.Spec.Inputs, &m); err != nil {
		return reconcileResult.FailedE(err)
	}
	m[MysqlRootPasswordKey] = req.GetStateData(MysqlRootPasswordKey)
	m[MysqlPasswordKey] = req.GetStateData(MysqlPasswordKey)
	marshal, err := json.Marshal(m)
	if err != nil {
		return reconcileResult.FailedE(err)
	}

	req.mysqlSvc.Spec.Inputs = marshal

	hash := req.mysqlSvc.Hash()

	if hash == req.mysqlSvc.Status.LastHash {
		return reconcileResult.OK()
	}

	b, err := templates.Parse(templates.MySqlStandalone, req.mysqlSvc)
	if err != nil {
		return reconcileResult.FailedE(err)
	}

	if _, err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply kubectl for mysql standalone"))
	}

	if err := r.reconcileOutput(ctx, req); err != nil {
		return reconcileResult.FailedE(err)
	}

	req.mysqlSvc.Status.LastHash = hash

	if err := r.statusUpdate(ctx, req); err != nil {
		return ctrl.Result{}, err
	}

	return reconcileResult.OK()
}

func (r *ServiceReconciler) reconcileOutput(ctx context.Context, req *ServiceReconReq) error {
	m, err := req.mysqlSvc.Spec.Inputs.MarshalJSON()
	if err != nil {
		return err
	}
	var j map[string]interface{}
	if err := json.Unmarshal(m, &j); err != nil {
		return err
	}
	hostUrl := fmt.Sprintf("%s.%s.svc.cluster.local", req.mysqlSvc.Name, req.mysqlSvc.Namespace)
	out := Output{
		RootPassword: req.GetStateData(MysqlRootPasswordKey),
		DbHosts:      hostUrl,
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", req.mysqlSvc.Name),
			Namespace: req.mysqlSvc.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(
		ctx, r.Client, scrt, func() error {

			outMap := map[string]string{
				MysqlRootPasswordKey: out.RootPassword,
				MysqlPasswordKey:     out.DbHosts,
			}

			if err := fn.Json.FromTo(out, &outMap); err != nil {
				return err
			}
			scrt.StringData = outMap
			return controllerutil.SetControllerReference(req.mysqlSvc, scrt, r.Scheme)
		},
	); err != nil {
		return err
	}
	return nil
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
		Watches(
			&source.Kind{
				Type: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": constants.MsvcApiVersion,
						"kind":       constants.HelmMySqlDBKind,
					},
				},
			}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					var svcList mysqlStandalone.ServiceList
					key, value := mysqlStandalone.Service{}.LabelRef()
					if err := r.List(
						context.TODO(), &svcList, &client.ListOptions{
							LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{key: value}),
						},
					); err != nil {
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
				},
			),
		).
		Watches(&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Complete(r)
}
