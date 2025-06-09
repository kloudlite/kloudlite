package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/controllers/environment"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&environment.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
