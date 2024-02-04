package controller

import (
	wireguardv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/device"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(wireguardv1.AddToScheme)
	mgr.RegisterControllers(
		&device.Reconciler{
			Name: "device-controller",
			Env:  ev,
		},
	)
}
