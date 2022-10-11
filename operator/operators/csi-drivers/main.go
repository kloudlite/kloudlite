package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/csi-drivers/internal/controller/driver"
	"operators.kloudlite.io/operators/csi-drivers/internal/env"
)

func main() {
	mgr := operator.New("csi-drivers")
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(
		&driver.Reconciler{Name: "driver", Env: ev},
	)
	mgr.Start()
}
