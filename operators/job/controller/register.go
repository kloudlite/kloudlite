package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/job/internal/env"
	job_controller "github.com/kloudlite/operator/operators/job/internal/job-controller"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&job_controller.Reconciler{Name: "job-controller", Env: ev},
	)
}
