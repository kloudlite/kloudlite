package reconciler

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type StepResult struct {
	toContinue bool
	requeue    ctrl.Result
	err        error
}

func (sr StepResult) NoRequeue() StepResult {
	sr.err = nil
	return sr
}

func (sr StepResult) ShouldProceed() bool {
	return sr.toContinue && sr.err == nil
}

func (sr StepResult) ReconcilerResponse() (ctrl.Result, error) {
	return sr.requeue, sr.err
}

func (sr StepResult) Continue(val bool) StepResult {
	sr.toContinue = val
	return sr
}

func (sr StepResult) RequeueAfter(d time.Duration) StepResult {
	sr.requeue = ctrl.Result{RequeueAfter: d}
	sr.err = nil // because, we can't have requeue, without setting error to nil, as per error logs
	return sr
}

func (sr StepResult) Err(err error) StepResult {
	sr.err = err
	return sr
}

func newStepResult() StepResult {
	return StepResult{}
}
