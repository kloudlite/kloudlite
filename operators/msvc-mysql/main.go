package main

import (
	mysqlMsvcv1 "operators.kloudlite.io/apis/mysql.msvc/v1"
	"operators.kloudlite.io/operator"
	clusterService "operators.kloudlite.io/operators/msvc-mysql/internal/controllers/cluster-service"
	"operators.kloudlite.io/operators/msvc-mysql/internal/controllers/database"
	standaloneService "operators.kloudlite.io/operators/msvc-mysql/internal/controllers/standalone-service"
	"operators.kloudlite.io/operators/msvc-mysql/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("msvc-mysql")
	mgr.AddToSchemes(
		mysqlMsvcv1.AddToScheme,
	)
	mgr.RegisterControllers(
		&standaloneService.ServiceReconciler{Name: "standalone-svc", Env: ev},
		&clusterService.Reconciler{Name: "cluster-svc", Env: ev},
		&database.Reconciler{Name: "database", Env: ev},
	)
	mgr.Start()
}
