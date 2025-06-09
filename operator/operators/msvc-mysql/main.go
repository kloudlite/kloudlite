package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-mysql/controller"
)

func main() {
	mgr := operator.New("msvc-mysql")
	controller.RegisterInto(mgr)
	mgr.Start()
}
