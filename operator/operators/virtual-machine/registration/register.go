package register

import (
	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/virtual-machine/internal/env"
	"github.com/kloudlite/operator/operators/virtual-machine/internal/vm-controller"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.LoadEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme, certmanagerv1.AddToScheme, acmev1.AddToScheme)
	mgr.RegisterControllers(
		&vm_controller.Reconciler{Name: "vm-controller", Env: ev},
	)
}
