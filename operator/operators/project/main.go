package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/project/internal/controllers/config"
	envController "github.com/kloudlite/operator/operators/project/internal/controllers/env"
	"github.com/kloudlite/operator/operators/project/internal/controllers/project"
	"github.com/kloudlite/operator/operators/project/internal/controllers/secret"
	"github.com/kloudlite/operator/operators/project/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("projects")
	mgr.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&project.Reconciler{Name: "project", Env: ev},
		&envController.Reconciler{Name: "env-reconciler", Env: ev},
		&config.Reconciler{Name: "config", Env: ev},
		&secret.Reconciler{Name: "secret", Env: ev},
	)
	mgr.Start()
}
