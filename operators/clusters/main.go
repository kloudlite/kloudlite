package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operator"

	"github.com/kloudlite/operator/operators/clusters/internal/controllers/platform-cluster"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/platform-node"

	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target-cluster"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target-node"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target-nodepool"

	"github.com/kloudlite/operator/operators/clusters/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("clusters")
	mgr.AddToSchemes(clustersv1.AddToScheme, artifactsv1.AddToScheme)
	switch ev.RunMode {
	case "platform":
		platformEnv := env.GetPlatformEnvOrDie()

		mgr.RegisterControllers(
			&platform_cluster.Reconciler{Name: "cluster", Env: ev, PlatformEnv: platformEnv},
			&platform_node.Reconciler{Name: "node", Env: ev, PlatformEnv: platformEnv},
		)
	case "target":
		targetEnv := env.GetTargetEnvOrDie()

		mgr.RegisterControllers(
			&target_cluster.Reconciler{Name: "cluster", Env: ev, TargetEnv: targetEnv},
			&target_nodepool.Reconciler{Name: "nodepool", Env: ev, TargetEnv: targetEnv},
			&target_node.Reconciler{Name: "node", Env: ev, TargetEnv: targetEnv},
		)
	default:
		panic("unknown RUN_MODE please provide one of [platform,targef]")
	}
	mgr.Start()
}
