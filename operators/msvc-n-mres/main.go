package main

import (
	"github.com/kloudlite/operator/operators/msvc-n-mres/controller"
	"github.com/kloudlite/operator/toolkit/operator"
)

func main() {
	mgr := operator.New("msvc-and-mres")
	controller.RegisterInto(mgr)
	mgr.Start()
}
