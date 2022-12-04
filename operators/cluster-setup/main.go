package main

import (
	"operators.kloudlite.io/apis/cluster-setup/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/cluster-setup/internal/controllers/primary"
	"operators.kloudlite.io/operators/cluster-setup/internal/env"
)

func main() {
	mgr := operator.New("cluster-setup")
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(v1.AddToScheme)
	mgr.RegisterControllers(&primary.Reconciler{Name: "primary-cluster", Env: ev})
	mgr.Start()
}
