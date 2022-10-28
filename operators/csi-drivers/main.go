package main

import (
	csiv1 "operators.kloudlite.io/apis/csi/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/csi-drivers/internal/controller/driver"
	"operators.kloudlite.io/operators/csi-drivers/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("csi-drivers")
	mgr.AddToSchemes(csiv1.AddToScheme)
	mgr.RegisterControllers(
		&driver.Reconciler{Name: "driver", Env: ev},
	)
	mgr.Start()
}
