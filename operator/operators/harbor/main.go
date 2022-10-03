package main

import (
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/harbor/internal/controllers/project"
	userAccount "operators.kloudlite.io/operators/harbor/internal/controllers/user-account"
)

func main() {
	mgr := operator.New("artifacts-harbor")

	harborCli, err := harbor.NewClient(
		harbor.Args{
			HarborAdminUsername: mgr.Env.HarborAdminUsername,
			HarborAdminPassword: mgr.Env.HarborAdminPassword,
			HarborRegistryHost:  mgr.Env.HarborImageRegistryHost,
			WebhookAddr:         mgr.Env.HarborWebhookAddr,
		},
	)
	if err != nil {
		panic(err)
	}

	mgr.RegisterControllers(
		&project.Reconciler{Name: "harbor-project", HarborCli: harborCli},
		&userAccount.Reconciler{Name: "harbor-user-account", HarborCli: harborCli},
	)
	mgr.Start()
}
