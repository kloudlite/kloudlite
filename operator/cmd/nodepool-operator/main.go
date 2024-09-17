package main

import (
	"github.com/kloudlite/operator/operator"
	// lifecycle "github.com/kloudlite/operator/operators/lifecycle/controller"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
)

func main() {
	mgr := operator.New("nodepool-operator")

	mgr.AddToSchemes(crdsv1.AddToScheme) // just for lifecycle resource type
	nodepool.RegisterInto(mgr)

	mgr.Start()
}
