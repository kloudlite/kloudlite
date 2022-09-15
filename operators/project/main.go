package main

import (
	"operators.kloudlite.io/operators/operator"
	"operators.kloudlite.io/operators/project/internal/controllers"
)

func main() {
	runner := operator.New("projects")
	runner.RegisterControllers(&controllers.ProjectReconciler{Name: "project"})
	runner.Start()
}
