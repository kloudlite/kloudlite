package main

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"

	"github.com/kloudlite/operator/operators/account/internal/account"
	"github.com/kloudlite/operator/operators/account/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("accounts")
	mgr.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
	mgr.RegisterControllers(
		&account.Reconciler{Name: "accounts", Env: ev},
	)
	mgr.Start()
}
