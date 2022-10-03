package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/app-n-lambda/internal/controllers/app"
	"operators.kloudlite.io/operators/app-n-lambda/internal/controllers/lambda"
)

func main() {
	runner := operator.New("app-n-lambda")

	runner.RegisterControllers(
		&app.Reconciler{Name: "app"},
		&lambda.Reconciler{Name: "lambda"},
	)
	runner.Start()
}
