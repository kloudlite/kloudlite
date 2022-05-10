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
	labels2 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	mongoStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
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
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrWhileReconciling",
		Message: err.Error(),
	})

	if err := r.Status().Update(ctx, r.mongoSvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.mongoSvc.NameRef()))
	}
	return reconcileResult.OK()
}

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var mongoSvc mongoStandalone.Service

	if err := r.Get(ctx, req.NamespacedName, &mongoSvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	r.mongoSvc = &mongoSvc

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

	return reconcileResult.OK()
}

type HelmMongoDB struct {
	Spec   map[string]interface{} `json:"spec,omitempty"`
	Status struct {
		Conditions []metav1.Condition `json:"conditions,omitempty"`
	} `json:"status"`
}

func (r *ServiceReconciler) getHelmMongoDB() (*HelmMongoDB, error) {
	b, err := fn.KubectlGet(r.mongoSvc.Namespace, fmt.Sprintf("helmmongodb.msvc.kloudlite.io/%s", r.mongoSvc.Name))
	if err != nil {
		return nil, err
	}
	var hmdb HelmMongoDB
	if err := json.Unmarshal(b, &hmdb); err != nil {
		return nil, err
	}
	return &hmdb, nil
}

func (r *ServiceReconciler) walk(ctx context.Context) error {
	hmdb, err := r.getHelmMongoDB()
	if err != nil {
		return err
	}
	r.mongoSvc.Status.Conditions.Build("helm", hmdb.Status.Conditions...)
	if !meta.IsStatusConditionTrue(hmdb.Status.Conditions, "Deployed") {
		return nil
	}

	// ASSERT: helm mongodb deployment is standalone
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
	r.logger.Infof("Hello available: %+v", meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)))
	if !meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)) {
		opts := &client.ListOptions{
			LabelSelector: labels2.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
			Namespace:     depl.Namespace,
		}
		r.logger.Infof("list.options: %+v\n", opts)
		var podsList corev1.PodList
		if err := r.List(ctx, &podsList, opts); err != nil {
			return errors.NewEf(err, "could not list pods for deployment")
		}

		for idx, pod := range podsList.Items {
			var podC []metav1.Condition
			for _, condition := range pod.Status.Conditions {
				podC = append(podC, metav1.Condition{
					Type:               string(condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Reason:             "NotSpecified",
					Message:            condition.Message,
				})
			}
			r.mongoSvc.Status.Conditions.Build(fmt.Sprintf("Pod%d", idx), podC...)
			var containerC []metav1.Condition
			for _, cs := range pod.Status.ContainerStatuses {
				p := metav1.Condition{
					Type:   fmt.Sprintf("Name%s", cs.Name),
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
			r.buildConditions("Container", containerC...)
			return nil
		}
		return nil
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongoStandalone.Service{}).
		Watches(&source.Kind{Type: &crdsv1.ManagedService{}},
			handler.EnqueueRequestsFromMapFunc(
				func(c client.Object) []reconcile.Request {
					if c.GetLabels()["msvc.kloudlite.io/type"] != "MongoDBStandalone" {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: types.NamespacedName{Namespace: c.GetNamespace(), Name: c.GetName()}},
					}
				},
			),
		).
		Complete(r)
}
