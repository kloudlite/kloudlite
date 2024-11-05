package main

import (
	neo4jMsvcv1 "github.com/kloudlite/operator/apis/neo4j.msvc/v1"
	"github.com/kloudlite/operator/operator"
	standaloneService "github.com/kloudlite/operator/operators/msvc-neo4j/internal/controllers/standalone-service"
	"github.com/kloudlite/operator/operators/msvc-neo4j/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("neo4j")
	mgr.AddToSchemes(neo4jMsvcv1.AddToScheme)
	mgr.RegisterControllers(&standaloneService.Reconciler{Name: "standalone-svc", Env: ev})
	mgr.Start()
}
