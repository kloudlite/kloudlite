package main

import (
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/controllers/project"
	userAccount "operators.kloudlite.io/operators/artifacts-harbor/internal/controllers/user-account"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/env"
	"operators.kloudlite.io/pkg/harbor"
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
