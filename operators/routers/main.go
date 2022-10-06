package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/routers/internal/controllers/router"
	"operators.kloudlite.io/operators/routers/internal/env"
)

func main() {
	op := operator.New("routers")
	ev := env.GetEnvOrDie()
	op.RegisterControllers(
		// &account_router.AccountRouterReconciler{Name: "acc-router", Env: ev},
		&router.Reconciler{Name: "router", Env: ev},
	)
	op.Start()
}
