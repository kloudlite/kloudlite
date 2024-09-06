package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	postgresv1 "github.com/kloudlite/operator/apis/postgres.msvc/v1"
	"github.com/kloudlite/operator/operator"
	standalone_database "github.com/kloudlite/operator/operators/msvc-postgres/internal/controllers/standalone-database"
	standalone "github.com/kloudlite/operator/operators/msvc-postgres/internal/controllers/standalone-service"
	"github.com/kloudlite/operator/operators/msvc-postgres/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(postgresv1.AddToScheme, crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&standalone_database.Reconciler{Name: "database", Env: ev},
		&standalone.Reconciler{Name: "standalone", Env: ev},
	)
}
