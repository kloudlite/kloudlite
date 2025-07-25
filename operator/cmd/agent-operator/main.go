package main

import (
	"github.com/kloudlite/operator/toolkit/operator"

	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	lifecycle "github.com/kloudlite/operator/operators/lifecycle/controller"

	msvcAndMres "github.com/kloudlite/operator/operators/msvc-n-mres/controller"
	networkingv1 "github.com/kloudlite/operator/operators/networking/register"
	project "github.com/kloudlite/operator/operators/project/controller"
	resourceWatcher "github.com/kloudlite/operator/operators/resource-watcher/controller"

	workmachine "github.com/kloudlite/operator/operators/workmachine/register"
	workspace "github.com/kloudlite/operator/operators/workspace/register"

	serviceIntercept "github.com/kloudlite/operator/operators/service-intercept/controller"
	pluginHelmChart "github.com/kloudlite/plugin-helm-chart/kloudlite"
	pluginMongoDB "github.com/kloudlite/plugin-mongodb/kloudlite"
)

func main() {
	mgr := operator.New("agent-operator")

	// kloudlite resources
	app.RegisterInto(mgr)
	project.RegisterInto(mgr)
	helmCharts.RegisterInto(mgr)
	// routers.RegisterInto(mgr)

	// kloudlite managed services
	msvcAndMres.RegisterInto(mgr)

	// msvcMongo.RegisterInto(mgr)
	// msvcRedis.RegisterInto(mgr)
	// msvcMysql.RegisterInto(mgr)
	// msvcPostgres.RegisterInto(mgr)

	lifecycle.RegisterInto(mgr)

	// kloudlite resource status updates
	resourceWatcher.RegisterInto(mgr)

	// distribution.RegisterInto(mgr)

	networkingv1.RegisterInto(mgr)
	serviceIntercept.RegisterInto(mgr)
	workmachine.RegisterInto(mgr)
	workspace.RegisterInto(mgr)

	pluginMongoDB.RegisterInto(mgr)
	pluginHelmChart.RegisterInto(mgr)

	mgr.Start()
}
