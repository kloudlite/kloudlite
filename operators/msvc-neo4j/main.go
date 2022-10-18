package main

import (
	neo4jMsvcv1 "operators.kloudlite.io/apis/neo4j.msvc/v1"
	"operators.kloudlite.io/operator"
	standaloneService "operators.kloudlite.io/operators/msvc-neo4j/internal/controllers/standalone-service"
	"operators.kloudlite.io/operators/msvc-neo4j/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("neo4j")
	mgr.AddToSchemes(neo4jMsvcv1.AddToScheme)
	mgr.RegisterControllers(&standaloneService.Reconciler{Name: "standalone-svc", Env: ev})
	mgr.Start()
}
