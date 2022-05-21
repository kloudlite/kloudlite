package mongodbclustermsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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

	mongoCluster "operators.kloudlite.io/apis/mongodb-cluster.msvc/v1"
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
	mongoSvc *mongoCluster.Service
}

type Output struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	DbHosts      string `json:"HOSTS"`
	DbUrl        string `json:"DB_URL"`
}

// +kubebuilder:rbac:groups=mongodb-cluster.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-cluster.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-cluster.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = crds.GetLogger(req.NamespacedName)
	r.logger.Infof("reconciling common service %s", req.NamespacedName)
	r.mongoSvc = new(mongoCluster.Service)
	if err := r.Get(ctx, req.NamespacedName, r.mongoSvc); err != nil {
		r.logger.Infof("err: %v", err)
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if !r.mongoSvc.HasLabels() {
		r.mongoSvc.EnsureLabels()
		if err := r.Update(ctx, r.mongoSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if r.mongoSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, r.mongoSvc)
	}

	var helmSecret corev1.Secret
	nn := types.NamespacedName{Namespace: r.mongoSvc.Namespace, Name: r.mongoSvc.Name}
	if err := r.Get(ctx, nn, &helmSecret); err != nil {
		r.logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", nn.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(r.mongoSvc.Spec.Inputs, &m); err != nil {
		return reconcileResult.FailedE(err)
	}
	x, ok := helmSecret.Data["mongodb-root-password"]
	m["root_password"] = fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40))
	y, ok := helmSecret.Data["mongodb-replica-set-key"]
	m["replica_set_key"] = fn.IfThenElse(ok, string(y), fn.CleanerNanoid(40))

	marshal, err := json.Marshal(m)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}
	r.mongoSvc.Spec.Inputs = marshal

	b, err := templates.Parse(templates.MongoDBCluster, r.mongoSvc)
	if err != nil {
		return ctrl.Result{}, err
	}

	if _, err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply kubectl for mongodb cluster"))
	}

	if err := r.walk(ctx); err != nil {
		r.logger.Error("###############################", err)
		return r.notifyAndDie(ctx, err)
	}
	if err := r.buildOutput(ctx); err != nil {
		return r.notifyAndDie(ctx, err)
	}
	return r.notify(ctx)
}

func (r *ServiceReconciler) finalize(ctx context.Context, m *mongoCluster.Service) (ctrl.Result, error) {
	r.logger.Infof("finalizing: %+v", m.NameRef())
	if err := r.Delete(ctx, &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": constants.MsvcApiVersion,
		"kind":       constants.HelmMongoDBKind,
		"metadata": map[string]interface{}{
			"name":      m.Name,
			"namespace": m.Namespace,
		},
	}}); err != nil {
		r.logger.Infof("could not delete helm resource: %+v", err)
		if !apiErrors.IsNotFound(err) {
			return reconcileResult.FailedE(err)
		}
	}
	controllerutil.RemoveFinalizer(m, finalizers.MsvcCommonService.String())
	if err := r.Update(ctx, m); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

func (r *ServiceReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	r.mongoSvc.Status.Conditions.Build("", metav1.Condition{
		Type:    constants.ConditionReady.Type,
		Status:  metav1.ConditionFalse,
		Reason:  constants.ConditionReady.ErrorReason,
		Message: err.Error(),
	})

	return r.notify(ctx)
}

func (r *ServiceReconciler) notify(ctx context.Context) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, r.mongoSvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.mongoSvc.NameRef()))
	}
	return reconcileResult.OK()
}

func (r *ServiceReconciler) walk(ctx context.Context) error {
	if err := r.mongoSvc.Status.Conditions.BuildFromHelmMsvc(ctx, r.Client, constants.HelmMongoDBKind, types.NamespacedName{Namespace: r.mongoSvc.Namespace, Name: r.mongoSvc.Name}); err != nil {
		return err
	}

	if err := r.mongoSvc.Status.Conditions.BuildFromStatefulset(ctx, r.Client, types.NamespacedName{Namespace: r.mongoSvc.Namespace, Name: r.mongoSvc.Name}); err != nil {
		return err
	}

	readyCond2 := meta.FindStatusCondition(r.mongoSvc.Status.Conditions.GetConditions(), constants.ConditionReady.Type)
	r.logger.Infof("POST: readyCond: %+v\n", readyCond2)
	return nil
}

func (r *ServiceReconciler) buildOutput(ctx context.Context) error {
	m, err := fn.Json.FromRawMessage(r.mongoSvc.Spec.Inputs)
	if err != nil {
		return err
	}
	replicaCount := m["replica_count"].(int)
	var hosts []string
	for i := 0; i < replicaCount; i++ {
		hosts = append(hosts, fmt.Sprintf("%s.%s.svc.cluster.local", r.mongoSvc.Name, r.mongoSvc.Namespace))
	}
	hosts = append(hosts, fmt.Sprintf("%s-arbiter.%s.svc.cluster.local", r.mongoSvc.Name, r.mongoSvc.Namespace))

	hostUrl := strings.Join(hosts, ",")
	out := Output{
		RootPassword: m["root_password"].(string),
		DbHosts:      hostUrl,
		DbUrl:        fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", m["root_password"], hostUrl),
	}

	var outMap map[string]string
	if err := fn.Json.FromTo(out, outMap); err != nil {
		return err
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", r.mongoSvc.Name),
			Namespace: r.mongoSvc.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, scrt, func() error {
		scrt.StringData = outMap
		return controllerutil.SetControllerReference(r.mongoSvc, scrt, r.Scheme)
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceReconciler) kWatcherMap(o client.Object) []reconcile.Request {
	labels := o.GetLabels()
	if s := labels["app.kubernetes.io/component"]; s != "mongodb" {
		return nil
	}
	if s := labels["app.kubernetes.io/name"]; s != "mongodb" {
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
		For(&mongoCluster.Service{}).
		Watches(&source.Kind{Type: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmMongoDBKind,
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(c client.Object) []reconcile.Request {
			var svcList mongoCluster.ServiceList
			key, value := mongoCluster.Service{}.LabelRef()
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
