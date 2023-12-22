package main

import (
	"github.com/kloudlite/operator/operator"
	app "github.com/kloudlite/operator/operators/app-n-lambda/controller"
	clusters "github.com/kloudlite/operator/operators/clusters/controller"
	helmCharts "github.com/kloudlite/operator/operators/helm-charts/controller"
	msvcMongo "github.com/kloudlite/operator/operators/msvc-mongo/controller"
	project "github.com/kloudlite/operator/operators/project/controller"
	routers "github.com/kloudlite/operator/operators/routers/controller"
	// routers "github.com/kloudlite/operator/operators/routers/controller"
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

	// kloudlite cluster management
	clusters.RegisterInto(mgr)

	mgr.Start()
}
