package main

import (
	"operators.kloudlite.io/operators/lambda/internal/controllers"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	op := operator.New("lambda")
	op.RegisterControllers(&controllers.LambdaReconciler{Name: "lambda"})
	op.Start()
}
