package operator

import ctrl "sigs.k8s.io/controller-runtime"

type StepResult interface {
	Err() error
	Result() ctrl.Result
	ShouldProceed() bool
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

func (s stepResult) ReconcilerResponse() (ctrl.Result, error) {
	return s.Result(), s.Err()
}
