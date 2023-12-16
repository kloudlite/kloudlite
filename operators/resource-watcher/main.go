package main

import (
	"flag"

	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/resource-watcher/controller"
)

func main() {
	var runningOnTenant bool
	flag.BoolVar(&runningOnTenant, "running-on-tenant", false, "--running-on-tenant")

	mgr := operator.New("resource-watcher")
	controller.RegisterInto(mgr, runningOnTenant)
	mgr.Start()
}
