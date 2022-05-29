package operator

import (
	"context"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"operators.kloudlite.io/lib/conditions"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logger"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

// +kubebuilder:object:generate=true

type Status struct {
	IsReady       bool                `json:"isReady"`
	GeneratedVars rawJson.KubeRawJson `json:"generatedVars,omitempty"`
	Conditions    []metav1.Condition  `json:"conditions,omitempty"`
	OpsConditions []metav1.Condition  `json:"opsConditions,omitempty"`
}

type Resource interface {
	client.Object
	runtime.Object
	GetStatus() *Status
	GetEnsuredLabels() map[string]string
}

type Request[T Resource] struct {
	ctx    context.Context
	client client.Client
	Object T
	Logger *zap.SugaredLogger
	locals map[string]any
}

type stepResult struct {
	result *ctrl.Result
	err    error
}

func GetLocal[T any, V Resource](r *Request[V], key string) (T, bool) {
	x := r.locals[key]
	t, ok := x.(T)
	if !ok {
		return *new(T), ok
	}
	return t, ok
}

func SetLocal[T any, V Resource](r *Request[V], key string, value T) {
	r.locals[key] = value
}

func NewStepResult(result *ctrl.Result, err error) StepResult {
	return &stepResult{result: result, err: err}
}

func (s stepResult) Err() error {
	return s.err
}

func (s stepResult) Result() ctrl.Result {
	if s.result == nil {
		return ctrl.Result{}
	}
	return *s.result
}

func (s stepResult) ShouldProceed() bool {
	return s.result == nil && s.err == nil
}

type StepResult interface {
	Err() error
	Result() ctrl.Result
	ShouldProceed() bool
}

func NewRequest[T Resource](ctx context.Context, c client.Client, nn types.NamespacedName,
	resInstance T) *Request[T] {
	if err := c.Get(ctx, nn, resInstance); err != nil {
		return nil
	}
	return &Request[T]{
		ctx:    ctx,
		client: c,
		Object: resInstance,
		Logger: logger.New(nn),
		locals: map[string]any{},
	}
}

func (r *Request[T]) EnsureLabels() StepResult {
	el := r.Object.GetEnsuredLabels()
	if fn.MapContains(r.Object.GetLabels(), el) {
		x := r.Object.GetLabels()
		if x == nil {
			x = map[string]string{}
		}

		for k, v := range el {
			x[k] = v
		}
		return NewStepResult(&ctrl.Result{}, r.client.Update(r.ctx, r.Object))
	}

	return NewStepResult(nil, nil)
}

func (r *Request[T]) FailWithStatusError(err error) StepResult {
	newConditions, _, err := conditions.Patch(
		r.Object.GetStatus().Conditions, []metav1.Condition{
			{
				Type:    "FailedWithErr",
				Status:  metav1.ConditionFalse,
				Reason:  "StatusFailedWithErr",
				Message: err.Error(),
			},
		},
	)
	if err != nil {
		return NewStepResult(nil, err)
	}

	r.Object.GetStatus().Conditions = newConditions
	return NewStepResult(&ctrl.Result{}, r.client.Status().Update(r.ctx, r.Object))
}

func (r *Request[T]) FailWithOpError(err error) StepResult {
	newConditions, _, err := conditions.Patch(
		r.Object.GetStatus().OpsConditions, []metav1.Condition{
			{
				Type:    "FailedWithErr",
				Status:  metav1.ConditionFalse,
				Reason:  "OpsFailedWithErr",
				Message: err.Error(),
			},
		},
	)
	if err != nil {
		return NewStepResult(nil, err)
	}

	r.Object.GetStatus().OpsConditions = newConditions
	return NewStepResult(&ctrl.Result{}, r.client.Status().Update(r.ctx, r.Object))
}

func (r *Request[T]) Context() context.Context {
	return r.ctx
}

func (r *Request[T]) Done(result ...*ctrl.Result) StepResult {
	if len(result) > 0 {
		return NewStepResult(result[0], nil)
	}
	return NewStepResult(&ctrl.Result{}, nil)
}

func (r *Request[T]) Next() StepResult {
	return NewStepResult(nil, nil)
}

func (r *Request[T]) Finalize() StepResult {
	controllerutil.RemoveFinalizer(r.Object, "finalizers.kloudlite.io")
	controllerutil.RemoveFinalizer(r.Object, "foregroundDeletion")
	return NewStepResult(&ctrl.Result{}, r.client.Update(r.ctx, r.Object))
}

// type Api interface {
// 	FailWithErr(ctx context.Context, err error) (ctrl.Result, error)
// 	Update(ctx context.Context) (ctrl.Result, error)
// 	UpdateStatus(ctx context.Context) (ctrl.Result, error)
// }
//
// func NewApiClient(client client.Client, object client.Object, status *Status, logger ...*zap.SugaredLogger) Api {
// 	return &cli{client: client, object: object, status: status}
// }
//
// func (c *cli) FailWithErr(ctx context.Context, err error) (ctrl.Result, error) {
// 	meta.SetStatusCondition(
// 		&c.status.OpsConditions, metav1.Condition{
// 			Type:    "ReconFailedWithErr",
// 			Status:  metav1.ConditionFalse,
// 			Reason:  "ErrDuringReconcile",
// 			Message: err.Error(),
// 		},
// 	)
// 	if err2 := c.client.Status().Update(ctx, c.object); err2 != nil {
// 		return ctrl.Result{}, err2
// 	}
// 	return ctrl.Result{}, err
// }
//
// func (c *cli) Update(ctx context.Context) (ctrl.Result, error) {
// 	return ctrl.Result{}, c.client.Update(ctx, c.object)
// }
//
// func (c *cli) UpdateStatus(ctx context.Context) (ctrl.Result, error) {
// 	return ctrl.Result{}, c.client.Status().Update(ctx, c.object)
// }
//
// type ResType[T any] interface {
// 	MakeList() []T
// 	New() T
// }

// func GetX(ctx context.Context, cli client.Client, nn types.NamespacedName) (T, error) {
//
// 	if err := cli.Get(ctx, nn, ts[0]); err != nil {
// 		return *new(T), err
// 	}
// 	return ts[0], nil
// }

func Get[T client.Object](ctx context.Context, cli client.Client, nn types.NamespacedName, obj T) (T, error) {
	if err := cli.Get(ctx, nn, obj); err != nil {
		return obj, err
	}
	return obj, nil
}
