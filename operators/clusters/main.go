package main

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	redpandav1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	account_s3_bucket "github.com/kloudlite/operator/operators/clusters/internal/controllers/account-s3-bucket"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("clusters")
	mgr.AddToSchemes(clustersv1.AddToScheme, redpandav1.AddToScheme)
	// switch ev.RunMode {
	// case "platform":
	// 	platformEnv := env.GetPlatformEnvOrDie()
	//
	// 	mgr.RegisterControllers(
	// 		&platform_cluster.Reconciler{Name: "cluster", Env: ev, PlatformEnv: platformEnv},
	// 		&platform_node.Reconciler{Name: "node", Env: ev, PlatformEnv: platformEnv},
	// 	)
	// case "target":
	// 	targetEnv := env.GetTargetEnvOrDie()
	//
	// 	mgr.RegisterControllers(
	// 		&target_cluster.Reconciler{Name: "cluster", Env: ev, TargetEnv: targetEnv},
	// 		&target_nodepool.Reconciler{Name: "nodepool", Env: ev, TargetEnv: targetEnv},
	// 		&target_node.Reconciler{Name: "node", Env: ev, TargetEnv: targetEnv},
	// 	)
	// default:
	// 	panic("unknown RUN_MODE please provide one of [platform,targef]")
	// }

	mgr.RegisterControllers(
		&target.ClusterReconciler{Name: "cluster", Env: ev},
		&account_s3_bucket.Reconciler{Name: "account-s3-bucket", Env: ev},
	)
	mgr.Start()
}
