package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/clusters/controller"
)

func main() {
	mgr := operator.New("clusters")
	controller.RegisterInto(mgr)
	mgr.Start()
}
