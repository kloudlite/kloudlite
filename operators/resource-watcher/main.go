package main

import (
	"flag"

	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/resource-watcher/controller"
)

func main() {
	mgr := operator.New("resource-watcher")
	controller.RegisterInto(mgr)
	mgr.Start()
}
