package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/lifecycle/internal/env"
	lifeycle_controller "github.com/kloudlite/operator/operators/lifecycle/internal/lifecycle-controller"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&lifeycle_controller.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
