package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-mongo/internal/controllers/database"
	"operators.kloudlite.io/operators/msvc-mongo/internal/controllers/standalone"
	"operators.kloudlite.io/operators/msvc-mongo/internal/env"
)

func main() {
	op := operator.New("mongodb")

	ev := env.GetEnvOrDie()

	op.RegisterControllers(
		&standalone.ServiceReconciler{Name: "mongodb-standalone-svc", Env: ev},
		&database.Reconciler{Name: "mongodb-database", Env: ev},
	)
	op.Start()
}
