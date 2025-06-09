package main

import (
	"github.com/kloudlite/operator/operators/nodepool/controller"
	"github.com/kloudlite/operator/toolkit/operator"
)

func main() {
	mgr := operator.New("nodepool")
	controller.RegisterInto(mgr)
	mgr.Start()
}
