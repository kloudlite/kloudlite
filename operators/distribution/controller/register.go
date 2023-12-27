package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operator"
	buildrun "github.com/kloudlite/operator/operators/distribution/internal/controllers/build-run"
	env "github.com/kloudlite/operator/operators/distribution/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(redisMsvcv1.AddToScheme, crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&buildrun.Reconciler{
			Name: "distribution:build-run",
			Env:  ev,
		},
	)
}
