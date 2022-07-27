package operator

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

// +kubebuilder:object:generate=true

type Status struct {
	IsReady         bool                `json:"isReady,omitempty"`
	DisplayVars     rawJson.KubeRawJson `json:"displayVars,omitempty"`
	GeneratedVars   rawJson.KubeRawJson `json:"generatedVars,omitempty"`
	Conditions      []metav1.Condition  `json:"conditions,omitempty"`
	ChildConditions []metav1.Condition  `json:"childConditions,omitempty"`
	OpsConditions   []metav1.Condition  `json:"opsConditions,omitempty"`
	Generation      int                 `json:"generation,omitempty"`
}

type Reconciler interface {
	reconcile.Reconciler
	SetupWithManager(manager ctrl.Manager) error
	GetName() string
}

type Resource interface {
	client.Object
	runtime.Object
	GetStatus() *Status
	GetEnsuredLabels() map[string]string
	GetEnsuredAnnotations() map[string]string
}

type Request[T Resource] struct {
	ctx    context.Context
	client client.Client
	Object T
	Logger logging.Logger
	locals map[string]any
}

type stepResult struct {
	result *ctrl.Result
	err    error
}

func (s stepResult) Raw() (ctrl.Result, error) {
	return s.Result(), s.Err()
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

type StepResult interface {
	Err() error
	Result() ctrl.Result
	ShouldProceed() bool
	Raw() (ctrl.Result, error)
}

func NewStepResult(result *ctrl.Result, err error) StepResult {
	return newStepResult(result, err)
}

func newStepResult(result *ctrl.Result, err error) StepResult {
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

func NewRequest[T Resource](ctx context.Context, c client.Client, nn types.NamespacedName,
	resInstance T) (*Request[T], error) {
	if err := c.Get(ctx, nn, resInstance); err != nil {
		return nil, err
	}
	logger, err := logging.New(
		&logging.Options{
			Name: nn.String(),
			Dev:  true,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Request[T]{
		ctx:    ctx,
		client: c,
		Object: resInstance,
		Logger: logger,
		locals: map[string]any{},
	}, nil
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

		if err := r.client.Update(r.ctx, r.Object); err != nil {
			return NewStepResult(&ctrl.Result{}, err)
		}
		return NewStepResult(&ctrl.Result{}, nil)
	}

	return NewStepResult(nil, nil)
}

func (r *Request[T]) FailWithStatusError(err error) StepResult {
	if err == nil {
		return r.Next()
	}
	newConditions, _, err2 := conditions.Patch(
		r.Object.GetStatus().Conditions, []metav1.Condition{
			{
				Type:    "FailedWithErr",
				Status:  metav1.ConditionFalse,
				Reason:  "StatusFailedWithErr",
				Message: err.Error(),
			},
		},
	)
	if err2 != nil {
		return NewStepResult(nil, err2)
	}

	r.Object.GetStatus().Conditions = newConditions
	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return newStepResult(&ctrl.Result{}, err2)
	}
	return newStepResult(&ctrl.Result{}, err)
}

func (r *Request[T]) FailWithOpError(err error) StepResult {
	if err == nil {
		return r.Next()
	}
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
	if err2 := r.client.Status().Update(r.ctx, r.Object); err2 != nil {
		return newStepResult(&ctrl.Result{}, err2)
	}
	return newStepResult(&ctrl.Result{}, err)
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
	return NewStepResult(&ctrl.Result{}, r.client.Update(r.ctx, r.Object))
}

func Get[T client.Object](ctx context.Context, cli client.Client, nn types.NamespacedName, obj T) (T, error) {
	if err := cli.Get(ctx, nn, obj); err != nil {
		return obj, err
	}
	return obj, nil
}
