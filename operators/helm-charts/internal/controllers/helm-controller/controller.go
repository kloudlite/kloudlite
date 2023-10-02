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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateJobRBAC             []byte
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
	ensureJobRBAC             string = "ensure-job-rbac"
)

const (
	LabelInstallOrUpgradeJob string = "kloudlite.io/chart-install-or-upgrade-job"
	LabelResourceGeneration  string = "kloudlite.io/resource-generation"
	LabelUninstallJob        string = "kloudlite.io/chart-uninstall-job"
	LabelHelmChartName       string = "kloudlite.io/helm-chart.name"
)

func getJobName(resName string) string {
	return fmt.Sprintf("helm-job-%s", resName)
}

func getJobSvcAccountName() string {
	return "helm-job-svc-account"
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

	if step := r.ensureJobRBAC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.startInstallJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	checkName := "finalize"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.startUninstallJob(req); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) ensureJobRBAC(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ensureJobRBAC)
	defer req.LogPostCheck(ensureJobRBAC)

	b, err := templates.ParseBytes(r.templateJobRBAC, map[string]any{
		"service-account-name":      getJobSvcAccountName(),
		"service-account-namespace": r.Env.RunningInNamespace,
	})
	if err != nil {
		return req.CheckFailed(ensureJobRBAC, check, err.Error()).Err(nil)
	}

	if _, err = r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(ensureJobRBAC, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[ensureJobRBAC] {
		obj.Status.Checks[ensureJobRBAC] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) startInstallJob(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(installOrUpgradeJob)
	defer req.LogPostCheck(installOrUpgradeJob)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(r.Env.RunningInNamespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		b, err := templates.ParseBytes(r.templateInstallOrUpgradeJob, map[string]any{
			"job-name":      getJobName(obj.Name),
			"job-namespace": r.Env.RunningInNamespace,
			"labels": map[string]string{
				LabelInstallOrUpgradeJob: "true",
				LabelHelmChartName:       obj.Name,
				LabelResourceGeneration:  fmt.Sprintf("%d", obj.Generation),
			},
			"service-account-name": getJobSvcAccountName(),

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

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(installOrUpgradeJob, check, err.Error())
		}

		req.AddToOwnedResources(rr...)
		return req.Done().RequeueAfter(1 * time.Second).Err(fmt.Errorf("waiting for job to be created")).Err(nil)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelInstallOrUpgradeJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(installOrUpgradeJob, check, fmt.Sprintf("waiting for previous jobs to finish execution")).Err(nil)
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(installOrUpgradeJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(installOrUpgradeJob, check, "waiting for job to finish execution").Err(nil)
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0

	if check != obj.Status.Checks[installOrUpgradeJob] {
		obj.Status.Checks[installOrUpgradeJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) startUninstallJob(req *rApi.Request[*crdsv1.HelmChart]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(uninstallJob)
	defer req.LogPostCheck(uninstallJob)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(r.Env.RunningInNamespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		b, err := templates.ParseBytes(r.templateUninstallJob, map[string]any{
			"job-name":      getJobName(obj.Name),
			"job-namespace": r.Env.RunningInNamespace,
			"labels": map[string]string{
				LabelHelmChartName:      obj.Name,
				LabelUninstallJob:       "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"service-account-name": getJobSvcAccountName(),

			"release-name":      obj.Name,
			"release-namespace": obj.Namespace,
		})
		if err != nil {
			return req.CheckFailed(uninstallJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(uninstallJob, check, err.Error()).Err(nil)
		}

		req.AddToOwnedResources(rr...)
		return req.Done().RequeueAfter(1 * time.Second).Err(fmt.Errorf("waiting for job to be created")).Err(nil)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelUninstallJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(uninstallJob, check, fmt.Sprintf("waiting for previous jobs to finish execution")).Err(nil)
		}
		// deleting that job
		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(uninstallJob, check, err.Error())
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(uninstallJob, check, "waiting for job to finish execution").Err(nil)
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[uninstallJob] {
		obj.Status.Checks[uninstallJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateJobRBAC, err = templatesDir.ReadFile("templates/job-rbac.yml.tpl")
	if err != nil {
		return err
	}

	r.templateInstallOrUpgradeJob, err = templatesDir.ReadFile("templates/install-or-upgrade-job.yml.tpl")
	if err != nil {
		return err
	}

	r.templateUninstallJob, err = templatesDir.ReadFile("templates/uninstall-job.yml.tpl")
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.HelmChart{})
	builder.Watches(
		&source.Kind{Type: &batchv1.Job{}},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			r.logger.Debugf("watch event received for job: %s/%s", obj.GetNamespace(), obj.GetName())
			if obj.GetNamespace() == r.Env.RunningInNamespace && obj.GetLabels()[LabelHelmChartName] != "" {
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetLabels()[LabelHelmChartName])}}
			}
			return nil
		}))
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
