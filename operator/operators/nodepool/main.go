package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/nodepool/controller"
)

func main() {
	mgr := operator.New("nodepool")
	controller.RegisterInto(mgr)
	mgr.Start()
}
