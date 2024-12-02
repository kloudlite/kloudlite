package reconciler

import (
	"time"

	fn "github.com/kloudlite/operator/toolkit/functions"
	step_result "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=true
type Check struct {
	Status     bool   `json:"status"`
	Message    string `json:"message,omitempty"`
	Generation int64  `json:"generation,omitempty"`

	State State `json:"state,omitempty"`

	StartedAt *metav1.Time `json:"startedAt,omitempty"`
	// CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	Info  string `json:"info,omitempty"`
	Debug string `json:"debug,omitempty"`
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:generate=true
type CheckMeta struct {
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Debug       bool    `json:"debug,omitempty"`
	Hide        bool    `json:"hide,omitempty"`
}

func AreChecksEqual(c1 Check, c2 Check) bool {
	c1.StartedAt = fn.New(fn.DefaultIfNil[metav1.Time](c1.StartedAt))
	c2.StartedAt = fn.New(fn.DefaultIfNil[metav1.Time](c1.StartedAt))

	// c1.CompletedAt = fn.New(fn.DefaultIfNil[metav1.Time](c1.StartedAt))
	// c2.CompletedAt = fn.New(fn.DefaultIfNil[metav1.Time](c1.StartedAt))

	return c1.Status == c2.Status &&
		c1.Message == c2.Message &&
		c1.Generation == c2.Generation &&
		c1.State == c2.State &&
		c1.StartedAt.Sub(c2.StartedAt.Time) == 0
}

type checkWrapper[T Resource] struct {
	checkName string
	request   *Request[T]
	Check     `json:",inline"`
}

func (cw *checkWrapper[T]) Failed(err error) step_result.Result {
	defer cw.request.LogPostCheck(cw.checkName)

	cw.Check.State = ErroredState
	cw.Check.Status = false
	if err != nil {
		if !apiErrors.IsConflict(err) {
			cw.Check.Message = err.Error()
		}
	}

	cw.request.Object.GetStatus().Checks[cw.checkName] = cw.Check
	cw.request.Object.GetStatus().IsReady = false

	if err2 := cw.request.client.Status().Update(cw.request.Context(), cw.request.Object); err != nil {
		return step_result.New().Err(err2)
	}
	// FIXME: change `Err(err)` to `Err(nil)`, once failed calls are checked on each of the controllers
	return step_result.New().Err(err)

	// return cw.request.updateStatus().Continue(false).Err(err)
}

func (cw *checkWrapper[T]) StillRunning(err error) step_result.Result {
	defer cw.request.LogPostCheck(cw.checkName)

	cw.Check.State = RunningState
	cw.Check.Status = false
	if err != nil {
		if !apiErrors.IsConflict(err) {
			cw.Check.Message = err.Error()
		}
	}

	cw.request.Object.GetStatus().Checks[cw.checkName] = cw.Check
	cw.request.Object.GetStatus().IsReady = false

	if err2 := cw.request.client.Status().Update(cw.request.Context(), cw.request.Object); err != nil {
		return step_result.New().Err(err2)
	}
	return step_result.New().Err(err)

	// return cw.request.updateStatus().Continue(false).Err(err)
}

func (cw *checkWrapper[T]) Completed() step_result.Result {
	defer cw.request.LogPostCheck(cw.checkName)

	cw.Check.State = CompletedState
	cw.Check.Status = true

	cw.request.Object.GetStatus().Checks[cw.checkName] = cw.Check
	cw.request.Object.GetStatus().IsReady = true

	if err := cw.request.client.Status().Update(cw.request.Context(), cw.request.Object); err != nil {
		return step_result.New().Err(err)
	}
	return step_result.New().Continue(true)
	// return cw.request.updateStatus().Continue(true)
}

func NewRunningCheck[T Resource](name string, req *Request[T]) *checkWrapper[T] {
	cw := &checkWrapper[T]{
		checkName: name,
		request:   req,
		Check: Check{
			Generation: req.Object.GetGeneration(),
			State:      RunningState,
			StartedAt: &metav1.Time{
				Time: time.Now(),
			},
			Info:  "",
			Debug: "",
			Error: "",
		},
	}

	cw.request.LogPreCheck(cw.checkName)
	return cw
}
