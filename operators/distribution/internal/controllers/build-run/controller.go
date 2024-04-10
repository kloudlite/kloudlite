package buildrun

import (
	"context"
	"fmt"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"github.com/kloudlite/operator/operators/distribution/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	JobCreated   string = "job-created"
	JobCompleted string = "job-completed"
	JobFailed    string = "job-failed"
	JobDeleted   string = "job-deleted"
)

var (
	ApplyChecklist = []rApi.CheckMeta{
		{Name: JobCreated, Title: "Job created for build"},
		{Name: JobCompleted, Title: "Job completed"},
	}

	DeleteChecklist = []rApi.CheckMeta{
		{Name: JobDeleted, Title: "Cleaning up resources"},
	}
)

// +kubebuilder:rbac:groups=distribution.kloudlite.io,resources=devices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=distribution.kloudlite.io,resources=devices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=distribution.kloudlite.io,resources=devices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &dbv1.BuildRun{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(ApplyChecklist); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureJobCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.provisionCreatedJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) ensureJobCreated(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(JobCreated, req)

	if obj.Status.Checks[JobCreated].Status {
		return check.Completed()
	}

	b, err := r.getBuildTemplate(req)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) provisionCreatedJob(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(JobCompleted, req)

	if obj.Status.Checks[JobCompleted].Status || obj.Status.Checks[JobFailed].Status {
		return req.Next()
	}

	j, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.BuildNamespace, fmt.Sprint("build-", obj.Name)), &batchv1.Job{})
	if err != nil {
		return check.Failed(err)
	}

	if j.Status.Active > 0 {
		return check.StillRunning(fmt.Errorf("job is running, and waiting for completion")).Err(nil)
	}

	if j.Status.Succeeded > 0 {
		return check.Completed()
	}

	if j.Status.Failed > 0 {
		return check.Failed(fmt.Errorf("job failed, please check logs")).Err(nil)
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(JobDeleted, req)

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	if sec, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), &corev1.Secret{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	} else if err := r.Delete(ctx, sec); err != nil {
		return check.Failed(err)
	}

	if job, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.BuildNamespace, fmt.Sprintf("build-%s", obj.Name)), &batchv1.Job{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
		return req.Finalize()
	} else if err := r.Delete(ctx, job); err != nil {
		return check.Failed(err)
	}

	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&dbv1.BuildRun{})

	watchList := []client.Object{
		&corev1.Secret{},
		&batchv1.Job{},
	}
	for _, object := range watchList {
		builder.Watches(
			object,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					if obj.GetNamespace() != r.Env.BuildNamespace {
						return nil
					}

					if brn, ok := obj.GetLabels()[constants.BuildRunNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetLabels()[buildrunNamespaceAnn], brn)}}
					}
					return nil
				}),
		)
	}

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
