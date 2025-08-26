package main

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/internal/app"
	"github.com/kloudlite/operator/internal/environment"
	"github.com/kloudlite/operator/internal/platform_service"
	"github.com/kloudlite/operator/internal/router"
	"github.com/kloudlite/operator/internal/service_intercept"
	"github.com/kloudlite/operator/internal/workmachine"
	"github.com/kloudlite/operator/internal/workspace"
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
