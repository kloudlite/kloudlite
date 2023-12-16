package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	WorkspaceRouteSwitcherService string `env:"WORKSPACE_ROUTE_SWITCHER_SERVICE" required:"true"`
	WorkspaceRouteSwitcherPort    uint16 `env:"WORKSPACE_ROUTE_SWITCHER_PORT" required:"true"`

	AcmeEmail                string `env:"ACME_EMAIL" required:"true"`
	WildcardCertName         any    `env:"WILDCARD_CERT_NAME" required:"false"`
	WildcardCertNamespace    any    `env:"WILDCARD_CERT_NAMESPACE" required:"false"`
	DefaultClusterIssuerName string `env:"DEFAULT_CLUSTER_ISSUER_NAME" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
