package controller

import (
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/acl-account"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/acl-configmap"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/standalone"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(redisMsvcv1.AddToScheme)
	mgr.RegisterControllers(
		&standalone.ServiceReconciler{Name: "msvc-redis:standalone-svc", Env: ev},
		&aclaccount.Reconciler{Name: "msvc-redis:acl-account", Env: ev},
		&acl_configmap.Reconciler{Name: "msvc-redis:acl-configmap", Env: ev},
	)
}
