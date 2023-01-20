package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/artifacts-harbor/internal/controllers/project"
	userAccount "github.com/kloudlite/operator/operators/artifacts-harbor/internal/controllers/user-account"
	"github.com/kloudlite/operator/operators/artifacts-harbor/internal/env"
	"github.com/kloudlite/operator/pkg/harbor"
)

func main() {
	ev := env.GetEnvOrDie()

	harborCli, err := harbor.NewClient(
		harbor.Args{
			HarborAdminUsername: ev.HarborAdminUsername,
			HarborAdminPassword: ev.HarborAdminPassword,
			HarborRegistryHost:  ev.HarborImageRegistryHost,
			HarborApiVersion:    ev.HarborApiVersion,
		},
	)
	if err != nil {
		panic(err)
	}

	mgr := operator.New("artifacts-harbor")
	mgr.AddToSchemes(artifactsv1.AddToScheme, crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&project.Reconciler{Name: "project", HarborCli: harborCli, Env: ev},
		&userAccount.Reconciler{Name: "user-account", HarborCli: harborCli, Env: ev},
	)
	mgr.Start()
}
