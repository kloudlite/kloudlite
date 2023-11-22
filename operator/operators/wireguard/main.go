package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"

	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/device"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("wireguard")
	mgr.AddToSchemes(wgv1.AddToScheme, artifactsv1.AddToScheme)

	mgr.RegisterControllers(
		&device.Reconciler{Name: "Device", Env: ev},
	)

	mgr.Start()
}
