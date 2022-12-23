package main

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"operators.kloudlite.io/apis/cluster-setup/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	redpandamsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/cluster-setup/internal/controllers/primary"
	"operators.kloudlite.io/operators/cluster-setup/internal/env"
)

func main() {
	mgr := operator.New("cluster-setup")
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(v1.AddToScheme, crdsv1.AddToScheme, redpandamsvcv1.AddToScheme, certmanagerv1.AddToScheme)
	mgr.RegisterControllers(&primary.Reconciler{Name: "primary-cluster", Env: ev})
	mgr.Start()
}
