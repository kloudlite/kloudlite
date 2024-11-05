package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-postgres/controller"
)

func main() {
	mgr := operator.New("msvc-postgres")
	controller.RegisterInto(mgr)
	mgr.Start()
}
