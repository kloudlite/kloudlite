package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/fatih/color"

	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=true
type CheckState string

const (
	WaitingState CheckState = "waiting" // yet to be reconciled
	RunningState CheckState = "running" // under reconcilation
	ErroredState CheckState = "errored" // transient error, might be retried in next reconcilation
	PassedState  CheckState = "passed"  // passed, no need for another run
	FailedState  CheckState = "failed"  // failed, no further runs
)

func (c CheckState) String() string {
	return string(c)
}

// +kubebuilder:object:generate=true
type CheckResult struct {
	Message    string `json:"message,omitempty"`
	Generation int64  `json:"generation,omitempty"`

	State CheckState `json:"state,omitempty"`

	StartedAt   *metav1.Time `json:"startedAt,omitempty"`
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`
}

func AreChecksEqual(c1 CheckResult, c2 CheckResult) bool {
	c1.StartedAt = fn.New(fn.DefaultIfNil(c1.StartedAt))
	c2.StartedAt = fn.New(fn.DefaultIfNil(c1.StartedAt))

	return c1.Generation == c2.Generation &&
		c1.State == c2.State &&
		c1.Message == c2.Message &&
		c1.StartedAt.Sub(c2.StartedAt.Time) == 0
}

type Check[T Resource] struct {
	name        string
	request     *Request[T]
	CheckResult `json:",inline"`
}

func (c *Check[T]) result() StepResult {
	return newStepResult()
}

func (c *Check[T]) preCheck() {
	blue := color.New(color.FgBlue).SprintFunc()
	c.request.internalLogger.Info(blue("check start"), "check.name", c.name, "check.state", c.State)
}

func (c *Check[T]) postCheck() {
	fg := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	args := []any{
		"check.name", c.name,
		"check.state", c.State,
		"check.time_taken", fmt.Sprintf("%.2fs", time.Since(c.StartedAt.Time).Seconds()),
	}
	if c.State == FailedState || c.State == ErroredState {
		fg = color.New(color.FgRed).SprintFunc()
		args = append(args, "check.message", c.Message)
	}
	c.request.internalLogger.Info(fg("check end"), args...)
}

// Query
func (c *Check[T]) Context() context.Context {
	return c.request.Context()
}

func (c *Check[T]) Logger() *slog.Logger {
	return c.request.Logger.With("check.name", c.name)
}

// Mutations on check
func (c *Check[T]) Errored(err error) StepResult {
	if apiErrors.IsConflict(err) {
		return newStepResult().RequeueAfter(100 * time.Millisecond)
	}

	defer c.postCheck()

	_, file, line, _ := runtime.Caller(1)
	c.request.Logger.Debug("check.errored", "err", err, "caller", fmt.Sprintf("%s:%d", file, line))

	c.State = ErroredState
	c.CheckResult.Message = err.Error()

	c.CompletedAt = &metav1.Time{Time: time.Now()}
	c.request.Object.GetStatus().Checks[c.name] = c.CheckResult
	c.request.Object.GetStatus().IsReady = false

	if err2 := c.request.statusUpdate(); err != nil {
		return c.result().Err(err2)
	}
	// FIXME: change `Err(err)` to `Err(nil)`, once failed calls are checked on each of the controllers
	return c.result().Err(err)
}

func (c *Check[T]) Failed(err error) StepResult {
	if apiErrors.IsConflict(err) {
		return c.Errored(err)
	}

	defer c.postCheck()

	_, file, line, _ := runtime.Caller(1)
	c.request.Logger.Debug("check.failed", "err", err, "caller", fmt.Sprintf("%s:%d", file, line))

	c.CheckResult.State = FailedState
	if err != nil {
		if !apiErrors.IsConflict(err) {
			c.CheckResult.Message = err.Error()
		}
	}

	c.CompletedAt = &metav1.Time{Time: time.Now()}
	c.request.Object.GetStatus().Checks[c.name] = c.CheckResult
	c.request.Object.GetStatus().IsReady = false

	if err2 := c.request.statusUpdate(); err != nil {
		return c.result().Err(err2)
	}

	return c.result().Err(nil)
}

func (c *Check[T]) Requeue(duration ...time.Duration) StepResult {
	if len(duration) > 0 {
		return c.result().RequeueAfter(duration[0])
	}
	return c.result()
}

// Abort aborts current check
func (c *Check[T]) Abort(msg string) StepResult {
	return c.Errored(fmt.Errorf(msg))
}

func (c *Check[T]) Passed() StepResult {
	defer c.postCheck()

	c.State = PassedState
	c.CompletedAt = &metav1.Time{Time: time.Now()}
	c.request.Object.GetStatus().Checks[c.name] = c.CheckResult

	if err := c.request.statusUpdate(); err != nil {
		return c.result().Err(err)
	}
	return c.result().Continue(true)
}

func NewRunningCheck[T Resource](name string, req *Request[T]) *Check[T] {
	check := &Check[T]{
		name:    name,
		request: req,
		CheckResult: CheckResult{
			Generation: req.Object.GetGeneration(),
			State:      RunningState,
			StartedAt: &metav1.Time{
				Time: time.Now(),
			},
		},
	}

	check.preCheck()
	return check
}
