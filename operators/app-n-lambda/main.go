package main

import (
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/app-n-lambda/internal/controllers/app"
	"operators.kloudlite.io/operators/app-n-lambda/internal/env"
)

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
