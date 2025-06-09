package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	helm_controller "github.com/kloudlite/operator/operators/helm-charts/internal/controllers/helm-controller"
	"github.com/kloudlite/operator/operators/helm-charts/internal/env"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&helm_controller.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
