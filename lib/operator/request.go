package operator

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/kubectl"
	"operators.kloudlite.io/lib/logging"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	rawJson "operators.kloudlite.io/lib/raw-json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Request[T Resource] struct {
	ctx    context.Context
	client client.Client
	Object T
	Logger logging.Logger
	locals map[string]any
}

func NewRequest[T Resource](ctx context.Context, c client.Client, nn types.NamespacedName, resource T) (*Request[T], error) {
	if err := c.Get(ctx, nn, resource); err != nil {
		return nil, err
	}
	logger, ok := ctx.Value("logger").(logging.Logger)
	if !ok {
		panic("no logger passed into NewRequest")
	}

	return &Request[T]{
		ctx:    ctx,
		client: c,
		Object: resource,
		Logger: logger.WithName(nn.String()).WithKV("NN", nn.String()),
		locals: map[string]any{},
	}, nil
}

// DebuggingOnlySetStatus only to be used in debugging environment, never in production
func (r *Request[T]) DebuggingOnlySetStatus(status Status) stepResult.Result {
	obj := r.Object
	ctx := r.ctx

	if value := obj.GetAnnotations()["kloudlite.io/reset-status"]; value == "true" {
		ann := obj.GetAnnotations()

		r.Object.GetStatus().OpsConditions = status.OpsConditions
		r.Object.GetStatus().Conditions = status.Conditions
		r.Object.GetStatus().ChildConditions = status.ChildConditions
		r.Object.GetStatus().DisplayVars = status.DisplayVars
		r.Object.GetStatus().GeneratedVars = status.GeneratedVars
		r.Object.GetStatus().IsReady = status.IsReady

		if err := r.client.Status().Update(ctx, obj); err != nil {
			return r.FailWithStatusError(err)
		}

		delete(ann, "kloudlite.io/reset-status")
		obj.SetAnnotations(ann)
		if err := r.client.Update(ctx, obj); err != nil {
			return r.FailWithStatusError(err)
		}
	}
	return nil
}

func (r *Request[T]) EnsureLabelsAndAnnotations() stepResult.Result {
	labels := r.Object.GetEnsuredLabels()
	annotations := r.Object.GetEnsuredAnnotations()

	hasAllLabels := fn.MapContains(r.Object.GetLabels(), labels)
	hasAllAnnotations := fn.MapContains(r.Object.GetAnnotations(), annotations)

	if !hasAllLabels || !hasAllAnnotations {
		x := r.Object.GetLabels()
		if x == nil {
			x = make(map[string]string, len(labels))
		}
		for k, v := range labels {
			x[k] = v
		}
		r.Object.SetLabels(x)

		y := r.Object.GetAnnotations()
		if y == nil {
			y = make(map[string]string, len(annotations))
		}
		for k, v := range annotations {
			y[k] = v
		}
		r.Object.SetAnnotations(y)
		return stepResult.New().Err(r.client.Update(r.ctx, r.Object))
	}

	return stepResult.New().Continue(true)
}

func (r *Request[T]) FailWithStatusError(err error, moreConditions ...metav1.Condition) stepResult.Result {
	if err == nil {
		return stepResult.New().Continue(true)
	}

	statusC := make([]metav1.Condition, 0, len(r.Object.GetStatus().Conditions)+len(moreConditions)+1)
	statusC = append(statusC, r.Object.GetStatus().Conditions...)
	statusC = append(statusC, moreConditions...)

	newConditions, _, err2 := conditions.Patch(
		r.Object.GetStatus().Conditions, append(
			statusC, metav1.Condition{
				Type:    "FailedWithErr",
				Status:  metav1.ConditionFalse,
				Reason:  "StatusFailedWithErr",
				Message: err.Error(),
			},
		),
	)

	if err2 != nil {
		return stepResult.New().Err(err2)
	}

	r.Object.GetStatus().IsReady = false
	r.Object.GetStatus().Conditions = newConditions
	r.Object.GetStatus().LastReconcileTime = metav1.Time{Time: time.Now()}

	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return stepResult.New().Err(err2)
	}
	return stepResult.New().Err(err)
}

func (r *Request[T]) FailWithOpError(err error, moreConditions ...metav1.Condition) stepResult.Result {
	if err == nil {
		return stepResult.New().Continue(true)
	}

	opsConditions := make([]metav1.Condition, 0, len(r.Object.GetStatus().OpsConditions)+len(moreConditions)+1)
	opsConditions = append(opsConditions, r.Object.GetStatus().OpsConditions...)
	opsConditions = append(opsConditions, moreConditions...)

	newConditions, _, err2 := conditions.Patch(
		r.Object.GetStatus().OpsConditions, append(
			opsConditions, metav1.Condition{
				Type:    "FailedWithErr",
				Status:  metav1.ConditionFalse,
				Reason:  "OpsFailedWithErr",
				Message: err.Error(),
			},
		),
	)
	if err2 != nil {
		return stepResult.New().Err(err2)
	}
	r.Object.GetStatus().IsReady = false
	r.Object.GetStatus().OpsConditions = newConditions
	r.Object.GetStatus().LastReconcileTime = metav1.Time{Time: time.Now()}

	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return stepResult.New().Err(err2)
	}
	return stepResult.New().Err(err)
}

func (r *Request[T]) EnsureChecks(names ...string) stepResult.Result {
	obj, ctx, checks := r.Object, r.Context(), r.Object.GetStatus().Checks
	nChecks := len(checks)

	if checks == nil {
		checks = map[string]Check{}
	}

	for i := range names {
		if _, ok := checks[names[i]]; !ok {
			checks[names[i]] = Check{}
		}
	}

	if nChecks != len(checks) {
		obj.GetStatus().Checks = checks
		if err := r.client.Status().Update(ctx, obj); err != nil {
			return r.FailWithOpError(err)
		}
		return stepResult.New().RequeueAfter(1 * time.Second)
	}

	return stepResult.New().Continue(true)
}

func (r *Request[T]) ClearStatusIfAnnotated() stepResult.Result {
	obj := r.Object

	ann := obj.GetAnnotations()
	if v := ann[constants.ClearStatusKey]; v == "true" {
		delete(ann, constants.ClearStatusKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(r.Context(), obj); err != nil {
			return r.FailWithOpError(err)
		}

		obj.GetStatus().IsReady = false
		obj.GetStatus().LastReconcileTime = metav1.Time{Time: time.Now()}
		obj.GetStatus().Checks = nil
		obj.GetStatus().Message = rawJson.RawJson{}
		obj.GetStatus().Messages = nil
		obj.GetStatus().Conditions = nil
		obj.GetStatus().OpsConditions = nil
		obj.GetStatus().ChildConditions = nil
		obj.GetStatus().DisplayVars = rawJson.RawJson{}

		if err := r.client.Status().Update(context.TODO(), obj); err != nil {
			return r.FailWithOpError(err)
		}
		return r.Done().RequeueAfter(2 * time.Second)
	}
	return r.Next()
}

func (r *Request[T]) RestartIfAnnotated() stepResult.Result {
	ctx, obj := r.Context(), r.Object
	ann := obj.GetAnnotations()
	if v := ann[constants.RestartKey]; v == "true" {
		delete(ann, constants.RestartKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(ctx, obj); err != nil {
			return r.FailWithOpError(err)
		}

		exitCode, err := kubectl.Restart(kubectl.Deployments, obj.GetNamespace(), obj.GetEnsuredLabels())
		if exitCode != 0 {
			// failed to restart deployments, with non-zero exit code
			r.Logger.Error(err)
		}
		exitCode, err = kubectl.Restart(kubectl.Statefulsets, obj.GetNamespace(), obj.GetEnsuredLabels())
		if exitCode != 0 {
			// failed to restart statefultset, with non-zero exit code
			r.Logger.Error(err)
		}
		return r.Done().RequeueAfter(2 * time.Second)
	}

	return r.Next()
}

func (r *Request[T]) EnsureFinalizers(finalizers ...string) stepResult.Result {
	obj := r.Object

	if !fn.ContainsFinalizers(obj, finalizers...) {
		for i := range finalizers {
			controllerutil.AddFinalizer(obj, finalizers[i])
		}
		if err := r.client.Update(r.Context(), obj); err != nil {
			return r.FailWithOpError(err)
		}
		return stepResult.New()
	}
	return stepResult.New().Continue(true)
}

func (r *Request[T]) CheckFailed(name string, check Check, msg string) stepResult.Result {
	check.Status = false
	check.Message = msg
	r.Object.GetStatus().Checks[name] = check
	r.Object.GetStatus().Message.Set(name, check.Message)
	r.Object.GetStatus().IsReady = false
	r.Object.GetStatus().LastReconcileTime = metav1.Time{Time: time.Now()}
	if err := r.client.Status().Update(r.ctx, r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	return stepResult.New()
}

func (r *Request[T]) Context() context.Context {
	return r.ctx
}

func (r *Request[T]) Done(result ...ctrl.Result) stepResult.Result {
	r.Object.GetStatus().LastReconcileTime = metav1.Time{Time: time.Now()}
	if err := r.client.Status().Update(context.TODO(), r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	if len(result) > 0 {
		return stepResult.New().RequeueAfter(result[0].RequeueAfter)
	}
	return stepResult.New()
}

func (r *Request[T]) Next() stepResult.Result {
	return stepResult.New().Continue(true)
}

func (r *Request[T]) UpdateStatus() stepResult.Result {
	r.Object.GetStatus().LastReconcileTime = metav1.Time{Time: time.Now()}
	checks := r.Object.GetStatus().Checks
	for name := range checks {
		if checks[name].Status {
			if err := r.Object.GetStatus().Message.Delete(name); err != nil {
				return stepResult.New().Err(err)
			}

			if r.Object.GetStatus().Message.Len() == 0 {
				r.Object.GetStatus().Message = rawJson.RawJson{RawMessage: nil}
			}
		}
	}

	if err := r.client.Status().Update(r.Context(), r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	return stepResult.New()
}

func (r *Request[T]) Finalize() stepResult.Result {
	controllerutil.RemoveFinalizer(r.Object, constants.CommonFinalizer)
	controllerutil.RemoveFinalizer(r.Object, constants.ForegroundFinalizer)
	return stepResult.New().Err(r.client.Update(r.ctx, r.Object))
}
