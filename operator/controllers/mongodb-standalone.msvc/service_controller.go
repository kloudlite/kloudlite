package mongodbstandalonemsvc

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

	mongoStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"
)

type Output struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	DbHosts      string `json:"HOSTS"`
	DbUrl        string `json:"DB_URL"`
}

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type serviceReq struct {
	ctrl.Request
	client.Client
	logger   *zap.SugaredLogger
	mongoSvc *mongoStandalone.Service
}

func (req *serviceReq) failWithErr(ctx context.Context, err error) (ctrl.Result, error) {
	if err != nil {
		req.mongoSvc.Status.Conditions.Build("", metav1.Condition{
			Type:    constants.ConditionReady.Type,
			Status:  metav1.ConditionFalse,
			Reason:  constants.ConditionReady.ErrorReason,
			Message: err.Error(),
		})
	}
	return req.updateStatus(ctx)
}

func (req *serviceReq) reconcileStatus(ctx context.Context) (*ctrl.Result, error) {
	prevStatus := req.mongoSvc.Status
	err := req.mongoSvc.Status.Conditions.FromHelmMsvc(ctx, req.Client, constants.HelmMongoDBKind, types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: req.mongoSvc.Name})
	if err != nil {
		return nil, nil
	}
	if err := req.mongoSvc.Status.Conditions.FromDeployment(ctx, req.Client, types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: req.mongoSvc.Name}); err != nil {
		if err := req.Status().Update(ctx, req.mongoSvc); err != nil {
			return nil, err
		}
		return &ctrl.Result{}, nil
	}

	if !cmp.Equal(prevStatus, req.mongoSvc.Status, cmpopts.IgnoreUnexported(t.Conditions{})) {
		req.logger.Infof("status is different, so updating status ...")
		req.mongoSvc.Status = prevStatus

		if err := req.Status().Update(ctx, req.mongoSvc); err != nil {
			return nil, err
		}
		return &ctrl.Result{}, nil
	}

	return nil, nil
}

func (req *serviceReq) reconcileOperations() error {
	b, err := templates.Parse(templates.MongoDBStandalone, req.mongoSvc)
	if err != nil {
		return err
	}

	if _, err := fn.KubectlApply(b); err != nil {
		return errors.NewEf(err, "could not apply kubectl for mongodb standalone")
	}

	return nil
}

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &serviceReq{Request: orgReq, Client: r.Client}
	req.logger = crds.GetLogger(req.NamespacedName)
	req.mongoSvc = new(mongoStandalone.Service)
	if err := r.Get(ctx, req.NamespacedName, req.mongoSvc); err != nil {
		req.logger.Infof("err: %v", err)
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
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

	if err := req.reconcileStatus(ctx); err != nil {
		return req.failWithErr(ctx, err)
	}

	req.logger.Infof("status is in sync, so proceeding with ops")

	// req.mongoSvc.Status.Conditions.Reset()

	var helmSecret corev1.Secret
	nn := types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: req.mongoSvc.Name}
	if err := r.Get(ctx, nn, &helmSecret); err != nil {
		req.logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", nn.String())
	}
	x, ok := helmSecret.Data["mongodb-root-password"]
	var m map[string]interface{}
	if err := json.Unmarshal(req.mongoSvc.Spec.Inputs, &m); err != nil {
		return reconcileResult.FailedE(err)
	}
	m["root_password"] = fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40))
	marshal, err := json.Marshal(m)
	if err != nil {
		return req.failWithErr(ctx, err)
	}

	req.mongoSvc.Spec.Inputs = marshal

	condition := meta.FindStatusCondition(req.mongoSvc.Status.Conditions.GetConditions(), constants.ConditionReady.Type)
	if condition == nil {
		req.mongoSvc.Status.Conditions.Build("", metav1.Condition{Type: constants.ConditionReady.Type, Reason: constants.ConditionReady.InProgressReason, Message: "reconcilation waiting up on child resources, to provide status updates"})
	}

	if err := req.reconcileOperations(); err != nil {
		return reconcileResult.OK()
	}

	hash, err := fn.Json.Hash(req.mongoSvc.FilterHashable())
	req.logger.Infof("\nhash: %s\n", hash)
	if err != nil {
		return req.failWithErr(ctx, err)
	}
	if hash == req.mongoSvc.Status.LastHash {
		return req.updateStatus(ctx)
	}

	req.mongoSvc.Status.LastHash = hash
	if err := r.buildOutput(ctx, req); err != nil {
		return req.failWithErr(ctx, err)
	}

	return req.updateStatus(ctx)
}

func (r *ServiceReconciler) buildOutput(ctx context.Context, req *serviceReq) error {
	m, err := req.mongoSvc.Spec.Inputs.MarshalJSON()
	if err != nil {
		return err
	}
	var j map[string]interface{}
	if err := json.Unmarshal(m, &j); err != nil {
		return err
	}
	hostUrl := fmt.Sprintf("%s.%s.svc.cluster.local", req.mongoSvc.Name, req.mongoSvc.Namespace)
	out := Output{
		RootPassword: j["root_password"].(string),
		DbHosts:      hostUrl,
		DbUrl:        fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", j["root_password"], hostUrl),
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", req.mongoSvc.Name),
			Namespace: req.mongoSvc.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, scrt, func() error {
		var outMap map[string]string
		if err := fn.Json.FromTo(out, &outMap); err != nil {
			return err
		}
		scrt.StringData = outMap
		return controllerutil.SetControllerReference(req.mongoSvc, scrt, r.Scheme)
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceReconciler) walk(ctx context.Context, req *serviceReq) error {
	if err := req.mongoSvc.Status.Conditions.FromHelmMsvc(ctx, r.Client, constants.HelmMongoDBKind, types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: req.mongoSvc.Name}); err != nil {
		return err
	}

	if err := req.mongoSvc.Status.Conditions.FromDeployment(ctx, r.Client, types.NamespacedName{Namespace: req.mongoSvc.Namespace, Name: req.mongoSvc.Name}); err != nil {
		return err
	}

	return nil
}

func (r *ServiceReconciler) finalize(ctx context.Context, req *serviceReq) (ctrl.Result, error) {
	req.logger.Infof("finalizing: %+v", req.mongoSvc.NameRef())
	if err := r.Delete(ctx, &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": constants.MsvcApiVersion,
		"kind":       constants.HelmMongoDBKind,
		"metadata": map[string]interface{}{
			"name":      req.mongoSvc.Name,
			"namespace": req.mongoSvc.Namespace,
		},
	}}); err != nil {
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
		Watches(&source.Kind{Type: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmMongoDBKind,
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(c client.Object) []reconcile.Request {
			var svcList mongoStandalone.ServiceList
			key, value := mongoStandalone.Service{}.LabelRef()
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
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Complete(r)
}
