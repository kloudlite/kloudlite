package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/extensions-cluster/internal/controllers/cluster"
	"operators.kloudlite.io/operators/extensions-cluster/internal/env"
)

func main() {
	mgr := operator.New("extensions-cluster")
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(&cluster.Reconciler{Name: "cluster", Env: ev})
	mgr.Start()
}
