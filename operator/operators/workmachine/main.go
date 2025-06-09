package main

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/workmachine/register"
	"github.com/kloudlite/operator/toolkit/operator"
)

func main() {
	mgr := operator.New("workmachine")
	mgr.AddToSchemes(crdsv1.AddToScheme)

	register.RegisterInto(mgr)
	mgr.Start()
}
