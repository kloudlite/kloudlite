package main

import (
	csiv1 "github.com/kloudlite/operator/apis/csi/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/csi-drivers/internal/controller/driver"
	"github.com/kloudlite/operator/operators/csi-drivers/internal/env"
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
