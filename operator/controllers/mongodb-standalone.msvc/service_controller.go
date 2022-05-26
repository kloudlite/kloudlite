package mongodbstandalonemsvc

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

	mongoStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
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
	Scheme *runtime.Scheme
}

type ServiceReconReq struct {
	req       ctrl.Request
	stateData map[string]string
	logger    *zap.SugaredLogger
	mongoSvc  *mongoStandalone.Service
}

const (
	MongoDbRootPasswordKey = "mongodb-root-password"
	StorageClassKey        = "storage-class"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &ServiceReconReq{
		req:       orgReq,
		logger:    crds.GetLogger(orgReq.NamespacedName),
		mongoSvc:  new(mongoStandalone.Service),
		stateData: map[string]string{},
	}

	if err := r.Get(ctx, orgReq.NamespacedName, req.mongoSvc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !req.mongoSvc.HasLabels() {
		req.mongoSvc.EnsureLabels()
		if err := r.Update(ctx, req.mongoSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if req.mongoSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req)
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

func (r *ServiceReconciler) failWithErr(ctx context.Context, req *ServiceReconReq, err error) (ctrl.Result, error) {
	fn.Conditions2.MarkNotReady(&req.mongoSvc.Status.OpsConditions, err, "ReconFailedWithErr")
	if err := r.Status().Update(ctx, req.mongoSvc); err != nil {
		return ctrl.Result{}, err
	}
	return reconcileResult.FailedE(err)
}

func (r *ServiceReconciler) reconcileStatus(ctx context.Context, req *ServiceReconReq) (*ctrl.Result, error) {
	var conditions []metav1.Condition

	err := fn.Conditions2.BuildFromHelmMsvc(
		&conditions,
		ctx,
		r.Client,
		constants.HelmMongoDBKind,
		fn.NamespacedName(req.mongoSvc),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	err = fn.Conditions2.BuildFromDeployment(
		&conditions,
		ctx,
		r.Client,
		fn.NamespacedName(req.mongoSvc),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	helmSecret := new(corev1.Secret)
	if err := r.Get(ctx, fn.NamespacedName(req.mongoSvc), helmSecret); err != nil {
		req.logger.Info(
			"helm release is not available yet, assuming resource not yet installed, so installing",
		)
		helmSecret = nil
	}

	if helmSecret != nil {
		req.stateData[MongoDbRootPasswordKey] = string(helmSecret.Data[MongoDbRootPasswordKey])
	}

	// check outputSecret exists
	outputSecret := new(corev1.Secret)
	if err := r.Get(
		ctx, types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: req.mongoSvc.Name}, outputSecret,
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}

		fn.Conditions2.MarkNotReady(
			&conditions, errors.NewEf(err, "outputSecret secret not found for resource"),
			"OutputSecretNotFound",
		)
		outputSecret = nil
	}

	if fn.Conditions2.Equal(conditions, req.mongoSvc.Status.Conditions) {
		req.logger.Infof("Status is already in sync, so moving forward with ops")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	req.mongoSvc.Status.Conditions = conditions
	if err := r.Status().Update(ctx, req.mongoSvc); err != nil {
		return nil, err
	}
	return reconcileResult.OKP()
}

func (r *ServiceReconciler) preOps(ctx context.Context, req *ServiceReconReq) error {
	if _, err := fn.JsonGet(req.mongoSvc.Status.GeneratedVars, MongoDbRootPasswordKey); err != nil {
		m, err := req.mongoSvc.Status.GeneratedVars.ToMap()
		if err != nil {
			return err
		}
		m[MongoDbRootPasswordKey] = fn.CleanerNanoid(40)
		m[StorageClassKey] = "do-block-storage-xfs"
		if err := req.mongoSvc.Status.GeneratedVars.FillFrom(m); err != nil {
			return err
		}
		return r.Status().Update(ctx, req.mongoSvc)
	}

	return nil
}

func (r *ServiceReconciler) reconcileOperations(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {

	if err := r.preOps(ctx, req); err != nil {
		return r.failWithErr(ctx, req, err)
	}

	// hash := req.mongoSvc.Hash()
	//
	// if hash == req.mongoSvc.Status.LastHash {
	// 	return reconcileResult.OK()
	// }

	b, err := templates.Parse(templates.MongoDBStandalone, req.mongoSvc)
	if err != nil {
		return reconcileResult.FailedE(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply kubectl for mongodb standalone"))
	}

	if err := r.reconcileOutput(ctx, req); err != nil {
		return r.failWithErr(ctx, req, err)
	}
	return reconcileResult.OK()
}

func (r *ServiceReconciler) reconcileOutput(ctx context.Context, req *ServiceReconReq) error {
	m, err := req.mongoSvc.Spec.Inputs.MarshalJSON()
	if err != nil {
		return err
	}
	var j map[string]interface{}
	if err := json.Unmarshal(m, &j); err != nil {
		return err
	}

	hostUrl := fmt.Sprintf("%s.%s.svc.cluster.local", req.mongoSvc.Name, req.mongoSvc.Namespace)
	authPasswd, err := fn.JsonGet[string](req.mongoSvc.Status.GeneratedVars, MongoDbRootPasswordKey)
	if err != nil {
		return err
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", req.mongoSvc.Name),
			Namespace: req.mongoSvc.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(req.mongoSvc, true),
			},
			Labels: req.mongoSvc.GetLabels(),
		},
		StringData: map[string]string{
			"ROOT_PASSWORD": authPasswd,
			"DB_HOSTS":      hostUrl,
			"DB_URL":        fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", authPasswd, hostUrl),
		},
	}

	err = fn.KubectlApply(ctx, r.Client, scrt)
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceReconciler) finalize(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	req.logger.Infof("finalizing: %+v", req.mongoSvc.NameRef())
	if err := r.Delete(
		ctx, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmMongoDBKind,
				"metadata": map[string]interface{}{
					"name":      req.mongoSvc.Name,
					"namespace": req.mongoSvc.Namespace,
				},
			},
		},
	); err != nil {
		req.logger.Infof("could not delete helm resource: %+v", err)
		if !apiErrors.IsNotFound(err) {
			return reconcileResult.FailedE(err)
		}
	}
	controllerutil.RemoveFinalizer(req.mongoSvc, finalizers.MsvcCommonService.String())
	if err := r.Update(ctx, req.mongoSvc); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
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
		For(&mongoStandalone.Service{}).
		Watches(
			&source.Kind{
				Type: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": constants.MsvcApiVersion,
						"kind":       constants.HelmMongoDBKind,
					},
				},
			}, handler.EnqueueRequestsFromMapFunc(
				func(c client.Object) []reconcile.Request {
					var svcList mongoStandalone.ServiceList
					key, value := mongoStandalone.Service{}.LabelRef()
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
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Complete(r)
}
