package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
	// batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

	if app.IsNewGeneration() {
		logger.Debugf("app.IsNewGeneration() %+v\n", app.Status)
		if app.HasJob() {
			logger.Debugf("app.HasJob() %+v\n", app.Status.Job)
			return reconcileResult.Retry(maxCoolingTime)
		}
		app.DefaultStatus()
		return r.updateStatus(ctx, app)
	}

	if app.HasJob() {
		logger.Debug("app.HasJob()")
		b, err := r.JobMgr.HasCompleted(ctx, app.Status.Job.Namespace, app.Status.Job.Name)
		if err != nil {
			return reconcileResult.Retry(minCoolingTime)
		}
		if b != nil {
			app.Status.JobCompleted = newBool(true)
			if !*b {
				return r.updateStatus(ctx, app)
			}
			app.Status.Job = nil
			return r.updateStatus(ctx, app)
		}
		return reconcileResult.Retry(minCoolingTime)
	}

	// STEP: do assertions for dependents

	// ASSERT: has configmaps/secrets
	if app.HasNotCheckedDependency() {
		logger.Debug("app.HasNotCheckedDependency()")
		checks := make(map[string]string)
		for _, container := range app.Spec.Containers {
			for _, env := range container.Env {
				if env.Value != "" {
					continue
				}
				sp := strings.Split(env.RefName, "/")
				if sp[0] == "config" {
					var cfg corev1.ConfigMap
					err := r.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: sp[1]}, &cfg)
					if err != nil {
						if apiErrors.IsNotFound(err) {
							checks[env.RefName] = "NotFound"
						}
						r.logger.Debug("failed as ", err)
						return reconcileResult.Retry(minCoolingTime)
					}
					if _, ok := cfg.Data[env.RefKey]; !ok {
						checks[fmt.Sprintf("%s/%s", env.RefName, env.RefKey)] = "NotFound"
					}
				}

				if sp[0] == "secret" {
					var scrt corev1.Secret
					err := r.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: sp[1]}, &scrt)
					if err != nil {
						if apiErrors.IsNotFound(err) {
							checks[env.RefName] = "NotFound"
						}
						r.logger.Debug("failed as", err)
						return reconcileResult.Retry(minCoolingTime)
					}
					if _, ok := scrt.Data[env.RefKey]; !ok {
						checks[fmt.Sprintf("%s/%s", env.RefName, env.RefKey)] = "NotFound"
					}
				}
			}
		}

		app.Status.DependencyChecked = &checks
		return r.updateStatus(ctx, app)
	}

	if app.HasNotCheckedImages() {
		logger.Debug("app.HasNotCheckedImages()")
		pContainers := []corev1.Container{}
		for _, container := range app.Spec.Containers {
			pContainers = append(pContainers, corev1.Container{
				Name:  container.Name,
				Image: container.Image,
			})
		}

		r.logger.Debugf("pContainers: %+v", pContainers)

		pod, err := r.ClientSet.CoreV1().Pods(app.Namespace).Create(ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "app-sanity-check-",
				Namespace:    app.Namespace,
			},
			Spec: corev1.PodSpec{
				Containers: pContainers,
			},
		}, metav1.CreateOptions{})

		if err != nil {
			return reconcileResult.RetryE(maxCoolingTime, errors.NewEf(err, "could not create pod"))
		}

		app.Status.ImagesCheckJob = &crdsv1.ReconPod{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		}
		return r.updateStatus(ctx, app)
	}

	if app.IsCheckingImages() {
		logger.Debug("app.IsCheckingImages()")
		// ASSERT: read status of pod to check for image pull error
		var pod corev1.Pod
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: app.Status.ImagesCheckJob.Namespace,
			Name:      app.Status.ImagesCheckJob.Name,
		}, &pod); err != nil {
			return reconcileResult.Failed()
		}

		hasImageErrors := []bool{}
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Running != nil {
				hasImageErrors = append(hasImageErrors, false)
			}
			if cs.State.Waiting != nil {
				if cs.State.Waiting.Reason == "ImagePullBackOff" {
					app.Status.ImagesCheckJob.Failed = fmt.Sprintf("container (%s) can not pull image", cs.Name)
					r.logger.Debug(app.Status.ImagesCheckJob.Failed)
					app.Status.ImagesCheckCompleted = newBool(true)
					if err := r.ClientSet.CoreV1().Pods(app.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
						return reconcileResult.Retry(minCoolingTime)
					}
					return r.updateStatus(ctx, app)
				}
			}
		}

		if len(hasImageErrors) == len(pod.Spec.Containers) {
			hasErr := false
			for _, item := range hasImageErrors {
				hasErr = hasErr || item
			}
			if !hasErr {
				if err := r.ClientSet.CoreV1().Pods(app.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
					return reconcileResult.Retry(minCoolingTime)
				}
				app.Status.ImagesCheckJob = nil
				app.Status.ImagesCheckCompleted = newBool(true)
				return r.updateStatus(ctx, app)
			}
		}
		return reconcileResult.Retry(minCoolingTime)
	}

	// STEP: create new job
	if app.ShouldCreateJob() {
		logger.Debug("app.ShouldCreateJob()")
		specB, err := json.Marshal(app.Spec)
		if err != nil {
			r.logger.Error(errors.New("could not unmarshal app spec into []byte"))
			return reconcileResult.Failed()
		}
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
		return r.updateStatus(ctx, app)
	}

	return reconcileResult.OK()
}

func (r *AppReconciler) updateStatus(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
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
			app.Status.JobCompleted = nil
			return r.updateStatus(ctx, app)
		}

		if app.ShouldCreateDeletionJob() {
			r.logger.Debug("app.ShouldCreateDeletionJob()")
			// specB, err := json.Marshal(app.Spec)
			_, err := json.Marshal(app.Spec)
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
				// Args: []string{
				// "delete",
				// "--name", app.Name,
				// "--namespace", app.Namespace,
				// "--spec", string(specB),
				// },
				Command: []string{
					"bash",
				},
				Args: []string{
					"-c",
					"sleep 5 && exit 1",
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

		if app.HasDeletionJob() {
			r.logger.Debug("app.HasDeletionJob()")
			//STEP:  WATCH for it
			jobStatus, err := r.JobMgr.HasCompleted(ctx, app.Status.DeletionJob.Namespace, app.Status.DeletionJob.Name)
			if err != nil {
				return reconcileResult.Retry(minCoolingTime)
			}
			if jobStatus != nil {
				app.Status.DeletionJobCompleted = newBool(true)
				if !*jobStatus {
					r.logger.Debugf("DELETION jobStatus %v", *jobStatus)
					return r.updateStatus(ctx, app)
				}
				r.logger.Debugf("DELETION jobStatus: %v", *jobStatus)
				app.Status.DeletionJob = nil
				return r.updateStatus(ctx, app)
			}
			return reconcileResult.Retry(minCoolingTime)
			// STEP: remove finalizer once done
		}

		if app.Status.DeletionJobCompleted != nil {
			r.logger.Debug("app.Status.DeletionJobCompleted")
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
