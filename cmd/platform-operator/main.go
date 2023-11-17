package main

import (
	"github.com/kloudlite/operator/operator"
	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	clusters "github.com/kloudlite/operator/operators/clusters/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	// msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	// msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"
	// msvcRedis "github.com/kloudlite/operator/operators/msvc-redis/controller"
	// msvcRedpanda "github.com/kloudlite/operator/operators/msvc-redpanda/controller"
	// nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
	project "github.com/kloudlite/operator/operators/project/controller"
)

func main() {
	mgr := operator.New("platform operator")
	app.RegisterInto(mgr)
	project.RegisterInto(mgr)
	clusters.RegisterInto(mgr)
	helmCharts.RegisterInto(mgr)
	// msvcAndMres.RegisterInto(mgr)
	// msvcMongo.RegisterInto(mgr)
	// msvcRedis.RegisterInto(mgr)
	// msvcRedpanda.RegisterInto(mgr)
	// nodepool.RegisterInto(mgr)
	mgr.Start()
}
