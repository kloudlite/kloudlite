package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/helm-charts/controller"
)

func main() {
	mgr := operator.New("helm-charts")
	controller.RegisterInto(mgr)
	mgr.Start()
}
