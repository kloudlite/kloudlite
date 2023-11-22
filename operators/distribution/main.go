package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"

	"github.com/kloudlite/operator/operators/distribution/internal/controllers/build-run"
	"github.com/kloudlite/operator/operators/distribution/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("wireguard")
	mgr.AddToSchemes(wgv1.AddToScheme, artifactsv1.AddToScheme)

	mgr.RegisterControllers(
		&buildrun.Reconciler{Name: "build", Env: ev},
	)

	mgr.Start()
}
