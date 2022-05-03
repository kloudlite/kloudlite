package controllers

import (
	"context"
	"fmt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels2 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	mongodb "operators.kloudlite.io/apis/mongodbs.msvc/v1"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crdsv1 "operators.kloudlite.io/api/v1"
	// mongodb "operators.kloudlite.io/apis/mongodbs.msvc/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
)

// ManagedResourceReconciler reconciles a ManagedResource object
type ManagedResourceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	lib.MessageSender
	JobMgr lib.Job
	logger *zap.SugaredLogger
	mres   *crdsv1.ManagedResource
}

const mresFinalizer = "finalizers.kloudlite.io/managed-resource"

func (r *ManagedResourceReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(&r.mres.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrWhileReconcilation",
		Message: err.Error(),
	})

	return r.notify(ctx)
}

func (r *ManagedResourceReconciler) notify(ctx context.Context) (ctrl.Result, error) {
	r.logger.Infof("Notify conditions: %+v", r.mres.Status.Conditions)
	err := r.SendMessage(r.mres.LogRef(), lib.MessageReply{
		Key:        r.mres.LogRef(),
		Conditions: r.mres.Status.Conditions,
		Status:     meta.IsStatusConditionTrue(r.mres.Status.Conditions, "Ready"),
	})
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not send message into kafka"))
	}

	if err := r.Status().Update(ctx, r.mres); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.mres.LogRef()))
	}
	return reconcileResult.OK()
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", "true")
	logger.Debug("Reconciling ManagedResource")

	mres := &crdsv1.ManagedResource{}
	if err := r.Get(ctx, req.NamespacedName, mres); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	if !mres.HasLabels() {
		mres.EnsureLabels()
		if err := r.Update(ctx, mres); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}
	r.mres = mres

	if mres.DeletionTimestamp != nil {
		logger.Debug("ManagedResource is being deleted")
		return r.finalize(ctx, mres)
	}

	managedSvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, types.NamespacedName{Name: mres.Spec.ManagedSvc, Namespace: mres.Namespace}, managedSvc); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "failing to get managed-svc(name=%s, namespace=%s), would start again when it is available", mres.Spec.ManagedSvc, mres.Namespace))
	}

	if !mres.OwnedByMsvc(managedSvc) {
		a := mres.OwnerReferences
		a = append(a, metav1.OwnerReference{
			APIVersion:         managedSvc.APIVersion,
			Kind:               managedSvc.Kind,
			Name:               managedSvc.Name,
			UID:                managedSvc.UID,
			Controller:         fn.NewBool(false),
			BlockOwnerDeletion: fn.NewBool(true),
		})
		mres.SetOwnerReferences(a)
		if err := r.Update(ctx, mres); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}

	// STEP: check if managedsvc is ready
	if ok := meta.IsStatusConditionTrue(managedSvc.Status.Conditions, "Ready"); !ok {
		return r.notifyAndDie(ctx, errors.Newf("%s is not ready, would start again when it is ready", managedSvc.LogRef()))
	}

	b, err := templates.Parse(templates.MongoDBResourceDatabase, mres)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}

	if err := fn.KubectlApply(b); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not apply mongodb resource %s", mres.LogRef()))
	}

	var mdb mongodb.Database
	if err := r.Get(ctx, req.NamespacedName, &mdb); err != nil {
		return ctrl.Result{}, err
	}
	if mdb.Name != "" {
		mres.Status.Conditions = mdb.Status.Conditions
		return r.notify(ctx)
	}

	return r.notify(ctx)

	// //STEP: check if managedsvc is ready
	// if ok := meta.IsStatusConditionTrue(managedSvc.Status.Conditions, "Ready"); !ok {
	//	return reconcileResult.FailedE(errors.Newf("managedSvc %s is not ready", toRefString(managedSvc)))
	// }

	// msvcSecretName := fmt.Sprintf("msvc-%s", mres.Spec.ManagedSvc)
	// var msvcSecret corev1.Secret
	// if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: msvcSecretName}, &msvcSecret); err != nil {
	//	logger.Errorf("ManagedSvc secret %s/%s not found, aborting reconcilation", mres.Namespace, msvcSecretName)
	//	return reconcileResult.Failed()
	// }

	// logger.Infof("Secret: %+v\n", string(msvcSecret.Data["mongodb-root-password"]))

	// return reconcileResult.OK()
}

func (r *ManagedResourceReconciler) finalize(ctx context.Context, mres *crdsv1.ManagedResource) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(mres, finalizers.ManagedResource.String()) {
		controllerutil.RemoveFinalizer(mres, finalizers.ManagedResource.String())
		if err := r.Update(ctx, mres); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}

	if controllerutil.ContainsFinalizer(mres, finalizers.Foreground.String()) {
		controllerutil.RemoveFinalizer(mres, finalizers.Foreground.String())
		if err := r.Update(ctx, mres); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}

	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedResource{}).
		Owns(&mongodb.Database{}).
		Watches(&source.Kind{
			Type: &crdsv1.ManagedService{},
		}, handler.EnqueueRequestsFromMapFunc(func(c client.Object) []reconcile.Request {
			if s := c.GetLabels()["msvc.kloudlite.io/type"]; s != MongoDBStandalone.String() {
				return nil
			}

			var mresList crdsv1.ManagedResourceList
			if err := r.List(context.TODO(), &mresList, &client.ListOptions{
				LabelSelector: labels2.SelectorFromValidatedSet(map[string]string{
					fmt.Sprintf("mres.kloudlite.io/of-msvc"): c.GetName(),
				}),
				Namespace: c.GetNamespace(),
			}); err != nil {
				return nil
			}

			var reqs []reconcile.Request

			for _, item := range mresList.Items {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
				})
			}

			fmt.Printf("Managed Resource Reconciler: %+v\n", reqs)
			return reqs
		})).
		Complete(r)
}
