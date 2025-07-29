package environment

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/api/v1"
)

func RegisterInto(mgr operator.Operator) {
	mgr.AddToSchemes(v1.AddToScheme)
	mgr.RegisterControllers(
		&Reconciler{},
	)
}
