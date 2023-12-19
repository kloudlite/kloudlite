package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	WorkspaceRouteSwitcherService string `env:"WORKSPACE_ROUTE_SWITCHER_SERVICE" required:"true"`
	WorkspaceRouteSwitcherPort    uint16 `env:"WORKSPACE_ROUTE_SWITCHER_PORT" required:"true"`

	IngressClass  string `env:"INGRESS_CLASS"`
	ClusterIssuer string `env:"CLUSTER_ISSUER"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	if ev.MaxConcurrentReconciles == 0 {
		ev.MaxConcurrentReconciles = 5
	}
	return &ev
}
