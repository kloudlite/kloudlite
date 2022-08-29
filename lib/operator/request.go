package operator

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
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

	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return stepResult.New().Err(err2)
	}
	return stepResult.New().Err(err)
}

func (r *Request[T]) CheckFailed(name string, check Check) stepResult.Result {
	r.Object.GetStatus().Message.Set(name, check.Error)
	r.Object.GetStatus().IsReady = false
	if err := r.client.Status().Update(r.ctx, r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	return stepResult.New()
}

func (r *Request[T]) Context() context.Context {
	return r.ctx
}

func (r *Request[T]) Done(result ...ctrl.Result) stepResult.Result {
	if len(result) > 0 {
		return stepResult.New().RequeueAfter(0)
	}
	return stepResult.New()
}

func (r *Request[T]) Next() stepResult.Result {
	return stepResult.New().Continue(true)
}

func (r *Request[T]) Finalize() stepResult.Result {
	controllerutil.RemoveFinalizer(r.Object, constants.CommonFinalizer)
	controllerutil.RemoveFinalizer(r.Object, constants.ForegroundFinalizer)
	return stepResult.New().Err(r.client.Update(r.ctx, r.Object))
}
