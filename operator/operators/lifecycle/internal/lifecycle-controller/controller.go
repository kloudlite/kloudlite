package lifecycle_controller

import (
	"context"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/lifecycle/internal/env"
	"github.com/kloudlite/operator/operators/lifecycle/internal/lifecycle-controller/templates"
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

	templateJobRBAC []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	EnsureJobRBAC string = "ensure-job-rbac"
	ApplyK8sJob   string = "apply-k8s-job"
	DeleteK8sJob  string = "delete-k8s-job"
)

const (
	AnnApplyJobPhase  string = "kloudlite.io/job-apply-phase"
	AnnDeleteJobPhase string = "kloudlite.io/job-delete-phase"
)

var ApplyCheckList = []rApi.CheckMeta{
	{Name: EnsureJobRBAC, Title: "Ensures K8s Lifecycle RBACs"},
	{Name: ApplyK8sJob, Title: "Apply Kubernetes Lifecycle"},
}

// DefaultsPatched string = "defaults-patched"
var DeleteCheckList = []rApi.CheckMeta{
	{Name: EnsureJobRBAC, Title: "Ensures K8s Lifecycle RBACs"},
	{Name: DeleteK8sJob, Title: "Delete Kubernetes Lifecycle"},
}

func getJobSvcAccountName() string {
	return "job-svc-account"
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.Lifecycle{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

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

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureJobRBAC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.applyK8sJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Lifecycle]) stepResult.Result {
	rApi.NewRunningCheck("finalizing", req)

	if step := req.EnsureCheckList(DeleteCheckList); !step.ShouldProceed() {
		return step
	}

	if step := r.deleteK8sJob(req); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) ensureJobRBAC(req *rApi.Request[*crdsv1.Lifecycle]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(EnsureJobRBAC, req)

	jobSvcAcc := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: getJobSvcAccountName(), Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, jobSvcAcc, func() error {
		if jobSvcAcc.Annotations == nil {
			jobSvcAcc.Annotations = make(map[string]string, 1)
		}
		jobSvcAcc.Annotations[constants.DescriptionKey] = "Service account used by kloudlite jobs to run apply/delete k8s jobs"
		return nil
	}); err != nil {
		return check.StillRunning(err)
	}

	crb := rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-rb", getJobSvcAccountName())}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &crb, func() error {
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
	}); err != nil {
		return check.StillRunning(err)
	}

	return check.Completed()
}

func (r *Reconciler) applyK8sJob(req *rApi.Request[*crdsv1.Lifecycle]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ApplyK8sJob, req)

	if v, ok := obj.Status.Checks[ApplyK8sJob]; ok && v.Generation == obj.Generation && (v.State == rApi.CompletedState || v.State == rApi.ErroredState) {
		return check.Completed()
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, obj.Name), job); err != nil {
		job = nil
	}

	if job == nil {
		obj.Spec.OnApply.PodSpec.ServiceAccountName = getJobSvcAccountName()
		if obj.Spec.OnApply.PodSpec.RestartPolicy == "" {
			obj.Spec.OnApply.PodSpec.RestartPolicy = corev1.RestartPolicyNever
		}
		if obj.Spec.OnApply.BackOffLimit == nil {
			obj.Spec.OnApply.BackOffLimit = fn.New(int32(1))
		}

		jobTemplate := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:            obj.Name,
				Namespace:       obj.Namespace,
				Labels:          obj.GetLabels(),
				Annotations:     fn.MapMerge(obj.GetAnnotations(), map[string]string{AnnApplyJobPhase: fmt.Sprintf("%d", obj.Generation)}),
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
			Spec: batchv1.JobSpec{
				BackoffLimit: obj.Spec.OnApply.BackOffLimit,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      obj.GetLabels(),
						Annotations: obj.GetAnnotations(),
					},
					Spec: obj.Spec.OnApply.PodSpec,
				},
			},
		}

		if err := r.Create(ctx, jobTemplate); err != nil {
			return check.Failed(err)
		}

		req.AddToOwnedResources(rApi.ParseResourceRef(jobTemplate))
		return req.Done().RequeueAfter(1 * time.Second).Err(fmt.Errorf("waiting for job to be created")).Err(nil)
	}

	isMyJob := job.Annotations[AnnApplyJobPhase] == fmt.Sprintf("%d", obj.Generation)

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return check.StillRunning(fmt.Errorf("waiting for previous generation job to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return check.StillRunning(err)
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if job.Status.Active > 0 {
		obj.Status.Phase = crdsv1.JobPhaseRunning
		return check.StillRunning(fmt.Errorf("waiting for job to finish execution"))
	}

	if job.Status.Succeeded > 0 {
		obj.Status.Phase = crdsv1.JobPhaseSucceeded
		return check.Completed()
	}

	// check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	if job.Status.Failed > 0 {
		obj.Status.Phase = crdsv1.JobPhaseFailed
		return check.Failed(fmt.Errorf("job failed"))
	}

	obj.Status.Phase = crdsv1.JobPhasePending
	return check.StillRunning(fmt.Errorf("job is pending, waiting for job to start"))
}

func (r *Reconciler) deleteK8sJob(req *rApi.Request[*crdsv1.Lifecycle]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(DeleteK8sJob, req)

	if obj.Spec.OnDelete == nil {
		return check.Completed()
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, obj.Name), job); err != nil {
		job = nil
	}

	if job == nil {
		obj.Spec.OnDelete.PodSpec.ServiceAccountName = getJobSvcAccountName()
		if obj.Spec.OnDelete.PodSpec.RestartPolicy == "" {
			obj.Spec.OnDelete.PodSpec.RestartPolicy = corev1.RestartPolicyNever
		}
		jobTemplate := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:            obj.Name,
				Namespace:       obj.Namespace,
				Labels:          obj.GetLabels(),
				Annotations:     fn.MapMerge(obj.GetAnnotations(), map[string]string{AnnDeleteJobPhase: fmt.Sprintf("%d", obj.Generation)}),
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
			Spec: batchv1.JobSpec{
				BackoffLimit: obj.Spec.OnDelete.BackOffLimit,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      obj.GetLabels(),
						Annotations: obj.GetAnnotations(),
					},
					Spec: obj.Spec.OnDelete.PodSpec,
				},
			},
		}

		if err := r.Create(ctx, jobTemplate); err != nil {
			return check.Failed(err)
		}

		req.AddToOwnedResources(rApi.ParseResourceRef(jobTemplate))
		return req.Done().RequeueAfter(1 * time.Second).Err(fmt.Errorf("waiting for deletion job to be created")).Err(nil)
	}

	isMyJob := job.Annotations[AnnDeleteJobPhase] == fmt.Sprintf("%d", obj.Generation)

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return check.StillRunning(fmt.Errorf("waiting for previous generation job to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return check.StillRunning(err)
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if job.Status.Active > 0 {
		obj.Status.Phase = crdsv1.JobPhaseRunning
		return check.StillRunning(fmt.Errorf("waiting for job to finish execution"))
	}

	if job.Status.Succeeded > 0 {
		obj.Status.Phase = crdsv1.JobPhaseSucceeded
		return check.Completed()
	}

	// check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	if job.Status.Failed > 0 {
		obj.Status.Phase = crdsv1.JobPhaseFailed
		return check.Failed(fmt.Errorf("job failed"))
	}

	obj.Status.Phase = crdsv1.JobPhasePending
	return check.StillRunning(fmt.Errorf("job is pending, waiting for job to start"))
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateJobRBAC, err = templates.Read(templates.JobRBACTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Lifecycle{}).Owns(&batchv1.Job{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
