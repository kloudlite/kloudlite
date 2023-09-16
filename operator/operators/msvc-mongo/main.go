package main

import (
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operator"
	// "github.com/kloudlite/operator/operators/msvc-mongo/internal/controllers/database"
	standaloneService "github.com/kloudlite/operator/operators/msvc-mongo/internal/controllers/standalone-service"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("mongodb")
	ev.IsDev = mgr.Operator().IsDev
	mgr.AddToSchemes(mongodbMsvcv1.AddToScheme)
	mgr.RegisterControllers(
		&standaloneService.Reconciler{Name: "standalone-svc", Env: ev},
		// &database.Reconciler{Name: "database", Env: ev},
	)
	mgr.RegisterWebhooks(&mongodbMsvcv1.StandaloneService{})
	mgr.Start()
}
