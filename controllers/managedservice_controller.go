package controllers

import (
	"context"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	msvcv1 "operators.kloudlite.io/apis/msvc/v1"
	watcherMsvc "operators.kloudlite.io/apis/watchers.msvc/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.msvc.LogRef()))
	}
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) notify() error {
	// fmt.Printf("Notify conditions: %+v", r.msvc.Status.Conditions)
	err := r.SendMessage(r.msvc.LogRef(), lib.MessageReply{
		Key:        r.msvc.LogRef(),
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
	logger := r.logger.With("RECONCILE", true)

	msvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, req.NamespacedName, msvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	r.msvc = msvc
	if msvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, msvc)
	}

	if msvc.Spec.Type != "MongoDBStandalone" {
		return reconcileResult.Failed()
	}

	mWatcher := &watcherMsvc.MongoDB{}
	if err := r.Get(ctx, req.NamespacedName, mWatcher); err != nil {
	}
	if mWatcher.Name != "" {
		msvc.Status.Conditions = mWatcher.Status.Conditions
		return r.notifyAndUpdate(ctx)
	}

	b, err := templates.Parse(templates.MongoDBStandalone, msvc)
	if err != nil {
		r.logger.Info(err)
		return r.notifyAndDie(ctx, err)
	}
	watcher, err := templates.Parse(templates.MongoDBWatcher, msvc)
	if err != nil {
		r.logger.Info(err)
		return r.notifyAndDie(ctx, err)
	}
	if err := fn.KubectlApply(b, watcher); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not apply managed service"))
	}
	logger.Infof("applied managed service")
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
		var mdb msvcv1.MongoDB
		if err := r.Get(ctx, types.NamespacedName{Namespace: msvc.Namespace, Name: "MongoDB"}, &mdb); err != nil {
			if apiErrors.IsNotFound(err) {
				controllerutil.RemoveFinalizer(msvc, finalizers.Foreground.String())
				if err := r.Update(ctx, msvc); err != nil {
					return reconcileResult.FailedE(err)
				}
				return reconcileResult.OK()
			}
		}
	}

	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedService{}).
		Owns(&msvcv1.MongoDB{}).
		Watches(&source.Kind{
			Type: &watcherMsvc.MongoDB{},
		}, handler.EnqueueRequestsFromMapFunc(func(c client.Object) []reconcile.Request {
			var reqs []reconcile.Request
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: c.GetNamespace(), Name: c.GetName()},
			})
			return reqs
		})).
		Complete(r)
}
