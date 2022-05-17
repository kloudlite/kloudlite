package mongodbstandalonemsvc

import (
	"context"
	"encoding/json"
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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mongoStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
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
	logger *zap.SugaredLogger
	lib.MessageSender
	mongoSvc *mongoStandalone.Service
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

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = crds.GetLogger(req.NamespacedName)
	r.logger.Infof("reconciling common service %s", req.NamespacedName)
	var mongoSvc mongoStandalone.Service

	if err := r.Get(ctx, req.NamespacedName, &mongoSvc); err != nil {
		r.logger.Infof("err: %v", err)
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	r.mongoSvc = &mongoSvc
	r.mongoSvc.Status.Conditions.Reset()

	if !mongoSvc.HasLabels() {
		mongoSvc.EnsureLabels()
		if err := r.Update(ctx, &mongoSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if mongoSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, &mongoSvc)
	}

	var helmSecret corev1.Secret
	nn := types.NamespacedName{Namespace: mongoSvc.Namespace, Name: fmt.Sprintf("%s-mongodb", mongoSvc.Name)}
	if err := r.Get(ctx, nn, &helmSecret); err != nil {
		r.logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", nn.String())
	}
	x, ok := helmSecret.Data["mongodb-root-password"]
	var m map[string]interface{}
	if err := json.Unmarshal(mongoSvc.Spec.Inputs, &m); err != nil {
		return reconcileResult.FailedE(err)
	}
	m["root_password"] = fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40))
	marshal, err := json.Marshal(m)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}
	mongoSvc.Spec.Inputs = marshal

	b, err := templates.Parse(templates.MongoDBStandalone, mongoSvc)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply kubectl for mongodb standalone"))
	}

	if err := r.walk(ctx); err != nil {
		return r.notifyAndDie(ctx, err)
	}
	return r.notify(ctx)
}

func (r *ServiceReconciler) walk(ctx context.Context) error {
	if err := r.mongoSvc.Status.Conditions.FromHelmMsvc(ctx, r.Client, constants.HelmMongoDBKind, types.NamespacedName{Namespace: r.mongoSvc.Namespace, Name: r.mongoSvc.Name}); err != nil {
		return err
	}

	if err := r.mongoSvc.Status.Conditions.FromDeployment(ctx, r.Client, types.NamespacedName{Namespace: r.mongoSvc.Namespace, Name: r.mongoSvc.Name}); err != nil {
		return err
	}

	return nil
	// // ASSERT: helm mongodb deployment is standalone
	// return r.walkDeployment(ctx)
}

func (r *ServiceReconciler) walkDeployment(ctx context.Context) error {
	var depl appsv1.Deployment
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: r.mongoSvc.Namespace, Name: fmt.Sprintf("%s-mongodb", r.mongoSvc.Name)}, &depl); err != nil {
		return errors.NewEf(err, "could not get deployment for helm resource")
	}

	var deplConditions []metav1.Condition
	for _, cond := range depl.Status.Conditions {
		deplConditions = append(deplConditions, metav1.Condition{
			Type:    string(cond.Type),
			Status:  metav1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	r.mongoSvc.Status.Conditions.Build("Deployment", deplConditions...)
	if !meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)) {
		opts := &client.ListOptions{
			LabelSelector: apiLabels.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
			Namespace:     depl.Namespace,
		}
		var podsList corev1.PodList
		if err := r.List(ctx, &podsList, opts); err != nil {
			return errors.NewEf(err, "could not list pods for deployment")
		}

		for idx, pod := range podsList.Items {
			var podC []metav1.Condition
			for _, condition := range pod.Status.Conditions {
				podC = append(podC, metav1.Condition{
					Type:               fmt.Sprintf("Pod-idx-%d-%s", idx, condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Reason:             fmt.Sprintf("Pod:Idx:%d:NotSpecified", idx),
					Message:            condition.Message,
				})
			}
			r.mongoSvc.Status.Conditions.Build("", podC...)
			var containerC []metav1.Condition
			for _, cs := range pod.Status.ContainerStatuses {
				p := metav1.Condition{
					Type:   fmt.Sprintf("Name-%s", cs.Name),
					Status: fn.StatusFromBool(cs.Ready),
				}
				if cs.State.Waiting != nil {
					p.Reason = cs.State.Waiting.Reason
					p.Message = cs.State.Waiting.Message
				}
				if cs.State.Running != nil {
					p.Reason = "Running"
					p.Message = fmt.Sprintf("Container running since %s", cs.State.Running.StartedAt.String())
				}
				containerC = append(containerC, p)
			}
			r.mongoSvc.Status.Conditions.Build("Container", containerC...)
			return nil
		}
		return nil
	}
	r.mongoSvc.Status.Conditions.Build("", metav1.Condition{
		Type:    constants.ConditionReady.Type,
		Status:  metav1.ConditionTrue,
		Reason:  constants.ConditionReady.SuccessReason,
		Message: "Deployment is Available",
	})
	return nil
}

func (r *ServiceReconciler) finalize(ctx context.Context, m *mongoStandalone.Service) (ctrl.Result, error) {
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
		Watches(&source.Kind{
			Type: &appsv1.Deployment{},
		}, handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			labels := o.GetLabels()

			if s := labels["app.kubernetes.io/component"]; s != "mongodb" {
				return nil
			}
			if s := labels["app.kubernetes.io/name"]; s != "mongodb" {
				return nil
			}
			resourceName := labels["app.kubernetes.io/instance"]
			nn := types.NamespacedName{Namespace: o.GetNamespace(), Name: resourceName}
			return []reconcile.Request{{NamespacedName: nn}}
		})).
		Complete(r)
}
