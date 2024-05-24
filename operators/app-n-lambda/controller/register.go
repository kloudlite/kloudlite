package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/controllers/app"
	external_app "github.com/kloudlite/operator/operators/app-n-lambda/internal/controllers/external-app"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&app.Reconciler{Name: "app", Env: ev},
		&external_app.ExternalAppReconciler{Name: "external-app", Env: ev},
	)
}
