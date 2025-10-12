package helmchart

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/controllers/helmchart/templates"
	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

const (
	helmChartFinalizer = "helmcharts.environments.kloudlite.io/finalizer"
	helmJobImage       = "alpine/helm:latest" // Helm 3 image

	jobServiceAccountName = "helm-job-sa"
)

// Reconciler reconciles HelmChart objects
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	HelmJobRunnerImage string
}

// Reconcile handles HelmChart events
func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	req, err := reconciler.NewRequest[*environmentsv1.HelmChart](ctx, r.Client, request.NamespacedName, &environmentsv1.HelmChart{})
	if err != nil {
		slog.Error("failed here", "err", err)
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*environmentsv1.HelmChart]{
		{
			Name:     "setup k8s job RBAC",
			Title:    "Setup Kubernetes RBAC for running helm job",
			OnCreate: r.createHelmJobRBAC,
			OnDelete: nil,
		},
		{
			Name:     "setup helm release job",
			Title:    "Setup Helm Release Job",
			OnCreate: r.createHelmInstallJob,
			OnDelete: r.createHelmUninstallJob,
		},
		// {
		// 	Name:     "process exports",
		// 	Title:    "Process helm chart exports",
		// 	OnCreate: r.createProcessExports,
		// 	OnDelete: r.deleteProcessedExports,
		// },
	})
}

type jobOpType string

const (
	InstallOp   jobOpType = "install"
	UninstallOp jobOpType = "uninstall"
)

func (r *Reconciler) createHelmJobRBAC(check *reconciler.Check[*environmentsv1.HelmChart], obj *environmentsv1.HelmChart) reconciler.StepResult {
	jobSvcAcc := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: jobServiceAccountName, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, jobSvcAcc, func() error {
		if jobSvcAcc.Annotations == nil {
			jobSvcAcc.Annotations = make(map[string]string, 1)
		}
		jobSvcAcc.Annotations[reconciler.AnnotationDescriptionKey] = "Service account used by helm charts to run helm release jobs"
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: jobServiceAccountName + "-rb"}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, &crb, func() error {
		if crb.Annotations == nil {
			crb.Annotations = make(map[string]string, 1)
		}
		crb.Annotations[reconciler.AnnotationDescriptionKey] = "Cluster role binding used by helm charts to run helm release jobs"

		crb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		}

		found := false
		for i := range crb.Subjects {
			if crb.Subjects[i].Namespace == obj.Namespace && crb.Subjects[i].Name == jobServiceAccountName {
				found = true
				break
			}
		}
		if !found {
			crb.Subjects = append(crb.Subjects, rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      jobServiceAccountName,
				Namespace: obj.Namespace,
			})
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func valuesToYaml(values map[string]any) (string, error) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	m := make(map[string]any, len(values))
	for _, k := range keys {
		m[k] = values[k]
	}

	slog.Info("values:", "m", m)

	b, err := yaml.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *Reconciler) runHelmChartJob(check *reconciler.Check[*environmentsv1.HelmChart], obj *environmentsv1.HelmChart, jobSpec []byte, op jobOpType) reconciler.StepResult {
	jobTypeAnnKey := "kloudlite.io/job.type"
	jobTypeAnnValue := fmt.Sprintf("%s/%d", op, obj.GetGeneration())
	name := fmt.Sprintf("%s-helmchart-job", obj.Name)

	job := &batchv1.Job{}
	if err := r.Client.Get(check.Context(), types.NamespacedName{Namespace: obj.Namespace, Name: name}, job); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		job := &batchv1.Job{
			TypeMeta: metav1.TypeMeta{Kind: "Job", APIVersion: "batch/v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: obj.Namespace,
				Labels:    obj.GetLabels(),
				Annotations: fn.MapMerge(
					fn.MapFilter(obj.GetAnnotations(), func(k, v string) bool {
						return strings.HasPrefix(k, reconciler.ObservabilityAnnotationKey)
					}),
					map[string]string{jobTypeAnnKey: jobTypeAnnValue},
				),
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
		}
		if err := yaml.Unmarshal(jobSpec, &job.Spec); err != nil {
			return check.Failed(err)
		}

		if err := r.Client.Create(check.Context(), job); err != nil {
			return check.Failed(err)
		}

		return check.Abort(fn.JobStatusPending.Message())
	}

	jobStatus := fn.ParseJobStatus(job)

	if job.Annotations[jobTypeAnnKey] != jobTypeAnnValue {
		// we need to wait for this job to finish first
		if jobStatus == fn.JobStatusSucceeded || jobStatus == fn.JobStatusFailed {
			if err := r.Delete(check.Context(), job); err != nil {
				return nil
			}

			return check.UpdateMsg("cleaning up previous/leftover job from old reconcilations").RequeueAfter(200 * time.Millisecond)
		}

		return check.Abort("waiting for previous/leftover job from old reconcilation to finish")
	}

	if jobStatus != fn.JobStatusSucceeded {
		return check.Abort(jobStatus.Message())
	}
	return check.Passed()
}

func (r *Reconciler) createHelmInstallJob(check *reconciler.Check[*environmentsv1.HelmChart], obj *environmentsv1.HelmChart) reconciler.StepResult {
	// Unmarshal helm values into a map
	helmValues := make(map[string]any)
	if len(obj.Spec.HelmValues.Raw) > 0 {
		if err := json.Unmarshal(obj.Spec.HelmValues.Raw, &helmValues); err != nil {
			return check.Failed(fmt.Errorf("unmarshaling helm values: %w", err))
		}
	}

	// Add timestamp to trigger upgrade
	helmValues["kloudlite"] = map[string]any{
		// INFO: this ensures that whenever this function runs, helm resource is actually upgraded,
		// i.e. missing helm resources or accidently deleted resources will be recreated
		"triggered-at": time.Now().Format(time.RFC3339),
	}

	values, err := valuesToYaml(helmValues)
	if err != nil {
		return check.Failed(fmt.Errorf("converting helm values to YAML: %w", err))
	}

	jobVars := obj.Spec.HelmJobVars
	if jobVars == nil {
		jobVars = &environmentsv1.HelmJobVars{}
	}

	b, err := templates.HelmInstallJobTemplate.Render(templates.HelmChartInstallJobSpecParams{
		PodAnnotations:     fn.MapFilter(obj.GetAnnotations(), reconciler.ObservabilityAnnotationFilter),
		ReleaseName:        obj.Name,
		ReleaseNamespace:   obj.Namespace,
		Image:              r.HelmJobRunnerImage,
		BackOffLimit:       1,
		ServiceAccountName: jobServiceAccountName,
		Tolerations:        jobVars.Tolerations,
		Affinity:           jobVars.Affinity,
		NodeSelector:       jobVars.NodeSelector,
		ChartRepoURL:       obj.Spec.Chart.URL,
		ChartName:          obj.Spec.Chart.Name,
		ChartVersion:       obj.Spec.Chart.Version,
		PreInstall:         obj.Spec.PreInstall,
		PostInstall:        obj.Spec.PostInstall,
		HelmValuesYAML:     values,
	})
	if err != nil {
		return check.Failed(err)
	}

	return r.runHelmChartJob(check, obj, b, InstallOp)
}

func (r *Reconciler) createHelmUninstallJob(check *reconciler.Check[*environmentsv1.HelmChart], obj *environmentsv1.HelmChart) reconciler.StepResult {
	jobVars := obj.Spec.HelmJobVars
	if jobVars == nil {
		jobVars = &environmentsv1.HelmJobVars{}
	}

	b, err := templates.HelmUninstallJobTemplate.Render(templates.HelmChartUninstallJobSpecParams{
		PodAnnotations:     fn.MapFilter(obj.GetAnnotations(), reconciler.ObservabilityAnnotationFilter),
		ReleaseName:        obj.Name,
		ReleaseNamespace:   obj.Namespace,
		Image:              r.HelmJobRunnerImage,
		BackOffLimit:       1,
		ServiceAccountName: jobServiceAccountName,
		Tolerations:        jobVars.Tolerations,
		Affinity:           jobVars.Affinity,
		NodeSelector:       jobVars.NodeSelector,
		ChartRepoURL:       obj.Spec.Chart.URL,
		ChartName:          obj.Spec.Chart.Name,
		ChartVersion:       obj.Spec.Chart.Version,
		PreUninstall:       obj.Spec.PreUninstall,
		PostUninstall:      obj.Spec.PostUninstall,
	})
	if err != nil {
		return check.Failed(err)
	}

	return r.runHelmChartJob(check, obj, b, UninstallOp)
}

// SetupWithManager sets up the controller with the Manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.HelmJobRunnerImage == "" {
		return fmt.Errorf("HelmJobRunnerImage is required but not set")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.HelmChart{}).
		Owns(&batchv1.Job{}). // Watch Jobs owned by HelmChart
		Complete(r)
}
