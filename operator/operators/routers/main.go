package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/routers/controller"
)

func main() {
	mgr := operator.New("routers")
	controller.RegisterInto(mgr)
	mgr.Start()
}
