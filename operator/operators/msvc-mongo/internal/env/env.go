package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	IsDev                   bool
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`
	ClusterInternalDNS      string        `env:"CLUSTER_INTERNAL_DNS" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
