package main

import (
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"

	"github.com/kloudlite/operator/operators/iot-device/internal/controllers/blueprint"
	"github.com/kloudlite/operator/operators/iot-device/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("iot-device")
	mgr.AddToSchemes(wgv1.AddToScheme)

	mgr.RegisterControllers(
		&blueprint.Reconciler{Name: "Blueprint", Env: ev},
	)

	mgr.Start()
}
