package operator

import (
	"time"

	step_result "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type checkWrapper[T Resource] struct {
	checkName string
	request   *Request[T]
	Check     `json:",inline"`
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

func (cw *checkWrapper[T]) Failed(err error) step_result.Result {
	defer cw.request.LogPostCheck(cw.checkName)

	cw.Check.State = ErroredState
	cw.Check.Status = false
  if err != nil {
    cw.Check.Message = err.Error()
    // if apiErrors.IsConflict(err) {
    // }
  }

	cw.request.Object.GetStatus().Checks[cw.checkName] = cw.Check

	// FIXME: change `Err(err)` to `Err(nil)`, once failed calls are checked on each of the controllers
	return cw.request.updateStatus().Continue(false).Err(err)
}

func (cw *checkWrapper[T]) StillRunning(err error) step_result.Result {
	defer cw.request.LogPostCheck(cw.checkName)

	cw.Check.State = RunningState
	cw.Check.Status = false
	if err != nil {
		cw.Check.Message = err.Error()
	}

	cw.request.Object.GetStatus().Checks[cw.checkName] = cw.Check

	return cw.request.updateStatus().Continue(false).Err(err)
}

func (cw *checkWrapper[T]) Completed() step_result.Result {
	defer cw.request.LogPostCheck(cw.checkName)

	cw.Check.State = CompletedState
	cw.Check.Status = true

	cw.request.Object.GetStatus().Checks[cw.checkName] = cw.Check

	return cw.request.updateStatus().Continue(true)
}
