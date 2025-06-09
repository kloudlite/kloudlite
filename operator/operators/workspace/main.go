package main

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
)

func main() {
	mgr := operator.New("workspace")
	mgr.AddToSchemes(crdsv1.AddToScheme)

	mgr.Start()
}
