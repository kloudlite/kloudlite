package main

import (
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	// acluser "github.com/kloudlite/operator/operators/msvc-redpanda/internal/controllers/acl-user"
	// "github.com/kloudlite/operator/operators/msvc-redpanda/internal/controllers/admin"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/controllers/topic"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/env"
)

func main() {
	mgr := operator.New("redpanda")
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(redpandaMsvcv1.AddToScheme)
	mgr.RegisterControllers(
		// &admin.Reconciler{Name: "admin", Env: ev},
		&topic.Reconciler{Name: "topic", Env: ev},
		// &acluser.Reconciler{Name: "acl", Env: ev},
	)
	mgr.Start()
}
