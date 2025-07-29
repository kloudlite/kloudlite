package main

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/internal/controllers/app"
	"github.com/kloudlite/operator/internal/controllers/environment"
	"github.com/kloudlite/operator/internal/controllers/platform_service"
	"github.com/kloudlite/operator/internal/controllers/router"
	"github.com/kloudlite/operator/internal/controllers/service_intercept"
	"github.com/kloudlite/operator/internal/controllers/workmachine"
	"github.com/kloudlite/operator/internal/controllers/workspace"
)

func main() {
	mgr := operator.New("platform-operator")

	environment.RegisterInto(mgr)
	app.RegisterInto(mgr)
	router.RegisterInto(mgr)
	service_intercept.RegisterInto(mgr)
	platform_service.RegisterInto(mgr)
	workspace.RegisterInto(mgr)
	workmachine.RegisterInto(mgr)

	mgr.Start()
}
