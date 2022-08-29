package step_result

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Result interface {
	// getter options
	ShouldProceed() bool
	ReconcilerResponse() (ctrl.Result, error)

	// builder options
	Continue(bool) Result
	RequeueAfter(time.Duration) Result
	Err(error) Result
}

type options struct {
	toContinue bool
	requeue    ctrl.Result
	err        error
}

func (opt options) ShouldProceed() bool {
	return opt.toContinue && opt.err == nil
}

func (opt options) ReconcilerResponse() (ctrl.Result, error) {
	return opt.requeue, opt.err
}

func (opt options) Continue(val bool) Result {
	opt.toContinue = val
	return opt
}

func (opt options) RequeueAfter(d time.Duration) Result {
	opt.requeue = ctrl.Result{RequeueAfter: d}
	return opt
}

func (opt options) Err(err error) Result {
	opt.err = err
	return opt
}

func New() Result {
	return &options{}
}
