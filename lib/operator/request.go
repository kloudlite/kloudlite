package operator

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
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

func NewRequest[T Resource](ctx context.Context, c client.Client, nn types.NamespacedName, resInstance T) (*Request[T], error) {
	if err := c.Get(ctx, nn, resInstance); err != nil {
		return nil, err
	}
	logger := logging.NewOrDie(
		&logging.Options{Name: nn.String(), Dev: true},
	)

	return &Request[T]{
		ctx:    ctx,
		client: c,
		Object: resInstance,
		Logger: logger,
		locals: map[string]any{},
	}, nil
}

func (r *Request[T]) CleanupLastRun() StepResult {
	status := r.Object.GetStatus()
	if len(status.OpsConditions) > 0 {
		status.OpsConditions = []metav1.Condition{}
		return newStepResult(&ctrl.Result{RequeueAfter: 0}, r.client.Status().Update(r.ctx, r.Object))
	}
	return newStepResult(nil, nil)
}

func (r *Request[T]) EnsureLabelsAndAnnotations() StepResult {
	labels := r.Object.GetEnsuredLabels()
	annotations := r.Object.GetEnsuredAnnotations()

	hasAllLabels := fn.MapContains(r.Object.GetLabels(), labels)
	hasAllAnnotations := fn.MapContains(r.Object.GetAnnotations(), annotations)

	if !hasAllLabels || !hasAllAnnotations {
		x := r.Object.GetLabels()
		if x == nil {
			x = map[string]string{}
		}
		for k, v := range labels {
			x[k] = v
		}
		r.Object.SetLabels(x)

		y := r.Object.GetAnnotations()
		if y == nil {
			y = map[string]string{}
		}
		for k, v := range annotations {
			y[k] = v
		}
		r.Object.SetAnnotations(y)

		return newStepResult(nil, r.client.Update(r.ctx, r.Object))
	}

	return newStepResult(nil, nil)
}

func (r *Request[T]) FailWithStatusError(err error, moreConditions ...metav1.Condition) StepResult {
	if err == nil {
		return r.Next()
	}

	newConditions, _, err2 := conditions.Patch(
		r.Object.GetStatus().Conditions, append(
			[]metav1.Condition{
				{
					Type:    "FailedWithErr",
					Status:  metav1.ConditionFalse,
					Reason:  "StatusFailedWithErr",
					Message: err.Error(),
				},
			}, moreConditions...,
		),
	)

	if err2 != nil {
		return newStepResult(&ctrl.Result{}, err2)
	}

	r.Object.GetStatus().IsReady = false
	r.Object.GetStatus().Conditions = newConditions
	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return newStepResult(&ctrl.Result{}, err2)
	}
	return newStepResult(&ctrl.Result{}, err)
}

func (r *Request[T]) FailWithOpError(err error, moreConditions ...metav1.Condition) StepResult {
	if err == nil {
		return r.Next()
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
		return newStepResult(&ctrl.Result{}, err2)
	}
	r.Object.GetStatus().IsReady = false
	r.Object.GetStatus().OpsConditions = newConditions
	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return newStepResult(&ctrl.Result{}, err2)
	}
	return newStepResult(&ctrl.Result{}, nil)
	// return newStepResult(&ctrl.Result{}, err)
}

func (r *Request[T]) Context() context.Context {
	return r.ctx
}

func (r *Request[T]) Done(result ...*ctrl.Result) StepResult {
	if len(result) > 0 {
		return newStepResult(result[0], nil)
	}
	return newStepResult(nil, nil)
}

func (r *Request[T]) Next() StepResult {
	return newStepResult(nil, nil)
}

func (r *Request[T]) Finalize() StepResult {
	controllerutil.RemoveFinalizer(r.Object, constants.CommonFinalizer)
	controllerutil.RemoveFinalizer(r.Object, constants.ForegroundFinalizer)
	return newStepResult(&ctrl.Result{}, r.client.Update(r.ctx, r.Object))
}
