package main

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/nodepool/internal/env"
	nodepool_controller "github.com/kloudlite/operator/operators/nodepool/internal/nodepool-controller"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("nodepool")
	mgr.AddToSchemes(clustersv1.AddToScheme)
	mgr.RegisterControllers(&nodepool_controller.Reconciler{
		Env:  ev,
		Name: "nodepool",
	})
	mgr.Start()
}
