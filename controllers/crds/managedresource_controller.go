package crds

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	labels2 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"

	// mongodb "operators.kloudlite.io/apis/mongodbs.msvc/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// ManagedResourceReconciler reconciles a ManagedResource object
type ManagedResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lib.MessageSender
	logger *zap.SugaredLogger
	lt     metav1.Time
}

func (r *ManagedResourceReconciler) notify(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	r.logger.Infof("Notify conditions: %+v", req.mres.Status.Conditions)
	err := r.SendMessage(
		req.mres.NameRef(), lib.MessageReply{
			Key:        req.mres.NameRef(),
			Conditions: req.mres.Status.Conditions.GetConditions(),
		},
	)
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not send message into kafka"))
	}

	if err := r.Status().Update(ctx, req.mres); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", req.mres.NameRef()))
	}
	return reconcileResult.OK()
}

func (r *ManagedResourceReconciler) failWithErr(ctx context.Context, req *ServiceReconReq, err error) (
	ctrl.Result,
	error,
) {
	req.mres.Status.Conditions.Build(
		"", metav1.Condition{
			Type:    constants.ConditionReady.Type,
			Status:  metav1.ConditionFalse,
			Reason:  constants.ConditionReady.ErrorReason,
			Message: err.Error(),
		},
	)
	err = r.Status().Update(ctx, req.mres)
	return reconcileResult.FailedE(err)
}

func (r *ManagedResourceReconciler) reconcileStatus(ctx context.Context, req *ServiceReconReq) (*ctrl.Result, error) {
	req.mres.Status.Conditions.Init(r.lt)

	msvc := new(v1.ManagedService)
	if err := r.Get(
		ctx,
		types.NamespacedName{Namespace: req.Namespace, Name: req.mres.Spec.ManagedSvcName},
		msvc,
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		req.mres.Status.Conditions.MarkNotReady(
			errors.NewEf(
				err,
				"could not find managed service (%s)",
				req.mres.Spec.ManagedSvcName,
			),
		)
	}

	if msvc.Status.Conditions.IsTrue(constants.ConditionReady.Type) {
		return nil, errors.Newf("managed service (%s) is not ready", msvc.Name)
	}

	prevStatus := req.mres.Status
	req.mres.Status.Conditions.Reset()

	resObj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": req.mres.Spec.ApiVersion,
			"kind":       req.mres.Spec.Kind,
		},
	}

	if err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.mres.Name}, &resObj); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		req.mres.Status.Conditions.MarkNotReady(
			errors.NewEf(
				err,
				"could not find %s/%s/(%s)",
				req.mres.Spec.ApiVersion,
				req.mres.Spec.Kind,
				req.mres.Spec.ManagedSvcName,
			),
		)
	}

	mj, err := resObj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var j struct {
		Status struct {
			Conditions t.Conditions `json:"conditions,omitempty"`
		} `json:"status,omitempty"`
	}
	if err := json.Unmarshal(mj, &j); err != nil {
		return nil, err
	}

	req.mres.Status.Conditions = j.Status.Conditions

	if req.mres.Status.Conditions.Equal(prevStatus.Conditions) {
		req.logger.Infof("Status is already in sync, so moving forward with ops")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	if err := r.Status().Update(ctx, req.mres); err != nil {
		return nil, err
	}

	return reconcileResult.OKP()
}

func (r *ManagedResourceReconciler) reconcileOperations(ctx context.Context, req *ServiceReconReq) (
	ctrl.Result,
	error,
) {
	b, err := templates.Parse(templates.CommonMres, req.mres)
	if err != nil {
		req.logger.Error(err)
		return r.failWithErr(ctx, req, err)
	}

	if _, err := fn.KubectlApply(b); err != nil {
		req.logger.Error(err)
		return r.failWithErr(ctx, req, errors.NewEf(err, "could not apply mongodb resource %s", req.mres.NameRef()))
	}
	hash := req.mres.Hash()

	if hash == req.mres.Status.LastHash {
		return reconcileResult.OK()
	}

	req.mres.Status.LastHash = hash
	if err = r.Status().Update(ctx, req.mres); err != nil {
		return r.failWithErr(ctx, req, err)
	}

	return reconcileResult.OK()
}

type ServiceReconReq struct {
	ctrl.Request
	logger    *zap.SugaredLogger
	stateData map[string]string
	mres      *v1.ManagedResource
}

func (req *ServiceReconReq) GetStateData(key string) string {
	if req.stateData == nil {
		req.stateData = map[string]string{}
	}
	return req.stateData[key]
}

func (req *ServiceReconReq) SetStateData(key string, val string) {
	if req.stateData == nil {
		req.stateData = map[string]string{}
	}
	req.stateData[key] = val
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &ServiceReconReq{
		Request: orgReq,
		logger:  GetLogger(orgReq.NamespacedName),
		mres:    new(v1.ManagedResource),
	}

	if err := r.Get(ctx, orgReq.NamespacedName, req.mres); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if !req.mres.HasLabels() {
		req.mres.EnsureLabels()
		if err := r.Update(ctx, req.mres); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if req.mres.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req.mres)
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

// func (r *ManagedResourceReconciler) ReconcileOld(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	r.logger = GetLogger(req.NamespacedName)
// 	logger := r.logger.With("RECONCILE", "true")
// 	logger.Debug("Reconciling ManagedResource")
//
// 	mres := &v1.ManagedResource{}
// 	if err := r.Get(ctx, req.NamespacedName, mres); err != nil {
// 		if apiErrors.IsNotFound(err) {
// 			return reconcileResult.OK()
// 		}
// 		return reconcileResult.Failed()
// 	}
// 	if !mres.HasLabels() {
// 		mres.EnsureLabels()
// 		if err := r.Update(ctx, mres); err != nil {
// 			return ctrl.Result{}, err
// 		}
// 		return reconcileResult.OK()
// 	}
// 	r.mres = mres
//
// 	if mres.DeletionTimestamp != nil {
// 		logger.Debug("ManagedResource is being deleted")
// 		return r.finalize(ctx, mres)
// 	}
//
// 	managedSvc := &v1.ManagedService{}
// 	if err := r.Get(ctx, types.NamespacedName{Name: mres.Spec.ManagedSvc, Namespace: mres.Namespace}, managedSvc); err != nil {
// 		return r.notifyAndDie(ctx, errors.NewEf(err, "failing to get managed-svc(name=%s, namespace=%s), would start again when it is available", mres.Spec.ManagedSvc, mres.Namespace))
// 	}
//
// 	if !mres.OwnedByMsvc(managedSvc) {
// 		a := mres.OwnerReferences
// 		a = append(a, metav1.OwnerReference{
// 			APIVersion:         managedSvc.APIVersion,
// 			Kind:               managedSvc.Kind,
// 			Name:               managedSvc.Name,
// 			UID:                managedSvc.UID,
// 			Controller:         fn.NewBool(false),
// 			BlockOwnerDeletion: fn.NewBool(true),
// 		})
// 		mres.SetOwnerReferences(a)
// 		if err := r.Update(ctx, mres); err != nil {
// 			return ctrl.Result{}, err
// 		}
// 		return reconcileResult.OK()
// 	}
//
// 	// STEP: check if managedsvc is ready
// 	if ok := meta.IsStatusConditionTrue(managedSvc.Status.conditions, "Ready"); !ok {
// 		return r.notifyAndDie(ctx, errors.Newf("%s is not ready, would start again when it is ready", managedSvc.NameRef()))
// 	}
//
// 	b, err := templates.Parse(templates.MongoDBResourceDatabase, mres)
// 	if err != nil {
// 		return r.notifyAndDie(ctx, err)
// 	}
//
// 	if err := fn.KubectlApply(b); err != nil {
// 		return r.notifyAndDie(ctx, errors.NewEf(err, "could not apply mongodb resource %s", mres.NameRef()))
// 	}
//
// 	var mdb mongodb.Database
// 	if err := r.Get(ctx, req.NamespacedName, &mdb); err != nil {
// 		return ctrl.Result{}, err
// 	}
// 	if mdb.Name != "" {
// 		mres.Status.conditions = mdb.Status.conditions
// 		return r.notify(ctx)
// 	}
//
// 	return r.notify(ctx)
//
// 	// //STEP: check if managedsvc is ready
// 	// if ok := meta.IsStatusConditionTrue(managedSvc.Status.conditions, "Ready"); !ok {
// 	//	return reconcileResult.FailedE(errors.Newf("managedSvc %s is not ready", toRefString(managedSvc)))
// 	// }
//
// 	// msvcSecretName := fmt.Sprintf("msvc-%s", mres.Spec.ManagedSvc)
// 	// var msvcSecret corev1.Secret
// 	// if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: msvcSecretName}, &msvcSecret); err != nil {
// 	//	logger.Errorf("ManagedSvc secret %s/%s not found, aborting reconcilation", mres.Namespace, msvcSecretName)
// 	//	return reconcileResult.Failed()
// 	// }
//
// 	// logger.Infof("Secret: %+v\n", string(msvcSecret.Data["mongodb-root-password"]))
//
// 	// return reconcileResult.OK()
// }

func (r *ManagedResourceReconciler) finalize(ctx context.Context, mres *v1.ManagedResource) (ctrl.Result, error) {
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
	r.lt = metav1.Time{Time: time.Now()}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ManagedResource{}).
		Watches(
			&source.Kind{
				Type: &v1.ManagedService{},
			}, handler.EnqueueRequestsFromMapFunc(
				func(c client.Object) []reconcile.Request {
					var mresList v1.ManagedResourceList
					msvcLabels := c.GetLabels()
					if msvcLabels == nil {
						msvcLabels = map[string]string{}
					}

					key, value := v1.ManagedResource{}.LabelRef()
					msvcLabels[key] = value
					if err := r.List(
						context.TODO(), &mresList, &client.ListOptions{
							LabelSelector: labels2.SelectorFromValidatedSet(msvcLabels),
							Namespace:     c.GetNamespace(),
						},
					); err != nil {
						return nil
					}

					var reqs []reconcile.Request

					for _, item := range mresList.Items {
						nn := types.NamespacedName{Name: item.GetName(), Namespace: item.GetNamespace()}
						for _, req := range reqs {
							if req.String() == nn.String() {
								return nil
							}
						}

						reqs = append(reqs, reconcile.Request{NamespacedName: nn})
					}
					return reqs
				},
			),
		).
		Complete(r)
}
