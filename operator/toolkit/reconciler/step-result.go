package reconciler

import (
	"time"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StepResult interface {
	ShouldProceed() bool
	ReconcilerResponse() (ctrl.Result, error)

	Continue(bool) StepResult
	RequeueAfter(time.Duration) StepResult
	Err(error) StepResult

	NoRequeue() StepResult
}

type stepResult struct {
	toContinue bool
	requeue    ctrl.Result
	err        error
}

// NoRequeue implements StepResult.
func (sr *stepResult) NoRequeue() StepResult {
	sr.err = nil
	return sr
}

func (sr *stepResult) ShouldProceed() bool {
	return sr.toContinue && sr.err == nil
}

func (sr *stepResult) ReconcilerResponse() (ctrl.Result, error) {
	return sr.requeue, sr.err
}

func (sr *stepResult) Continue(val bool) StepResult {
	sr.toContinue = val
	return sr
}

func (sr *stepResult) RequeueAfter(d time.Duration) StepResult {
	sr.requeue = ctrl.Result{RequeueAfter: d}
	sr.err = nil // because, we can't have requeue, without setting error to nil, as per error logs
	return sr
}

func (sr *stepResult) Err(err error) StepResult {
	if apiErrors.IsConflict(err) {
		sr.requeue = ctrl.Result{RequeueAfter: 100 * time.Millisecond}
		return sr
	}

	sr.err = err
	return sr
}

func newStepResult() StepResult {
	return &stepResult{}
}
