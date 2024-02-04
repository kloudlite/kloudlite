package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-redis/controller"
)

func main() {
	mgr := operator.New("redis")
	controller.RegisterInto(mgr)
	mgr.Start()
}
