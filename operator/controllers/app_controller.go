package controllers

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// applyAppsV1 "k8s.io/client-go/applyconfigurations/apps/v1"
	// applyCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	// applyMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func newBool(b bool) *bool {
	return &b
}

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
	logger    *zap.SugaredLogger
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

const maxCoolingTime = 5
const minCoolingTime = 2

const appFinalizer = "finalizers.kloudlite.io/app"

func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)
	logger := GetLogger(req.NamespacedName)
	r.logger = logger

	app := &crdsv1.App{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if app.HasToBeDeleted() {
		logger.Debugf("app.HasToBeDeleted()")
		return r.finalizeApp(ctx, app)
	}

	// logger.Debugf("app.IsNewGeneration() %v\n", app.IsNewGeneration())
	// logger.Debugf("app.HasJob() %v\n", app.HasJob())
	if app.IsNewGeneration() {
		logger.Debug("app.IsNewGeneration()")
		if app.HasJob() {
			return reconcileResult.Retry(maxCoolingTime)
		}
	}

	if app.HasJob() {
		logger.Debug("app.HasJob()")
		b, err := r.JobMgr.HasCompleted(ctx, app.Status.Job.Namespace, app.Status.Job.Name)
		if err != nil {
			return reconcileResult.Retry(minCoolingTime)
		}
		if b != nil {
			app.Status.JobCompleted = true
			if !*b {
				return r.updateStatus(ctx, app)
			}
			app.Status.Job = nil
			return r.updateStatus(ctx, app)
		}
		return reconcileResult.Retry(minCoolingTime)
	}

	// STEP: create new job
	specB, err := json.Marshal(app.Spec)
	if err != nil {
		r.logger.Error(errors.New("could not unmarshal app spec into []byte"))
		return reconcileResult.Failed()
	}

	if app.ShouldCreateJob() {
		logger.Debug("app.ShouldCreateJob()")
		job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
			Name:            "create-job",
			Namespace:       "hotspot",
			ServiceAccount:  "hotspot-cluster-svc-account",
			Image:           "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest",
			ImagePullPolicy: "Always",
			Args: []string{
				"create",
				"--name", app.Name,
				"--namespace", app.Namespace,
				"--spec", string(specB),
			},
		})

		if err != nil {
			return reconcileResult.Failed()
		}

		app.Status.Job = &crdsv1.ReconJob{
			Namespace: job.Namespace,
			Name:      job.Name,
		}
		app.Status.JobCompleted = false
		return r.updateStatus(ctx, app)
	}

	return reconcileResult.OK()
}

func (r *AppReconciler) updateStatus(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
	app.Status.Generation = &app.Generation
	err := r.Status().Update(ctx, app)
	if err != nil {
		return reconcileResult.RetryE(maxCoolingTime, errors.StatusUpdate(err))
	}
	if app.Status.Job == nil {
		r.logger.Debug("app has been updated nil")
	} else {
		r.logger.Debugf("app has been updated %+v", *app.Status.Job)
	}
	return reconcileResult.Retry(minCoolingTime)
}

func (r *AppReconciler) finalizeApp(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(app, appFinalizer) {
		if app.HasJob() {
			r.logger.Debug("app.HasJob() (deletion)")
			// STEP: cleaning currently executing jobs
			err := r.JobMgr.Delete(ctx, app.Status.Job.Namespace, app.Status.Job.Name)
			if err != nil {
				return reconcileResult.RetryE(minCoolingTime, err)
			}
			app.Status.Job = nil
			app.Status.JobCompleted = false
			return r.updateStatus(ctx, app)
		}

		if app.ShouldCreateDeletionJob() {
			r.logger.Debug("!app.ShouldCreateDeletionJob()")
			specB, err := json.Marshal(app.Spec)
			if err != nil {
				r.logger.Error(errors.New("could not unmarshal app spec into []byte"))
				return reconcileResult.Failed()
			}

			job, err := r.JobMgr.Create(ctx, "hotspot", &lib.JobVars{
				Name:            "delete-job",
				Namespace:       "hotspot",
				ServiceAccount:  "hotspot-cluster-svc-account",
				Image:           "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest",
				ImagePullPolicy: "Always",
				Args: []string{
					"delete",
					"--name", app.Name,
					"--namespace", app.Namespace,
					"--spec", string(specB),
				},
			})

			if err != nil {
				return reconcileResult.Failed()
			}

			app.Status.DeletionJob = &crdsv1.ReconJob{
				Name:      job.Name,
				Namespace: job.Namespace,
			}
			return r.updateStatus(ctx, app)
		}

		if app.HasRunningDeletionJob() {
			r.logger.Debug("app.HasDeletionJob()")
			//STEP:  WATCH for it
			jobStatus, err := r.JobMgr.HasCompleted(ctx, app.Status.DeletionJob.Namespace, app.Status.DeletionJob.Name)
			if err != nil {
				return reconcileResult.Retry(minCoolingTime)
			}
			if jobStatus != nil {
				app.Status.DeletionJobCompleted = true
				if !*jobStatus {
					r.logger.Debugf("jobStatus, false: %v", *jobStatus)
					return r.updateStatus(ctx, app)
				}
				r.logger.Debugf("jobStatus: %v", *jobStatus)
				app.Status.DeletionJob = nil
				return r.updateStatus(ctx, app)
			}
			return reconcileResult.Retry(minCoolingTime)
			// STEP: remove finalizer once done
		}

		if app.Status.DeletionJobCompleted {
			r.logger.Debug("app.Status.JobCompleted")
			controllerutil.RemoveFinalizer(app, appFinalizer)
			err := r.Update(ctx, app)
			if err != nil {
				r.logger.Error(errors.NewEf(err, "could not remove finalizers from app"))
			}
			return reconcileResult.OK()
		}
	}

	r.logger.Debug("contains no finalizers")
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Complete(r)
}
