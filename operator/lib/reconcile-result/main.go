package reconcileResult

import (
	"fmt"
	"math"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func OK() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func OKP() (*reconcile.Result, error) {
	return &reconcile.Result{}, nil
}

func Failed() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func FailedE(err error) (reconcile.Result, error) {
	return reconcile.Result{}, err
}

func Retry(after ...int) (reconcile.Result, error) {
	a := 0
	if len(after) > 0 {
		a = after[0]
	}
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Second * time.Duration(math.Max(float64(a), 1)),
	}, nil
}

func RetryE(after int, err error) (reconcile.Result, error) {
	fmt.Printf("RetryE: %+v", err)
	return Retry(after)
}
