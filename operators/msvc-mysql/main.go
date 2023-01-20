package main

import (
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	"github.com/kloudlite/operator/operator"
	clusterService "github.com/kloudlite/operator/operators/msvc-mysql/internal/controllers/cluster-service"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/controllers/database"
	standaloneService "github.com/kloudlite/operator/operators/msvc-mysql/internal/controllers/standalone-service"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/env"
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
