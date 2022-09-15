package main

import (
	"operators.kloudlite.io/operators/app/internal/controllers"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	runner := operator.New("apps")
	runner.RegisterControllers(&controllers.AppReconciler{Name: "app"})
	runner.Start()
}
