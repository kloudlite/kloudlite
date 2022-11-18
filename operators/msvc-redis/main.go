package main

import (
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/acl-account"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/acl-configmap"
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/standalone"
	"operators.kloudlite.io/operators/msvc-redis/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	op := operator.New("redis")
	op.AddToSchemes(redisMsvcv1.AddToScheme)
	op.RegisterControllers(
		&standalone.ServiceReconciler{Name: "standalone-svc", Env: ev},
		&aclaccount.Reconciler{Name: "acl-account", Env: ev},
		&acl_configmap.Reconciler{Name: "acl-configmap", Env: ev},
	)
	op.Start()
}
