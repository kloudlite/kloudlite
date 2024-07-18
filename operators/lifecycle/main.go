package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/lifecycle/controller"
)

func main() {
	mgr := operator.New("lifecycle")
	controller.RegisterInto(mgr)
	mgr.Start()
}
