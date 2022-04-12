package controllers

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
	logger    *zap.SugaredLogger
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/finalizers,verbs=update

const msvcFinalizer = "finalizers.kloudlite.io/msvc"

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	r.logger = GetLogger(req.NamespacedName)

	msvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, req.NamespacedName, msvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if msvc.HasToBeDeleted() {
		r.logger.Debugf("msvc.HasToBeDeleted()")
		return r.finalizeMsvc(ctx, msvc)
	}

	if msvc.IsNewGeneration() {
		if msvc.HasJob() {
			return reconcileResult.Retry(minCoolingTime)
		}
		msvc.DefaultStatus()
		return r.updateStatus(ctx, msvc)
	}

	if msvc.HasJob() {
		r.logger.Debug("msvc.HasJob()")
		b, err := r.JobMgr.HasCompleted(ctx, msvc.Status.Job.Namespace, msvc.Status.Job.Name)
		if err != nil {
			return reconcileResult.Retry(minCoolingTime)
		}
		if b != nil {
			msvc.Status.JobCompleted = newBool(true)
			if !*b {
				return r.updateStatus(ctx, msvc)
			}
			msvc.Status.Job = nil
			return r.updateStatus(ctx, msvc)
		}
		return reconcileResult.Retry(minCoolingTime)
	}

	// if msvc.HasNotCheckedDependency() {
	// }

	if msvc.ShouldCreateJob() {
		r.logger.Debug("msvc.ShouldCreateJob()")
		specB, err := json.Marshal(msvc.Spec)
		if err != nil {
			r.logger.Error(errors.New("could not unmarshal app spec into []byte"))
			return reconcileResult.Failed()
		}

		action := "create"
		if msvc.Generation > 1 {
			action = "update"
		}

		dockerI, err := r.getDockerImage(ctx, msvc.Spec.TemplateName, action)
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
				"--name", msvc.Name,
				"--namespace", msvc.Namespace,
				"--spec", string(specB),
			},
		})

		if err != nil {
			return reconcileResult.Failed()
		}

		msvc.Status.Job = &crdsv1.ReconJob{
			Namespace: job.Namespace,
			Name:      job.Name,
		}
		return r.updateStatus(ctx, msvc)
	}

	return ctrl.Result{}, nil
}

func (r *ManagedServiceReconciler) finalizeMsvc(ctx context.Context, msvc *crdsv1.ManagedService) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(msvc, msvcFinalizer) {
		if msvc.HasJob() {
			r.logger.Debug("msvc.HasJob() (deletion)")
			// STEP: cleaning currently executing jobs
			err := r.JobMgr.Delete(ctx, msvc.Status.Job.Namespace, msvc.Status.Job.Name)
			if err != nil {
				return reconcileResult.RetryE(minCoolingTime, err)
			}
			msvc.Status.Job = nil
			msvc.Status.JobCompleted = nil
			return r.updateStatus(ctx, msvc)
		}

		if msvc.ShouldCreateDeletionJob() {
			r.logger.Debug("app.ShouldCreateDeletionJob()")
			specB, err := json.Marshal(msvc.Spec)
			if err != nil {
				r.logger.Error(errors.New("could not unmarshal app spec into []byte"))
				return reconcileResult.Failed()
			}

			dockerI, err := r.getDockerImage(ctx, msvc.Spec.TemplateName, "delete")
			if err != nil {
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
					"--name", msvc.Name,
					"--namespace", msvc.Namespace,
					"--spec", string(specB),
				},
			})

			if err != nil {
				return reconcileResult.Failed()
			}

			msvc.Status.DeletionJob = &crdsv1.ReconJob{
				Name:      job.Name,
				Namespace: job.Namespace,
			}
			return r.updateStatus(ctx, msvc)
		}

		if msvc.HasDeletionJob() {
			r.logger.Debug("msvc.HasDeletionJob()")
			//STEP:  WATCH for it
			jobStatus, err := r.JobMgr.HasCompleted(ctx, msvc.Status.DeletionJob.Namespace, msvc.Status.DeletionJob.Name)
			if err != nil {
				return reconcileResult.Retry(minCoolingTime)
			}
			if jobStatus != nil {
				msvc.Status.DeletionJobCompleted = newBool(true)
				if !*jobStatus {
					r.logger.Debugf("DELETION jobStatus %v", *jobStatus)
					return r.updateStatus(ctx, msvc)
				}
				r.logger.Debugf("DELETION jobStatus: %v", *jobStatus)
				msvc.Status.DeletionJob = nil
				return r.updateStatus(ctx, msvc)
			}
			return reconcileResult.Retry(minCoolingTime)
			// STEP: remove finalizer once done
		}

		if msvc.Status.DeletionJobCompleted != nil {
			r.logger.Debug("app.Status.DeletionJobCompleted")
			controllerutil.RemoveFinalizer(msvc, appFinalizer)
			err := r.Update(ctx, msvc)
			if err != nil {
				r.logger.Error(errors.NewEf(err, "could not remove finalizers from app"))
			}
			return reconcileResult.OK()
		}
	}

	r.logger.Debug("contains no finalizers")
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) getDockerImage(ctx context.Context, templateName, action string) (string, error) {
	// READ configMap
	var cfg corev1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{Namespace: "hotspot", Name: "available-msvc"}, &cfg); err != nil {
		return "", errors.NewEf(err, "Could not get available managed svc list")
	}

	var msvcT MSvcTemplate
	err := yaml.Unmarshal([]byte(cfg.Data[templateName]), &msvcT)
	if err != nil {
		return "", errors.NewEf(err, "could not YAML unmarshal services into MsvcTemplate")
	}

	switch action {
	case "create":
		{
			return msvcT.Operations.Create, nil
		}
	case "update":
		{
			return msvcT.Operations.Update, nil
		}
	case "delete":
		{
			return msvcT.Operations.Update, nil
		}
	default:
		return "", errors.Newf("unknown action should be one of create|update|delete")
	}
}

func (r *ManagedServiceReconciler) updateStatus(ctx context.Context, msvc *crdsv1.ManagedService) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, msvc); err != nil {
		return reconcileResult.RetryE(maxCoolingTime, errors.StatusUpdate(err))
	}
	r.logger.Debug("ManagedService has been updated")
	return reconcileResult.Retry(minCoolingTime)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedService{}).
		Complete(r)
}
