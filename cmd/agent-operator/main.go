package main

import (
	"github.com/kloudlite/operator/operator"

	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	lifecycle "github.com/kloudlite/operator/operators/lifecycle/controller"

	msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	msvcMysql "github.com/kloudlite/operator/operators/msvc-mysql/controller"
	msvcPostgres "github.com/kloudlite/operator/operators/msvc-postgres/controller"
	msvcRedis "github.com/kloudlite/operator/operators/msvc-redis/controller"

	// distribution "github.com/kloudlite/operator/operators/distribution/controller"
	msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"
	networkingv1 "github.com/kloudlite/operator/operators/networking/register"
	project "github.com/kloudlite/operator/operators/project/controller"
	resourceWatcher "github.com/kloudlite/operator/operators/resource-watcher/controller"
	routers "github.com/kloudlite/operator/operators/routers/controller"
	// nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
	// wfreguard "github.com/kloudlite/operator/operators/wireguard/controller"
)

func main() {
	mgr := operator.New("agent-operator")

	// // kloudlite resources
	app.RegisterInto(mgr)
	project.RegisterInto(mgr)
	helmCharts.RegisterInto(mgr)
	routers.RegisterInto(mgr)
	//
	// // kloudlite managed services
	msvcAndMres.RegisterInto(mgr)
	msvcMongo.RegisterInto(mgr)
	msvcRedis.RegisterInto(mgr)
	msvcMysql.RegisterInto(mgr)
	msvcPostgres.RegisterInto(mgr)

	// kloudlite cluster management
	// nodepool.RegisterInto(mgr)

	lifecycle.RegisterInto(mgr)

	// kloudlite resource status updates
	resourceWatcher.RegisterInto(mgr)

	// distribution.RegisterInto(mgr)

	networkingv1.RegisterInto(mgr)

	mgr.Start()
}
