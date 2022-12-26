package main

import (
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/project/internal/controllers/config"
	envC "operators.kloudlite.io/operators/project/internal/controllers/env"
	"operators.kloudlite.io/operators/project/internal/controllers/project"
	secondary_env "operators.kloudlite.io/operators/project/internal/controllers/secondary-env"
	"operators.kloudlite.io/operators/project/internal/controllers/secret"
	"operators.kloudlite.io/operators/project/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("projects")
	mgr.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&project.Reconciler{Name: "project", Env: ev, IsDev: mgr.IsDev},
		&envC.Reconciler{Name: "env", Env: ev, IsDev: mgr.IsDev},
		&secondary_env.Reconciler{Name: "secondary_env", Env: ev},
		&config.Reconciler{Name: "config", Env: ev},
		&secret.Reconciler{Name: "secret", Env: ev},
	)
	mgr.Start()
}
