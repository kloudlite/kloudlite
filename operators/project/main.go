package main

import (
	"time"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/project/internal/controllers/config"
	"github.com/kloudlite/operator/operators/project/internal/controllers/project"
	"github.com/kloudlite/operator/operators/project/internal/controllers/secret"
	"github.com/kloudlite/operator/operators/project/internal/controllers/workspace"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/pkg/profile"
)

func main() {
	profiler := profile.Start(profile.MemProfile)
	time.AfterFunc(1*time.Minute, func() {
		profiler.Stop()
	})

	ev := env.GetEnvOrDie()
	mgr := operator.New("projects")
	mgr.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&project.Reconciler{Name: "project", Env: ev},
		&workspace.Reconciler{Name: "workspace", Env: ev},
		&config.Reconciler{Name: "config", Env: ev},
		&secret.Reconciler{Name: "secret", Env: ev},
	)
	mgr.Start()
}
