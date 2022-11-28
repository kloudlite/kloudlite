package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/cluster-setup/internal/env"
)

func main() {
	mgr := operator.New("cluster-setup")
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes()
	mgr.Start()
}
