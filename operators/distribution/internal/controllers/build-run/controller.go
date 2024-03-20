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
	"k8s.io/apimachinery/pkg/labels"
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
	// CredsAvailable string = "creds-available"

	PVCReady     string = "pvc-ready"
	JobCreated   string = "job-created"
	JobCompleted string = "job-completed"
	JobFailed    string = "job-failed"
	JobDeleted   string = "job-deleted"
)

var (
	B_CHECKLIST = []rApi.CheckMeta{
		{Name: PVCReady, Title: "PVC ready for cache"},
		{Name: JobCreated, Title: "Job created for build"},
		{Name: JobCompleted, Title: "Job completed"},
		// {Name: CredsAvailable, Title: "credentials available"},
	}
	B_DESTROY_CHECKLIST = []rApi.CheckMeta{
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

	if step := req.EnsureCheckList(B_CHECKLIST); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(PVCReady, JobCreated, JobCompleted, JobFailed); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensurePvcCreated(req); !step.ShouldProceed() {
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

func (r *Reconciler) ensurePvcCreated(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(PVCReady, check, err.Error())
	}

	if obj.Spec.CacheKeyName == nil {
		check.Status = true
		check.Message = "no cache key name specified, so not required"
		check.State = rApi.CompletedState
		check.Info = check.Message
		if check != obj.Status.Checks[PVCReady] {
			fn.MapSet(&obj.Status.Checks, PVCReady, check)
			if sr := req.UpdateStatus(); !sr.ShouldProceed() {
				return sr
			}
		}
		return req.Next()
	}

	_, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.BuildNamespace, *obj.Spec.CacheKeyName), &corev1.PersistentVolumeClaim{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
		return failed(fmt.Errorf("pvc, required for cache not found, please create a pvc with name %s", *obj.Spec.CacheKeyName))
	}

	check.Status = true
	check.State = rApi.CompletedState
	if check != obj.Status.Checks[PVCReady] {
		fn.MapSet(&obj.Status.Checks, PVCReady, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureJobCreated(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(JobCreated, check, err.Error())
	}

	if obj.Status.Checks[JobCreated].Status {
		return req.Next()
	}

	var jobs batchv1.JobList

	if obj.Spec.CacheKeyName != nil {
		if err := r.List(ctx, &jobs,
			&client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.BuildNameKey: *obj.Spec.CacheKeyName}),
				Namespace:     r.Env.BuildNamespace,
			},
		); err != nil {
			return failed(err)
		}

		for _, j := range jobs.Items {
			if j.Status.Active > 0 {
				return failed(fmt.Errorf("cache is in use, currently building %s", j.Name)).Err(nil)
			}
		}
	}

	b, err := r.getBuildTemplate(req)
	if err != nil {
		return failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return failed(err).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	check.State = rApi.CompletedState
	if check != obj.Status.Checks[JobCreated] {
		fn.MapSet(&obj.Status.Checks, JobCreated, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) provisionCreatedJob(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(JobCompleted, check, err.Error())
	}

	if obj.Status.Checks[JobCompleted].Status || obj.Status.Checks[JobFailed].Status {
		return req.Next()
	}

	j, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.BuildNamespace, fmt.Sprint("build-", obj.Name)), &batchv1.Job{})
	if err != nil {
		return failed(err)
	}

	if j.Status.Active > 0 {
		return failed(fmt.Errorf("job is running, and waiting for completion")).Err(nil)
	}

	if j.Status.Succeeded > 0 {
		check.Status = true
		check.State = rApi.CompletedState
		if check != obj.Status.Checks[JobCompleted] {
			fn.MapSet(&obj.Status.Checks, JobCompleted, check)
			if sr := req.UpdateStatus(); !sr.ShouldProceed() {
				return sr
			}
		}
	}

	if j.Status.Failed > 0 {
		check.Status = true
		check.State = rApi.ErroredState
		check.Error = "job failed, please check logs"
		check.Message = "job failed, please check logs"
		if check != obj.Status.Checks[JobFailed] {
			fn.MapSet(&obj.Status.Checks, JobFailed, check)
			if sr := req.UpdateStatus(); !sr.ShouldProceed() {
				return sr
			}
		}
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(JobDeleted, check, err.Error())
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	s, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
		return req.Finalize()
	}

	if err := r.Delete(ctx, s); err != nil {
		return failed(err)
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

					if brn, ok := obj.GetAnnotations()[constants.BuildRunNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetAnnotations()[buildrunNamespaceAnn], brn)}}
					}
					return nil
				}),
		)
	}

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
