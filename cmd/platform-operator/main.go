package main

import (
	"github.com/kloudlite/operator/operator"
	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	clusters "github.com/kloudlite/operator/operators/clusters/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	lifecycle "github.com/kloudlite/operator/operators/lifecycle/controller"
	msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"
	msvcRedis "github.com/kloudlite/operator/operators/msvc-redis/controller"
	nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
	project "github.com/kloudlite/operator/operators/project/controller"
	routers "github.com/kloudlite/operator/operators/routers/controller"
	virtualMachine "github.com/kloudlite/operator/operators/virtual-machine/registration"
	wireguard "github.com/kloudlite/operator/operators/wireguard/controller"
)

func main() {
	mgr := operator.New("platform-operator")

	app.RegisterInto(mgr)
	routers.RegisterInto(mgr)
	project.RegisterInto(mgr)
	helmCharts.RegisterInto(mgr)

	msvcMongo.RegisterInto(mgr)
	msvcRedis.RegisterInto(mgr)
	msvcAndMres.RegisterInto(mgr)

	lifecycle.RegisterInto(mgr)

	clusters.RegisterInto(mgr)
	nodepool.RegisterInto(mgr)
	virtualMachine.RegisterInto(mgr)

	wireguard.RegisterInto(mgr)

	mgr.Start()
}
