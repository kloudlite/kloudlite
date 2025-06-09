package controller

import (
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	"github.com/kloudlite/operator/operator"
	// clusterService "github.com/kloudlite/operator/operators/msvc-mysql/internal/controllers/cluster-service"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	standalone_database "github.com/kloudlite/operator/operators/msvc-mysql/internal/controllers/standalone-database"
	standaloneService "github.com/kloudlite/operator/operators/msvc-mysql/internal/controllers/standalone-service"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mysqlMsvcv1.AddToScheme,
	)
	mgr.RegisterControllers(
		&standaloneService.ServiceReconciler{Name: "standalone-svc", Env: ev},
		// &clusterService.Reconciler{Name: "cluster-svc", Env: ev},
		&standalone_database.Reconciler{Name: "standalone-database", Env: ev},
	)
}
