package main

import (
	"github.com/kloudlite/operator/operator"
	resourceWatcher "github.com/kloudlite/operator/operators/resource-watcher/controller"
	// routers "github.com/kloudlite/operator/operators/routers/controller"
)

func main() {
	mgr := operator.New("resource-watcher-operator")
	// kloudlite resource status updates
	resourceWatcher.RegisterInto(mgr, false)

	mgr.Start()
}
