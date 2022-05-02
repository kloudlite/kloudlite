package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fn "operators.kloudlite.io/lib/functions"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

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
	Scheme      *runtime.Scheme
	ClientSet   *kubernetes.Clientset
	SendMessage func(key string, msg lib.MessageReply) error
	JobMgr      lib.Job
	logger      *zap.SugaredLogger
	mres        *crdsv1.ManagedResource
}

const mresFinalizer = "finalizers.kloudlite.io/managed-resource"

func (r *ManagedResourceReconciler) NotifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(&r.mres.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrUnknown",
		Message: err.Error(),
	})

	return r.Notify(ctx)
}

func (r *ManagedResourceReconciler) Notify(ctx context.Context) (ctrl.Result, error) {
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

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", "true")
	logger.Debug("Reconciling ManagedResource")

	mres := &crdsv1.ManagedResource{}
	if err := r.Get(ctx, req.NamespacedName, mres); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		//return reconcileResult.Retry()
		return reconcileResult.Failed()
	}

	if mres.DeletionTimestamp != nil {
		logger.Debug("ManagedResource is being deleted")
		return r.finalize(ctx, mres)
	}

	b, err := templates.Parse(templates.MongoDBResourceDatabase, mres)
	if err != nil {
		return r.NotifyAndDie(ctx, err)
	}

	if err := fn.KubectlApply(b); err != nil {
		return r.NotifyAndDie(ctx, errors.NewEf(err, "could not apply mongodb resource %s", mres.LogRef()))
	}

	return r.Notify(ctx)

	////STEP: check if managedsvc is ready
	//if ok := meta.IsStatusConditionTrue(managedSvc.Status.Conditions, "Ready"); !ok {
	//	return reconcileResult.FailedE(errors.Newf("managedSvc %s is not ready", toRefString(managedSvc)))
	//}

	//msvcSecretName := fmt.Sprintf("msvc-%s", mres.Spec.ManagedSvc)
	//var msvcSecret corev1.Secret
	//if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: msvcSecretName}, &msvcSecret); err != nil {
	//	logger.Errorf("ManagedSvc secret %s/%s not found, aborting reconcilation", mres.Namespace, msvcSecretName)
	//	return reconcileResult.Failed()
	//}

	//logger.Infof("Secret: %+v\n", string(msvcSecret.Data["mongodb-root-password"]))

	//return reconcileResult.OK()
}

func (r *ManagedResourceReconciler) finalize(ctx context.Context, mres *crdsv1.ManagedResource) (ctrl.Result, error) {
	return reconcileResult.OK()
}

// func (r *ManagedResourceReconciler) Reconcile2(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	r.logger = GetLogger(req.NamespacedName)

// 	mres := &crdsv1.ManagedResource{}
// 	if err := r.Get(ctx, req.NamespacedName, mres); err != nil {
// 		if apiErrors.IsNotFound(err) {
// 			return reconcileResult.OK()
// 		}
// 		return reconcileResult.Failed()
// 	}

// 	if mres.HasToBeDeleted() {
// 		return r.finalizeMres(ctx, mres)
// 	}

// 	if mres.IsNewGeneration() {
// 		r.logger.Debug("mres.IsNewGeneration()")
// 		if mres.Status.ApplyJobCheck.IsRunning() {
// 			return reconcileResult.Retry(minCoolingTime)
// 		}
// 		mres.DefaultStatus()
// 		return r.updateStatus(ctx, mres)
// 	}

// 	if mres.Status.ManagedSvcDepCheck.ShouldCheck() {
// 		r.logger.Debugf("mres.Status.ManagedSvcDepCheck.ShouldCheck()")
// 		mres.Status.ManagedSvcDepCheck.SetStarted()
// 		msvc := &crdsv1.ManagedService{}
// 		if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: mres.Spec.ManagedSvc}, msvc); err != nil {
// 			mres.Status.ManagedSvcDepCheck.SetFinishedWith(false, errors.NewEf(err, "could not get managed svc").Error())
// 			return r.updateStatus(ctx, mres)
// 		}

// 		c := meta.FindStatusCondition(msvc.Status.Conditions, "Ready")
// 		if c == nil || c.Status != metav1.ConditionTrue {
// 			mres.Status.ManagedSvcDepCheck.SetFinishedWith(false, errors.Newf("managed svc (%s/%s) is not ready, yet", msvc.Namespace, msvc.Name).Error())
// 			return r.updateStatus(ctx, mres)
// 		}
// 		mres.Status.ManagedSvcDepCheck.SetFinishedWith(true, fmt.Sprintf("managed svc (%s/%s) is ready", msvc.Namespace, msvc.Name))
// 		return r.updateStatus(ctx, mres)
// 	}

// 	if !mres.Status.ManagedSvcDepCheck.Status {
// 		r.logger.Debugf("ManagedSvc Dependency Check failed ..., would retry soon")
// 		if mres.Status.ManagedSvcDepCheck.ShouldRetry(maxCoolingTime) {
// 			return r.updateStatus(ctx, mres)
// 		}
// 		return reconcileResult.Retry(1)
// 	}

// 	if mres.Status.ApplyJobCheck.ShouldCheck() {
// 		mres.Status.ApplyJobCheck.SetStarted()
// 		specB, err := json.Marshal(mres.Spec)
// 		if err != nil {
// 			e := errors.New("could not unmarshal mres Spec into []byte")
// 			r.logger.Error(e)
// 			mres.Status.ApplyJobCheck.SetFinishedWith(false, e.Error())
// 			return r.updateStatus(ctx, mres)
// 		}

// 		job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
// 			Name:            fmt.Sprintf("create-job-%s", mres.Name),
// 			Namespace:       mres.Namespace,
// 			ServiceAccount:  SvcAccountName,
// 			Image:           mres.Spec.Operations.Apply,
// 			ImagePullPolicy: "Always",
// 			Args: []string{
// 				"apply",
// 				"--name", mres.Name,
// 				"--namespace", mres.Namespace,
// 				"--spec", string(specB),
// 			},
// 		})

// 		if err != nil {
// 			e := errors.NewEf(err, "could not create job")
// 			r.logger.Error(e)
// 			mres.Status.ApplyJobCheck.SetFinishedWith(false, e.Error())
// 			return r.updateStatus(ctx, mres)
// 		}

// 		mres.Status.ApplyJobCheck.Job = &crdsv1.ReconJob{
// 			Namespace: job.Namespace,
// 			Name:      job.Name,
// 		}

// 		return r.updateStatus(ctx, mres)
// 	}

// 	if mres.Status.ApplyJobCheck.IsRunning() {
// 		j := mres.Status.ApplyJobCheck.Job
// 		b, err := r.JobMgr.HasCompleted(ctx, j.Namespace, j.Name)
// 		if err != nil {
// 			mres.Status.ApplyJobCheck.SetFinishedWith(false, errors.NewEf(err, "job failed").Error())
// 			return r.updateStatus(ctx, mres)
// 		}

// 		if b != nil {
// 			if !*b {
// 				mres.Status.ApplyJobCheck.SetFinishedWith(false, "job failed")
// 				return r.updateStatus(ctx, mres)
// 			}
// 			mres.Status.ApplyJobCheck.SetFinishedWith(true, "job failed")
// 			return r.updateStatus(ctx, mres)
// 		}

// 		return reconcileResult.Retry(minCoolingTime)
// 	}

// 	if !mres.Status.ApplyJobCheck.Status {
// 		r.logger.Debugf("ApplyJobCheck status has failed")
// 		return reconcileResult.Failed()
// 	}

// 	r.logger.Infof("Managed Resource reconcile complete ...")
// 	return reconcileResult.OK()
// }

// func (r *ManagedResourceReconciler) finalizeMres(ctx context.Context, mres *crdsv1.ManagedResource) (ctrl.Result, error) {
// 	logger := r.logger.With("FINALIZER", "true")

// 	if !controllerutil.ContainsFinalizer(mres, mresFinalizer) {
// 		return reconcileResult.OK()
// 	}

// 	if mres.Status.ApplyJobCheck.IsRunning() {
// 		j := mres.Status.ApplyJobCheck.Job
// 		err := r.JobMgr.Delete(ctx, j.Namespace, j.Name)
// 		if err != nil {
// 			mres.Status.ApplyJobCheck.SetFinishedWith(false, err.Error())
// 			return r.updateStatus(ctx, mres)
// 		}
// 		mres.Status.ApplyJobCheck.SetFinishedWith(true, "job deleted")
// 		return r.updateStatus(ctx, mres)
// 	}

// 	if mres.Status.DeleteJobCheck.ShouldCheck() {
// 		mres.Status.DeleteJobCheck.SetStarted()
// 		logger.Debug("mres.ShouldCreateDeletionJob()")
// 		specB, err := json.Marshal(mres.Spec)
// 		if err != nil {
// 			mres.Status.DeleteJobCheck.SetFinishedWith(false, errors.New("could not unmarshal mres spec into []byte").Error())
// 			return r.updateStatus(ctx, mres)
// 		}

// 		job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
// 			Name:            fmt.Sprintf("delete-job-%s", mres.Name),
// 			Namespace:       mres.Namespace,
// 			ServiceAccount:  SvcAccountName,
// 			Image:           mres.Spec.Operations.Delete,
// 			ImagePullPolicy: "Always",
// 			Args: []string{
// 				"delete",
// 				"--name", mres.Name,
// 				"--namespace", mres.Namespace,
// 				"--spec", string(specB),
// 			},
// 		})

// 		if err != nil {
// 			mres.Status.DeleteJobCheck.SetFinishedWith(false, errors.NewEf(err, "could not create deletion job").Error())
// 			return reconcileResult.Failed()
// 		}

// 		mres.Status.DeleteJobCheck.Job = &crdsv1.ReconJob{
// 			Name:      job.Name,
// 			Namespace: job.Namespace,
// 		}
// 		return r.updateStatus(ctx, mres)
// 	}

// 	if mres.Status.DeleteJobCheck.IsRunning() {
// 		j := mres.Status.DeleteJobCheck.Job
// 		jst, err := r.JobMgr.HasCompleted(ctx, j.Namespace, j.Name)

// 		if err != nil {
// 			return reconcileResult.Retry(minCoolingTime)
// 		}

// 		if jst != nil {
// 			if !*jst {
// 				mres.Status.DeleteJobCheck.SetFinishedWith(false, "job failed")
// 				return r.updateStatus(ctx, mres)
// 			}
// 			mres.Status.DeleteJobCheck.SetFinishedWith(true, "job succeeded")
// 			return r.updateStatus(ctx, mres)
// 		}
// 		return reconcileResult.Retry(minCoolingTime)
// 	}

// 	if !mres.Status.DeleteJobCheck.Status {
// 		logger.Infof("mres.Status.DeleteJobCheck.Status has failed, still letting pass through though ...")
// 	}

// 	logger.Debug("[Finalizer]: all deletion checks completed, removing finalizer ...")
// 	controllerutil.RemoveFinalizer(mres, mresFinalizer)
// 	err := r.Update(ctx, mres)
// 	if err != nil {
// 		logger.Error(errors.NewEf(err, "could not remove finalizers from app"))
// 		return reconcileResult.Retry(minCoolingTime)
// 	}
// 	return reconcileResult.OK()
// }

// func (r *ManagedResourceReconciler) updateStatus(ctx context.Context, mres *crdsv1.ManagedResource) (ctrl.Result, error) {
// 	mres.BuildConditions()
// 	if err := r.Status().Update(ctx, mres); err != nil {
// 		return reconcileResult.OK()
// 		// r.logger.Infof("Status Update Failed ... as %w", err)
// 		// return reconcileResult.RetryE(0, errors.StatusUpdate(err))
// 	}
// 	r.logger.Debug("ManagedResource has been updated")
// 	return reconcileResult.Retry()
// }

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedResource{}).
		Complete(r)
}
