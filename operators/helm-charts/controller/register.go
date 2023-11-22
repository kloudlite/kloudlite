package controller

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	helm_controller "github.com/kloudlite/operator/operators/helm-charts/internal/controllers/helm-controller"
	"github.com/kloudlite/operator/operators/helm-charts/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&helm_controller.Reconciler{Name: "helm-controller", Env: ev},
	)
}
