package main

import (
	"operators.kloudlite.io/operator"
	acluser "operators.kloudlite.io/operators/msvc.redpanda/internal/controllers/acl-user"
	"operators.kloudlite.io/operators/msvc.redpanda/internal/controllers/topic"
)

func main() {
	mgr := operator.New("redpanda")
	mgr.RegisterControllers(
		&topic.Reconciler{Name: "topic"},
		&acluser.Reconciler{Name: "acl"},
	)
	mgr.Start()
}
