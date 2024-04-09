package main

import (
	"github.com/kloudlite/operator/operator"
	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	clusters "github.com/kloudlite/operator/operators/clusters/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	job "github.com/kloudlite/operator/operators/job/controller"
	msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"
	msvcRedis "github.com/kloudlite/operator/operators/msvc-redis/controller"
	nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
	project "github.com/kloudlite/operator/operators/project/controller"
	routers "github.com/kloudlite/operator/operators/routers/controller"
	wireguard "github.com/kloudlite/operator/operators/wireguard/controller"
)

func main() {
	mgr := operator.New("platform-operator")

	// kloudlite resources
	app.RegisterInto(mgr)
	routers.RegisterInto(mgr)
	project.RegisterInto(mgr)
	helmCharts.RegisterInto(mgr)

	// kloudlite managed services
	msvcMongo.RegisterInto(mgr)
	msvcRedis.RegisterInto(mgr)
	msvcAndMres.RegisterInto(mgr)

	job.RegisterInto(mgr)

	// kloudlite cluster management
	clusters.RegisterInto(mgr)
	nodepool.RegisterInto(mgr)
	wireguard.RegisterInto(mgr)

	mgr.Start()
}
