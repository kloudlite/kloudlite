package main

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/internal/controllers/app"
	"github.com/kloudlite/operator/internal/controllers/router"
	"github.com/kloudlite/operator/internal/controllers/service_intercept"
)

func main() {
	mgr := operator.New("platform-operator")

	app.RegisterInto(mgr)
	router.RegisterInto(mgr)
	service_intercept.RegisterInto(mgr)

	mgr.Start()
}
