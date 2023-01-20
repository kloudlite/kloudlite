package main

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	csiv1 "github.com/kloudlite/operator/apis/csi/v1"
	extensionsv1 "github.com/kloudlite/operator/apis/extensions/v1"
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/extensions/internal/controllers/cluster"
	edgeWatcher "github.com/kloudlite/operator/operators/extensions/internal/controllers/edge-watcher"
	edgeWorker "github.com/kloudlite/operator/operators/extensions/internal/controllers/edge-worker"
	"github.com/kloudlite/operator/operators/extensions/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("extensions-cluster")
	mgr.AddToSchemes(extensionsv1.AddToScheme, crdsv1.AddToScheme, csiv1.AddToScheme, redpandaMsvcv1.AddToScheme)
	mgr.RegisterControllers(
		&cluster.Reconciler{Name: "cluster", Env: ev},
		&edgeWatcher.Reconciler{Name: "edge-watcher", Env: ev},
		&edgeWorker.Reconciler{Name: "edge-worker", Env: ev},
	)
	mgr.Start()
}
