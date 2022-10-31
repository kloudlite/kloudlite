package main

import (
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/project/internal/controllers"
	"operators.kloudlite.io/operators/project/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("projects")
	mgr.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(&controllers.ProjectReconciler{Name: "project", Env: ev})
	mgr.Start()
}
