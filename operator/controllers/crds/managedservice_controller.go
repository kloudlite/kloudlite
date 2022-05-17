package crds

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
	lib.MessageSender
	logger *zap.SugaredLogger
	msvc   *crdsv1.ManagedService
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/finalizers,verbs=update

const msvcFinalizer = "finalizers.kloudlite.io/managed-service"

func (r *ManagedServiceReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(&r.msvc.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrUnknown",
		Message: err.Error(),
	})

	if err := r.notify(); err != nil {
		return reconcileResult.FailedE(err)
	}
	if err := r.Status().Update(ctx, r.msvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.msvc.NameRef()))
	}
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) notify() error {
	err := r.SendMessage(r.msvc.NameRef(), lib.MessageReply{
		Key:        r.msvc.NameRef(),
		Conditions: r.msvc.Status.Conditions,
		Status:     meta.IsStatusConditionTrue(r.msvc.Status.Conditions, "Ready"),
	})
	if err != nil {
		return errors.NewEf(err, "could not send message into kafka")
	}
	return nil
}

func (r *ManagedServiceReconciler) notifyAndUpdate(ctx context.Context) (ctrl.Result, error) {
	if err := r.notify(); err != nil {
		return reconcileResult.FailedE(err)
	}
	if err := r.Status().Update(ctx, r.msvc); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)

	msvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, req.NamespacedName, msvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if !msvc.HasLabels() {
		msvc.EnsureLabels()
		if err := r.Update(ctx, msvc); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}

	r.msvc = msvc
	if msvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, msvc)
	}

	b, err := templates.Parse(templates.CommonMsvcService, &msvc)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}

	if err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply service %s:%s", msvc.APIVersion, msvc.Kind))
	}

	svcObj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": msvc.Spec.ApiVersion,
			"kind":       "Service",
		},
	}

	if err := r.Get(ctx, types.NamespacedName{Namespace: msvc.Namespace, Name: msvc.Name}, &svcObj); err != nil {
		return reconcileResult.FailedE(err)
	}

	mj, err := svcObj.MarshalJSON()

	if err != nil {
		return r.notifyAndDie(ctx, err)
	}
	var j struct {
		Status struct {
			Conditions t.Conditions `json:"conditions,omitempty"`
		} `json:"status,omitempty"`
	}
	if err := json.Unmarshal(mj, &j); err != nil {
		return r.notifyAndDie(ctx, err)
	}
	r.msvc.Status.Conditions = j.Status.Conditions.GetConditions()

	return r.notifyAndUpdate(ctx)
}

func (r *ManagedServiceReconciler) finalize(ctx context.Context, msvc *crdsv1.ManagedService) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(msvc, msvcFinalizer) {
		controllerutil.RemoveFinalizer(msvc, msvcFinalizer)
		if err := r.Update(ctx, msvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if controllerutil.ContainsFinalizer(msvc, finalizers.Foreground.String()) {
		resource := unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": msvc.Spec.ApiVersion,
				"kind":       "Service",
			},
		}

		if err := r.Get(ctx, types.NamespacedName{Namespace: msvc.Namespace, Name: msvc.Name}, &resource); err != nil {
			r.logger.Infof("ERR: %+v", err)
			if apiErrors.IsNotFound(err) {
				controllerutil.RemoveFinalizer(msvc, finalizers.Foreground.String())
				if err := r.Update(ctx, msvc); err != nil {
					return reconcileResult.FailedE(err)
				}
				return reconcileResult.OK()
			}
		}

		r.logger.Infof("resource: %+v", resource)
		if resource.GetName() != "" {
			if err := r.Delete(ctx, &resource); err != nil {
				return r.notifyAndDie(ctx, err)
			}
		}
	}

	// return reconcileResult.FailedE(errors.New("trying to finalize managed service"))
	return reconcileResult.Failed()
}

func (r *ManagedServiceReconciler) watcherFuncMap(c client.Object) []reconcile.Request {
	var msvcList crdsv1.ManagedServiceList

	key, _ := crdsv1.ManagedService{}.LabelRef()
	if err := r.List(context.TODO(), &msvcList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{key: c.GetObjectKind().GroupVersionKind().Group}),
	}); err != nil {
		return nil
	}
	var reqs []reconcile.Request
	for _, item := range msvcList.Items {
		nn := types.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}
		for _, req := range reqs {
			if req.NamespacedName.String() == nn.String() {
				return nil
			}
		}
		reqs = append(reqs, reconcile.Request{NamespacedName: nn})
	}
	return reqs
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedService{}).
		Watches(&source.Kind{
			Type: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": fmt.Sprintf("mongodb-standalone.%s", constants.MsvcApiVersion),
					"kind":       "Service",
				},
			},
		}, handler.EnqueueRequestsFromMapFunc(r.watcherFuncMap)).
		Watches(&source.Kind{
			Type: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": fmt.Sprintf("mongodb-cluster.%s", constants.MsvcApiVersion),
					"kind":       "Service",
				},
			},
		}, handler.EnqueueRequestsFromMapFunc(r.watcherFuncMap)).
		Watches(&source.Kind{
			Type: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": fmt.Sprintf("mysql-standalone.%s", constants.MsvcApiVersion),
					"kind":       "Service",
				},
			},
		}, handler.EnqueueRequestsFromMapFunc(r.watcherFuncMap)).
		Complete(r)
}
