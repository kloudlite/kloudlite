package main

import (
	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	edgeRouter "github.com/kloudlite/operator/operators/routers/internal/controllers/edge-router"
	"github.com/kloudlite/operator/operators/routers/internal/controllers/router"
	"github.com/kloudlite/operator/operators/routers/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("routers")
	mgr.AddToSchemes(crdsv1.AddToScheme, certmanagerv1.AddToScheme, acmev1.AddToScheme)
	mgr.RegisterControllers(
		// &account_router.AccountRouterReconciler{Name: "acc-router", Env: ev},
		&router.Reconciler{Name: "router", Env: ev},
		&edgeRouter.Reconciler{Name: "edge-router", Env: ev},
	)
	mgr.Start()
}
