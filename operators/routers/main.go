package main

import (
	"operators.kloudlite.io/operators/operator"
	accRouter "operators.kloudlite.io/operators/routers/internal/account-router"
	"operators.kloudlite.io/operators/routers/internal/router"
)

func main() {
	op := operator.New("routers")
	op.RegisterControllers(
		&accRouter.AccountRouterReconciler{Name: "acc-router"},
		&router.Reconciler{Name: "router"},
	)
	op.Start()
}
