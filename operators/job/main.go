package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/helm-charts/controller"
)

func main() {
	mgr := operator.New("job-controller")
	controller.RegisterInto(mgr)
	mgr.Start()
}
