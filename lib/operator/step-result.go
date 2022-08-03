package operator

import ctrl "sigs.k8s.io/controller-runtime"

type StepResult interface {
	ShouldProceed() bool
	NoErr() StepResult
	ReconcilerResponse() (ctrl.Result, error)
}

func NewStepResult(result *ctrl.Result, err error) StepResult {
	return newStepResult(result, err)
}

func newStepResult(result *ctrl.Result, err error) StepResult {
	return &stepResult{result: result, err: err}
}

type stepResult struct {
	result *ctrl.Result
	err    error
}

func (s stepResult) ShouldProceed() bool {
	return s.result == nil && s.err == nil
}

func (s stepResult) NoErr() StepResult {
	s.result = &ctrl.Result{}
	s.err = nil
	return s
}

func (s stepResult) ReconcilerResponse() (ctrl.Result, error) {
	if s.result == nil {
		return ctrl.Result{}, s.err
	}
	return *s.result, s.err
}
