package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/service-intercept/controller"
)

func main() {
	mgr := operator.New("service-intercept")
	controller.RegisterInto(mgr)
	mgr.Start()
}
