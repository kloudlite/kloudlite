package main

import (
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/acl-account"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/acl-configmap"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/standalone"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
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
