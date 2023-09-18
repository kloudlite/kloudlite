package helm_controller

import (
	"context"
	"embed"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/helm-charts/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	job_manager "github.com/kloudlite/operator/pkg/job-helper"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateInstallOrUpgradeJob []byte
	templateUninstallJob        []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	installOrUpgradeJob       string = "install-or-upgrade-job"
	uninstallJob              string = "uninstall-job"
	checkJobStatus            string = "check-job-status"
	waitForPrevJobsToComplete string = "wait-for-prev-jobs-to-complete"
)

const (
	LabelInstallOrUpgradeJob string = "kloudlite.io/chart-install-or-upgrade-job"
	LabelUninstallJob        string = "kloudlite.io/chart-uninstall-job"
)

func getJobName(resName string) string {
	return fmt.Sprintf("helm-job-%s", resName)
}

//go:embed templates/*
var templatesDir embed.FS

// +kubebuilder:rbac:groups=helm.kloudlite.io,resources=helmcharts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helm.kloudlite.io,resources=helmcharts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=helm.kloudlite.io,resources=helmcharts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.HelmChart{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.startInstallJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	if step := req.UpdateStatus(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	checkName := "finalize"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	check := rApi.Check{Generation: obj.Generation}

	if step := r.startUninstallJob(req); !step.ShouldProceed() {
		return step
	}

	for i := range obj.Status.Resources {
		res := obj.Status.Resources[i]
		resObj := unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": res.APIVersion,
				"kind":       res.Kind,
				"metadata": map[string]any{
					"name":      res.Name,
					"namespace": res.Namespace,
				},
			},
		}
		_ = resObj
		req.Logger.Infof("deleting child resource: apiVersion: %s, kind: %s, %s/%s", res.APIVersion, res.Kind, res.Namespace, res.Name)
		if err := r.Delete(ctx, &resObj); err != nil {
			if !errors.IsNotFound(err) {
				return req.CheckFailed(checkName, check, err.Error())
			}
		}
	}

	return req.Finalize()
}

func (r *Reconciler) startInstallJob(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(installOrUpgradeJob)
	defer req.LogPostCheck(installOrUpgradeJob)

	b, err := templates.ParseBytes(r.templateInstallOrUpgradeJob, map[string]any{
		"job-name":      getJobName(obj.Name),
		"job-namespace": obj.Namespace,
		"labels": map[string]string{
			LabelInstallOrUpgradeJob: "true",
		},
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

		"repo-url":  obj.Spec.ChartRepo.Url,
		"repo-name": obj.Spec.ChartRepo.Name,

		"chart-name":    obj.Spec.ChartName,
		"chart-version": obj.Spec.ChartVersion,

		"release-name":      obj.Name,
		"release-namespace": obj.Namespace,
		"values-yaml":       obj.Spec.ValuesYaml,
	})
	if err != nil {
		return req.CheckFailed(installOrUpgradeJob, check, err.Error()).Err(nil)
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	jobFound := false

	if job != nil {
		// handle job exists
		if job.Generation == obj.Generation && job.Labels[LabelInstallOrUpgradeJob] == "true" {
			jobFound = true
			// this is our job
			if !job_manager.HasJobFinished(ctx, r.Client, job) {
				return req.CheckFailed(installOrUpgradeJob, check, "waiting for job to finish execution")
			}
			tlog := job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
			check.Message = tlog
		} else {
			// it is someone else's job, wait for it to complete
			if !job_manager.HasJobFinished(ctx, r.Client, job) {
				return req.CheckFailed(installOrUpgradeJob, check, fmt.Sprintf("waiting for previous jobs to finish execution"))
			}

			if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
				return req.CheckFailed(installOrUpgradeJob, check, err.Error())
			}
		}
	}

	if !jobFound {
		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(installOrUpgradeJob, check, err.Error())
		}

		req.AddToOwnedResources(rr...)
		return req.Done().RequeueAfter(1 * time.Second).Err(fmt.Errorf("waiting for job to be created"))
	}

	check.Status = true
	if obj.Status.Checks == nil {
		obj.Status.Checks = map[string]rApi.Check{}
	}
	if check != obj.Status.Checks[installOrUpgradeJob] {
		obj.Status.Checks[installOrUpgradeJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) startUninstallJob(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(uninstallJob)
	defer req.LogPostCheck(uninstallJob)

	b, err := templates.ParseBytes(r.templateUninstallJob, map[string]any{
		"job-name":      getJobName(obj.Name),
		"job-namespace": obj.Namespace,
		"labels": map[string]string{
			LabelUninstallJob: "true",
		},
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

		"release-name":      obj.Name,
		"release-namespace": obj.Namespace,
	})
	if err != nil {
		return req.CheckFailed(uninstallJob, check, err.Error()).Err(nil)
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	jobFound := false
	if job != nil {
		// job exists
		if job.Generation == obj.Generation && job.Labels[LabelUninstallJob] == "true" {
			jobFound = true
			// this is our job
			if !job_manager.HasJobFinished(ctx, r.Client, job) {
				return req.CheckFailed(uninstallJob, check, "waiting for job to finish execution")
			}
			tlog := job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
			check.Message = tlog
		} else {
			// it is someone else's job, wait for it to complete
			if !job_manager.HasJobFinished(ctx, r.Client, job) {
				return req.CheckFailed(uninstallJob, check, fmt.Sprintf("waiting for previous jobs to finish execution"))
			}
			// deleting that job
			if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
				return req.CheckFailed(uninstallJob, check, err.Error())
			}
		}
	}

	if !jobFound {
		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(uninstallJob, check, err.Error()).Err(nil)
		}

		req.AddToOwnedResources(rr...)
	}

	check.Status = true
	if check != obj.Status.Checks[uninstallJob] {
		obj.Status.Checks[uninstallJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateInstallOrUpgradeJob, err = templatesDir.ReadFile("templates/install-or-upgrade-job.yml.tpl")
	if err != nil {
		return err
	}

	r.templateUninstallJob, err = templatesDir.ReadFile("templates/uninstall-job.yml.tpl")
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.HelmChart{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
