package main

import (
	"github.com/kloudlite/operator/operator"
	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"

	msvcRedis "github.com/kloudlite/operator/operators/msvc-redis/controller"
	msvcRedpanda "github.com/kloudlite/operator/operators/msvc-redpanda/controller"
	nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
	project "github.com/kloudlite/operator/operators/project/controller"
	resourceWatcher "github.com/kloudlite/operator/operators/resource-watcher/controller"
	routers "github.com/kloudlite/operator/operators/routers/controller"
)

func main() {
	mgr := operator.New("agent-operator")

	// kloudlite resources
	app.RegisterInto(mgr)
	project.RegisterInto(mgr)
	helmCharts.RegisterInto(mgr)
	routers.RegisterInto(mgr)

	// kloudlite managed services
	msvcAndMres.RegisterInto(mgr)
	msvcMongo.RegisterInto(mgr)
	msvcRedis.RegisterInto(mgr)
	msvcRedpanda.RegisterInto(mgr)

	// kloudlite cluster management
	nodepool.RegisterInto(mgr)

	// kloudlite resource status updates
	resourceWatcher.RegisterInto(mgr, true)

	mgr.Start()
}
