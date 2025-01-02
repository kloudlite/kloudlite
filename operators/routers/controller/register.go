package controller

import (
	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/routers/internal/env"
	router_controller "github.com/kloudlite/operator/operators/routers/internal/router-controller"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme, certmanagerv1.AddToScheme, acmev1.AddToScheme)
	mgr.RegisterControllers(
		&router_controller.Reconciler{Name: "router", Env: ev, YAMLClient: mgr.Operator().KubeYAMLClient()},
	)
}
