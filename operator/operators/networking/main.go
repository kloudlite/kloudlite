package main

import (
	"github.com/kloudlite/operator/operators/networking/register"
	"github.com/kloudlite/operator/toolkit/operator"
)

func main() {
	mgr := operator.New("networking")
	register.RegisterInto(mgr)
	mgr.Start()
}
