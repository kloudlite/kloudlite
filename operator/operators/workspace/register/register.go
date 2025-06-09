package register

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/workspace/internal/controllers/workspace"
	"github.com/kloudlite/operator/operators/workspace/internal/env"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&workspace.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
