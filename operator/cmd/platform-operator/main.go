package main

import (
	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	"github.com/kloudlite/operator/toolkit/operator"

	// clusters "github.com/kloudlite/operator/operators/clusters/controller"
	// helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	lifecycle "github.com/kloudlite/operator/operators/lifecycle/controller"
	msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"

	// msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	// msvcRedis "github.com/kloudlite/operator/operators/msvc-redis/controller"

	// networkingv1 "github.com/kloudlite/operator/operators/networking/register"
	// nodepool "github.com/kloudlite/operator/operators/nodepool/controller"
	project "github.com/kloudlite/operator/operators/project/controller"
	routers "github.com/kloudlite/operator/operators/routers/controller"

	// virtualMachine "github.com/kloudlite/operator/operators/virtual-machine/registration"

	// wireguard "github.com/kloudlite/operator/operators/wireguard/controller"

	serviceIntercept "github.com/kloudlite/operator/operators/service-intercept/controller"
	workspace "github.com/kloudlite/operator/operators/workspace/register"
	pluginK3sCluster "github.com/kloudlite/plugin-k3s-cluster/kloudlite"
	// pluginMongoDB "github.com/kloudlite/plugin-mongodb/kloudlite"
)

func main() {
	mgr := operator.New("platform-operator")

	app.RegisterInto(mgr)
	routers.RegisterInto(mgr)
	project.RegisterInto(mgr)
	// helmCharts.RegisterInto(mgr)

	msvcAndMres.RegisterInto(mgr)

	// msvcMongo.RegisterInto(mgr)
	// msvcRedis.RegisterInto(mgr)
	// msvcMysql.RegisterInto(mgr)
	// msvcPostgres.RegisterInto(mgr)

	lifecycle.RegisterInto(mgr)

	serviceIntercept.RegisterInto(mgr)

	// clusters.RegisterInto(mgr)
	// nodepool.RegisterInto(mgr) // MIGRATE
	// virtualMachine.RegisterInto(mgr)

	// wireguard.RegisterInto(mgr)    // MIGRATE
	// networkingv0.RegisterInto(mgr) // MIGRATE

	// pluginMongoDB.RegisterInto(mgr)
	pluginK3sCluster.RegisterInto(mgr)
	workspace.RegisterInto(mgr)

	mgr.Start()
}
