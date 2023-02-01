package main

import (
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/controllers/app"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
)

func Register(mgr operator.Operator) {
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&app.Reconciler{Name: "app", Env: &env.Env{
			ReconcilePeriod:         30 * time.Second,
			MaxConcurrentReconciles: 1,
		}},
	)
}

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("app-n-lambda")
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&app.Reconciler{Name: "app", Env: ev},
		// &lambda.Reconciler{Name: "lambda", Env: ev},
	)
	mgr.Start()
}
