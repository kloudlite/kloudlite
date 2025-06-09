package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/app-n-lambda/controller"
)

func main() {
	mgr := operator.New("app-n-lambda")
	controller.RegisterInto(mgr)
	mgr.Start()
}
