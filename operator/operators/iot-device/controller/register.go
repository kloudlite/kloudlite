package controller

import (
	wireguardv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/iot-device/internal/controllers/blueprint"

	"github.com/kloudlite/operator/operators/iot-device/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(wireguardv1.AddToScheme)
	mgr.RegisterControllers(
		&blueprint.Reconciler{Name: "Blueprint", Env: ev},
	)
}
