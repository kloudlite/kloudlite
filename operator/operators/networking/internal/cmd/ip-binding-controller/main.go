package main

import (
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/env"
	pod_pinger "github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/pod-pinger"
	service_binding "github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/service-binding"
)

func main() {
	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}
	mgr := operator.New("ip-binding")
	mgr.AddToSchemes(networkingv1.AddToScheme)
	mgr.RegisterControllers(
		&service_binding.Reconciler{Env: ev, Name: "svc-binding"},
		&pod_pinger.Reconciler{Env: ev, Name: "pod-pinger"},
	)
	mgr.Start()
}
