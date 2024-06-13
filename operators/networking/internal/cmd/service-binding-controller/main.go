package main

import (
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/service-binding-controller/env"
)

func main() {
	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}
	mgr := operator.New("service-binding")
	mgr.AddToSchemes(networkingv1.AddToScheme)
	mgr.RegisterControllers(&Reconciler{Env: ev, Name: "controller"})
	mgr.Start()
}
