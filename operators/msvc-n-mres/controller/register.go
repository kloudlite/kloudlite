package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	"github.com/kloudlite/operator/toolkit/operator"

	cluster_msvc "github.com/kloudlite/operator/operators/msvc-n-mres/internal/cluster-msvc"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/msvc"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(
		crdsv1.AddToScheme,
	)
	mgr.RegisterControllers(
		&msvc.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
		&cluster_msvc.Reconciler{Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
