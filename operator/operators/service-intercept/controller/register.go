package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/service-intercept/internal/controllers/svci"
	"github.com/kloudlite/operator/operators/service-intercept/internal/env"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&svci.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
