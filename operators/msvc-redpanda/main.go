package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-redpanda/controller"
)

func main() {
	mgr := operator.New("redpanda")
	controller.RegisterInto(mgr)
	mgr.Start()
}
