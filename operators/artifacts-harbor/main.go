package main

import (
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/controllers/project"
	userAccount "operators.kloudlite.io/operators/artifacts-harbor/internal/controllers/user-account"
	"operators.kloudlite.io/operators/artifacts-harbor/internal/env"
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
	mgr.AddToSchemes(artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&project.Reconciler{Name: "harbor-project", HarborCli: harborCli, Env: ev},
		&userAccount.Reconciler{Name: "harbor-user-account", HarborCli: harborCli, Env: ev},
	)
	mgr.Start()
}
