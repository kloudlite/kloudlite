package main

import (
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operator"
	edgeRouter "operators.kloudlite.io/operators/routers/internal/controllers/edge-router"
	"operators.kloudlite.io/operators/routers/internal/controllers/router"
	"operators.kloudlite.io/operators/routers/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("routers")
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		// &account_router.AccountRouterReconciler{Name: "acc-router", Env: ev},
		&router.Reconciler{Name: "router", Env: ev},
		&edgeRouter.Reconciler{Name: "edge-router", Env: ev},
	)
	mgr.Start()
}
