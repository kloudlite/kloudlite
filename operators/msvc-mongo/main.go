package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-mongo/controller"
)

func main() {
	mgr := operator.New("mongodb")
	controller.RegisterInto(mgr)
	mgr.Start()
}
