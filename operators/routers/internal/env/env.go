package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod          time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles  int           `env:"MAX_CONCURRENT_RECONCILES"`
	DefaultClusterIssuerName string        `env:"DEFAULT_CLUSTER_ISSUER_NAME" required:"true"`
	AcmeEmail                string        `env:"ACME_EMAIL" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
