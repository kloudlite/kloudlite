package main

import (
	"operators.kloudlite.io/operator"
	standaloneService "operators.kloudlite.io/operators/msvc-neo4j/internal/controllers/standalone-service"
	"operators.kloudlite.io/operators/msvc-neo4j/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("neo4j")
	mgr.RegisterControllers(&standaloneService.Reconciler{Name: "standalone-svc", Env: ev})
	mgr.Start()
}
