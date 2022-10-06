package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/project/internal/controllers"
	"operators.kloudlite.io/operators/project/internal/env"
)

func main() {
	runner := operator.New("projects")
	ev := env.GetEnvOrDie()
	runner.RegisterControllers(&controllers.ProjectReconciler{Name: "project", Env: ev})
	runner.Start()
}
