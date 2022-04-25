package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"fmt"

	"go.uber.org/zap"
	// batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	// appsv1 "k8s.io/api/apps/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

func newBool(b bool) *bool {
	return &b
}

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ClientSet   *kubernetes.Clientset
	JobMgr      lib.Job
	logger      *zap.SugaredLogger
	SendMessage func(key string, msg lib.MessageReply) error

	HarborUserName string
	HarborPassword string
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := GetLogger(req.NamespacedName)
	r.logger = logger.With("Name", req.NamespacedName)

	app := &crdsv1.App{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.FailedE(err)
	}

	if app.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, app)
	}

	kt, err := templates.UseTemplate(templates.App)
	if err != nil {
		logger.Error(err, "could not useTemplate, aborting...")
		return reconcileResult.Failed()
	}
	b, err := kt.WithValues(app)
	if err != nil {
		logger.Info(b, err)
	}

	kt2, err := templates.UseTemplate(templates.Service)
	if err != nil {
		logger.Error(err, "could not useTemplate, aborting...")
		return reconcileResult.Failed()
	}
	b2, err := kt2.WithValues(app)
	if err != nil {
		logger.Info(b, err)
	}

	logger.Info("####################################################################################################################33")
	logger.Infof("app:\n%+v\n", string(b))
	logger.Infof("service:\n%+v\n", string(b2))
	logger.Info("####################################################################################################################33")

	var ry unstructured.Unstructured
	if err = yaml.Unmarshal(b, &ry.Object); err != nil {
		logger.Infof(errors.NewEf(err, "could not convert template %s []byte into mongodb", templates.App).Error())
		return reconcileResult.Failed()
	}
	m := new(unstructured.Unstructured)
	m.Object = map[string]interface{}{
		"apiVersion": ry.Object["apiVersion"],
		"kind":       ry.Object["kind"],
		"metadata":   ry.Object["metadata"],
		"spec":       ry.Object["spec"],
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, m, func() error {
		m = m.DeepCopy()
		m.Object["spec"] = ry.Object["spec"]

		if err = controllerutil.SetControllerReference(app, m, r.Scheme); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not create/update resource"))
	}

	var ry2 unstructured.Unstructured
	if err = yaml.Unmarshal(b, &ry2.Object); err != nil {
		logger.Infof(errors.NewEf(err, "could not convert template %s []byte into mongodb", templates.App).Error())
		return reconcileResult.Failed()
	}
	m2 := new(unstructured.Unstructured)
	m2.Object = map[string]interface{}{
		"apiVersion": ry2.Object["apiVersion"],
		"kind":       ry2.Object["kind"],
		"metadata":   ry2.Object["metadata"],
		"spec":       ry2.Object["spec"],
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, m2, func() error {
		m2 = m2.DeepCopy()
		m2.Object["spec"] = ry2.Object["spec"]

		if err = controllerutil.SetControllerReference(app, m2, r.Scheme); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not create/update resource"))
	}

	// d = appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMea,
	// }

	return reconcileResult.OK()
}

func (r *AppReconciler) finalize(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
	return reconcileResult.OK()
}

func (r *AppReconciler) Reconcile2(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)
	logger := GetLogger(req.NamespacedName)
	r.logger = logger.With("Name", req.NamespacedName)

	app := &crdsv1.App{}

	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if app.HasToBeDeleted() {
		return r.finalizeApp(ctx, app)
	}

	if app.IsNewGeneration() {
		if app.Status.ApplyJob.IsRunning() {
			return reconcileResult.Retry()
		}

		app.DefaultStatus()
		return r.updateStatus(ctx, app)
	}

	if app.Status.DependencyCheck.ShouldCheck() {
		r.logger.Debugf("app.Status.DependencyCheck.ShouldCheck()")
		app.Status.DependencyCheck.SetStarted()

		checks := r.checkDependency(ctx, app)

		if checks != nil {
			r.logger.Infof("Dependenct check failed ...")
			b, err := json.Marshal(*checks)
			if err != nil {
				r.logger.Infof("faied to marshal map[string]string into JSON []byte")
			}
			app.Status.DependencyCheck.SetFinishedWith(false, string(b))
			return r.updateStatus(ctx, app)
		}

		// ASSERT: DependencyCheck  passed
		app.Status.DependencyCheck.SetFinishedWith(true, "Dependency check passed")
		r.logger.Infof("Dependency check passed ...")
		return r.updateStatus(ctx, app)
	}

	if !app.Status.DependencyCheck.Status {
		r.logger.Infof("Dependency Check failed")
		if app.Status.DependencyCheck.ShouldRetry(maxCoolingTime) {
			return r.updateStatus(ctx, app)
		}
		return reconcileResult.Retry()
	}

	if app.Status.ImagesAvailability.ShouldCheck() {
		r.logger.Debugf("app.Status.ImagesAvailability.ShouldCheck()")
		app.Status.ImagesAvailability.SetStarted()

		checks, err := r.checkImagesAvailable(ctx, app)
		if err != nil {
			app.Status.ImagesAvailability.SetFinishedWith(false, err.Error())
			return r.updateStatus(ctx, app)
		}

		r.logger.Infof("CHECKS: %+v", checks)

		if checks == nil {
			return reconcileResult.Failed()
		}

		for k, c := range checks {
			if !c {
				r.logger.Infof("CHECKS c: %v\n", c)
				app.Status.ImagesAvailability.SetFinishedWith(false, fmt.Sprintf("image %s not available to cluster", k))
				return r.updateStatus(ctx, app)
			}
		}

		app.Status.ImagesAvailability.SetFinishedWith(true, "all images available to the cluster")
		return r.updateStatus(ctx, app)
	}

	// ASSERT: keep retying until you find the image
	if !app.Status.ImagesAvailability.Status {
		r.logger.Infof("ImagesAvailability check failed")
		if app.Status.ImagesAvailability.ShouldRetry(maxCoolingTime) {
			return r.updateStatus(ctx, app)
		}
		return reconcileResult.Retry()
	}

	if app.Status.ApplyJob.ShouldCheck() {
		r.logger.Debugf("app.Status.ApplyJob.ShouldCheck()")
		app.Status.ApplyJob.SetStarted()
		err := r.createAppJob(ctx, app)
		if err != nil {
			app.Status.ApplyJob.SetFinishedWith(false, "")
			return r.updateStatus(ctx, app)
		}
		return r.updateStatus(ctx, app)
	}

	if app.Status.ApplyJob.IsRunning() {
		r.logger.Debugf("app.Status.ApplyJob.IsRunning()")
		j := app.Status.ApplyJob.Job

		b, err := r.JobMgr.HasCompleted(ctx, j.Namespace, j.Name)
		if err != nil {
			app.Status.ApplyJob.SetFinishedWith(false, errors.NewEf(err, "job failed").Error())
			return r.updateStatus(ctx, app)
		}

		if b != nil {
			if !*b {
				app.Status.ApplyJob.SetFinishedWith(false, "")
				return r.updateStatus(ctx, app)
			}
			app.Status.ApplyJob.SetFinishedWith(true, "Ready")
			return r.updateStatus(ctx, app)
		}

		return reconcileResult.Retry()
	}

	return reconcileResult.OK()
}

func (r *AppReconciler) updateStatus(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
	app.BuildConditions()
	// fmt.Printf("####################\nAPP.Conditions: %+v\n", app.Status.Conditions)
	// fmt.Printf("#############\nAPP.Status: %+v\n", app.Status)

	b, err := json.Marshal(app.Status.Conditions)
	if err != nil {
		r.logger.Debug(err)
		b = []byte("")
	}

	err = r.SendMessage(fmt.Sprintf("%s/%s/%s", app.Namespace, "app", app.Name), lib.MessageReply{
		Message: string(b),
		Status:  false,
	})

	if err != nil {
		r.logger.Infof("unable to send kafka reply message")
	}

	if err := r.Status().Update(ctx, app); err != nil {
		return reconcileResult.RetryE(maxCoolingTime, errors.StatusUpdate(err))
	}

	r.logger.Debug("App has been updated")
	return ctrl.Result{}, nil
}

func (r *AppReconciler) finalizeApp(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
	logger := r.logger.With("FINALIZER", "true")
	if controllerutil.ContainsFinalizer(app, finalizers.App.String()) {
		// STEP: cleaning currently executing jobs
		if app.Status.ApplyJob.IsRunning() {
			r.logger.Debugf("[Finalizer]: killing app.Status.ApplyJob.IsRunning()")
			j := app.Status.ApplyJob.Job
			err := r.JobMgr.Delete(ctx, j.Namespace, j.Name)
			if err != nil {
				logger.Error(errors.NewEf(err, "error deleting job %s/%s, silently exiting ", j.Namespace, j.Name))
			}
			app.Status.ApplyJob.SetFinishedWith(false, "killed by APP finalizer")
			return r.updateStatus(ctx, app)
		}

		if app.Status.DeleteJob.ShouldCheck() {
			logger.Debugf("[Finalizer]: app.Status.DeleteJob.ShouldCheck()")
			app.Status.DeleteJob.SetStarted()
			specB, err := json.Marshal(app.Spec)
			if err != nil {
				app.Status.DeleteJob.SetFinishedWith(false, "could not unmarshal app spec into []byte")
				return r.updateStatus(ctx, app)
			}

			job, err := r.JobMgr.Create(ctx, app.Namespace, &lib.JobVars{
				Name:           "delete-job",
				Namespace:      app.Namespace,
				ServiceAccount: SvcAccountName,
				// Image:           "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest",
				Image:           "harbor.dev.madhouselabs.io/ci/jobs/app:latest",
				ImagePullPolicy: "Always",
				Args: []string{
					"delete",
					"--name", app.Name,
					"--namespace", app.Namespace,
					"--spec", string(specB),
				},
			})

			if err != nil {
				app.Status.DeleteJob.SetFinishedWith(false, fmt.Sprintf("could not create deletion job as %s", err.Error()))
				return r.updateStatus(ctx, app)
			}

			app.Status.DeleteJob.Job = &crdsv1.ReconJob{
				Name:      job.Name,
				Namespace: job.Namespace,
			}
			return r.updateStatus(ctx, app)
		}

		if app.Status.DeleteJob.IsRunning() {
			r.logger.Debugf("[Finalizer]: app.Status.DeleteJob.IsRunning()")

			j := app.Status.DeleteJob.Job
			jobStatus, err := r.JobMgr.HasCompleted(ctx, j.Namespace, j.Name)
			if err != nil {
				// means child job is running
				app.Status.DeleteJob.SetFinishedWith(false, errors.NewEf(err, "job failed").Error())
				return r.updateStatus(ctx, app)
			}

			if jobStatus != nil {
				if !*jobStatus {
					r.logger.Debugf("DELETION jobStatus %v", *jobStatus)
					app.Status.DeleteJob.SetFinishedWith(false, "finalzing deletion job failed")
					return r.updateStatus(ctx, app)
				}
				r.logger.Debugf("DELETION jobStatus: %v", *jobStatus)
				app.Status.DeleteJob.SetFinishedWith(true, "finalzing deletion job succeeded")
				return r.updateStatus(ctx, app)
			}

			return reconcileResult.Retry()
		}

		r.logger.Debug("[Finalizer]: all deletion checks completed ...")
		controllerutil.RemoveFinalizer(app, finalizers.App.String())
		err := r.Update(ctx, app)
		if err != nil {
			eMsg := errors.NewEf(err, "could not remove finalizers from app")
			r.logger.Error(eMsg)
			return reconcileResult.Retry()
		}

		return reconcileResult.OK()
	}

	r.logger.Debug("contains no finalizers")
	return reconcileResult.OK()
}

func (r *AppReconciler) createAppJob(ctx context.Context, app *crdsv1.App) error {
	specB, err := json.Marshal(app.Spec)
	if err != nil {
		return errors.New("could not unmarshal app spec into []byte")
	}

	job, err := r.JobMgr.Create(ctx, app.Namespace, &lib.JobVars{
		Name:           "create-job",
		Namespace:      app.Namespace,
		ServiceAccount: SvcAccountName,
		// Image:           "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest",
		Image:           "harbor.dev.madhouselabs.io/ci/jobs/app:latest",
		ImagePullPolicy: "Always",
		Args: []string{
			"create",
			"--name", app.Name,
			"--namespace", app.Namespace,
			"--spec", string(specB),
		},
	})

	if err != nil {
		return errors.NewEf(err, "could not create job")
	}

	app.Status.ApplyJob.Job = &crdsv1.ReconJob{Namespace: job.Namespace, Name: job.Name}
	return nil
}

const ImageRegistry = "harbor.dev.madhouselabs.io"

func (r *AppReconciler) checkImagesAvailable(ctx context.Context, app *crdsv1.App) (map[string]bool, error) {
	checks := map[string]bool{}
	for _, c := range app.Spec.Containers {
		if !strings.HasPrefix(c.Image, ImageRegistry) {
			checks[c.Image] = true
			continue
		}

		imageName := strings.Replace(c.Image, fmt.Sprintf("%s/ci/", ImageRegistry), "", 1)
		artifact := strings.Split(imageName, ":")
		artifactName := artifact[0]
		tag := artifact[1]

		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("https://harbor.dev.madhouselabs.io/api/v2.0/projects/ci/repositories/%s/artifacts/%s/tags", url.QueryEscape(artifactName), tag),
			nil,
		)
		if err != nil {
			return nil, errors.NewEf(err, "could not create http object")
		}
		req.Header.Add("Content-Type", "application/vnd.oci.image.index.v1+json")
		req.SetBasicAuth(r.HarborUserName, r.HarborPassword)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.NewEf(err, "could not make http request")
		}
		r.logger.Infof("resp.StatusCode=%v", resp.StatusCode)
		checks[c.Image] = resp.StatusCode == 200
	}

	return checks, nil
}

func (r *AppReconciler) checkDependency(ctx context.Context, app *crdsv1.App) *map[string]string {
	checks := map[string]string{}
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
						return &checks
					}
					checks[env.RefName] = err.Error()
					return &checks
				}
				if _, ok := cfg.Data[env.RefKey]; !ok {
					checks[fmt.Sprintf("%s/%s", env.RefName, env.RefKey)] = "NotFound"
					return &checks
				}
			}

			if sp[0] == "secret" {
				var scrt corev1.Secret
				err := r.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: sp[1]}, &scrt)
				if err != nil {
					if apiErrors.IsNotFound(err) {
						checks[env.RefName] = "NotFound"
						return &checks
					}
					checks[env.RefName] = err.Error()
					return &checks
				}
				if _, ok := scrt.Data[env.RefKey]; !ok {
					checks[fmt.Sprintf("%s/%s", env.RefName, env.RefKey)] = "NotFound"
					return &checks
				}
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Complete(r)
}
