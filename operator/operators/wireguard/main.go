package main

import (
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"

	cc "github.com/kloudlite/operator/operators/wireguard/internal/controllers/cluster-connection"
	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/device"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("wireguard")
	mgr.AddToSchemes(wgv1.AddToScheme)

	mgr.RegisterControllers(
		&device.Reconciler{Name: "Device", Env: ev},
		&cc.Reconciler{Name: "ClusterConnection", Env: ev},
	)

	mgr.Start()
}
