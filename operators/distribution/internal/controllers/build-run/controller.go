package buildrun

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"github.com/kloudlite/operator/operators/distribution/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
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
	CreadsAvailable string = "creds-available"

	PVCReady     string = "pvc-ready"
	JobCreated   string = "job-created"
	JobCompleted string = "job-completed"
	JobFailed    string = "job-failed"
	JobDeleted   string = "job-deleted"
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

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(CreadsAvailable, PVCReady, JobCreated, JobCompleted, JobFailed); !step.ShouldProceed() {
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

func (r Reconciler) ensurePvcCreated(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(PVCReady, check, err.Error())
	}

	if obj.Spec.CacheKeyName == nil {
		check.Status = true
		check.Message = "no cache key name specified, so not required"
		if check != checks[PVCReady] {
			checks[PVCReady] = check
			req.UpdateStatus()
		}
		return req.Next()
	}

	_, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, *obj.Spec.CacheKeyName), &corev1.PersistentVolumeClaim{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
		return failed(fmt.Errorf("pvc, required for cache not found, please create a pvc with name %s", *obj.Spec.CacheKeyName))
	}

	check.Status = true
	if check != checks[PVCReady] {
		checks[PVCReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureJobCreated(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(JobCreated, check, err.Error())
	}

	if checks[JobCreated].Status == true {
		return req.Next()
	}

	var jobs batchv1.JobList

	if obj.Spec.CacheKeyName != nil {
		if err := r.List(ctx, &jobs,
			&client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.BuildNameKey: *obj.Spec.CacheKeyName}),
				Namespace:     obj.Namespace,
			},
		); err != nil {
			return failed(err)
		}

		for _, j := range jobs.Items {
			if j.Status.Active > 0 {
				return failed(fmt.Errorf("cache is in use, currently building %s", j.Name))
			}
		}
	}

	b, err := r.getBuildTemplate(req)
	if err != nil {
		return failed(err)
	}

	if _, err = r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[JobCreated] {
		checks[JobCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) provisionCreatedJob(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(JobCompleted, check, err.Error())
	}

	if checks[JobCompleted].Status == true || checks[JobFailed].Status == true {
		return req.Next()
	}

	j, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &batchv1.Job{})
	if err != nil {
		return failed(err)
	}

	if j.Status.Active > 0 {
		return failed(fmt.Errorf("job is running, and waiting for completion"))
	}

	if j.Status.Succeeded > 0 {
		check.Status = true
		if check != checks[JobCompleted] {
			checks[JobCompleted] = check
			req.UpdateStatus()
		}
	}

	if j.Status.Failed > 0 {
		check.Status = true
		if check != checks[JobFailed] {
			checks[JobFailed] = check
			req.UpdateStatus()
		}
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*dbv1.BuildRun]) stepResult.Result {

	ctx, obj, _ := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(JobDeleted, check, err.Error())
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
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&dbv1.BuildRun{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	watchList := []client.Object{}
	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if dev, ok := obj.GetLabels()[constants.WGDeviceNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN("", dev)}}
					}
					return nil
				}),
		)
	}

	return builder.Complete(r)
}
