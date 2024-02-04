package main

import (
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"github.com/kloudlite/operator/operator"

	"github.com/kloudlite/operator/operators/distribution/internal/controllers/build-run"
	"github.com/kloudlite/operator/operators/distribution/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("distribution")
	mgr.AddToSchemes(distributionv1.AddToScheme)

	mgr.RegisterControllers(
		&buildrun.Reconciler{Name: "distribution:build-run", Env: ev},
	)

	mgr.Start()
}
