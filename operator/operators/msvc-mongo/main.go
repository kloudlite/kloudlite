package main

import (
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-mongo/internal/controllers/database"
	standaloneService "operators.kloudlite.io/operators/msvc-mongo/internal/controllers/standalone-service"
	"operators.kloudlite.io/operators/msvc-mongo/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("mongodb")
	mgr.AddToSchemes(mongodbMsvcv1.AddToScheme)
	mgr.RegisterControllers(&standaloneService.ServiceReconciler{Name: "standalone-svc", Env: ev}, &database.Reconciler{Name: "database", Env: ev})
	mgr.Start()
}
