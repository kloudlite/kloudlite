package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc.redis/internal/controllers/acl-account"
	"operators.kloudlite.io/operators/msvc.redis/internal/controllers/acl-configmap"
	"operators.kloudlite.io/operators/msvc.redis/internal/controllers/standalone"
)

func main() {
	op := operator.New("redis")
	op.RegisterControllers(
		&standalone.ServiceReconciler{Name: "redis-standalone"},
		&aclaccount.Reconciler{Name: "acl-account"},
		&acl_configmap.Reconciler{Name: "acl-configmap"},
	)
	op.Start()
}
