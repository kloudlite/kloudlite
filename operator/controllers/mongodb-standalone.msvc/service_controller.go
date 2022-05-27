package mongodbstandalonemsvc

import (
	"context"
	"fmt"
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
	"operators.kloudlite.io/lib/conditions"
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
		req.logger.Error(err)
		meta.SetStatusCondition(
			&req.mongoSvc.Status.Conditions, metav1.Condition{Type: "Ready",
				Status: "False", Reason: "StatusReconcliationFailed", Message: err.Error(),
			},
		)
		if err := r.Status().Update(ctx, req.mongoSvc); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.FailedE(err)
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
	isReady := true

	cs, err := conditions.FromResource(
		ctx,
		r.Client,
		constants.HelmMongoDBGroup,
		"Helm",
		fn.NamespacedName(req.mongoSvc),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		isReady = false
	}

	deploymentConditions, err := conditions.FromResource(
		ctx,
		r.Client,
		constants.DeploymentGroup,
		"Deployment",
		fn.NamespacedName(req.mongoSvc),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		isReady = false
	}

	isReady = meta.IsStatusConditionTrue(deploymentConditions, "Deployment-Available")

	cs = append(cs, deploymentConditions...)

	helmSecret := new(corev1.Secret)
	if err := r.Get(ctx, fn.NamespacedName(req.mongoSvc), helmSecret); err != nil {
		helmSecret = nil
		cs = append(
			cs, metav1.Condition{
				Type:    "HelmSecretAvailable",
				Status:  "False",
				Reason:  "HelmSecretNotFound",
				Message: "Helm secret not found, so assuming resource not yet installed, so installing",
			},
		)
		isReady = false
	}

	if _, ok := req.mongoSvc.Status.GeneratedVars.GetString(MongoDbRootPasswordKey); !ok {
		cs = append(
			cs, metav1.Condition{
				Type:    "GeneratedVars",
				Status:  "False",
				Reason:  "NotGeneratedYet",
				Message: "",
			},
		)
		isReady = false
	}

	outputSecret := new(corev1.Secret)
	if err := r.Get(
		ctx, types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: fmt.Sprintf(
			"msvc-%s",
			req.mongoSvc.Name,
		)}, outputSecret,
	); err != nil {
		isReady = false
	}

	//req.logger.Debugf("req.mongoSvc.Status: %+v", req.mongoSvc.Status)
	newConditions, updated, err := conditions.Patch(req.mongoSvc.Status.Conditions, cs)

	if !updated {
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	req.mongoSvc.Status.IsReady = isReady
	req.mongoSvc.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, req.mongoSvc); err != nil {
		req.logger.Debugf("err: %v", err)
		return nil, err
	}

	return reconcileResult.OKP()
}

func (r *ServiceReconciler) reconcileOperations(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	if meta.IsStatusConditionFalse(req.mongoSvc.Status.Conditions, "GeneratedVars") {
		if err := req.mongoSvc.Status.GeneratedVars.Merge(
			map[string]any{
				MongoDbRootPasswordKey: fn.CleanerNanoid(40),
				//StorageClassKey:        "do-block-storage-xfs",
				StorageClassKey: "local-path-xfs",
			},
		); err != nil {
			return ctrl.Result{}, err
		}
		req.logger.Debugf("req.mongoSvc.Status.GeneratedVars: %s", req.mongoSvc.Status.GeneratedVars)
		return ctrl.Result{}, r.Status().Update(ctx, req.mongoSvc)
	}

	//req.logger.Error("gVArs:", req.mongoSvc.Status.GeneratedVars)
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
	hostUrl := fmt.Sprintf("%s.%s.svc.cluster.local", req.mongoSvc.Name, req.mongoSvc.Namespace)
	authPasswd, ok := req.mongoSvc.Status.GeneratedVars.GetString(MongoDbRootPasswordKey)
	if !ok {
		return errors.Newf("could not find key(%s) in req.mongoSvc.Status.GeneratedVars", MongoDbRootPasswordKey)
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

	if err := fn.KubectlApply(ctx, r.Client, scrt); err != nil {
		return err
	}
	return nil
}

func (r *ServiceReconciler) finalize(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	req.logger.Infof("finalizing: %+v", req.mongoSvc.NameRef())
	controllerutil.RemoveFinalizer(req.mongoSvc, finalizers.MsvcCommonService.String())
	if err := r.Update(ctx, req.mongoSvc); err != nil {
		return reconcileResult.FailedE(err)
	}

	controllerutil.RemoveFinalizer(req.mongoSvc, finalizers.Foreground.String())
	return ctrl.Result{}, r.Update(ctx, req.mongoSvc)
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
