package main

import (
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	csiv1 "operators.kloudlite.io/apis/csi/v1"
	extensionsv1 "operators.kloudlite.io/apis/extensions/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/extensions/internal/controllers/cluster"
	edgeWatcher "operators.kloudlite.io/operators/extensions/internal/controllers/edge-watcher"
	"operators.kloudlite.io/operators/extensions/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("extensions-cluster")
	mgr.AddToSchemes(extensionsv1.AddToScheme, crdsv1.AddToScheme, csiv1.AddToScheme)
	mgr.RegisterControllers(
		&cluster.Reconciler{Name: "cluster", Env: ev},
		&edgeWatcher.Reconciler{Name: "edge-watcher", Env: ev},
	)
	mgr.Start()
}
