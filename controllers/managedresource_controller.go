package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// ManagedResourceReconciler reconciles a ManagedResource object
type ManagedResourceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
	logger    *zap.SugaredLogger
}

const mresFinalizer = "finalizers.kloudlite.io/mres"

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)

	mres := &crdsv1.ManagedResource{}
	if err := r.Get(ctx, req.NamespacedName, mres); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if mres.HasToBeDeleted() {
		return r.finalizeMres(ctx, mres)
	}

	if mres.IsNewGeneration() {
		r.logger.Debug("mres.IsNewGeneration()")
		if mres.Status.ApplyJobCheck.IsRunning() {
			return reconcileResult.Retry(minCoolingTime)
		}
		mres.DefaultStatus()
		return r.updateStatus(ctx, mres)
	}

	if mres.Status.ManagedSvcDepCheck.ShouldCheck() {
		mres.Status.ManagedSvcDepCheck.SetStarted()
		msvc := &crdsv1.ManagedService{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: mres.Spec.ManagedSvc}, msvc); err != nil {
			mres.Status.ManagedSvcDepCheck.SetFinishedWith(false, errors.NewEf(err, "could not get managed svc").Error())
			return r.updateStatus(ctx, mres)
		}

		c := meta.FindStatusCondition(msvc.Status.Conditions, "Ready")
		if c == nil || c.Status != metav1
	}

	if !mres.Status.ManagedSvcDepCheck.Status {
		r.logger.Debugf("ManagedSvc Dependency Check failed ...")
		time.AfterFunc(time.Second*maxCoolingTime, func() {
			mres.Status.ManagedSvcDepCheck = crdsv1.Recon{}
			r.updateStatus(ctx, mres)
		})
		return reconcileResult.Retry(minCoolingTime)
	}

	if mres.Status.ApplyJobCheck.ShouldCheck() {
	}

	if mres.Status.ApplyJobCheck.IsRunning() {
	}

	if !mres.Status.ApplyJobCheck.Status {
		return reconcileResult.Failed()
	}

	return reconcileResult.OK()
}

func (r *ManagedResourceReconciler) Reconcile2(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	r.logger = GetLogger(req.NamespacedName)

	mres := &crdsv1.ManagedResource{}
	if err := r.Get(ctx, req.NamespacedName, mres); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if mres.HasToBeDeleted() {
		r.logger.Debugf("mres.HasToBeDeleted()")
		return r.finalizeMres(ctx, mres)
	}

	if mres.IsNewGeneration() {
		r.logger.Debug("mres.IsNewGeneration()")
		if mres.HasJob() {
			fmt.Printf("HERE %+v\n", mres.HasJob())
			return reconcileResult.Retry(minCoolingTime)
		}
		mres.DefaultStatus()
		return r.updateStatus(ctx, mres)
	}

	if mres.HasJob() {
		r.logger.Debug("mres.HasJob()")
		b, err := r.JobMgr.HasCompleted(ctx, mres.Status.Job.Namespace, mres.Status.Job.Name)
		if err != nil {
			return reconcileResult.Retry(minCoolingTime)
		}
		if b != nil {
			mres.Status.JobCompleted = newBool(true)
			if !*b {
				return r.updateStatus(ctx, mres)
			}
			mres.Status.Job = nil
			return r.updateStatus(ctx, mres)
		}
		return reconcileResult.Retry(minCoolingTime)
	}

	if mres.HasNotCheckedDependency() {
		r.logger.Debug("mres.HasNotCheckedDependency()")
		checks := make(map[string]string)
		msvc := &crdsv1.ManagedService{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: mres.Spec.ManagedSvc}, msvc); err != nil {
			r.logger.Debug(types.NamespacedName{Namespace: mres.Namespace, Name: mres.Spec.ManagedSvc})
			r.logger.Debug(mres.Spec.ManagedSvc)
			r.logger.Error(err)
			return reconcileResult.Failed()
		}
		// ASSERT: no job running on msvc
		r.logger.Debugf("msvc.Status: %+v\n", msvc.Status)
		if msvc.Status.Job != nil {
			checks[fmt.Sprintf("msvc/%s", msvc.Name)] = "Not Ready"
		}
		mres.Status.DependencyChecked = &checks
		return r.updateStatus(ctx, mres)
	}

	if mres.ShouldCreateJob() {
		r.logger.Debug("mres.ShouldCreateJob()")

		specB, err := json.Marshal(mres.Spec)
		if err != nil {
			r.logger.Error(errors.New("could not unmarshal mres spec into []byte"))
			return reconcileResult.Failed()
		}

		action := "create"
		if mres.Generation > 1 {
			action = "update"
		}

		dockerI, err := r.getDockerImage(ctx, mres, action)
		if err != nil {
			r.logger.Debug(err)
			return reconcileResult.Failed()
		}

		job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
			Name:            "create-job",
			Namespace:       "hotspot",
			ServiceAccount:  "hotspot-cluster-svc-account",
			Image:           dockerI,
			ImagePullPolicy: "Always",
			Args: []string{
				action,
				"--name", mres.Name,
				"--namespace", mres.Namespace,
				"--spec", string(specB),
			},
		})

		if err != nil {
			return reconcileResult.Failed()
		}

		mres.Status.Job = &crdsv1.ReconJob{
			Namespace: job.Namespace,
			Name:      job.Name,
		}
		return r.updateStatus(ctx, mres)
	}

	return reconcileResult.OK()
}

func (r *ManagedResourceReconciler) finalizeMres(ctx context.Context, mres *crdsv1.ManagedResource) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(mres, mresFinalizer) {
		if mres.HasJob() {
			r.logger.Debug("mres.HasJob() (deletion)")
			// STEP: cleaning currently executing jobs
			err := r.JobMgr.Delete(ctx, mres.Status.Job.Namespace, mres.Status.Job.Name)
			if err != nil {
				return reconcileResult.RetryE(minCoolingTime, err)
			}
			mres.Status.Job = nil
			mres.Status.JobCompleted = nil
			return r.updateStatus(ctx, mres)
		}

		if mres.ShouldCreateDeletionJob() {
			r.logger.Debug("mres.ShouldCreateDeletionJob()")
			specB, err := json.Marshal(mres.Spec)
			if err != nil {
				r.logger.Error(errors.New("could not unmarshal mres spec into []byte"))
				return reconcileResult.Failed()
			}

			dockerI, err := r.getDockerImage(ctx, mres, "delete")
			if err != nil {
				r.logger.Debug(err)
				return reconcileResult.Failed()
			}

			job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
				Name:            "delete-job",
				Namespace:       "hotspot",
				ServiceAccount:  "hotspot-cluster-svc-account",
				Image:           dockerI,
				ImagePullPolicy: "Always",
				Args: []string{
					"delete",
					"--name", mres.Name,
					"--namespace", mres.Namespace,
					"--spec", string(specB),
				},
			})

			if err != nil {
				return reconcileResult.Failed()
			}

			mres.Status.DeletionJob = &crdsv1.ReconJob{
				Name:      job.Name,
				Namespace: job.Namespace,
			}
			return r.updateStatus(ctx, mres)
		}

		if mres.HasDeletionJob() {
			r.logger.Debug("mres.HasDeletionJob()")
			//STEP:  WATCH for it
			jobStatus, err := r.JobMgr.HasCompleted(ctx, mres.Status.DeletionJob.Namespace, mres.Status.DeletionJob.Name)
			if err != nil {
				return reconcileResult.Retry(minCoolingTime)
			}
			if jobStatus != nil {
				mres.Status.DeletionJobCompleted = newBool(true)
				if !*jobStatus {
					r.logger.Debugf("DELETION jobStatus %v", *jobStatus)
					return r.updateStatus(ctx, mres)
				}
				r.logger.Debugf("DELETION jobStatus: %v", *jobStatus)
				mres.Status.DeletionJob = nil
				return r.updateStatus(ctx, mres)
			}
			return reconcileResult.Retry(minCoolingTime)
			// STEP: remove finalizer once done
		}

		if mres.Status.DeletionJobCompleted != nil {
			r.logger.Debug("mres.Status.DeletionJobCompleted")
			controllerutil.RemoveFinalizer(mres, appFinalizer)
			err := r.Update(ctx, mres)
			if err != nil {
				r.logger.Error(errors.NewEf(err, "could not remove finalizers from mres"))
			}
			return reconcileResult.OK()
		}
	}

	r.logger.Debug("contains no finalizers")
	return reconcileResult.OK()
}

func (r *ManagedResourceReconciler) getDockerImage(ctx context.Context, mres *crdsv1.ManagedResource, action string) (string, error) {
	var m crdsv1.ManagedService
	if err := r.Get(ctx, types.NamespacedName{Namespace: mres.Namespace, Name: mres.Spec.ManagedSvc}, &m); err != nil {
		return "", errors.NewEf(err, "could not get managedsvc(%s) for this resource", mres.Spec.ManagedSvc)
	}

	// READ configMap
	var cfg corev1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{Namespace: "hotspot", Name: "available-msvc"}, &cfg); err != nil {
		return "", errors.NewEf(err, "Could not get available managed svc list")
	}

	var msvcT MSvcTemplate
	err := yaml.Unmarshal([]byte(cfg.Data[m.Spec.TemplateName]), &msvcT)

	if err != nil {
		return "", errors.NewEf(err, "could not YAML unmarshal services into MsvcTemplate")
	}

	var res *MresTemplate
	for _, item := range msvcT.Resources {
		if item.Type == mres.Spec.Type {
			res = &item
			break
		}
	}
	if res == nil {
		return "", errors.Newf("could not find resource(%s) in svc(%s) definition", mres.Name, m.Spec.TemplateName)
	}

	switch action {
	case "create":
		{
			return res.Operations.Create, nil
		}
	case "update":
		{
			return res.Operations.Update, nil
		}
	case "delete":
		{
			return res.Operations.Update, nil
		}
	default:
		return "", errors.Newf("unknown action should be one of create|update|delete")
	}
}

func (r *ManagedResourceReconciler) updateStatus(ctx context.Context, mres *crdsv1.ManagedResource) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, mres); err != nil {
		return reconcileResult.RetryE(maxCoolingTime, errors.StatusUpdate(err))
	}
	r.logger.Debug("ManagedResource has been updated")
	return reconcileResult.Retry(minCoolingTime)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedResource{}).
		Complete(r)
}
