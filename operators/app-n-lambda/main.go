package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/app-n-lambda/internal/controllers/app"
	"operators.kloudlite.io/operators/app-n-lambda/internal/env"
)

func main() {
	runner := operator.New("app-n-lambda")

	ev := env.GetEnvOrDie()

	runner.RegisterControllers(
		&app.Reconciler{Name: "app", Env: ev},
		// &lambda.Reconciler{Name: "lambda", Env: ev},
	)
	runner.Start()
}
