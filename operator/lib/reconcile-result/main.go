package reconcileResult

import (
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func OK() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func Failed() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func retry(after int, err error) (reconcile.Result, error) {
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Second * time.Duration(after),
	}, nil
}

func Retry(after int) (reconcile.Result, error) {
	return retry(after, nil)
}

func RetryE(after int, err error) (reconcile.Result, error) {
	fmt.Printf("RetryE: %+v", err)
	return retry(after, err)
}
