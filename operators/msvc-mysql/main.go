package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-mysql/internal/controllers/database"
	"operators.kloudlite.io/operators/msvc-mysql/internal/controllers/standalone"
	"operators.kloudlite.io/operators/msvc-mysql/internal/env"
)

func main() {
	mgr := operator.New("msvc-mysql")
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(
		&standalone.ServiceReconciler{Name: "standalone-service", Env: ev},
		&database.Reconciler{Name: "database", Env: ev},
	)
	mgr.Start()
}
