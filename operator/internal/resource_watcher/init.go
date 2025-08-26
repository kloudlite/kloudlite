package resource_watcher

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/resource_watcher/internal/controller"
)

func RegisterInto(mgr operator.Operator) {
	mgr.AddToSchemes(v1.AddToScheme)
	mgr.RegisterControllers(
		&controller.Reconciler{},
	)
}
