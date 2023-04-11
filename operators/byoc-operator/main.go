package main

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/byoc-operator/internal/controller"
	"github.com/kloudlite/operator/operators/byoc-operator/internal/env"
)

func main() {
	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}
	mgr := operator.New("byoc")
	mgr.AddToSchemes(clustersv1.AddToScheme, redpandaMsvcv1.AddToScheme)
	mgr.RegisterControllers(&controller.Reconciler{Name: "controller", Env: ev})
	mgr.Start()
}
