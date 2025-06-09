package main

import (
	"github.com/kloudlite/operator/operators/helm-charts/controller"
	"github.com/kloudlite/operator/toolkit/operator"
)

func main() {
	mgr := operator.New("helm-charts")
	controller.RegisterInto(mgr)
	mgr.Start()
}
