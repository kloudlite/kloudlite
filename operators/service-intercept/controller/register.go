package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/service-intercept/internal/controllers/svci"
	"github.com/kloudlite/operator/operators/service-intercept/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&svci.Reconciler{Name: "service-intercept", Env: ev},
	)
}
