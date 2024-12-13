package main

import (
	"flag"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	lifecycle "github.com/kloudlite/operator/operators/lifecycle/controller"
	nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
)

func main() {
	var enableLifecycleController bool
	flag.BoolVar(&enableLifecycleController, "lifecycle-controller", false, "enable lifecycle controller")

	mgr := operator.New("nodepool-operator")

	mgr.AddToSchemes(crdsv1.AddToScheme) // just for lifecycle resource type

	nodepool.RegisterInto(mgr)

	if enableLifecycleController {
		lifecycle.RegisterInto(mgr)
	}

	mgr.Start()
}
