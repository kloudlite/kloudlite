package main

import (
	mysqlMsvcv1 "operators.kloudlite.io/apis/mysql.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-mysql/internal/controllers/database"
	"operators.kloudlite.io/operators/msvc-mysql/internal/controllers/standalone"
	"operators.kloudlite.io/operators/msvc-mysql/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("msvc-mysql")
	mgr.AddToSchemes(
		mysqlMsvcv1.AddToScheme,
	)
	mgr.RegisterControllers(
		&standalone.ServiceReconciler{Name: "standalone-service", Env: ev},
		&database.Reconciler{Name: "database", Env: ev},
	)
	mgr.Start()
}
