package main

import (
	csiv1 "github.com/kloudlite/operator/apis/csi/v1"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/kloudlite/operator/apis/cluster-setup/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	extensionsv1 "github.com/kloudlite/operator/apis/extensions/v1"
	redpandamsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	// "github.com/kloudlite/operator/operators/cluster-setup/internal/controllers/edge-watcher"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/controllers/edge-worker"
	// "github.com/kloudlite/operator/operators/cluster-setup/internal/controllers/managed"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/env"
)

func main() {
	mgr := operator.New("cluster-setup")
	ev := env.GetEnvOrDie()

	mgr.AddToSchemes(v1.AddToScheme, crdsv1.AddToScheme, redpandamsvcv1.AddToScheme, certmanagerv1.AddToScheme, extensionsv1.AddToScheme, csiv1.AddToScheme)

	mgr.RegisterControllers(
		&edgeWorker.Reconciler{
			Name: "edge-worker",
			Env:  ev,
		},
		// &edgeWatcher.Reconciler{
		// 	Name: "edge-watcher",
		// 	Env:  ev,
		// },
		// &managed.Reconciler{
		// 	Name: "managed-cluster",
		// 	Env:  ev,
		// }
	)
	mgr.Start()
}
