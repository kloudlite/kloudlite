package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/acl-account"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/acl-configmap"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/standalone"
	"operators.kloudlite.io/operators/msvc-redis/internal/env"
)

func main() {
	op := operator.New("redis")

	ev := env.GetEnvOrDie()

	op.RegisterControllers(
		&standalone.ServiceReconciler{Name: "redis-standalone", Env: ev},
		&aclaccount.Reconciler{Name: "acl-account", Env: ev},
		&acl_configmap.Reconciler{Name: "acl-configmap", Env: ev},
	)
	op.Start()
}
