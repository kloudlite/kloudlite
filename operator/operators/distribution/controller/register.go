package controller

import (
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"

	"github.com/kloudlite/operator/operator"
	buildrun "github.com/kloudlite/operator/operators/distribution/internal/controllers/build-run"
	env "github.com/kloudlite/operator/operators/distribution/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(distributionv1.AddToScheme)
	mgr.RegisterControllers(
		&buildrun.Reconciler{
			Name: "distribution:build-run",
			Env:  ev,
		},
	)
}
