package router

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/kloudlite/kloudlite/operator/toolkit/operator"
	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/router/internal/controller"
)

func RegisterInto(mgr operator.Operator) {
	mgr.AddToSchemes(v1.AddToScheme, certmanagerv1.AddToScheme)
	mgr.RegisterControllers(&controller.Reconciler{})
}
