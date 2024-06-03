package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/networking/register"
)

func main() {
	mgr := operator.New("networking")
	register.RegisterInto(mgr)
	mgr.Start()
}
