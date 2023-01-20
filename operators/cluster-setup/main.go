package main

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/kloudlite/operator/apis/cluster-setup/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	redpandamsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/controllers/primary"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/env"
)

func main() {
	mgr := operator.New("cluster-setup")
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(v1.AddToScheme, crdsv1.AddToScheme, redpandamsvcv1.AddToScheme, certmanagerv1.AddToScheme)
	mgr.RegisterControllers(&primary.Reconciler{Name: "primary-cluster", Env: ev})
	mgr.Start()
}
