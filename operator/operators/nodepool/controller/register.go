package controller

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/nodepool/internal/env"
	node_controller "github.com/kloudlite/operator/operators/nodepool/internal/node-controller"
	nodepool_controller "github.com/kloudlite/operator/operators/nodepool/internal/nodepool-controller"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()

	if !ev.EnableNodepools {
		mgr.Operator().Logger().Info("nodepools controller is disabled")
		return
	}

	mgr.AddToSchemes(clustersv1.AddToScheme)
	mgr.RegisterControllers(
		&nodepool_controller.Reconciler{Env: ev, Name: "nodepool", YAMLClient: mgr.Operator().KubeYAMLClient()},
		&node_controller.Reconciler{Env: ev, Name: "nodepool:node", YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
