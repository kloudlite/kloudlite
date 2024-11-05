package register

import (
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/networking/internal/env"
	"github.com/kloudlite/operator/operators/networking/internal/gateway"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(networkingv1.AddToScheme)
	mgr.RegisterControllers(
		&gateway.Reconciler{Env: ev, Name: "gateway"},
	)
}
