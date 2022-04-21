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

const msvcFinalizer = "finalizers.kloudlite.io/managed-service"

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)

	msvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, req.NamespacedName, msvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if msvc.HasToBeDeleted() {
		return r.finalizeMsvc(ctx, msvc)
	}

	if msvc.IsNewGeneration() {
		if msvc.Status.ApplyJobCheck.IsRunning() {
			return reconcileResult.Retry(minCoolingTime)
		}
		msvc.DefaultStatus()
		return r.updateStatus(ctx, msvc)
	}

	if msvc.Status.ApplyJobCheck.ShouldCheck() {
		r.logger.Debugf("msvc.Status.ApplyJobCheck.ShouldCheck()")
		msvc.Status.ApplyJobCheck.SetStarted()

		specB, err := json.Marshal(msvc.Spec)
		if err != nil {
			e := errors.Newf("could not unmarshal Spec into []byte")
			msvc.Status.ApplyJobCheck.SetFinishedWith(false, e.Error())
			return r.updateStatus(ctx, msvc)
		}

		action := "create"
		if msvc.Generation > 1 {
			action = "update"
		}

		dockerI, err := r.getDockerImage(ctx, msvc.Spec.TemplateName, action)
		if err != nil {
			e := errors.NewEf(err, "could not find docker image for (template=%s)", msvc.Spec.TemplateName)
			msvc.Status.ApplyJobCheck.SetFinishedWith(false, e.Error())
			return r.updateStatus(ctx, msvc)
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
			e := errors.NewEf(err, "could not create job")
			r.logger.Error(e)
			msvc.Status.ApplyJobCheck.SetFinishedWith(false, e.Error())
			return r.updateStatus(ctx, msvc)
		}

		msvc.Status.ApplyJobCheck.Job = &crdsv1.ReconJob{Namespace: job.Namespace, Name: job.Name}
		return r.updateStatus(ctx, msvc)
	}

	if msvc.Status.ApplyJobCheck.IsRunning() {
		r.logger.Debugf("msvc.Status.ApplyJobCheck.IsRunning()")

		j := msvc.Status.ApplyJobCheck.Job

		b, err := r.JobMgr.HasCompleted(ctx, j.Namespace, j.Name)
		if err != nil {
			msvc.Status.ApplyJobCheck.SetFinishedWith(false, errors.NewEf(err, "job failed").Error())
			return r.updateStatus(ctx, msvc)
		}

		if b != nil {
			if !*b {
				msvc.Status.ApplyJobCheck.SetFinishedWith(false, "job failed")
				return r.updateStatus(ctx, msvc)
			}
			msvc.Status.ApplyJobCheck.SetFinishedWith(true, "job succeeded")
			return r.updateStatus(ctx, msvc)
		}

		return reconcileResult.Retry(minCoolingTime)
	}

	if !msvc.Status.ApplyJobCheck.Status {
		r.logger.Debugf("ManagedSvc ApplyJob failed, aborting reconcilation ...")
		return reconcileResult.Failed()
	}

	r.logger.Infof("ManagedService reconcile completed ...")
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) finalizeMsvc(ctx context.Context, msvc *crdsv1.ManagedService) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(msvc, msvcFinalizer) {
		return reconcileResult.OK()
	}
	logger := r.logger.With("FINALIZER", "true")

	if msvc.Status.ApplyJobCheck.IsRunning() {
		logger.Debugf("[Finalizer]: killing app.Status.ApplyJob.IsRunning()")
		j := msvc.Status.ApplyJobCheck.Job
		err := r.JobMgr.Delete(ctx, j.Namespace, j.Name)
		if err != nil {
			logger.Error(errors.NewEf(err, "error deleting job %s/%s, silently exiting ", j.Namespace, j.Name))
		}
		msvc.Status.ApplyJobCheck.SetFinishedWith(false, "killed by APP finalizer")
		return r.updateStatus(ctx, msvc)
	}

	if msvc.Status.DeleteJobCheck.ShouldCheck() {
		logger.Debugf("[Finalizer]: msvc.Status.DeleteJobCheck.ShouldCheck()")
		msvc.Status.DeleteJobCheck.SetStarted()

		specB, err := json.Marshal(msvc.Spec)
		if err != nil {
			msvc.Status.DeleteJobCheck.SetFinishedWith(false, "could not unmarshal Spec into []byte")
			return r.updateStatus(ctx, msvc)
		}

		dockerI, err := r.getDockerImage(ctx, msvc.Spec.TemplateName, "delete")
		if err != nil {
			msvc.Status.DeleteJobCheck.SetFinishedWith(false, errors.NewEf(err, "could not find docker image (template=%s) for action=delete", msvc.Spec.TemplateName).Error())
			return r.updateStatus(ctx, msvc)
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
			msvc.Status.DeleteJobCheck.SetFinishedWith(false, errors.NewEf(err, "could not create deletion job").Error())
			return r.updateStatus(ctx, msvc)
		}

		msvc.Status.DeleteJobCheck.Job = &crdsv1.ReconJob{
			Name:      job.Name,
			Namespace: job.Namespace,
		}
		return r.updateStatus(ctx, msvc)
	}

	if msvc.Status.DeleteJobCheck.IsRunning() {
		logger.Debug("msvc.Status.DeleteJobCheck.IsRunning()")
		j := msvc.Status.DeleteJobCheck.Job
		jst, err := r.JobMgr.HasCompleted(ctx, j.Namespace, j.Name)

		if err != nil {
			msvc.Status.DeleteJobCheck.SetFinishedWith(false, errors.NewEf(err, "job failed").Error())
			return r.updateStatus(ctx, msvc)
		}

		if jst != nil {
			if !*jst {
				msvc.Status.DeleteJobCheck.SetFinishedWith(false, "job failed")
				return r.updateStatus(ctx, msvc)
			}
			msvc.Status.DeleteJobCheck.SetFinishedWith(true, "job succeeded")
			return r.updateStatus(ctx, msvc)
		}
		return reconcileResult.Retry(minCoolingTime)
	}

	if !msvc.Status.DeleteJobCheck.Status {
		logger.Infof("msvc.Status.DeleteJobCheck.Status has failed, letting pass through though ...")
	}

	logger.Debug("[Finalizer]: all deletion checks completed ...")
	controllerutil.RemoveFinalizer(msvc, msvcFinalizer)
	err := r.Update(ctx, msvc)
	if err != nil {
		logger.Error(errors.NewEf(err, "could not remove finalizers from app"))
		return reconcileResult.Retry(minCoolingTime)
	}
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) getDockerImage(ctx context.Context, templateName, action string) (string, error) {
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
	msvc.BuildConditions()

	if err := r.Status().Update(ctx, msvc); err != nil {
		return reconcileResult.RetryE(maxCoolingTime, errors.StatusUpdate(err))
	}
	r.logger.Debug("ManagedService has been updated")
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedService{}).
		Complete(r)
}
