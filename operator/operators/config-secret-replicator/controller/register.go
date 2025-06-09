package controller

import (
	"github.com/kloudlite/operator/operator"
	config_replicator "github.com/kloudlite/operator/operators/config-secret-replicator/internal/config-replicator"
	"github.com/kloudlite/operator/operators/config-secret-replicator/internal/env"
	secret_replicator "github.com/kloudlite/operator/operators/config-secret-replicator/internal/secret-replicator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(
		&config_replicator.Reconciler{
			Env:  ev,
			Name: "configmap-replicator",
		},
		&secret_replicator.Reconciler{
			Env:  ev,
			Name: "secret-replicator",
		},
	)
}
