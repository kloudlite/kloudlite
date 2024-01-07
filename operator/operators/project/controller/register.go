package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/project/internal/controllers/environment"
	"github.com/kloudlite/operator/operators/project/internal/controllers/project"
	"github.com/kloudlite/operator/operators/project/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&project.Reconciler{Name: "project", Env: ev},
		&environment.Reconciler{Name: "workspace", Env: ev},
	)
}
