package step_result

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Result interface {
	// getter result
	ShouldProceed() bool
	ReconcilerResponse() (ctrl.Result, error)

	// builder result
	Continue(bool) Result
	RequeueAfter(time.Duration) Result
	Err(error) Result
}

type result struct {
	toContinue bool
	requeue    ctrl.Result
	err        error
}

func (opt result) ShouldProceed() bool {
	return opt.toContinue && opt.err == nil
}

func (opt result) ReconcilerResponse() (ctrl.Result, error) {
	return opt.requeue, opt.err
}

func (opt result) Continue(val bool) Result {
	opt.toContinue = val
	return opt
}

func (opt result) RequeueAfter(d time.Duration) Result {
	opt.requeue = ctrl.Result{RequeueAfter: d}
	return opt
}

func (opt result) Err(err error) Result {
	opt.err = err
	return opt
}

func New() Result {
	return &result{}
}
