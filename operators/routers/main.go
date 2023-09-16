package main

import (
	"time"

	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	edgeRouter "github.com/kloudlite/operator/operators/routers/internal/controllers/edge-router"
	"github.com/kloudlite/operator/operators/routers/internal/controllers/router"
	"github.com/kloudlite/operator/operators/routers/internal/env"
	"github.com/pkg/profile"
)

func main() {
	profiler := profile.Start(profile.MemProfile)
	time.AfterFunc(1*time.Minute, func() {
		profiler.Stop()
	})
	ev := env.GetEnvOrDie()
	mgr := operator.New("routers")
	mgr.AddToSchemes(crdsv1.AddToScheme, certmanagerv1.AddToScheme, acmev1.AddToScheme)
	mgr.RegisterControllers(
		&router.Reconciler{Name: "router", Env: ev},
		&edgeRouter.Reconciler{Name: "edge-router", Env: ev},
	)
	mgr.Start()
}
