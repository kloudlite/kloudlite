package main

import (
	"operators.kloudlite.io/operator"
	acluser "operators.kloudlite.io/operators/msvc-redpanda/internal/controllers/acl-user"
	"operators.kloudlite.io/operators/msvc-redpanda/internal/controllers/admin"
	"operators.kloudlite.io/operators/msvc-redpanda/internal/controllers/topic"
	"operators.kloudlite.io/operators/msvc-redpanda/internal/env"
)

func main() {
	mgr := operator.New("redpanda")
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(
		&admin.Reconciler{Name: "admin", Env: ev},
		&topic.Reconciler{Name: "topic", Env: ev},
		&acluser.Reconciler{Name: "acl", Env: ev},
	)
	mgr.Start()
}
