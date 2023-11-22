package helm_controller

import (
	"context"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/helm-charts/internal/controllers/helm-controller/templates"
	"github.com/kloudlite/operator/operators/helm-charts/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	job_manager "github.com/kloudlite/operator/pkg/job-helper"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	return ctrl.Result{}, nil
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

	jobSvcAcc := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: getJobSvcAccountName(), Namespace: obj.Namespace}}

	controllerutil.CreateOrUpdate(ctx, r.Client, jobSvcAcc, func() error {
		if jobSvcAcc.Annotations == nil {
			jobSvcAcc.Annotations = make(map[string]string, 1)
		}
		jobSvcAcc.Annotations[constants.DescriptionKey] = "Service account used by helm charts to run helm release jobs"
		return nil
	})

	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-rb", getJobSvcAccountName())}}
	controllerutil.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		if crb.Annotations == nil {
			crb.Annotations = make(map[string]string, 1)
		}
		crb.Annotations[constants.DescriptionKey] = "Cluster role binding used by helm charts to run helm release jobs"

		crb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		}

		found := false
		for i := range crb.Subjects {
			if crb.Subjects[i].Namespace == obj.Namespace && crb.Subjects[i].Name == getJobSvcAccountName() {
				found = true
				break
			}
		}
		if !found {
			crb.Subjects = append(crb.Subjects, rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      getJobSvcAccountName(),
				Namespace: obj.Namespace,
			})
		}
		return nil
	})

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
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		b, err := templates.ParseBytes(r.templateInstallOrUpgradeJob, map[string]any{
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelInstallOrUpgradeJob: "true",
				LabelHelmChartName:       obj.Name,
				LabelResourceGeneration:  fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"service-account-name": getJobSvcAccountName(),
			"tolerations":          obj.Spec.JobVars.Tolerations,
			"affinity":             obj.Spec.JobVars.Affinity,
			"node-selector":        obj.Spec.JobVars.NodeSelector,
			"backoff-limit":        obj.Spec.JobVars.BackOffLimit,

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
			return req.CheckFailed(installOrUpgradeJob, check, "waiting for previous jobs to finish execution").Err(nil)
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
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		b, err := templates.ParseBytes(r.templateUninstallJob, map[string]any{
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelHelmChartName:      obj.Name,
				LabelUninstallJob:       "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs":           []metav1.OwnerReference{fn.AsOwner(obj, true)},
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
			return req.CheckFailed(uninstallJob, check, "waiting for previous jobs to finish execution").Err(nil)
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
	r.templateJobRBAC, err = templates.Read(templates.JobRBACTemplate)
	if err != nil {
		return err
	}

	r.templateInstallOrUpgradeJob, err = templates.Read(templates.HelmInstallOrUpgradeJobTemplate)
	if err != nil {
		return err
	}

	r.templateUninstallJob, err = templates.Read(templates.HelmUninstallJobTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.HelmChart{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
