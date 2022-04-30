package controllers

import (
	"context"
	fn "operators.kloudlite.io/lib/functions"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/api/v1"
	msvcv1 "operators.kloudlite.io/apis/msvc/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ClientSet   *kubernetes.Clientset
	JobMgr      lib.Job
	SendMessage func(key string, msg lib.MessageReply) error
	logger      *zap.SugaredLogger
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/finalizers,verbs=update

const msvcFinalizer = "finalizers.kloudlite.io/managed-service"

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

	if msvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, msvc)
	}

	if msvc.Spec.Type != "MongoDBStandalone" {
		return reconcileResult.Failed()
	}

	kt, err := templates.UseTemplate(templates.MongoDBStandalone)
	if err != nil {
		logger.Info("could not useTemplate, aborting...")
		return reconcileResult.Failed()
	}
	b, err := kt.WithValues(msvc)
	if err != nil {
		logger.Info(b, err)
	}

	if err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply managed service"))
	}
	logger.Infof("applied managed service")

	if err := r.SendMessage(msvc.String(), lib.MessageReply{
		Conditions: msvc.Status.Conditions,
		Status:     false,
	}); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not send message to kafka"))
	}

	return reconcileResult.OK()
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
		Complete(r)
}
