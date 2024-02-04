package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-n-mres/controller"
)

func main() {
	mgr := operator.New("msvc-and-mres")
	controller.RegisterInto(mgr)
	mgr.Start()
}
