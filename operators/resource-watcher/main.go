package main

import (
	"flag"

	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/resource-watcher/controller"
)

func main() {
	var runningOnPlatform bool
	flag.BoolVar(&runningOnPlatform, "running-on-platform", false, "--running-on-platform")

	mgr := operator.New("resource-watcher")
	controller.RegisterInto(mgr, runningOnPlatform)
	mgr.Start()
}
