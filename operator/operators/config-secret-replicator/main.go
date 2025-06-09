package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/config-secret-replicator/controller"
)

func main() {
	mgr := operator.New("config-secret-replicator")
	controller.RegisterInto(mgr)
	mgr.Start()
}
