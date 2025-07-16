package main

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/controllers/app"
	"github.com/kloudlite/operator/internal/controllers/router"
)

func main() {
	mgr := operator.New("platform-operator")

	mgr.AddToSchemes(v1.AddToScheme)
	mgr.RegisterControllers(
		&app.Reconciler{
			Env: app.Env{
				MaxConcurrentReconciles: 5,
			},
		},
		&router.Reconciler{
			Env: router.Env{
				MaxConcurrentReconciles: 5,
			},
		},
	)

	mgr.Start()
}
