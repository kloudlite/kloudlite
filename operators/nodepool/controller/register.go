package controller

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/nodepool/internal/env"
	node_controller "github.com/kloudlite/operator/operators/nodepool/internal/node-controller"
	nodepool_controller "github.com/kloudlite/operator/operators/nodepool/internal/nodepool-controller"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(clustersv1.AddToScheme)
	mgr.RegisterControllers(
		&nodepool_controller.Reconciler{
			Env:  ev,
			Name: "nodepool",
		},
		&node_controller.Reconciler{
			Env:  ev,
			Name: "node",
		},
	)
}
