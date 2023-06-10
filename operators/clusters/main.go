package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/node"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/nodepool"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("projects")
	mgr.AddToSchemes(clustersv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&nodepool.Reconciler{Name: "nodepool", Env: ev},
		&node.Reconciler{Name: "node", Env: ev},
	)
	mgr.Start()
}
