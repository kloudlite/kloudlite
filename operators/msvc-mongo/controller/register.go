package controller

import (
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operator"
	clusterService "github.com/kloudlite/operator/operators/msvc-mongo/internal/controllers/cluster-service"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/controllers/database"
	standaloneService "github.com/kloudlite/operator/operators/msvc-mongo/internal/controllers/standalone-service"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(mongodbMsvcv1.AddToScheme)
	mgr.RegisterControllers(
		&clusterService.Reconciler{Name: "msvc-mongo:cluster-service", Env: ev},
		&standaloneService.Reconciler{Name: "msvc-mongo:standalone-svc", Env: ev},
		&database.Reconciler{Name: "msvc-mongo:database", Env: ev},
	)
}
